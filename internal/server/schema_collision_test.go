// Regression guard: asserts that every Huma operation across every
// converted handler package can coexist on a single huma.API without
// colliding on OpenAPI component names (registry.go:130 panic) or
// operation IDs (openapi.go:1520 panic).
//
// These panics used to surface only at `bin/brokle-server` boot —
// `make build` and `make test` both passed because the build is
// static, the panic is a runtime check, and the humatest harness per
// handler package never loads more than one domain at a time.
//
// The test wires every handler's RegisterRoutes against a single
// shared humatest.TestAPI. Services are passed as embedded-interface
// stub structs (e.g. `type dummyUserService struct{ user.UserService }`)
// rather than bare `nil` interfaces. The embedded-interface pattern:
//
//   - RegisterRoutes succeeds because the stub satisfies the interface
//     at the type level.
//   - Any method call on the stub panics with a nil-deref because the
//     embedded interface value is nil.
//
// So a future handler that dereferences a service inside RegisterRoutes
// (e.g. eager config validation, schema introspection) will make the
// test fail loudly with a clear panic stack — instead of silently
// passing because RegisterRoutes was given a raw nil interface.
package server_test

import (
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"

	annotationDomain "brokle/internal/core/domain/annotation"
	billingDomain "brokle/internal/core/domain/billing"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	orgDomain "brokle/internal/core/domain/organization"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/testing/humax"
	annotationHandler "brokle/internal/transport/http/handlers/annotation"
	billingHandler "brokle/internal/transport/http/handlers/billing"
	evaluationHandler "brokle/internal/transport/http/handlers/evaluation"
	observabilityHandler "brokle/internal/transport/http/handlers/observability"
	organizationHandler "brokle/internal/transport/http/handlers/organization"
)

// ---- interface stubs -------------------------------------------------
//
// Each stub embeds the corresponding service interface so the outer
// struct satisfies it at compile time. The embedded value is the
// zero-value nil interface; calling any method panics with the
// familiar "invalid memory address or nil pointer dereference" — the
// signal we want if a handler dereferences a service at registration
// time.

type dummyOrg struct{ orgDomain.OrganizationService }
type dummyMember struct{ orgDomain.MemberService }
type dummyAnnotationQueue struct{ annotationDomain.QueueService }
type dummyAnnotationItem struct{ annotationDomain.ItemService }
type dummyAnnotationAssignment struct{ annotationDomain.AssignmentService }
type dummyBillingUsage struct{ billingDomain.BillableUsageService }
type dummyBillingBudget struct{ billingDomain.BudgetService }
type dummyBillingContract struct{ billingDomain.ContractService }
type dummyBillingPricing struct{ billingDomain.PricingService }
type dummyInvitation struct{ orgDomain.InvitationService }
type dummyOrgSettings struct{ orgDomain.OrganizationSettingsService }
type dummyEvalScoreConfig struct{ evaluationDomain.ScoreConfigService }
type dummyEvalDataset struct{ evaluationDomain.DatasetService }
type dummyEvalDatasetItem struct{ evaluationDomain.DatasetItemService }
type dummyEvalDatasetVersion struct{ evaluationDomain.DatasetVersionService }
type dummyEvalExperiment struct{ evaluationDomain.ExperimentService }
type dummyEvalExperimentItem struct{ evaluationDomain.ExperimentItemService }
type dummyEvalExperimentWizard struct{ evaluationDomain.ExperimentWizardService }
type dummyEvalEvaluator struct{ evaluationDomain.EvaluatorService }
type dummyEvalEvaluatorExec struct{ evaluationDomain.EvaluatorExecutionService }

// TestHandlers_NoSchemaOrOperationIDCollisions is the boot-time
// regression guard. It must pass on every commit; a new duplicate
// name / op-ID surfaces either as a Huma panic during
// RegisterRoutes (fails t.Fatal) or as a duplicate in the
// operation-ID uniqueness walk below.
func TestHandlers_NoSchemaOrOperationIDCollisions(t *testing.T) {
	// Wiring all handlers into ONE API deliberately differs from the
	// production split (apiPublic + apiAdmin). The flat-namespace
	// stress test catches the superset of real-world collisions
	// without having to reproduce the full server.New Deps graph.
	api := humax.NewAPI(t)
	logger := slog.Default()

	// websiteHandler is chi-native (off Huma).
	// overviewHandler is chi-native (off Huma).
	// apikeyHandler, userHandler are chi-native (off Huma).
	// credentialsHandler is chi-native (off Huma); it no longer
	// shares the Huma schema registry and is excluded here. Chi has
	// no schema namespace, so collisions are structurally impossible.
	annotationHandler.RegisterRoutes(api, dummyAnnotationQueue{}, dummyAnnotationItem{}, dummyAnnotationAssignment{}, logger)
	billingHandler.RegisterRoutes(api, dummyBillingUsage{}, dummyBillingBudget{}, dummyBillingContract{}, dummyBillingPricing{}, logger)
	organizationHandler.RegisterRoutes(api, dummyOrg{}, dummyMember{}, dummyInvitation{}, dummyOrgSettings{}, logger)

	// Observability takes concrete *TraceService / *ScoreService /
	// *ScoreAnalyticsService / *FilterPresetService pointers (not
	// interfaces). Typed-nil pointers satisfy the signature and
	// panic on method call exactly like embedded-interface stubs —
	// same invariant, different Go syntax.
	observabilityHandler.RegisterRoutes(api,
		(*obsServices.TraceService)(nil),
		(*obsServices.ScoreService)(nil),
		(*obsServices.ScoreAnalyticsService)(nil),
		(*obsServices.FilterPresetService)(nil),
		logger)

	evaluationHandler.RegisterRoutes(api,
		dummyEvalScoreConfig{},
		dummyEvalDataset{},
		dummyEvalDatasetItem{},
		dummyEvalDatasetVersion{},
		dummyEvalExperiment{},
		dummyEvalExperimentItem{},
		dummyEvalExperimentWizard{},
		dummyEvalEvaluator{},
		dummyEvalEvaluatorExec{},
		logger)

	assertUniqueOperationIDs(t, api)
}

func assertUniqueOperationIDs(t *testing.T, api huma.API) {
	t.Helper()
	b, err := json.Marshal(api.OpenAPI())
	require.NoError(t, err)

	var doc struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(b, &doc))

	seen := map[string]string{}
	for path, methods := range doc.Paths {
		for method, op := range methods {
			if op.OperationID == "" {
				continue
			}
			if prev, dup := seen[op.OperationID]; dup {
				t.Errorf("duplicate operation ID %q — used by %s and %s %s",
					op.OperationID, prev, method, path)
			}
			seen[op.OperationID] = method + " " + path
		}
	}
	require.NotEmpty(t, seen, "expected at least one operation registered")
}
