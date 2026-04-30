package server

import (
	annotation "brokle/internal/transport/http/handlers/annotation"
	apikey "brokle/internal/transport/http/handlers/apikey"
	auth "brokle/internal/transport/http/handlers/auth"
	billing "brokle/internal/transport/http/handlers/billing"
	comment "brokle/internal/transport/http/handlers/comment"
	credentials "brokle/internal/transport/http/handlers/credentials"
	dashboard "brokle/internal/transport/http/handlers/dashboard"
	evaluation "brokle/internal/transport/http/handlers/evaluation"
	observability "brokle/internal/transport/http/handlers/observability"
	organization "brokle/internal/transport/http/handlers/organization"
	overview "brokle/internal/transport/http/handlers/overview"
	playground "brokle/internal/transport/http/handlers/playground"
	project "brokle/internal/transport/http/handlers/project"
	projectmember "brokle/internal/transport/http/handlers/projectmember"
	prompt "brokle/internal/transport/http/handlers/prompt"
	rbac "brokle/internal/transport/http/handlers/rbac"
	user "brokle/internal/transport/http/handlers/user"
	website "brokle/internal/transport/http/handlers/website"
)

// Handlers bundles every per-package handler instance. Constructed
// once at server startup (NewHandlers below) from the Deps service
// graph, then referenced by addRoutes for route declaration.
//
// Mat Ryer's "ask for what you need" rule applies here: each handler
// package's New() takes only the services it actually consumes. The
// bundle struct exists so addRoutes can reference handler methods
// uniformly (h.Auth.Login, h.Observability.Dashboard.ListTraces,
// etc.) without threading individual service references through every
// route line.
//
// Mirrors SigNoz's `aH.Signoz.Handlers.X` pattern and Brokle's own
// pre-migration `s.handlers.X` shape.
type Handlers struct {
	Auth         *auth.Handler
	AuthSDK      *auth.SDKHandler
	User         *user.Handler
	Organization *organization.Handler
	Project      *project.Handler
	APIKey        *apikey.Handler
	ProjectMember *projectmember.Handler
	Overview      *overview.Handler
	Dashboard    *dashboard.Handler
	Annotation   *annotation.Handler
	Billing      *billing.Handler
	Credentials  *credentials.Handler
	RBAC         *rbac.Handler
	Prompt       *prompt.Handler
	Playground   *playground.Handler
	Evaluation   *evaluation.Handler
	Comment      *comment.Handler
	Website      *website.Handler

	// Observability has three plane-specific handler structs (dashboard,
	// SDK span query, OTLP raw protobuf ingestion). Bundle them under
	// one field so addRoutes references them as h.Observability.Dashboard
	// / h.Observability.SDK / h.Observability.OTLP.
	Observability ObservabilityHandlers
}

// ObservabilityHandlers groups the three observability-plane handlers
// since the package exports three distinct structs (one per plane).
type ObservabilityHandlers struct {
	Dashboard *observability.DashboardHandler
	SDK       *observability.SDKHandler
	OTLP      *observability.OTLPHandler
}

// NewHandlers constructs the full Handlers bundle from the Deps
// service graph. Called once during server startup. Each handler
// package's New() receives only the services it consumes — see each
// handler's New() signature for the canonical dependency set.
func NewHandlers(d Deps) Handlers {
	return Handlers{
		Auth: auth.New(
			d.Auth, d.User, d.Profile, d.Registration, d.Session,
			d.OAuthProvider, d.APIKey, d.Config, d.Logger,
		),
		AuthSDK:      auth.NewSDK(d.APIKey, d.Logger),
		User:         user.New(d.User, d.Profile, d.Organization, d.ProjectMember, d.Logger),
		Organization: organization.New(d.Organization, d.OrgMemberOrg, d.Invitation, d.OrgSettings, d.Logger),
		Project:      project.New(d.Project, d.Logger),
		APIKey:        apikey.New(d.APIKey, d.Logger),
		ProjectMember: projectmember.New(d.ProjectMember, d.Logger),
		Overview:     overview.New(d.Overview, d.Logger),
		Dashboard:    dashboard.New(d.Dashboard, d.DashboardQuery, d.DashboardTemplate, d.Logger),
		Annotation:   annotation.New(d.AnnotationQueue, d.AnnotationItem, d.AnnotationAssignment, d.Logger),
		Billing:      billing.New(d.BillingUsage, d.BillingBudget, d.BillingContract, d.BillingPricing, d.Logger),
		Credentials:  credentials.New(d.Credential, d.CredentialModelCatalog, d.Logger),
		RBAC:         rbac.New(d.Role, d.Permission, d.OrgMember, d.Logger),
		Prompt:       prompt.New(d.Prompt, d.PromptCompiler, d.Logger),
		Playground:   playground.New(d.Playground, d.Project, d.Logger),
		Evaluation: evaluation.New(
			d.EvalScoreConfig, d.EvalDataset, d.EvalDatasetItem, d.EvalDatasetVersion,
			d.EvalExperiment, d.EvalExperimentItem, d.EvalExperimentWizard,
			d.EvalEvaluator, d.EvalEvaluatorExecution,
			d.Observability.ScoreService,
			d.Logger,
		),
		Comment: comment.New(d.Comment, d.Logger),
		Website: website.New(d.Website, d.Logger),
		Observability: ObservabilityHandlers{
			Dashboard: observability.NewDashboard(
				d.Observability.TraceService,
				d.Observability.ScoreService,
				d.Observability.ScoreAnalyticsService,
				d.Observability.FilterPresetService,
				d.Logger,
			),
			SDK: observability.NewSDK(d.Observability.SpanQueryService, d.Logger),
			OTLP: observability.NewOTLP(observability.OTLPDeps{
				StreamProducer:       d.Observability.StreamProducer,
				DeduplicationService: d.Observability.DeduplicationService,
				OTLPConverter:        d.Observability.OTLPConverterService,
				LogsConverter:        d.Observability.OTLPLogsConverterService,
				EventsConverter:      d.Observability.OTLPEventsConverterService,
				MetricsConverter:     d.Observability.OTLPMetricsConverterService,
				Logger:               d.Logger,
			}),
		},
	}
}
