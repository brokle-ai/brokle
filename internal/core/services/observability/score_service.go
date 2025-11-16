package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// ScoreService implements business logic for quality score management
type ScoreService struct {
	scoreRepo observability.ScoreRepository
	traceRepo observability.TraceRepository
	spanRepo  observability.SpanRepository
}

// NewScoreService creates a new score service instance
func NewScoreService(
	scoreRepo observability.ScoreRepository,
	traceRepo observability.TraceRepository,
	spanRepo observability.SpanRepository,
) *ScoreService {
	return &ScoreService{
		scoreRepo: scoreRepo,
		traceRepo: traceRepo,
		spanRepo:  spanRepo,
	}
}

// CreateScore creates a new quality score with validation
func (s *ScoreService) CreateScore(ctx context.Context, score *observability.Score) error {
	// Validate required fields
	if score.ProjectID == "" {
		return appErrors.NewValidationError("project_id is required", "score must have a valid project_id")
	}
	if score.Name == "" {
		return appErrors.NewValidationError("name is required", "score name cannot be empty")
	}

	// Validate trace and span IDs are set
	if score.TraceID == "" {
		return appErrors.NewValidationError("trace_id is required", "score must have a trace_id")
	}
	if score.SpanID == "" {
		return appErrors.NewValidationError("span_id is required", "score must have a span_id")
	}

	// Validate data type and value consistency
	if err := s.validateScoreData(score); err != nil {
		return err
	}

	// Generate new ID if not provided
	if score.ID == "" {
		score.ID = ulid.New().String()
	}

	// Set timestamp if not provided
	if score.Timestamp.IsZero() {
		score.Timestamp = time.Now()
	}

	// Validate targets exist
	if err := s.validateScoreTargets(ctx, score); err != nil {
		return err
	}

	// Create score
	if err := s.scoreRepo.Create(ctx, score); err != nil {
		return appErrors.NewInternalError("failed to create score", err)
	}

	return nil
}

// UpdateScore updates an existing score
func (s *ScoreService) UpdateScore(ctx context.Context, score *observability.Score) error {
	// Validate score exists
	existing, err := s.scoreRepo.GetByID(ctx, score.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError("score " + score.ID)
		}
		return appErrors.NewInternalError("failed to get score", err)
	}

	// Merge non-zero fields from incoming score into existing
	mergeScoreFields(existing, score)

	// Preserve version for increment in repository layer

	// Validate data type and value consistency
	if err := s.validateScoreData(existing); err != nil {
		return err
	}

	// Update score
	if err := s.scoreRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update score", err)
	}

	return nil
}

// mergeScoreFields merges non-zero fields from src into dst
func mergeScoreFields(dst *observability.Score, src *observability.Score) {
	// Update optional fields only if non-zero
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.Value != nil {
		dst.Value = src.Value
	}
	if src.StringValue != nil {
		dst.StringValue = src.StringValue
	}
	if src.DataType != "" {
		dst.DataType = src.DataType
	}
	if src.Source != "" {
		dst.Source = src.Source
	}
	if src.Comment != nil {
		dst.Comment = src.Comment
	}
	if src.EvaluatorName != nil {
		dst.EvaluatorName = src.EvaluatorName
	}
	if src.EvaluatorVersion != nil {
		dst.EvaluatorVersion = src.EvaluatorVersion
	}
	if src.EvaluatorConfig != nil {
		dst.EvaluatorConfig = src.EvaluatorConfig
	}
	if src.AuthorUserID != nil && *src.AuthorUserID != "" {
		dst.AuthorUserID = src.AuthorUserID
	}
	if !src.Timestamp.IsZero() {
		dst.Timestamp = src.Timestamp
	}
}

// DeleteScore soft deletes a score
func (s *ScoreService) DeleteScore(ctx context.Context, id string) error {
	// Validate score exists
	_, err := s.scoreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError("score " + id)
		}
		return appErrors.NewInternalError("failed to get score", err)
	}

	// Delete score
	if err := s.scoreRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete score", err)
	}

	return nil
}

// GetScoreByID retrieves a score by ID
func (s *ScoreService) GetScoreByID(ctx context.Context, id string) (*observability.Score, error) {
	score, err := s.scoreRepo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.NewNotFoundError("score " + id)
	}

	return score, nil
}

// GetScoresByTraceID retrieves all scores for a trace
func (s *ScoreService) GetScoresByTraceID(ctx context.Context, traceID string) ([]*observability.Score, error) {
	if traceID == "" {
		return nil, appErrors.NewValidationError("trace_id is required", "scores query requires a valid trace_id")
	}

	scores, err := s.scoreRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores", err)
	}

	return scores, nil
}

// GetScoresBySpanID retrieves all scores for a span
func (s *ScoreService) GetScoresBySpanID(ctx context.Context, spanID string) ([]*observability.Score, error) {
	if spanID == "" {
		return nil, appErrors.NewValidationError("span_id is required", "scores query requires a valid span_id")
	}

	scores, err := s.scoreRepo.GetBySpanID(ctx, spanID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores", err)
	}

	return scores, nil
}

// GetScoresByFilter retrieves scores matching filter criteria
func (s *ScoreService) GetScoresByFilter(ctx context.Context, filter *observability.ScoreFilter) ([]*observability.Score, error) {
	scores, err := s.scoreRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores by filter", err)
	}

	return scores, nil
}

// CreateScoreBatch creates multiple scores in a single batch operation
func (s *ScoreService) CreateScoreBatch(ctx context.Context, scores []*observability.Score) error {
	if len(scores) == 0 {
		return appErrors.NewValidationError("scores array cannot be empty", "batch create requires at least one score")
	}

	// Validate all scores
	for i, score := range scores {
		if score.ProjectID == "" {
			return appErrors.NewValidationError(
				fmt.Sprintf("score[%d]: project_id is required", i),
				"all scores must have valid project_id",
			)
		}
		if score.Name == "" {
			return appErrors.NewValidationError(
				fmt.Sprintf("score[%d]: name is required", i),
				"all scores must have a name",
			)
		}

		// Validate trace and span IDs are set
		if score.TraceID == "" {
			return appErrors.NewValidationError(
				fmt.Sprintf("score[%d]: trace_id is required", i),
				"all scores must have trace_id",
			)
		}
		if score.SpanID == "" {
			return appErrors.NewValidationError(
				fmt.Sprintf("score[%d]: span_id is required", i),
				"all scores must have span_id",
			)
		}

		// Validate data type and value
		if err := s.validateScoreData(score); err != nil {
			return err
		}

		// Generate ID if not provided
		if score.ID == "" {
			score.ID = ulid.New().String()
		}

		// Set timestamp if not provided
		if score.Timestamp.IsZero() {
			score.Timestamp = time.Now()
		}
	}

	// Create batch
	if err := s.scoreRepo.CreateBatch(ctx, scores); err != nil {
		return appErrors.NewInternalError("failed to create score batch", err)
	}

	return nil
}

// CountScores returns the count of scores matching the filter
func (s *ScoreService) CountScores(ctx context.Context, filter *observability.ScoreFilter) (int64, error) {
	count, err := s.scoreRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count scores", err)
	}

	return count, nil
}

// Helper methods

// validateScoreData validates score data type and value consistency
func (s *ScoreService) validateScoreData(score *observability.Score) error {
	switch score.DataType {
	case observability.ScoreDataTypeNumeric:
		if score.Value == nil {
			return appErrors.NewValidationError("numeric score must have value", "value is required for NUMERIC data type")
		}
		if score.StringValue != nil {
			return appErrors.NewValidationError("numeric score cannot have string_value", "string_value not allowed for NUMERIC data type")
		}

	case observability.ScoreDataTypeCategorical:
		if score.StringValue == nil {
			return appErrors.NewValidationError("categorical score must have string_value", "string_value is required for CATEGORICAL data type")
		}
		if score.Value != nil {
			return appErrors.NewValidationError("categorical score cannot have numeric value", "value not allowed for CATEGORICAL data type")
		}

	case observability.ScoreDataTypeBoolean:
		if score.Value == nil {
			return appErrors.NewValidationError("boolean score must have value", "value is required for BOOLEAN data type")
		}
		if *score.Value != 0 && *score.Value != 1 {
			return appErrors.NewValidationError("boolean score value must be 0 or 1", "value must be 0 (false) or 1 (true)")
		}
		if score.StringValue != nil {
			return appErrors.NewValidationError("boolean score cannot have string_value", "string_value not allowed for BOOLEAN data type")
		}

	default:
		return appErrors.NewValidationError("invalid data type", "data_type must be NUMERIC, CATEGORICAL, or BOOLEAN")
	}

	return nil
}

// validateScoreTargets validates that score targets exist
func (s *ScoreService) validateScoreTargets(ctx context.Context, score *observability.Score) error {
	// Validate trace exists
	if score.TraceID != "" {
		_, err := s.traceRepo.GetByID(ctx, score.TraceID)
		if err != nil {
			return appErrors.NewNotFoundError("trace " + score.TraceID)
		}
	}

	// Validate span exists
	if score.SpanID != "" {
		_, err := s.spanRepo.GetByID(ctx, score.SpanID)
		if err != nil {
			return appErrors.NewNotFoundError("span " + score.SpanID)
		}
	}

	return nil
}
