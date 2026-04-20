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
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

// ---- handler aggregate ----------------------------------------------

type handler struct {
	scoreConfigSvc      evaluationDomain.ScoreConfigService
	datasetSvc          evaluationDomain.DatasetService
	datasetItemSvc      evaluationDomain.DatasetItemService
	datasetVersionSvc   evaluationDomain.DatasetVersionService
	experimentSvc       evaluationDomain.ExperimentService
	experimentItemSvc   evaluationDomain.ExperimentItemService
	experimentWizardSvc evaluationDomain.ExperimentWizardService
	evaluatorSvc        evaluationDomain.EvaluatorService
	evaluatorExecSvc    evaluationDomain.EvaluatorExecutionService
	scoreSvc            *obsServices.ScoreService
	logger              *slog.Logger
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

// Shared cross-feature DTOs (pageList, ScoreType, DatasetItemResponse,
// ExperimentItemResponse, BulkImportResponse, KeysMappingRequest,
// CountResponse, EmptyOutput) live in types.go.
// Feature ops live in their own files: dataset.go, experiment.go,
// evaluator.go, execution.go, wizard.go, score_config.go, score.go.
