package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// observationService implements the ObservationService interface
type observationService struct {
	observationRepo observability.ObservationRepository
	traceRepo       observability.TraceRepository
	eventPublisher  observability.EventPublisher
}

// NewObservationService creates a new observation service
func NewObservationService(
	observationRepo observability.ObservationRepository,
	traceRepo observability.TraceRepository,
	eventPublisher observability.EventPublisher,
) observability.ObservationService {
	return &observationService{
		observationRepo: observationRepo,
		traceRepo:       traceRepo,
		eventPublisher:  eventPublisher,
	}
}

// CreateObservation creates a new observation
func (s *observationService) CreateObservation(ctx context.Context, observation *observability.Observation) (*observability.Observation, error) {
	if observation == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"observation cannot be nil",
		)
	}

	// Generate ID if not provided
	if observation.ID.IsZero() {
		observation.ID = ulid.New()
	}

	// Validate required fields
	if err := s.validateObservation(observation); err != nil {
		return nil, err
	}

	// Verify trace exists
	if _, err := s.traceRepo.GetByID(ctx, observation.TraceID); err != nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeObservationTraceNotFound,
			"trace not found for observation",
		).WithDetail("trace_id", observation.TraceID.String())
	}

	// Set timestamps
	now := time.Now()
	if observation.CreatedAt.IsZero() {
		observation.CreatedAt = now
	}
	observation.UpdatedAt = now

	// Create observation in repository
	if err := s.observationRepo.Create(ctx, observation); err != nil {
		return nil, fmt.Errorf("failed to create observation: %w", err)
	}

	// Publish observation created event
	event := observability.NewObservationCreatedEvent(observation, nil)
	if publishErr := s.eventPublisher.Publish(ctx, event); publishErr != nil {
		_ = publishErr
	}

	return observation, nil
}

// GetObservation retrieves an observation by ID
func (s *observationService) GetObservation(ctx context.Context, id ulid.ULID) (*observability.Observation, error) {
	if id.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidObservationID,
			"observation ID cannot be empty",
		)
	}

	observation, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	return observation, nil
}

// GetObservationByExternalID retrieves an observation by external ID
func (s *observationService) GetObservationByExternalID(ctx context.Context, externalObservationID string) (*observability.Observation, error) {
	if externalObservationID == "" {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"external observation ID cannot be empty",
		)
	}

	observation, err := s.observationRepo.GetByExternalObservationID(ctx, externalObservationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observation by external ID: %w", err)
	}

	return observation, nil
}

// UpdateObservation updates an existing observation
func (s *observationService) UpdateObservation(ctx context.Context, observation *observability.Observation) (*observability.Observation, error) {
	if observation == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"observation cannot be nil",
		)
	}

	if observation.ID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidObservationID,
			"observation ID cannot be empty",
		)
	}

	// Validate observation data
	if err := s.validateObservation(observation); err != nil {
		return nil, err
	}

	// Update timestamp
	observation.UpdatedAt = time.Now()

	// Update in repository
	if err := s.observationRepo.Update(ctx, observation); err != nil {
		return nil, fmt.Errorf("failed to update observation: %w", err)
	}

	return observation, nil
}

// CompleteObservation completes an observation with final data
func (s *observationService) CompleteObservation(ctx context.Context, id ulid.ULID, completionData *observability.ObservationCompletion) (*observability.Observation, error) {
	if id.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidObservationID,
			"observation ID cannot be empty",
		)
	}

	if completionData == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"completion data cannot be nil",
		)
	}

	// Get current observation
	observation, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	// Check if already completed
	if observation.EndTime != nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeObservationAlreadyCompleted,
			"observation is already completed",
		)
	}

	// Apply completion data
	observation.EndTime = &completionData.EndTime
	observation.UpdatedAt = time.Now()

	if completionData.Output != nil {
		observation.Output = completionData.Output
	}

	if completionData.Usage != nil {
		observation.PromptTokens = completionData.Usage.PromptTokens
		observation.CompletionTokens = completionData.Usage.CompletionTokens
		observation.TotalTokens = completionData.Usage.TotalTokens
	}

	if completionData.Cost != nil {
		inputCost := completionData.Cost.InputCost
		outputCost := completionData.Cost.OutputCost
		totalCost := completionData.Cost.TotalCost

		observation.InputCost = &inputCost
		observation.OutputCost = &outputCost
		observation.TotalCost = &totalCost
	}

	if completionData.QualityScore != nil {
		observation.QualityScore = completionData.QualityScore
	}

	if completionData.StatusMessage != nil {
		observation.StatusMessage = completionData.StatusMessage
	}

	// Calculate latency automatically (will be handled by database trigger, but set here too)
	if observation.EndTime != nil {
		latency := int(completionData.EndTime.Sub(observation.StartTime).Milliseconds())
		observation.LatencyMs = &latency
	}

	// Update in repository
	if err := s.observationRepo.Update(ctx, observation); err != nil {
		return nil, fmt.Errorf("failed to complete observation: %w", err)
	}

	// Publish observation completed event
	event := observability.NewObservationCompletedEvent(observation, nil)
	if publishErr := s.eventPublisher.Publish(ctx, event); publishErr != nil {
		_ = publishErr
	}

	return observation, nil
}

// DeleteObservation deletes an observation by ID
func (s *observationService) DeleteObservation(ctx context.Context, id ulid.ULID) error {
	if id.IsZero() {
		return observability.NewObservabilityError(
			observability.ErrCodeInvalidObservationID,
			"observation ID cannot be empty",
		)
	}

	// Delete from repository
	if err := s.observationRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete observation: %w", err)
	}

	return nil
}

// ListObservations retrieves observations based on filter criteria
func (s *observationService) ListObservations(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, int, error) {
	if filter == nil {
		filter = &observability.ObservationFilter{}
	}

	// Set default pagination if not provided
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	observations, total, err := s.observationRepo.SearchObservations(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list observations: %w", err)
	}

	return observations, total, nil
}

// GetObservationsByTrace retrieves all observations for a specific trace
func (s *observationService) GetObservationsByTrace(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	if traceID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidTraceID,
			"trace ID cannot be empty",
		)
	}

	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observations by trace: %w", err)
	}

	return observations, nil
}

// GetChildObservations retrieves child observations for a parent observation
func (s *observationService) GetChildObservations(ctx context.Context, parentID ulid.ULID) ([]*observability.Observation, error) {
	if parentID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidObservationID,
			"parent observation ID cannot be empty",
		)
	}

	observations, err := s.observationRepo.GetByParentObservationID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child observations: %w", err)
	}

	return observations, nil
}

// CreateObservationsBatch creates multiple observations in a batch operation
func (s *observationService) CreateObservationsBatch(ctx context.Context, observations []*observability.Observation) ([]*observability.Observation, error) {
	if len(observations) == 0 {
		return []*observability.Observation{}, nil
	}

	// Validate and prepare observations
	now := time.Now()
	for _, obs := range observations {
		if obs.ID.IsZero() {
			obs.ID = ulid.New()
		}

		if err := s.validateObservation(obs); err != nil {
			return nil, err
		}

		if obs.CreatedAt.IsZero() {
			obs.CreatedAt = now
		}
		obs.UpdatedAt = now
	}

	// Create batch in repository
	if err := s.observationRepo.CreateBatch(ctx, observations); err != nil {
		return nil, fmt.Errorf("failed to create observations batch: %w", err)
	}

	// Publish events for created observations
	var events []*observability.Event
	for _, obs := range observations {
		event := observability.NewObservationCreatedEvent(obs, nil)
		events = append(events, event)
	}

	if len(events) > 0 {
		if publishErr := s.eventPublisher.PublishBatch(ctx, events); publishErr != nil {
			_ = publishErr
		}
	}

	return observations, nil
}

// UpdateObservationsBatch updates multiple observations in a batch operation
func (s *observationService) UpdateObservationsBatch(ctx context.Context, observations []*observability.Observation) ([]*observability.Observation, error) {
	if len(observations) == 0 {
		return []*observability.Observation{}, nil
	}

	// Validate observations
	now := time.Now()
	for _, obs := range observations {
		if obs.ID.IsZero() {
			return nil, observability.NewObservabilityError(
				observability.ErrCodeInvalidObservationID,
				"observation ID cannot be empty",
			)
		}

		if err := s.validateObservation(obs); err != nil {
			return nil, err
		}

		obs.UpdatedAt = now
	}

	// Update batch in repository
	if err := s.observationRepo.UpdateBatch(ctx, observations); err != nil {
		return nil, fmt.Errorf("failed to update observations batch: %w", err)
	}

	return observations, nil
}

// GetObservationStats retrieves statistics for observations (implementation placeholder)
func (s *observationService) GetObservationStats(ctx context.Context, filter *observability.ObservationFilter) (*observability.ObservationStats, error) {
	// This would implement stats aggregation logic
	return &observability.ObservationStats{}, nil
}

// GetObservationAnalytics retrieves analytics data for observations (implementation placeholder)
func (s *observationService) GetObservationAnalytics(ctx context.Context, filter *observability.AnalyticsFilter) (*observability.ObservationAnalytics, error) {
	// This would implement analytics aggregation logic
	return &observability.ObservationAnalytics{}, nil
}

// CalculateCost calculates the cost for an observation (implementation placeholder)
func (s *observationService) CalculateCost(ctx context.Context, observation *observability.Observation) (*observability.CostCalculation, error) {
	// This would implement cost calculation logic based on provider, model, and token usage
	provider := ""
	model := ""
	if observation.Provider != nil {
		provider = *observation.Provider
	}
	if observation.Model != nil {
		model = *observation.Model
	}

	return &observability.CostCalculation{
		Currency: "USD",
		Provider: provider,
		Model:    model,
	}, nil
}

// GetCostBreakdown retrieves cost breakdown data (implementation placeholder)
func (s *observationService) GetCostBreakdown(ctx context.Context, filter *observability.AnalyticsFilter) ([]*observability.CostBreakdown, error) {
	// This would implement cost breakdown aggregation logic
	return []*observability.CostBreakdown{}, nil
}

// GetLatencyPercentiles retrieves latency percentile data (implementation placeholder)
func (s *observationService) GetLatencyPercentiles(ctx context.Context, filter *observability.ObservationFilter) (*observability.LatencyPercentiles, error) {
	// This would implement latency percentile calculation
	return &observability.LatencyPercentiles{}, nil
}

// GetThroughputMetrics retrieves throughput metrics (implementation placeholder)
func (s *observationService) GetThroughputMetrics(ctx context.Context, filter *observability.AnalyticsFilter) (*observability.ThroughputMetrics, error) {
	// This would implement throughput calculation
	return &observability.ThroughputMetrics{}, nil
}

// Helper methods

// validateObservation validates an observation object
func (s *observationService) validateObservation(observation *observability.Observation) error {
	if observation.TraceID.IsZero() {
		return observability.NewValidationError("trace_id", "trace ID is required")
	}

	if observation.ExternalObservationID == "" {
		return observability.NewValidationError("external_observation_id", "external observation ID is required")
	}

	if observation.Name == "" {
		return observability.NewValidationError("name", "observation name is required")
	}

	if observation.Type == "" {
		return observability.NewValidationError("type", "observation type is required")
	}

	// Validate observation type
	validTypes := []observability.ObservationType{
		observability.ObservationTypeLLM,
		observability.ObservationTypeSpan,
		observability.ObservationTypeEvent,
		observability.ObservationTypeGeneration,
		observability.ObservationTypeRetrieval,
		observability.ObservationTypeEmbedding,
		observability.ObservationTypeAgent,
		observability.ObservationTypeTool,
		observability.ObservationTypeChain,
	}

	isValidType := false
	for _, validType := range validTypes {
		if observation.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return observability.NewValidationError("type", "invalid observation type")
	}

	if observation.StartTime.IsZero() {
		return observability.NewValidationError("start_time", "start time is required")
	}

	return nil
}