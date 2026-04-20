package evaluation

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/uid"
)

// registerSDKScoreRoutes wires the SDK-plane score ingestion endpoints.
// Called from RegisterSDKRoutes in handlers.go.
func registerSDKScoreRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "sdk-create-score",
		Method:        http.MethodPost,
		Path:          "/v1/scores",
		Tags:          []string{"SDK - scores"},
		Summary:       "Create a score (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkCreateScore)

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-create-score-batch",
		Method:        http.MethodPost,
		Path:          "/v1/scores/batch",
		Tags:          []string{"SDK - scores"},
		Summary:       "Batch-create scores (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkCreateScoreBatch)
}

func (h *handler) sdkCreateScore(ctx context.Context, in *SDKCreateScoreInput) (*SDKCreateScoreOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)

	if err := h.validateScoreAgainstConfig(ctx, projectID, in.Body.Name, in.Body.Type, in.Body.Value, in.Body.StringValue); err != nil {
		return nil, err
	}

	score := h.buildScore(projectID, &in.Body)
	if err := h.scoreSvc.CreateScore(ctx, score); err != nil {
		return nil, err
	}

	h.logger.InfoContext(ctx, "evaluation: score created via SDK",
		"score_id", score.ID,
		"project_id", projectID,
		"trace_id", score.TraceID,
		"name", score.Name,
	)

	return &SDKCreateScoreOutput{Body: toSubmittedScoreResponse(score)}, nil
}

func (h *handler) sdkCreateScoreBatch(ctx context.Context, in *SDKCreateScoreBatchInput) (*SDKCreateScoreBatchOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)

	if len(in.Body.Scores) == 0 {
		return nil, appErrors.NewValidationError("Invalid request body", "scores array cannot be empty")
	}

	scores := make([]*observability.Score, 0, len(in.Body.Scores))
	for i, sr := range in.Body.Scores {
		if err := h.validateScoreAgainstConfig(ctx, projectID, sr.Name, sr.Type, sr.Value, sr.StringValue); err != nil {
			h.logger.WarnContext(ctx, "evaluation: score validation failed",
				"index", i, "name", sr.Name, "error", err.Error(),
			)
			return nil, err
		}
		scores = append(scores, h.buildScore(projectID, &sr))
	}

	if err := h.scoreSvc.CreateScoreBatch(ctx, scores); err != nil {
		return nil, err
	}

	h.logger.InfoContext(ctx, "evaluation: batch scores created via SDK",
		"project_id", projectID, "count", len(scores),
	)

	return &SDKCreateScoreBatchOutput{Body: &BatchScoreResponse{Created: len(scores)}}, nil
}

// validateScoreAgainstConfig verifies a score against the matching ScoreConfig
// when one exists. Absence of a config is allowed (unvalidated ingestion);
// lookup errors other than not-found fail closed.
func (h *handler) validateScoreAgainstConfig(
	ctx context.Context,
	projectID uuid.UUID,
	name, scoreType string,
	value *float64,
	stringValue *string,
) error {
	cfg, err := h.scoreConfigSvc.GetByName(ctx, projectID, name)
	if err != nil {
		if appErrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if string(cfg.Type) != scoreType {
		return appErrors.NewValidationError("type",
			"must match score config (expected: "+string(cfg.Type)+")")
	}

	switch cfg.Type {
	case evaluationDomain.ScoreTypeNumeric:
		if value == nil {
			return appErrors.NewValidationError("value", "required for NUMERIC type")
		}
		if cfg.MinValue != nil && *value < *cfg.MinValue {
			return appErrors.NewValidationError("value", "below minimum configured value")
		}
		if cfg.MaxValue != nil && *value > *cfg.MaxValue {
			return appErrors.NewValidationError("value", "above maximum configured value")
		}
	case evaluationDomain.ScoreTypeCategorical:
		if stringValue == nil {
			return appErrors.NewValidationError("string_value", "required for CATEGORICAL type")
		}
		found := false
		for _, cat := range cfg.Categories {
			if cat == *stringValue {
				found = true
				break
			}
		}
		if !found {
			return appErrors.NewValidationError("string_value", "not in allowed categories")
		}
	case evaluationDomain.ScoreTypeBoolean:
		if value == nil && stringValue == nil {
			return appErrors.NewValidationError("value", "required for BOOLEAN type (0 or 1)")
		}
		if value != nil && *value != 0 && *value != 1 {
			return appErrors.NewValidationError("value", "must be 0 or 1 for BOOLEAN type")
		}
	}
	return nil
}

func (h *handler) buildScore(projectID uuid.UUID, req *CreateScoreRequest) *observability.Score {
	metadata := json.RawMessage("{}")
	if req.Metadata != nil {
		if b, err := json.Marshal(req.Metadata); err == nil {
			metadata = b
		}
	}

	score := &observability.Score{
		ID:               uid.New(),
		ProjectID:        projectID,
		TraceID:          req.TraceID,
		Name:             req.Name,
		Type:             req.Type,
		Value:            req.Value,
		StringValue:      req.StringValue,
		Source:           observability.ScoreSourceAPI,
		Reason:           req.Reason,
		Metadata:         metadata,
		ExperimentID:     req.ExperimentID,
		ExperimentItemID: req.ExperimentItemID,
		Timestamp:        time.Now(),
	}

	// Span defaults to the trace for trace-linked scores; experiment-only
	// scores leave both nil.
	if req.SpanID != nil {
		score.SpanID = req.SpanID
	} else if req.TraceID != nil {
		score.SpanID = req.TraceID
	}
	return score
}

func toSubmittedScoreResponse(s *observability.Score) *SubmittedScoreResponse {
	// Metadata on the entity is json.RawMessage; sanitize any malformed
	// legacy bytes by JSON-escaping them into a string. Valid JSON passes
	// through unchanged.
	metadata := s.Metadata
	if len(metadata) > 0 && !json.Valid(metadata) {
		metadata, _ = json.Marshal(string(metadata))
	}
	return &SubmittedScoreResponse{
		ID:               s.ID,
		ProjectID:        s.ProjectID,
		TraceID:          s.TraceID,
		SpanID:           s.SpanID,
		Name:             s.Name,
		Value:            s.Value,
		StringValue:      s.StringValue,
		Type:             s.Type,
		Source:           s.Source,
		Reason:           s.Reason,
		Metadata:         metadata,
		ExperimentID:     s.ExperimentID,
		ExperimentItemID: s.ExperimentItemID,
		Timestamp:        s.Timestamp,
	}
}
