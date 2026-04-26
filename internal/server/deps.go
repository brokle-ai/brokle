package server

import (
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	httprateredis "github.com/go-chi/httprate-redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"brokle/internal/config"
	analyticsService "brokle/internal/core/services/analytics"
	annotationService "brokle/internal/core/services/annotation"
	authService "brokle/internal/core/services/auth"
	billingService "brokle/internal/core/services/billing"
	commentService "brokle/internal/core/services/comment"
	credentialsService "brokle/internal/core/services/credentials"
	dashboardService "brokle/internal/core/services/dashboard"
	evaluationService "brokle/internal/core/services/evaluation"
	observabilityService "brokle/internal/core/services/observability"
	organizationService "brokle/internal/core/services/organization"
	playgroundService "brokle/internal/core/services/playground"
	promptService "brokle/internal/core/services/prompt"
	"brokle/internal/core/services/registration"
	userService "brokle/internal/core/services/user"
	websiteService "brokle/internal/core/services/website"
	"brokle/internal/transport/http/middleware"
)

// Deps groups every dependency the HTTP server needs to wire its
// route table. Populated by internal/app/providers.go and passed
// once to server.New.
//
// Per-domain RegisterRoutes functions (in
// internal/transport/http/handlers/<domain>/) take the explicit
// services they need rather than reaching into the whole Deps struct —
// Mat Ryer's "ask for what you need" rule applied at the registration
// boundary. Adding a new handler domain means adding the service
// field here and threading it into the domain's RegisterRoutes call
// in routes.go.
type Deps struct {
	Config *config.Config
	Logger *slog.Logger

	// Infrastructure handles for health pings + rate-limit backend.
	DB         *pgxpool.Pool
	Redis      *redis.Client
	ClickHouse driver.Conn

	// Auth/identity services consumed by RequireAuth, RequireSDKAuth,
	// RequireProjectAccess, and RequirePermission middleware.
	JWT           *authService.JWTService
	Blacklist     *authService.BlacklistedTokenService
	OrgMember     *authService.OrganizationMemberService
	ProjectMember *authService.ProjectMemberService
	APIKey        *authService.APIKeyService

	// Project service used by RequireProjectAccess and the project
	// handler domain.
	Project *organizationService.ProjectService

	// OrgMemberOrg is the organization-domain member service (distinct
	// from OrgMember above, which is auth.OrganizationMemberService).
	// Consumed by the project handler for per-org membership checks on
	// list/create.
	OrgMemberOrg *organizationService.MemberService

	// Domain services consumed by handler RegisterRoutes calls in
	// internal/server/routes.go. Per CLAUDE.md scaffolded-but-
	// unreachable rule, no service is listed here without an active
	// caller.

	// Auth domain handlers — login, signup, logout, refresh, password
	// mgmt, /me, profile, sessions, OAuth. Distinct from the
	// middleware-facing JWT/Blacklist/OrgMember services above:
	// those carry invariant checks for every protected route; the
	// services below are the business-logic services the auth handler
	// methods invoke.
	Auth          *authService.AuthService
	User          *userService.UserService
	Profile       *userService.ProfileService
	Registration  *registration.RegistrationService
	Session       *authService.SessionService
	OAuthProvider *authService.OAuthProviderService

	// Organization service is used by the user handler's
	// get-user-profile op to render the org hierarchy in the
	// dashboard sidebar.
	Organization *organizationService.OrganizationService

	// Comment service powers the trace-attached discussion threads.
	Comment *commentService.CommentService

	// Overview service powers the project-overview dashboard page.
	Overview *analyticsService.OverviewService

	// Credentials: AI provider credential CRUD + connection-test +
	// model-catalog discovery (available models derived from
	// configured providers).
	Credential             *credentialsService.ProviderCredentialService
	CredentialModelCatalog *credentialsService.ModelCatalogService

	// Website contact-form handler.
	Website *websiteService.WebsiteService

	// Dashboard domain: dashboards, widget query execution, and
	// pre-defined templates. Three-service trio to match the
	// pre-migration gin layout.
	Dashboard         *dashboardService.DashboardService
	DashboardQuery    *dashboardService.WidgetQueryService
	DashboardTemplate *dashboardService.TemplateService

	// Annotation domain: HITL review queues, items, assignments.
	AnnotationQueue      *annotationService.QueueService
	AnnotationItem       *annotationService.ItemService
	AnnotationAssignment *annotationService.AssignmentService

	// Billing domain: usage tracking, budgets, contracts, pricing.
	BillingUsage    *billingService.BillableUsageService
	BillingBudget   *billingService.BudgetService
	BillingContract *billingService.ContractService
	BillingPricing  *billingService.PricingService

	// RBAC / auth-extended domain services consumed by the rbac handler
	// (read-only role/permission discovery + custom-role lifecycle +
	// scope checks). Distinct from the middleware-facing invariants
	// (JWT / Blacklist / OrgMember) already above.
	Role       *authService.RoleService
	Permission *authService.PermissionService
	Scope      *authService.ScopeService

	// Organization: invitations + settings. The OrganizationService +
	// MemberService (OrgMemberOrg) are already declared above.
	Invitation  *organizationService.InvitationService
	OrgSettings *organizationService.OrganizationSettingsService

	// Prompt domain: prompt + version CRUD + compile/preview.
	Prompt         *promptService.PromptService
	PromptCompiler *promptService.CompilerService

	// Playground domain: execute + session CRUD + streaming.
	Playground *playgroundService.PlaygroundService

	// Evaluation domain: datasets, experiments, evaluators, score
	// configs, executions. Largest service surface in the codebase.
	EvalScoreConfig        *evaluationService.ScoreConfigService
	EvalDataset            *evaluationService.DatasetService
	EvalDatasetItem        *evaluationService.DatasetItemService
	EvalDatasetVersion     *evaluationService.DatasetVersionService
	EvalExperiment         *evaluationService.ExperimentService
	EvalExperimentItem     *evaluationService.ExperimentItemService
	EvalExperimentWizard   *evaluationService.ExperimentWizardService
	EvalEvaluator          *evaluationService.EvaluatorService
	EvalEvaluatorExecution *evaluationService.EvaluatorExecutionService

	// Observability: trace/span/score/filter-preset + OTLP ingest.
	// We carry the observability ServiceRegistry pointer directly —
	// both handler-plane services and the OTLP plain-chi mount draw
	// from the same registry.
	Observability *observabilityService.ServiceRegistry

	// Handlers bundles every per-package handler instance. Populated
	// inside server.New() via NewHandlers(deps) right before addRoutes
	// runs — addRoutes references handler methods uniformly through
	// d.Handlers.X.Method.
	Handlers Handlers
}

// authMiddlewareDeps assembles the middleware.AuthDeps struct from
// the matching Deps fields. Centralised so route registration in
// addRoutes doesn't repeat the field mapping.
func (d Deps) authMiddlewareDeps() middleware.AuthDeps {
	return middleware.AuthDeps{
		JWT:           d.JWT,
		Blacklist:     d.Blacklist,
		OrgMember:     d.OrgMember,
		ProjectMember: d.ProjectMember,
		Project:       d.Project,
		Logger:        d.Logger,
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
