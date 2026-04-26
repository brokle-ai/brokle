package evaluation

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
	"brokle/pkg/uid"
)

func (h *Handler) SdkCreateScore(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	var body CreateScoreRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.validateScoreAgainstConfig(r.Context(), projectID, body.Name, body.Type, body.Value, body.StringValue); err != nil {
		response.WriteError(w, err)
		return
	}

	score := h.buildScore(projectID, &body)
	if err := h.scoreSvc.CreateScore(r.Context(), score); err != nil {
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "evaluation: score created via SDK",
		"score_id", score.ID,
		"project_id", projectID,
		"trace_id", score.TraceID,
		"name", score.Name,
	)

	response.Created(w, toSubmittedScoreResponse(score))
}

func (h *Handler) SdkCreateScoreBatch(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	var body BatchScoreRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if len(body.Scores) == 0 {
		response.WriteError(w, appErrors.NewValidationError(
			"Invalid request body", "scores array cannot be empty"))
		return
	}

	scores := make([]*observability.Score, 0, len(body.Scores))
	for i, sr := range body.Scores {
		if err := h.validateScoreAgainstConfig(r.Context(), projectID, sr.Name, sr.Type, sr.Value, sr.StringValue); err != nil {
			h.logger.WarnContext(r.Context(), "evaluation: score validation failed",
				"index", i, "name", sr.Name, "error", err.Error(),
			)
			response.WriteError(w, err)
			return
		}
		scores = append(scores, h.buildScore(projectID, &sr))
	}

	if err := h.scoreSvc.CreateScoreBatch(r.Context(), scores); err != nil {
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "evaluation: batch scores created via SDK",
		"project_id", projectID, "count", len(scores),
	)

	response.Created(w, &BatchScoreResponse{Created: len(scores)})
}

// validateScoreAgainstConfig verifies a score against the matching
// ScoreConfig when one exists. Absence of a config is allowed
// (unvalidated ingestion); lookup errors other than not-found fail
// closed.
func (h *Handler) validateScoreAgainstConfig(
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

func (h *Handler) buildScore(projectID uuid.UUID, req *CreateScoreRequest) *observability.Score {
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

	// Span defaults to the trace for trace-linked scores; experiment-
	// only scores leave both nil.
	if req.SpanID != nil {
		score.SpanID = req.SpanID
	} else if req.TraceID != nil {
		score.SpanID = req.TraceID
	}
	return score
}

func toSubmittedScoreResponse(s *observability.Score) *SubmittedScoreResponse {
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
