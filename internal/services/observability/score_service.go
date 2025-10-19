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

// ScoreService implements business logic for quality score management
type ScoreService struct {
	scoreRepo       observability.ScoreRepository
	traceRepo       observability.TraceRepository
	observationRepo observability.ObservationRepository
	sessionRepo     observability.SessionRepository
}

// NewScoreService creates a new score service instance
func NewScoreService(
	scoreRepo observability.ScoreRepository,
	traceRepo observability.TraceRepository,
	observationRepo observability.ObservationRepository,
	sessionRepo observability.SessionRepository,
) *ScoreService {
	return &ScoreService{
		scoreRepo:       scoreRepo,
		traceRepo:       traceRepo,
		observationRepo: observationRepo,
		sessionRepo:     sessionRepo,
	}
}

// CreateScore creates a new quality score with validation
func (s *ScoreService) CreateScore(ctx context.Context, score *observability.Score) error {
	// Validate required fields
	if score.ProjectID.IsZero() {
		return appErrors.NewValidationError("project_id is required", "score must have a valid project_id")
	}
	if score.Name == "" {
		return appErrors.NewValidationError("name is required", "score name cannot be empty")
	}

	// Validate at least one target is set (trace, observation, or session)
	if score.TraceID == nil && score.ObservationID == nil && score.SessionID == nil {
		return appErrors.NewValidationError(
			"target is required",
			"score must be attached to trace_id, observation_id, or session_id",
		)
	}

	// Validate data type and value consistency
	if err := s.validateScoreData(score); err != nil {
		return err
	}

	// Generate new ID if not provided
	if score.ID.IsZero() {
		score.ID = ulid.New()
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
		// Only convert sql.ErrNoRows to 404, propagate infrastructure errors as 500
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("score %s", score.ID.String()))
		}
		return appErrors.NewInternalError("failed to get score", err)
	}

	// Merge non-zero fields from incoming score into existing
	mergeScoreFields(existing, score)

	// Preserve version for increment in repository layer
	existing.Version = existing.Version

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
// This prevents zero-value corruption from partial JSON updates
func mergeScoreFields(dst *observability.Score, src *observability.Score) {
	// Immutable fields (never update):
	// - ID (primary key)
	// - ProjectID (security boundary)
	// - TraceID (foreign key, typically immutable)
	// - ObservationID (foreign key, typically immutable)
	// - SessionID (foreign key, typically immutable)
	// - Version (managed by repository)
	// - EventTs (managed by repository)
	// - IsDeleted (managed by Delete method)

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
	// Allow clearing evaluator config by sending empty map {}
	// nil = not sent (preserve), {} = clear, {...} = update
	if src.EvaluatorConfig != nil {
		dst.EvaluatorConfig = src.EvaluatorConfig
	}
	if src.AuthorUserID != nil && !src.AuthorUserID.IsZero() {
		dst.AuthorUserID = src.AuthorUserID
	}
	if !src.Timestamp.IsZero() {
		dst.Timestamp = src.Timestamp
	}
}

// DeleteScore soft deletes a score
func (s *ScoreService) DeleteScore(ctx context.Context, id ulid.ULID) error {
	// Validate score exists
	_, err := s.scoreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("score %s", id.String()))
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
func (s *ScoreService) GetScoreByID(ctx context.Context, id ulid.ULID) (*observability.Score, error) {
	score, err := s.scoreRepo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("score %s", id.String()))
	}

	return score, nil
}

// GetScoresByTraceID retrieves all scores for a trace
func (s *ScoreService) GetScoresByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.Score, error) {
	if traceID.IsZero() {
		return nil, appErrors.NewValidationError("trace_id is required", "scores query requires a valid trace_id")
	}

	scores, err := s.scoreRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores", err)
	}

	return scores, nil
}

// GetScoresByObservationID retrieves all scores for an observation
func (s *ScoreService) GetScoresByObservationID(ctx context.Context, observationID ulid.ULID) ([]*observability.Score, error) {
	if observationID.IsZero() {
		return nil, appErrors.NewValidationError("observation_id is required", "scores query requires a valid observation_id")
	}

	scores, err := s.scoreRepo.GetByObservationID(ctx, observationID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get scores", err)
	}

	return scores, nil
}

// GetScoresBySessionID retrieves all scores for a session
func (s *ScoreService) GetScoresBySessionID(ctx context.Context, sessionID ulid.ULID) ([]*observability.Score, error) {
	if sessionID.IsZero() {
		return nil, appErrors.NewValidationError("session_id is required", "scores query requires a valid session_id")
	}

	scores, err := s.scoreRepo.GetBySessionID(ctx, sessionID)
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
		if score.ProjectID.IsZero() {
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

		// Validate at least one target
		if score.TraceID == nil && score.ObservationID == nil && score.SessionID == nil {
			return appErrors.NewValidationError(
				fmt.Sprintf("score[%d]: target is required", i),
				"all scores must be attached to trace_id, observation_id, or session_id",
			)
		}

		// Validate data type and value
		if err := s.validateScoreData(score); err != nil {
			return err
		}

		// Generate ID if not provided
		if score.ID.IsZero() {
			score.ID = ulid.New()
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
	// Validate trace if provided
	if score.TraceID != nil && !score.TraceID.IsZero() {
		_, err := s.traceRepo.GetByID(ctx, *score.TraceID)
		if err != nil {
			return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", score.TraceID.String()))
		}
	}

	// Validate observation if provided
	if score.ObservationID != nil && !score.ObservationID.IsZero() {
		_, err := s.observationRepo.GetByID(ctx, *score.ObservationID)
		if err != nil {
			return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", score.ObservationID.String()))
		}
	}

	// Validate session if provided
	if score.SessionID != nil && !score.SessionID.IsZero() {
		_, err := s.sessionRepo.GetByID(ctx, *score.SessionID)
		if err != nil {
			return appErrors.NewNotFoundError(fmt.Sprintf("session %s", score.SessionID.String()))
		}
	}

	return nil
}
