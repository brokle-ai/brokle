// Package evaluation exposes evaluation-domain operations on both surfaces:
//
//   - Dashboard plane (RequireAuth) — score-configs, evaluators,
//     executions, experiment wizard, plus project-scoped dataset /
//     experiment CRUD mounted under /api/v1/projects/{projectId}/...
//   - SDK plane (RequireSDKAuth) — dataset + experiment + score
//     ingestion under /v1/... with the project derived from the API key.
//
// The two registration entrypoints (RegisterRoutes / RegisterSDKRoutes)
// share handler-method implementations; the only difference is how the
// project ID is sourced (path parameter vs. SDK auth context).
package evaluation

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	evaluationService "brokle/internal/core/services/evaluation"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
)

// ---- handler aggregate ----------------------------------------------

type handler struct {
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

// RegisterRoutes mounts the dashboard-plane evaluation routes on r.
// Expected mount context: the authed dashboard chi group.
func RegisterRoutes(
	r chi.Router,
	scoreConfigSvc *evaluationService.ScoreConfigService,
	datasetSvc *evaluationService.DatasetService,
	datasetItemSvc *evaluationService.DatasetItemService,
	datasetVersionSvc *evaluationService.DatasetVersionService,
	experimentSvc *evaluationService.ExperimentService,
	experimentItemSvc *evaluationService.ExperimentItemService,
	experimentWizardSvc *evaluationService.ExperimentWizardService,
	evaluatorSvc *evaluationService.EvaluatorService,
	evaluatorExecSvc *evaluationService.EvaluatorExecutionService,
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

	// All evaluation dashboard routes sit under a single
	// r.Route("/api/v1/projects/{projectId}", ...) Mount. The wizard +
	// execution sub-registrations reference each other's prefixes
	// (/experiments/wizard, /datasets/{id}/fields,
	// /evaluators/{id}/executions), so collapsing them into one Mount
	// subrouter keeps the trie clean: each sub-file Mounts a distinct
	// deeper prefix (/score-configs, /datasets, /experiments,
	// /evaluators) and the cross-referencing wizard/execution nodes
	// live as siblings inside that same subrouter.
	//
	// Observability's project-scoped routes (/scores, /sessions,
	// /filter-presets) Mount as siblings on the parent tree with
	// distinct deeper prefixes. chi's radix trie routes static-deeper
	// before wildcard-shallower, so both coexist without collision.
	r.Route("/api/v1/projects/{projectId}", func(r chi.Router) {
		registerScoreConfigRoutes(r, h)
		registerDashboardDatasetRoutes(r, h)
		registerDashboardExperimentRoutes(r, h)
		registerEvaluatorRoutes(r, h)
		registerExecutionRoutes(r, h)
		registerWizardRoutes(r, h)
	})
}

// RegisterSDKRoutes mounts the SDK-plane evaluation routes on r. The
// project is derived from the API key (MustGetProjectID).
func RegisterSDKRoutes(
	r chi.Router,
	scoreConfigSvc *evaluationService.ScoreConfigService,
	datasetSvc *evaluationService.DatasetService,
	datasetItemSvc *evaluationService.DatasetItemService,
	datasetVersionSvc *evaluationService.DatasetVersionService,
	experimentSvc *evaluationService.ExperimentService,
	experimentItemSvc *evaluationService.ExperimentItemService,
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

	registerSDKDatasetRoutes(r, h)
	registerSDKExperimentRoutes(r, h)
	registerSDKScoreRoutes(r, h)
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
