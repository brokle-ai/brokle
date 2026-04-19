package server

import (
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	httprateredis "github.com/go-chi/httprate-redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"brokle/internal/config"
	analyticsDomain "brokle/internal/core/domain/analytics"
	authDomain "brokle/internal/core/domain/auth"
	annotationDomain "brokle/internal/core/domain/annotation"
	billingDomain "brokle/internal/core/domain/billing"
	commentDomain "brokle/internal/core/domain/comment"
	credentialsDomain "brokle/internal/core/domain/credentials"
	dashboardDomain "brokle/internal/core/domain/dashboard"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	orgDomain "brokle/internal/core/domain/organization"
	playgroundDomain "brokle/internal/core/domain/playground"
	promptDomain "brokle/internal/core/domain/prompt"
	userDomain "brokle/internal/core/domain/user"
	websiteDomain "brokle/internal/core/domain/website"
	authService "brokle/internal/core/services/auth"
	credentialsService "brokle/internal/core/services/credentials"
	observabilityService "brokle/internal/core/services/observability"
	"brokle/internal/core/services/registration"
	"brokle/internal/transport/http/middleware"
)

// Deps groups every dependency the HTTP server needs to wire its
// route table. Populated by internal/app/providers.go and passed
// once to server.New.
//
// The set grows incrementally as handler domains are converted to
// Huma operations (Step 4 of the chi+Huma migration). Each new
// domain adds its service field here; per-domain RegisterRoutes
// functions (in internal/transport/http/handlers/<domain>/routes.go)
// take the explicit services they need rather than reaching into the
// whole Deps struct — Mat Ryer's "ask for what you need" rule
// applied at the registration boundary.
type Deps struct {
	Config *config.Config
	Logger *slog.Logger

	// Infrastructure handles for health pings + rate-limit backend.
	DB         *pgxpool.Pool
	Redis      *redis.Client
	ClickHouse driver.Conn

	// Auth/identity services consumed by RequireAuth, RequireSDKAuth,
	// RequireProjectAccess, and RequirePermission middleware.
	JWT       authDomain.JWTService
	Blacklist authDomain.BlacklistedTokenService
	OrgMember authDomain.OrganizationMemberService
	APIKey    authDomain.APIKeyService

	// Project service used by RequireProjectAccess and the project
	// handler domain.
	Project orgDomain.ProjectService

	// OrgMemberOrg is the organization-domain member service (distinct
	// from OrgMember above, which is auth.OrganizationMemberService).
	// Consumed by the project handler for per-org membership checks on
	// list/create.
	OrgMemberOrg orgDomain.MemberService

	// Domain services. Add as handler domains migrate to Huma. The
	// list grows with each vertical slice; domains not yet converted
	// to Huma operations are NOT listed here (CLAUDE.md scaffolded-
	// but-unreachable rule).

	// Auth domain handlers — login, signup, logout, refresh, password
	// mgmt, /me, profile, sessions, OAuth. Distinct from the
	// middleware-facing JWT/Blacklist/OrgMember services above:
	// those carry invariant checks for every protected route;
	// Auth/User/Profile/Registration/Session/OAuth are the
	// "business logic" services the auth handler operations invoke.
	Auth          authDomain.AuthService
	User          userDomain.UserService
	Profile       userDomain.ProfileService
	Registration  registration.RegistrationService
	Session       authDomain.SessionService
	OAuthProvider *authService.OAuthProviderService

	// Organization service is used by the user handler's
	// get-user-profile op to render the org hierarchy in the
	// dashboard sidebar.
	Organization orgDomain.OrganizationService

	// Comment service powers the trace-attached discussion threads.
	Comment commentDomain.Service

	// Overview service powers the project-overview dashboard page.
	Overview analyticsDomain.OverviewService

	// Credentials: AI provider credential CRUD + connection-test +
	// model-catalog discovery (available models derived from
	// configured providers).
	Credential           credentialsDomain.ProviderCredentialService
	CredentialModelCatalog credentialsService.ModelCatalogService

	// Website contact-form handler.
	Website websiteDomain.WebsiteService

	// Dashboard domain: dashboards, widget query execution, and
	// pre-defined templates. Three-service trio to match the
	// pre-migration gin layout.
	Dashboard         dashboardDomain.DashboardService
	DashboardQuery    dashboardDomain.WidgetQueryService
	DashboardTemplate dashboardDomain.TemplateService

	// Annotation domain: HITL review queues, items, assignments.
	AnnotationQueue      annotationDomain.QueueService
	AnnotationItem       annotationDomain.ItemService
	AnnotationAssignment annotationDomain.AssignmentService

	// Billing domain: usage tracking, budgets, contracts, pricing.
	BillingUsage    billingDomain.BillableUsageService
	BillingBudget   billingDomain.BudgetService
	BillingContract billingDomain.ContractService
	BillingPricing  billingDomain.PricingService

	// RBAC / auth-extended domain services consumed by the rbac handler
	// (read-only role/permission discovery + custom-role lifecycle +
	// scope checks). Distinct from the middleware-facing invariants
	// (JWT / Blacklist / OrgMember) already above.
	Role       authDomain.RoleService
	Permission authDomain.PermissionService
	Scope      authDomain.ScopeService

	// Organization: invitations + settings. The OrganizationService +
	// MemberService (OrgMemberOrg) are already declared above.
	Invitation       orgDomain.InvitationService
	OrgSettings      orgDomain.OrganizationSettingsService

	// Prompt domain: prompt + version CRUD + compile/preview.
	Prompt         promptDomain.PromptService
	PromptCompiler promptDomain.CompilerService

	// Playground domain: execute + session CRUD + streaming.
	Playground playgroundDomain.PlaygroundService

	// Evaluation domain: datasets, experiments, evaluators, score
	// configs, executions. Largest service surface in the codebase.
	EvalScoreConfig        evaluationDomain.ScoreConfigService
	EvalDataset            evaluationDomain.DatasetService
	EvalDatasetItem        evaluationDomain.DatasetItemService
	EvalDatasetVersion     evaluationDomain.DatasetVersionService
	EvalExperiment         evaluationDomain.ExperimentService
	EvalExperimentItem     evaluationDomain.ExperimentItemService
	EvalExperimentWizard   evaluationDomain.ExperimentWizardService
	EvalEvaluator          evaluationDomain.EvaluatorService
	EvalEvaluatorExecution evaluationDomain.EvaluatorExecutionService

	// Observability: trace/span/score/filter-preset + OTLP ingest.
	// We carry the observability ServiceRegistry pointer directly —
	// both handler-plane services and the OTLP plain-chi mount draw
	// from the same registry.
	Observability *observabilityService.ServiceRegistry
}

// authMiddlewareDeps assembles the middleware.AuthDeps struct from
// the matching Deps fields. Centralised so route registration in
// addRoutes doesn't repeat the field mapping.
func (d Deps) authMiddlewareDeps() middleware.AuthDeps {
	return middleware.AuthDeps{
		JWT:       d.JWT,
		Blacklist: d.Blacklist,
		OrgMember: d.OrgMember,
		Project:   d.Project,
		Logger:    d.Logger,
	}
}

// sdkAuthMiddlewareDeps assembles the middleware.SDKAuthDeps struct
// for SDK route protection.
func (d Deps) sdkAuthMiddlewareDeps() middleware.SDKAuthDeps {
	return middleware.SDKAuthDeps{
		APIKey: d.APIKey,
		Logger: d.Logger,
	}
}

// rateLimitMiddlewareDeps builds the RateLimitDeps with a Redis
// backend pointed at the same instance the rest of the application
// uses. httprateredis.Config accepts a Client field of type
// redis.UniversalClient (which *redis.Client satisfies) — passing
// our shared client avoids a duplicate connection pool and keeps
// rate-limit traffic correlated with the rest of our Redis usage in
// dashboards.
func (d Deps) rateLimitMiddlewareDeps() middleware.RateLimitDeps {
	return middleware.RateLimitDeps{
		Redis: &httprateredis.Config{
			Client:    d.Redis,
			PrefixKey: "ratelimit:",
		},
		Auth:   &d.Config.Auth,
		Logger: d.Logger,
	}
}

// healthDeps assembles the readiness-check dependencies. Nil
// pointers skip the corresponding check (used by tests).
func (d Deps) healthDeps() healthDeps {
	return healthDeps{
		DB:         d.DB,
		Redis:      d.Redis,
		ClickHouse: d.ClickHouse,
		Logger:     d.Logger,
	}
}
