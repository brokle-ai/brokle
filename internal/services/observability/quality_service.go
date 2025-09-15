package observability

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// qualityService implements the QualityService interface
type qualityService struct {
	qualityScoreRepo observability.QualityScoreRepository
	traceRepo        observability.TraceRepository
	observationRepo  observability.ObservationRepository
	eventPublisher   observability.EventPublisher
	evaluators       map[string]observability.QualityEvaluator
}

// NewQualityService creates a new quality service
func NewQualityService(
	qualityScoreRepo observability.QualityScoreRepository,
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	eventPublisher observability.EventPublisher,
) observability.QualityService {
	return &qualityService{
		qualityScoreRepo: qualityScoreRepo,
		traceRepo:        traceRepo,
		observationRepo:  observationRepo,
		eventPublisher:   eventPublisher,
		evaluators:       make(map[string]observability.QualityEvaluator),
	}
}

// CreateQualityScore creates a new quality score
func (s *qualityService) CreateQualityScore(ctx context.Context, score *observability.QualityScore) (*observability.QualityScore, error) {
	if score == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"quality score cannot be nil",
		)
	}

	// Generate ID if not provided
	if score.ID.IsZero() {
		score.ID = ulid.New()
	}

	// Validate required fields
	if err := s.validateQualityScore(score); err != nil {
		return nil, err
	}

	// Verify trace exists
	if _, err := s.traceRepo.GetByID(ctx, score.TraceID); err != nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeObservationTraceNotFound,
			"trace not found for quality score",
		).WithDetail("trace_id", score.TraceID.String())
	}

	// Verify observation exists if specified
	if score.ObservationID != nil && !score.ObservationID.IsZero() {
		if _, err := s.observationRepo.GetByID(ctx, *score.ObservationID); err != nil {
			return nil, observability.NewObservabilityError(
				observability.ErrCodeObservationNotFound,
				"observation not found for quality score",
			).WithDetail("observation_id", score.ObservationID.String())
		}
	}

	// Set timestamps
	now := time.Now()
	if score.CreatedAt.IsZero() {
		score.CreatedAt = now
	}
	score.UpdatedAt = now

	// Create quality score in repository
	if err := s.qualityScoreRepo.Create(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to create quality score: %w", err)
	}

	// Publish quality score added event
	event := observability.NewQualityScoreAddedEvent(score, score.AuthorUserID)
	if publishErr := s.eventPublisher.Publish(ctx, event); publishErr != nil {
		_ = publishErr
	}

	return score, nil
}

// GetQualityScore retrieves a quality score by ID
func (s *qualityService) GetQualityScore(ctx context.Context, id ulid.ULID) (*observability.QualityScore, error) {
	if id.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidQualityScoreID,
			"quality score ID cannot be empty",
		)
	}

	score, err := s.qualityScoreRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality score: %w", err)
	}

	return score, nil
}

// UpdateQualityScore updates an existing quality score
func (s *qualityService) UpdateQualityScore(ctx context.Context, score *observability.QualityScore) (*observability.QualityScore, error) {
	if score == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"quality score cannot be nil",
		)
	}

	if score.ID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidQualityScoreID,
			"quality score ID cannot be empty",
		)
	}

	// Validate quality score data
	if err := s.validateQualityScore(score); err != nil {
		return nil, err
	}

	// Update timestamp
	score.UpdatedAt = time.Now()

	// Update in repository
	if err := s.qualityScoreRepo.Update(ctx, score); err != nil {
		return nil, fmt.Errorf("failed to update quality score: %w", err)
	}

	return score, nil
}

// DeleteQualityScore deletes a quality score by ID
func (s *qualityService) DeleteQualityScore(ctx context.Context, id ulid.ULID) error {
	if id.IsZero() {
		return observability.NewObservabilityError(
			observability.ErrCodeInvalidQualityScoreID,
			"quality score ID cannot be empty",
		)
	}

	// Delete from repository
	if err := s.qualityScoreRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete quality score: %w", err)
	}

	return nil
}

// GetQualityScoresByTrace retrieves all quality scores for a specific trace
func (s *qualityService) GetQualityScoresByTrace(ctx context.Context, traceID ulid.ULID) ([]*observability.QualityScore, error) {
	if traceID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidTraceID,
			"trace ID cannot be empty",
		)
	}

	scores, err := s.qualityScoreRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality scores by trace: %w", err)
	}

	return scores, nil
}

// GetQualityScoresByObservation retrieves all quality scores for a specific observation
func (s *qualityService) GetQualityScoresByObservation(ctx context.Context, observationID ulid.ULID) ([]*observability.QualityScore, error) {
	if observationID.IsZero() {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeInvalidObservationID,
			"observation ID cannot be empty",
		)
	}

	scores, err := s.qualityScoreRepo.GetByObservationID(ctx, observationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality scores by observation: %w", err)
	}

	return scores, nil
}

// ListQualityScores retrieves quality scores based on filter criteria
func (s *qualityService) ListQualityScores(ctx context.Context, filter *observability.QualityScoreFilter) ([]*observability.QualityScore, int, error) {
	if filter == nil {
		filter = &observability.QualityScoreFilter{}
	}

	// Set default pagination if not provided
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	// For now, use a simple approach - this would need proper implementation
	// based on available repository methods
	return []*observability.QualityScore{}, 0, fmt.Errorf("list quality scores not fully implemented")
}

// EvaluateTrace evaluates a trace using the specified evaluator
func (s *qualityService) EvaluateTrace(ctx context.Context, traceID ulid.ULID, evaluatorName string) (*observability.QualityScore, error) {
	// Get the evaluator
	evaluator, err := s.GetEvaluator(ctx, evaluatorName)
	if err != nil {
		return nil, err
	}

	// Get the trace with observations
	trace, err := s.traceRepo.GetByID(ctx, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace for evaluation: %w", err)
	}

	// Get observations for the trace
	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observations for evaluation: %w", err)
	}

	// Convert pointers to values
	var obsValues []observability.Observation
	for _, obs := range observations {
		obsValues = append(obsValues, *obs)
	}
	trace.Observations = obsValues

	// Prepare evaluation input
	input := &observability.EvaluationInput{
		TraceID: &traceID,
		Trace:   trace,
		Context: map[string]any{
			"evaluator": evaluatorName,
			"timestamp": time.Now(),
		},
	}

	// Validate input
	if err := evaluator.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("evaluation input validation failed: %w", err)
	}

	// Perform evaluation
	score, err := evaluator.Evaluate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	// Set additional fields
	score.TraceID = traceID
	score.Source = observability.ScoreSourceAuto
	name := evaluatorName
	version := evaluator.Version()
	score.EvaluatorName = &name
	score.EvaluatorVersion = &version

	// Save the score
	createdScore, err := s.CreateQualityScore(ctx, score)
	if err != nil {
		return nil, fmt.Errorf("failed to save evaluation result: %w", err)
	}

	return createdScore, nil
}

// EvaluateObservation evaluates an observation using the specified evaluator
func (s *qualityService) EvaluateObservation(ctx context.Context, observationID ulid.ULID, evaluatorName string) (*observability.QualityScore, error) {
	// Get the evaluator
	evaluator, err := s.GetEvaluator(ctx, evaluatorName)
	if err != nil {
		return nil, err
	}

	// Get the observation
	observation, err := s.observationRepo.GetByID(ctx, observationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get observation for evaluation: %w", err)
	}

	// Prepare evaluation input
	input := &observability.EvaluationInput{
		ObservationID: &observationID,
		Observation:   observation,
		Context: map[string]any{
			"evaluator": evaluatorName,
			"timestamp": time.Now(),
		},
	}

	// Validate input
	if err := evaluator.ValidateInput(input); err != nil {
		return nil, fmt.Errorf("evaluation input validation failed: %w", err)
	}

	// Perform evaluation
	score, err := evaluator.Evaluate(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	// Set additional fields
	score.TraceID = observation.TraceID
	score.ObservationID = &observationID
	score.Source = observability.ScoreSourceAuto
	name := evaluatorName
	version := evaluator.Version()
	score.EvaluatorName = &name
	score.EvaluatorVersion = &version

	// Save the score
	createdScore, err := s.CreateQualityScore(ctx, score)
	if err != nil {
		return nil, fmt.Errorf("failed to save evaluation result: %w", err)
	}

	return createdScore, nil
}

// BulkEvaluate performs bulk evaluation (implementation placeholder)
func (s *qualityService) BulkEvaluate(ctx context.Context, request *observability.BulkEvaluationRequest) (*observability.BulkEvaluationResult, error) {
	if request == nil {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"bulk evaluation request cannot be nil",
		)
	}

	result := &observability.BulkEvaluationResult{
		Scores: []*observability.QualityScore{},
		Errors: []observability.BulkEvaluationError{},
	}

	// Evaluate traces
	for i, traceID := range request.TraceIDs {
		for _, evaluatorName := range request.EvaluatorNames {
			score, err := s.EvaluateTrace(ctx, traceID, evaluatorName)
			if err != nil {
				result.FailedCount++
				result.Errors = append(result.Errors, observability.BulkEvaluationError{
					ItemID:  traceID,
					Error:   err.Error(),
					Details: map[string]any{"type": "trace", "index": i},
				})
				continue
			}
			result.ProcessedCount++
			result.Scores = append(result.Scores, score)
		}
	}

	// Evaluate observations
	for i, obsID := range request.ObservationIDs {
		for _, evaluatorName := range request.EvaluatorNames {
			score, err := s.EvaluateObservation(ctx, obsID, evaluatorName)
			if err != nil {
				result.FailedCount++
				result.Errors = append(result.Errors, observability.BulkEvaluationError{
					ItemID:  obsID,
					Error:   err.Error(),
					Details: map[string]any{"type": "observation", "index": i},
				})
				continue
			}
			result.ProcessedCount++
			result.Scores = append(result.Scores, score)
		}
	}

	// If async, would return a job ID here
	if request.Async {
		jobID := ulid.New().String()
		result.JobID = &jobID
	}

	return result, nil
}

// RegisterEvaluator registers a new quality evaluator
func (s *qualityService) RegisterEvaluator(ctx context.Context, evaluator observability.QualityEvaluator) error {
	if evaluator == nil {
		return observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"evaluator cannot be nil",
		)
	}

	name := evaluator.Name()
	if name == "" {
		return observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"evaluator name cannot be empty",
		)
	}

	s.evaluators[name] = evaluator
	return nil
}

// GetEvaluator retrieves a registered evaluator by name
func (s *qualityService) GetEvaluator(ctx context.Context, name string) (observability.QualityEvaluator, error) {
	if name == "" {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeValidationFailed,
			"evaluator name cannot be empty",
		)
	}

	evaluator, exists := s.evaluators[name]
	if !exists {
		return nil, observability.NewObservabilityError(
			observability.ErrCodeEvaluatorNotFound,
			"evaluator not found",
		).WithDetail("evaluator_name", name)
	}

	return evaluator, nil
}

// ListEvaluators retrieves information about all registered evaluators
func (s *qualityService) ListEvaluators(ctx context.Context) ([]observability.QualityEvaluatorInfo, error) {
	var evaluators []observability.QualityEvaluatorInfo

	for _, evaluator := range s.evaluators {
		evaluators = append(evaluators, observability.QualityEvaluatorInfo{
			Name:           evaluator.Name(),
			Version:        evaluator.Version(),
			Description:    evaluator.Description(),
			SupportedTypes: evaluator.SupportedTypes(),
			IsBuiltIn:      true, // Could be configured differently
			Configuration:  map[string]any{},
		})
	}

	return evaluators, nil
}

// GetQualityAnalytics retrieves quality analytics data (implementation placeholder)
func (s *qualityService) GetQualityAnalytics(ctx context.Context, filter *observability.AnalyticsFilter) (*observability.QualityAnalytics, error) {
	// This would implement quality analytics aggregation logic
	return &observability.QualityAnalytics{}, nil
}

// GetQualityTrends retrieves quality trend data (implementation placeholder)
func (s *qualityService) GetQualityTrends(ctx context.Context, filter *observability.AnalyticsFilter, interval string) ([]*observability.QualityTrendPoint, error) {
	// This would implement quality trends calculation
	return []*observability.QualityTrendPoint{}, nil
}

// GetScoreDistribution retrieves score distribution data (implementation placeholder)
func (s *qualityService) GetScoreDistribution(ctx context.Context, scoreName string, filter *observability.QualityScoreFilter) (map[string]int, error) {
	// This would implement score distribution calculation
	return map[string]int{}, nil
}

// Helper methods

// validateQualityScore validates a quality score object
func (s *qualityService) validateQualityScore(score *observability.QualityScore) error {
	if score.TraceID.IsZero() {
		return observability.NewValidationError("trace_id", "trace ID is required")
	}

	if score.ScoreName == "" {
		return observability.NewValidationError("score_name", "score name is required")
	}

	// Validate data type and corresponding value
	switch score.DataType {
	case observability.ScoreDataTypeNumeric:
		if score.ScoreValue == nil {
			return observability.NewValidationError("score_value", "score value is required for numeric data type")
		}
	case observability.ScoreDataTypeCategorical:
		if score.StringValue == nil || *score.StringValue == "" {
			return observability.NewValidationError("string_value", "string value is required for categorical data type")
		}
	case observability.ScoreDataTypeBoolean:
		if score.ScoreValue == nil {
			return observability.NewValidationError("score_value", "score value is required for boolean data type")
		}
		// For boolean, score_value should be 0 or 1
		if *score.ScoreValue != 0 && *score.ScoreValue != 1 {
			return observability.NewValidationError("score_value", "score value for boolean type must be 0 or 1")
		}
	default:
		return observability.NewValidationError("data_type", "invalid data type")
	}

	return nil
}