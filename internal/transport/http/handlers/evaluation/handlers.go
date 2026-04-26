// Package evaluation exposes evaluation-domain operations on both surfaces:
//
//   - Dashboard plane (RequireAuth) — score-configs, evaluators,
//     executions, experiment wizard, plus project-scoped dataset /
//     experiment CRUD mounted under /api/v1/projects/{projectId}/...
//   - SDK plane (RequireSDKAuth) — dataset + experiment + score
//     ingestion under /v1/... with the project derived from the API key.
//
// Both surfaces share handler-method implementations; the only difference
// is how the project ID is sourced (path parameter vs. SDK auth context).
// Routes are wired centrally in internal/server/routes.go.
package evaluation

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	evaluationService "brokle/internal/core/services/evaluation"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
)

// ---- handler aggregate ----------------------------------------------

type Handler struct {
	scoreConfigSvc      *evaluationService.ScoreConfigService
	datasetSvc          *evaluationService.DatasetService
	datasetItemSvc      *evaluationService.DatasetItemService
	datasetVersionSvc   *evaluationService.DatasetVersionService
	experimentSvc       *evaluationService.ExperimentService
	experimentItemSvc   *evaluationService.ExperimentItemService
	experimentWizardSvc *evaluationService.ExperimentWizardService
	evaluatorSvc        *evaluationService.EvaluatorService
	evaluatorExecSvc    *evaluationService.EvaluatorExecutionService
	scoreSvc            *obsServices.ScoreService
	logger              *slog.Logger
}

// New constructs a Handler with all services any evaluation Register*
// function might need (dashboard + SDK).
func New(
	scoreConfigSvc *evaluationService.ScoreConfigService,
	datasetSvc *evaluationService.DatasetService,
	datasetItemSvc *evaluationService.DatasetItemService,
	datasetVersionSvc *evaluationService.DatasetVersionService,
	experimentSvc *evaluationService.ExperimentService,
	experimentItemSvc *evaluationService.ExperimentItemService,
	experimentWizardSvc *evaluationService.ExperimentWizardService,
	evaluatorSvc *evaluationService.EvaluatorService,
	evaluatorExecSvc *evaluationService.EvaluatorExecutionService,
	scoreSvc *obsServices.ScoreService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		scoreConfigSvc:      scoreConfigSvc,
		datasetSvc:          datasetSvc,
		datasetItemSvc:      datasetItemSvc,
		datasetVersionSvc:   datasetVersionSvc,
		experimentSvc:       experimentSvc,
		experimentItemSvc:   experimentItemSvc,
		experimentWizardSvc: experimentWizardSvc,
		evaluatorSvc:        evaluatorSvc,
		evaluatorExecSvc:    evaluatorExecSvc,
		scoreSvc:            scoreSvc,
		logger:              logger,
	}
}

// ---- shared helpers -------------------------------------------------

func userIDPtr(ctx context.Context) *uuid.UUID {
	id, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &id
}

// projectIDForSDK returns the project UUID stored in the SDK auth
// context. The RequireSDKAuth middleware guarantees it is set.
func projectIDForSDK(r *http.Request) uuid.UUID {
	return httpctx.MustGetProjectID(r.Context())
}

// readPagination reads ?page=&limit= with evaluation-specific defaults
// (PageSize ≈ pagination.DefaultPageSize). Kept here because
// pkg/request.QueryPagination uses limit 1..1000 which is too
// permissive for the evaluation endpoints.
func readPagination(r *http.Request) (page, limit int, err error) {
	page, err = request.QueryInt(r, "page", 1)
	if err != nil {
		return 0, 0, err
	}
	if page < 1 {
		page = 1
	}
	limit, err = request.QueryInt(r, "limit", 0)
	if err != nil {
		return 0, 0, err
	}
	if limit <= 0 || !pagination.IsValidPageSize(limit) {
		limit = pagination.DefaultPageSize
	}
	return page, limit, nil
}
