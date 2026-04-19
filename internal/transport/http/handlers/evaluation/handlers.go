// Package evaluation exposes evaluation-domain operations on both surfaces:
//
//   - Dashboard plane (apiAdmin, RequireAuth) — score-configs, evaluators,
//     executions, experiment wizard, plus project-scoped dataset /
//     experiment CRUD mounted under /api/v1/projects/{projectId}/...
//   - SDK plane (apiPublic, RequireSDKAuth) — dataset + experiment +
//     score ingestion under /v1/... with the project derived from the
//     API key.
//
// The two registration entrypoints (RegisterRoutes / RegisterSDKRoutes)
// share handler-method implementations; the only difference is how the
// project ID is sourced (path parameter vs. SDK auth context).
package evaluation

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/uid"
)

// ---- handler aggregate ----------------------------------------------

type handler struct {
	scoreConfigSvc       evaluationDomain.ScoreConfigService
	datasetSvc           evaluationDomain.DatasetService
	datasetItemSvc       evaluationDomain.DatasetItemService
	datasetVersionSvc    evaluationDomain.DatasetVersionService
	experimentSvc        evaluationDomain.ExperimentService
	experimentItemSvc    evaluationDomain.ExperimentItemService
	experimentWizardSvc  evaluationDomain.ExperimentWizardService
	evaluatorSvc         evaluationDomain.EvaluatorService
	evaluatorExecSvc     evaluationDomain.EvaluatorExecutionService
	scoreSvc             *obsServices.ScoreService
	logger               *slog.Logger
}

// RegisterRoutes wires the dashboard-plane evaluation operations on apiAdmin.
func RegisterRoutes(
	api huma.API,
	scoreConfigSvc evaluationDomain.ScoreConfigService,
	datasetSvc evaluationDomain.DatasetService,
	datasetItemSvc evaluationDomain.DatasetItemService,
	datasetVersionSvc evaluationDomain.DatasetVersionService,
	experimentSvc evaluationDomain.ExperimentService,
	experimentItemSvc evaluationDomain.ExperimentItemService,
	experimentWizardSvc evaluationDomain.ExperimentWizardService,
	evaluatorSvc evaluationDomain.EvaluatorService,
	evaluatorExecSvc evaluationDomain.EvaluatorExecutionService,
	logger *slog.Logger,
) {
	h := &handler{
		scoreConfigSvc:      scoreConfigSvc,
		datasetSvc:          datasetSvc,
		datasetItemSvc:      datasetItemSvc,
		datasetVersionSvc:   datasetVersionSvc,
		experimentSvc:       experimentSvc,
		experimentItemSvc:   experimentItemSvc,
		experimentWizardSvc: experimentWizardSvc,
		evaluatorSvc:        evaluatorSvc,
		evaluatorExecSvc:    evaluatorExecSvc,
		logger:              logger,
	}

	registerScoreConfigRoutes(api, h)
	registerDatasetRoutes(api, h)
	registerExperimentRoutes(api, h)
	registerEvaluatorRoutes(api, h)
	registerExecutionRoutes(api, h)
	registerWizardRoutes(api, h)
}

// RegisterSDKRoutes wires the SDK-plane evaluation operations on apiPublic.
// The project is derived from the API key (MustGetProjectID).
func RegisterSDKRoutes(
	api huma.API,
	scoreConfigSvc evaluationDomain.ScoreConfigService,
	datasetSvc evaluationDomain.DatasetService,
	datasetItemSvc evaluationDomain.DatasetItemService,
	datasetVersionSvc evaluationDomain.DatasetVersionService,
	experimentSvc evaluationDomain.ExperimentService,
	experimentItemSvc evaluationDomain.ExperimentItemService,
	scoreSvc *obsServices.ScoreService,
	logger *slog.Logger,
) {
	h := &handler{
		scoreConfigSvc:    scoreConfigSvc,
		datasetSvc:        datasetSvc,
		datasetItemSvc:    datasetItemSvc,
		datasetVersionSvc: datasetVersionSvc,
		experimentSvc:     experimentSvc,
		experimentItemSvc: experimentItemSvc,
		scoreSvc:          scoreSvc,
		logger:            logger,
	}

	registerSDKDatasetRoutes(api, h)
	registerSDKExperimentRoutes(api, h)
	registerSDKScoreRoutes(api, h)
}

// ---- shared parsers ---------------------------------------------------

func parseProjectID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	return id, nil
}

func parseDatasetID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid dataset ID", "datasetId must be a valid UUID")
	}
	return id, nil
}

func parseVersionID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid version ID", "versionId must be a valid UUID")
	}
	return id, nil
}

func parseItemID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid item ID", "itemId must be a valid UUID")
	}
	return id, nil
}

func parseExperimentID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid experiment ID", "experimentId must be a valid UUID")
	}
	return id, nil
}

func parseEvaluatorID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid evaluator ID", "evaluatorId must be a valid UUID")
	}
	return id, nil
}

func parseExecutionID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid execution ID", "executionId must be a valid UUID")
	}
	return id, nil
}

func parseScoreConfigID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid score config ID", "configId must be a valid UUID")
	}
	return id, nil
}

func userIDPtr(ctx context.Context) *uuid.UUID {
	id, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &id
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || !pagination.IsValidPageSize(limit) {
		limit = pagination.DefaultPageSize
	}
	return page, limit
}

// pageList is the canonical paginated list envelope for evaluation responses.
type pageList[T any] struct {
	Data  []T   `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// ---- shared DTOs (also used by SDK payloads) ------------------------

// ScoreType is the score-config enum carried in request bodies.
type ScoreType string

const (
	ScoreTypeNumeric     ScoreType = "NUMERIC"
	ScoreTypeCategorical ScoreType = "CATEGORICAL"
	ScoreTypeBoolean     ScoreType = "BOOLEAN"
)

// DatasetItemResponse is the serialisation shape used by all item-returning
// endpoints. Identical to the underlying evaluationDomain.DatasetItemResponse
// but with `source` pre-stringified for stable JSON regardless of future
// enum changes.
type DatasetItemResponse struct {
	ID            uuid.UUID      `json:"id"`
	DatasetID     uuid.UUID      `json:"dataset_id"`
	Input         map[string]any `json:"input"`
	Expected      map[string]any `json:"expected,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Source        string         `json:"source"`
	SourceTraceID *string        `json:"source_trace_id,omitempty"`
	SourceSpanID  *string        `json:"source_span_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

func toDatasetItemResponse(item *evaluationDomain.DatasetItem) *DatasetItemResponse {
	r := item.ToResponse()
	return &DatasetItemResponse{
		ID:            r.ID,
		DatasetID:     r.DatasetID,
		Input:         r.Input,
		Expected:      r.Expected,
		Metadata:      r.Metadata,
		Source:        string(r.Source),
		SourceTraceID: r.SourceTraceID,
		SourceSpanID:  r.SourceSpanID,
		CreatedAt:     r.CreatedAt,
	}
}

// ExperimentItemResponse mirrors evaluationDomain.ExperimentItemResponse
// so the handler can keep a single shape across endpoints.
type ExperimentItemResponse struct {
	ID            uuid.UUID      `json:"id"`
	ExperimentID  uuid.UUID      `json:"experiment_id"`
	DatasetItemID *uuid.UUID     `json:"dataset_item_id,omitempty"`
	TraceID       *string        `json:"trace_id,omitempty"`
	Input         map[string]any `json:"input"`
	Output        any            `json:"output,omitempty"`
	Expected      any            `json:"expected,omitempty"`
	TrialNumber   int            `json:"trial_number"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

func toExperimentItemResponse(item *evaluationDomain.ExperimentItem) *ExperimentItemResponse {
	r := item.ToResponse()
	return &ExperimentItemResponse{
		ID:            r.ID,
		ExperimentID:  r.ExperimentID,
		DatasetItemID: r.DatasetItemID,
		TraceID:       r.TraceID,
		Input:         r.Input,
		Output:        r.Output,
		Expected:      r.Expected,
		TrialNumber:   r.TrialNumber,
		Metadata:      r.Metadata,
		CreatedAt:     r.CreatedAt,
	}
}

// BulkImportResponse is the result of a bulk dataset-item import.
type BulkImportResponse struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

func toBulkImportResponse(r *evaluationDomain.BulkImportResult) *BulkImportResponse {
	return &BulkImportResponse{
		Created: r.Created,
		Skipped: r.Skipped,
		Errors:  r.Errors,
	}
}

// KeysMappingRequest is the optional field-extraction mapping shared by
// import endpoints.
type KeysMappingRequest struct {
	InputKeys    []string `json:"input_keys,omitempty"`
	ExpectedKeys []string `json:"expected_keys,omitempty"`
	MetadataKeys []string `json:"metadata_keys,omitempty"`
}

func (k *KeysMappingRequest) toDomain() *evaluationDomain.KeysMapping {
	if k == nil {
		return nil
	}
	return &evaluationDomain.KeysMapping{
		InputKeys:    k.InputKeys,
		ExpectedKeys: k.ExpectedKeys,
		MetadataKeys: k.MetadataKeys,
	}
}

// CountResponse is returned by SDK batch creates (items + experiment items).
type CountResponse struct {
	Created int `json:"created"`
}

// ---- score config (dashboard only) ---------------------------------

type CreateScoreConfigRequest struct {
	Name        string         `json:"name" minLength:"1" maxLength:"100"`
	Description *string        `json:"description,omitempty"`
	Type        ScoreType      `json:"type" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateScoreConfigRequest struct {
	Name        *string        `json:"name,omitempty" minLength:"1" maxLength:"100"`
	Description *string        `json:"description,omitempty"`
	Type        *ScoreType     `json:"type,omitempty" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	MinValue    *float64       `json:"min_value,omitempty"`
	MaxValue    *float64       `json:"max_value,omitempty"`
	Categories  []string       `json:"categories,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func registerScoreConfigRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-score-config",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/score-configs",
		Tags:          []string{"score-configs"},
		Summary:       "Create a score config",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createScoreConfig)

	huma.Register(api, huma.Operation{
		OperationID: "list-score-configs",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/score-configs",
		Tags:        []string{"score-configs"},
		Summary:     "List score configs",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listScoreConfigs)

	huma.Register(api, huma.Operation{
		OperationID: "get-score-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/score-configs/{configId}",
		Tags:        []string{"score-configs"},
		Summary:     "Get a score config",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getScoreConfig)

	huma.Register(api, huma.Operation{
		OperationID: "update-score-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/score-configs/{configId}",
		Tags:        []string{"score-configs"},
		Summary:     "Update a score config",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateScoreConfig)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-score-config",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/score-configs/{configId}",
		Tags:          []string{"score-configs"},
		Summary:       "Delete a score config",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteScoreConfig)
}

type CreateScoreConfigInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CreateScoreConfigRequest
}
type ScoreConfigOutput struct {
	Body *evaluationDomain.ScoreConfigResponse
}

func (h *handler) createScoreConfig(ctx context.Context, in *CreateScoreConfigInput) (*ScoreConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	domainReq := &evaluationDomain.CreateScoreConfigRequest{
		Name:        in.Body.Name,
		Description: in.Body.Description,
		Type:        evaluationDomain.ScoreType(in.Body.Type),
		MinValue:    in.Body.MinValue,
		MaxValue:    in.Body.MaxValue,
		Categories:  in.Body.Categories,
		Metadata:    in.Body.Metadata,
	}
	cfg, err := h.scoreConfigSvc.Create(ctx, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &ScoreConfigOutput{Body: cfg.ToResponse()}, nil
}

type ListScoreConfigsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}
type ListScoreConfigsOutput struct {
	Body pageList[*evaluationDomain.ScoreConfigResponse]
}

func (h *handler) listScoreConfigs(ctx context.Context, in *ListScoreConfigsInput) (*ListScoreConfigsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	page, limit := normalizePagination(in.Page, in.Limit)
	cfgs, total, err := h.scoreConfigSvc.List(ctx, projectID, page, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*evaluationDomain.ScoreConfigResponse, len(cfgs))
	for i, c := range cfgs {
		out[i] = c.ToResponse()
	}
	return &ListScoreConfigsOutput{Body: pageList[*evaluationDomain.ScoreConfigResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

type GetScoreConfigInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ConfigID  string `path:"configId" format:"uuid"`
}

func (h *handler) getScoreConfig(ctx context.Context, in *GetScoreConfigInput) (*ScoreConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	configID, err := parseScoreConfigID(in.ConfigID)
	if err != nil {
		return nil, err
	}
	cfg, err := h.scoreConfigSvc.GetByID(ctx, configID, projectID)
	if err != nil {
		return nil, err
	}
	return &ScoreConfigOutput{Body: cfg.ToResponse()}, nil
}

type UpdateScoreConfigInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	ConfigID  string `path:"configId" format:"uuid"`
	Body      UpdateScoreConfigRequest
}

func (h *handler) updateScoreConfig(ctx context.Context, in *UpdateScoreConfigInput) (*ScoreConfigOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	configID, err := parseScoreConfigID(in.ConfigID)
	if err != nil {
		return nil, err
	}

	var scoreType *evaluationDomain.ScoreType
	if in.Body.Type != nil {
		st := evaluationDomain.ScoreType(*in.Body.Type)
		scoreType = &st
	}
	domainReq := &evaluationDomain.UpdateScoreConfigRequest{
		Name:        in.Body.Name,
		Description: in.Body.Description,
		Type:        scoreType,
		MinValue:    in.Body.MinValue,
		MaxValue:    in.Body.MaxValue,
		Categories:  in.Body.Categories,
		Metadata:    in.Body.Metadata,
	}
	cfg, err := h.scoreConfigSvc.Update(ctx, configID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &ScoreConfigOutput{Body: cfg.ToResponse()}, nil
}

type DeleteScoreConfigInput = GetScoreConfigInput
type EmptyOutput struct{}

func (h *handler) deleteScoreConfig(ctx context.Context, in *DeleteScoreConfigInput) (*EmptyOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	configID, err := parseScoreConfigID(in.ConfigID)
	if err != nil {
		return nil, err
	}
	if err := h.scoreConfigSvc.Delete(ctx, configID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

// ---- SDK-plane score ingestion --------------------------------------

type CreateScoreRequest struct {
	TraceID          *string        `json:"trace_id,omitempty"`
	SpanID           *string        `json:"span_id,omitempty"`
	Name             string         `json:"name"`
	Value            *float64       `json:"value,omitempty"`
	StringValue      *string        `json:"string_value,omitempty"`
	Type             string         `json:"type" enum:"NUMERIC,CATEGORICAL,BOOLEAN"`
	Reason           *string        `json:"reason,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	ExperimentID     *uuid.UUID     `json:"experiment_id,omitempty"`
	ExperimentItemID *string        `json:"experiment_item_id,omitempty"`
}

type BatchScoreRequest struct {
	Scores []CreateScoreRequest `json:"scores"`
}

// ScoreResponse exposes the observability.Score DTO. Metadata is a
// json.RawMessage on the domain entity — we sanitize it for safe JSON
// marshaling in toScoreResponse.
type ScoreResponse struct {
	ID               uuid.UUID       `json:"id"`
	ProjectID        uuid.UUID       `json:"project_id"`
	TraceID          *string         `json:"trace_id,omitempty"`
	SpanID           *string         `json:"span_id,omitempty"`
	Name             string          `json:"name"`
	Value            *float64        `json:"value,omitempty"`
	StringValue      *string         `json:"string_value,omitempty"`
	Type             string          `json:"type"`
	Source           string          `json:"source"`
	Reason           *string         `json:"reason,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	ExperimentID     *uuid.UUID      `json:"experiment_id,omitempty"`
	ExperimentItemID *string         `json:"experiment_item_id,omitempty"`
	Timestamp        time.Time       `json:"timestamp"`
}

type BatchScoreResponse struct {
	Created int `json:"created"`
}

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

type SDKCreateScoreInput struct {
	Body CreateScoreRequest
}
type SDKCreateScoreOutput struct {
	Body *ScoreResponse
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

	return &SDKCreateScoreOutput{Body: toScoreResponse(score)}, nil
}

type SDKCreateScoreBatchInput struct {
	Body BatchScoreRequest
}
type SDKCreateScoreBatchOutput struct {
	Body *BatchScoreResponse
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

func toScoreResponse(s *observability.Score) *ScoreResponse {
	// Metadata on the entity is json.RawMessage; sanitize any malformed
	// legacy bytes by JSON-escaping them into a string. Valid JSON passes
	// through unchanged.
	metadata := s.Metadata
	if len(metadata) > 0 && !json.Valid(metadata) {
		metadata, _ = json.Marshal(string(metadata))
	}
	return &ScoreResponse{
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

