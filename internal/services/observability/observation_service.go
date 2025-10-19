package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// ObservationService implements business logic for observation management with cost calculation
type ObservationService struct {
	observationRepo observability.ObservationRepository
	traceRepo       observability.TraceRepository
	scoreRepo       observability.ScoreRepository
}

// NewObservationService creates a new observation service instance
func NewObservationService(
	observationRepo observability.ObservationRepository,
	traceRepo observability.TraceRepository,
	scoreRepo observability.ScoreRepository,
) *ObservationService {
	return &ObservationService{
		observationRepo: observationRepo,
		traceRepo:       traceRepo,
		scoreRepo:       scoreRepo,
	}
}

// CreateObservation creates a new observation with validation
func (s *ObservationService) CreateObservation(ctx context.Context, obs *observability.Observation) error {
	// Validate required fields
	if obs.TraceID.IsZero() {
		return appErrors.NewValidationError("trace_id is required", "observation must be linked to a trace")
	}
	if obs.ProjectID.IsZero() {
		return appErrors.NewValidationError("project_id is required", "observation must have a valid project_id")
	}
	if obs.Name == "" {
		return appErrors.NewValidationError("name is required", "observation name cannot be empty")
	}

	// Generate new ID if not provided
	if obs.ID.IsZero() {
		obs.ID = ulid.New()
	}

	// Validate trace exists
	_, err := s.traceRepo.GetByID(ctx, obs.TraceID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", obs.TraceID.String()))
	}

	// Validate parent observation exists if provided
	if obs.ParentObservationID != nil && !obs.ParentObservationID.IsZero() {
		_, err := s.observationRepo.GetByID(ctx, *obs.ParentObservationID)
		if err != nil {
			return appErrors.NewNotFoundError(fmt.Sprintf("parent observation %s", obs.ParentObservationID.String()))
		}
	}

	// Calculate total cost and tokens if details provided
	s.calculateAggregates(obs)

	// Create observation
	if err := s.observationRepo.Create(ctx, obs); err != nil {
		return appErrors.NewInternalError("failed to create observation", err)
	}

	return nil
}

// UpdateObservation updates an existing observation
func (s *ObservationService) UpdateObservation(ctx context.Context, obs *observability.Observation) error {
	// Validate observation exists
	existing, err := s.observationRepo.GetByID(ctx, obs.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", obs.ID.String()))
		}
		return appErrors.NewInternalError("failed to get observation", err)
	}

	// Merge non-zero fields from incoming observation into existing
	mergeObservationFields(existing, obs)

	// Preserve version for increment in repository layer
	existing.Version = existing.Version

	// Recalculate aggregates if cost/usage details changed
	s.calculateAggregates(existing)

	// Update observation
	if err := s.observationRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update observation", err)
	}

	return nil
}

// mergeObservationFields merges non-zero fields from src into dst
// This prevents zero-value corruption from partial JSON updates
func mergeObservationFields(dst *observability.Observation, src *observability.Observation) {
	// Immutable fields (never update):
	// - ID (primary key)
	// - TraceID (foreign key, security boundary)
	// - ProjectID (security boundary)
	// - Version (managed by repository)
	// - EventTs (managed by repository)
	// - IsDeleted (managed by Delete method)

	// Update optional fields only if non-zero
	if src.ParentObservationID != nil && !src.ParentObservationID.IsZero() {
		dst.ParentObservationID = src.ParentObservationID
	}
	if src.Type != "" {
		dst.Type = src.Type
	}
	if src.Name != "" {
		dst.Name = src.Name
	}
	if !src.StartTime.IsZero() {
		dst.StartTime = src.StartTime
	}
	if src.EndTime != nil && !src.EndTime.IsZero() {
		dst.EndTime = src.EndTime
	}
	if src.Model != nil {
		dst.Model = src.Model
	}
	// Allow clearing maps by sending empty map {}
	// nil = not sent (preserve), {} = clear, {...} = update
	if src.ModelParameters != nil {
		dst.ModelParameters = src.ModelParameters
	}
	if src.Input != nil {
		dst.Input = src.Input
	}
	if src.Output != nil {
		dst.Output = src.Output
	}
	if src.Metadata != nil {
		dst.Metadata = src.Metadata
	}
	if src.CostDetails != nil {
		dst.CostDetails = src.CostDetails
	}
	if src.UsageDetails != nil {
		dst.UsageDetails = src.UsageDetails
	}
	if src.Level != "" {
		dst.Level = src.Level
	}
	if src.StatusMessage != nil {
		dst.StatusMessage = src.StatusMessage
	}
	if src.CompletionStartTime != nil && !src.CompletionStartTime.IsZero() {
		dst.CompletionStartTime = src.CompletionStartTime
	}
	if src.TimeToFirstTokenMs != nil {
		dst.TimeToFirstTokenMs = src.TimeToFirstTokenMs
	}
}

// DeleteObservation soft deletes an observation
func (s *ObservationService) DeleteObservation(ctx context.Context, id ulid.ULID) error {
	// Validate observation exists
	_, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", id.String()))
		}
		return appErrors.NewInternalError("failed to get observation", err)
	}

	// Delete observation
	if err := s.observationRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete observation", err)
	}

	return nil
}

// GetObservationByID retrieves an observation by ID
func (s *ObservationService) GetObservationByID(ctx context.Context, id ulid.ULID) (*observability.Observation, error) {
	obs, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("observation %s", id.String()))
	}

	return obs, nil
}

// GetObservationsByTraceID retrieves all observations for a trace
func (s *ObservationService) GetObservationsByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	if traceID.IsZero() {
		return nil, appErrors.NewValidationError("trace_id is required", "observations query requires a valid trace_id")
	}

	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observations", err)
	}

	return observations, nil
}

// GetObservationTreeByTraceID retrieves observations in hierarchical tree structure
func (s *ObservationService) GetObservationTreeByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Observation, error) {
	if traceID.IsZero() {
		return nil, appErrors.NewValidationError("trace_id is required", "observation tree query requires a valid trace_id")
	}

	// Validate trace exists
	_, err := s.traceRepo.GetByID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("trace %s", traceID.String()))
	}

	observations, err := s.observationRepo.GetTreeByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observation tree", err)
	}

	return observations, nil
}

// GetChildObservations retrieves child observations of a parent
func (s *ObservationService) GetChildObservations(ctx context.Context, parentObservationID ulid.ULID) ([]*observability.Observation, error) {
	if parentObservationID.IsZero() {
		return nil, appErrors.NewValidationError("parent_observation_id is required", "parent_observation_id cannot be empty")
	}

	// Validate parent exists
	_, err := s.observationRepo.GetByID(ctx, parentObservationID)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("parent observation %s", parentObservationID.String()))
	}

	observations, err := s.observationRepo.GetChildren(ctx, parentObservationID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get child observations", err)
	}

	return observations, nil
}

// GetObservationsByFilter retrieves observations matching filter criteria
func (s *ObservationService) GetObservationsByFilter(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, error) {
	observations, err := s.observationRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observations by filter", err)
	}

	return observations, nil
}

// CreateObservationBatch creates multiple observations in a single batch operation
func (s *ObservationService) CreateObservationBatch(ctx context.Context, observations []*observability.Observation) error {
	if len(observations) == 0 {
		return appErrors.NewValidationError("observations array cannot be empty", "batch create requires at least one observation")
	}

	// Validate all observations
	for i, obs := range observations {
		if obs.TraceID.IsZero() {
			return appErrors.NewValidationError(
				fmt.Sprintf("observation[%d]: trace_id is required", i),
				"all observations must be linked to a trace",
			)
		}
		if obs.ProjectID.IsZero() {
			return appErrors.NewValidationError(
				fmt.Sprintf("observation[%d]: project_id is required", i),
				"all observations must have valid project_id",
			)
		}
		if obs.Name == "" {
			return appErrors.NewValidationError(
				fmt.Sprintf("observation[%d]: name is required", i),
				"all observations must have a name",
			)
		}

		// Generate ID if not provided
		if obs.ID.IsZero() {
			obs.ID = ulid.New()
		}

		// Calculate aggregates
		s.calculateAggregates(obs)
	}

	// Create batch
	if err := s.observationRepo.CreateBatch(ctx, observations); err != nil {
		return appErrors.NewInternalError("failed to create observation batch", err)
	}

	return nil
}

// CountObservations returns the count of observations matching the filter
func (s *ObservationService) CountObservations(ctx context.Context, filter *observability.ObservationFilter) (int64, error) {
	count, err := s.observationRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count observations", err)
	}

	return count, nil
}

// SetObservationCost sets cost details for an observation with automatic total calculation
func (s *ObservationService) SetObservationCost(ctx context.Context, observationID ulid.ULID, inputCost, outputCost float64) error {
	// Get observation
	obs, err := s.observationRepo.GetByID(ctx, observationID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", observationID.String()))
	}

	// Set cost details using domain helper
	obs.SetCostDetails(inputCost, outputCost)

	// Update observation
	if err := s.observationRepo.Update(ctx, obs); err != nil {
		return appErrors.NewInternalError("failed to update observation cost", err)
	}

	return nil
}

// SetObservationUsage sets usage details for an observation with automatic total calculation
func (s *ObservationService) SetObservationUsage(ctx context.Context, observationID ulid.ULID, promptTokens, completionTokens uint64) error {
	// Get observation
	obs, err := s.observationRepo.GetByID(ctx, observationID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", observationID.String()))
	}

	// Set usage details using domain helper
	obs.SetUsageDetails(promptTokens, completionTokens)

	// Update observation
	if err := s.observationRepo.Update(ctx, obs); err != nil {
		return appErrors.NewInternalError("failed to update observation usage", err)
	}

	return nil
}

// CalculateTraceCost aggregates total cost from all observations in a trace
func (s *ObservationService) CalculateTraceCost(ctx context.Context, traceID ulid.ULID) (float64, error) {
	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to get observations for cost calculation", err)
	}

	var totalCost float64
	for _, obs := range observations {
		totalCost += obs.GetTotalCost()
	}

	return totalCost, nil
}

// CalculateTraceTokens aggregates total tokens from all observations in a trace
func (s *ObservationService) CalculateTraceTokens(ctx context.Context, traceID ulid.ULID) (uint64, error) {
	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to get observations for token calculation", err)
	}

	var totalTokens uint64
	for _, obs := range observations {
		totalTokens += obs.GetTotalTokens()
	}

	return totalTokens, nil
}

// Helper methods

// calculateAggregates ensures cost and usage totals are calculated from details maps
func (s *ObservationService) calculateAggregates(obs *observability.Observation) {
	// Calculate total cost if individual costs provided
	if obs.CostDetails != nil {
		if _, hasTotal := obs.CostDetails["total"]; !hasTotal {
			inputCost := obs.CostDetails["input"]
			outputCost := obs.CostDetails["output"]
			obs.CostDetails["total"] = inputCost + outputCost
		}
	}

	// Calculate total tokens if individual counts provided
	if obs.UsageDetails != nil {
		if _, hasTotal := obs.UsageDetails["total_tokens"]; !hasTotal {
			promptTokens := obs.UsageDetails["prompt_tokens"]
			completionTokens := obs.UsageDetails["completion_tokens"]
			obs.UsageDetails["total_tokens"] = promptTokens + completionTokens
		}
	}
}
