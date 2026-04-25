package server

import (
	"github.com/go-chi/chi/v5"

	annotationHandler "brokle/internal/transport/http/handlers/annotation"
	apikeyHandler "brokle/internal/transport/http/handlers/apikey"
	authHandler "brokle/internal/transport/http/handlers/auth"
	billingHandler "brokle/internal/transport/http/handlers/billing"
	commentHandler "brokle/internal/transport/http/handlers/comment"
	credentialsHandler "brokle/internal/transport/http/handlers/credentials"
	dashboardHandler "brokle/internal/transport/http/handlers/dashboard"
	evaluationHandler "brokle/internal/transport/http/handlers/evaluation"
	observabilityHandler "brokle/internal/transport/http/handlers/observability"
	organizationHandler "brokle/internal/transport/http/handlers/organization"
	overviewHandler "brokle/internal/transport/http/handlers/overview"
	playgroundHandler "brokle/internal/transport/http/handlers/playground"
	projectHandler "brokle/internal/transport/http/handlers/project"
	promptHandler "brokle/internal/transport/http/handlers/prompt"
	rbacHandler "brokle/internal/transport/http/handlers/rbac"
	userHandler "brokle/internal/transport/http/handlers/user"
	websiteHandler "brokle/internal/transport/http/handlers/website"
	"brokle/internal/transport/http/middleware"
)

// addRoutes wires the full HTTP surface area onto a chi router. Run
// once at server start; never per-request. The function is the single
// authoritative dependency map for the service — when a route 404s,
// this is the file to grep.
//
// Architecture:
//
//   - Cross-cutting middleware (CORS, CSRF, request ID, logger,
//     recoverer, metrics) is installed at the chi mux level by
//     installGlobalMiddleware. CORS + CSRF are path-scoped to
//     /api/v1 so the SDK plane at /v1 doesn't pay the cookie-domain
//     tax.
//
//   - Auth + rate-limit are applied at the chi-group level. Sub-
//     routers (r.Group, r.Route) have their own middleware stacks;
//     they don't violate chi's mux-level Use invariant (all top-
//     level middleware lives in installGlobalMiddleware).
//
//   - Rate-limit scoping follows the GitHub/Stripe/OpenAI pattern
//     (CLAUDE.md gotcha #37a): IP buckets for pre-auth surfaces only,
//     principal buckets (user ID, API-key ID) for authed surfaces.
//     NEVER layer IP on top of a principal counter — shared-egress
//     clients (NAT, CGNAT, BFF/SSR pods, serverless) collapse all
//     users into one IP bucket and cross-throttle. Concretely:
//
//   - sdkPublic  — LimitByIP + LimitByKeyPrefix  (validate-key;
//     IP for flood defence, key-prefix for brute force)
//
//   - sdkAuth    — LimitByAPIKey only            (authed SDK)
//
//   - dashPublic — LimitByIP only                (login, signup, OAuth, etc.)
//
//   - dashAuth   — LimitByUser only              (authed dashboard)
//
// Plane layout:
//
//	/v1 (SDK plane, X-API-Key auth)
//	  ├── sdkPublic  chi.Group  → LimitByIP + LimitByKeyPrefix
//	  │     └── auth.RegisterSDKRoutes (validate-key)
//	  └── sdkAuth    chi.Group  → RequireSDKAuth + LimitByAPIKey
//	        ├── observability OTLP (raw protobuf)
//	        ├── annotation, prompt, playground SDK
//	        ├── observability SDK (span query)
//	        └── evaluation SDK
//
//	/api/v1 (dashboard plane, cookie+JWT auth)
//	  ├── dashPublic chi.Group  → LimitByIP
//	  │     ├── auth.RegisterPublicRoutes (login, signup, OAuth, refresh)
//	  │     └── website.RegisterRoutes (contact form)
//	  └── dashAuth   chi.Group  → RequireAuth + LimitByUser
//	        └── auth (protected), user, apikey, comment, overview,
//	            credentials, project, dashboard, annotation, billing,
//	            organization, prompt, rbac, playground,
//	            observability, evaluation
//
// Auth spans both posture groups because /api/v1/auth/* has public
// endpoints (login, signup) and protected endpoints (me, logout)
// under one prefix. The two Register functions register concrete
// paths (r.Post / r.Get), not an r.Route subtree, so they coexist on
// the shared routing tree without chi Mount collision. See CLAUDE.md
// Transport gotcha #35 and the 2026-04-24 Lessons Learned entry.
//
// NEVER call r.Use(...) on the top-level chi.Mux here — that belongs
// in installGlobalMiddleware; chi panics if mux-level middleware is
// registered after any route has been mounted.
func addRoutes(r chi.Router, d Deps) {
	rateLimitD := d.rateLimitMiddlewareDeps()
	sdkAuthD := d.sdkAuthMiddlewareDeps()

	// -------------------- /v1 SDK plane --------------------

	// validate-key — pre-auth, dual rate-limit defence.
	r.Group(func(r chi.Router) {
		r.Use(middleware.LimitByIP(rateLimitD))
		r.Use(middleware.LimitByKeyPrefix(rateLimitD))
		authHandler.RegisterSDKRoutes(r, d.APIKey, d.Logger)
	})

	// Authed SDK surface.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireSDKAuth(sdkAuthD))
		r.Use(middleware.LimitByAPIKey(rateLimitD))

		// OTLP protobuf ingestion.
		observabilityHandler.RegisterOTLPRoutes(r, observabilityHandler.OTLPDeps{
			StreamProducer:       d.Observability.StreamProducer,
			DeduplicationService: d.Observability.DeduplicationService,
			OTLPConverter:        d.Observability.OTLPConverterService,
			LogsConverter:        d.Observability.OTLPLogsConverterService,
			EventsConverter:      d.Observability.OTLPEventsConverterService,
			MetricsConverter:     d.Observability.OTLPMetricsConverterService,
			Logger:               d.Logger,
		})

		playgroundHandler.RegisterSDKRoutes(r, d.Playground, d.Logger)
		promptHandler.RegisterSDKRoutes(r, d.Prompt, d.Logger)
		annotationHandler.RegisterSDKRoutes(r, d.AnnotationItem, d.Logger)
		observabilityHandler.RegisterSDKRoutes(r, d.Observability.SpanQueryService, d.Logger)
		evaluationHandler.RegisterSDKRoutes(
			r,
			d.EvalScoreConfig, d.EvalDataset, d.EvalDatasetItem, d.EvalDatasetVersion,
			d.EvalExperiment, d.EvalExperimentItem,
			d.Observability.ScoreService,
			d.Logger,
		)
	})

	// -------------------- /api/v1 dashboard plane --------------------

	// Pre-auth dashboard routes — login, signup, password reset, OAuth,
	// token refresh, website contact form.
	r.Group(func(r chi.Router) {
		r.Use(middleware.LimitByIP(rateLimitD))
		authHandler.RegisterPublicRoutes(r,
			d.Auth, d.User, d.Registration, d.Session,
			d.OAuthProvider, d.Config, d.Logger)
		websiteHandler.RegisterRoutes(r, d.Website, d.Logger)
	})

	// Authed dashboard routes.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.authMiddlewareDeps()))
		r.Use(middleware.LimitByUser(rateLimitD))

		authHandler.RegisterProtectedRoutes(r,
			d.Auth, d.User, d.Profile, d.Registration, d.Session,
			d.OAuthProvider, d.Config, d.Logger)

		userHandler.RegisterRoutes(r, d.User, d.Profile, d.Organization, d.Logger)
		apikeyHandler.RegisterRoutes(r, d.APIKey, d.Logger)
		commentHandler.RegisterRoutes(r, d.Comment, d.Logger)
		overviewHandler.RegisterRoutes(r, d.Overview, d.Logger)
		credentialsHandler.RegisterRoutes(r, d.Credential, d.CredentialModelCatalog, d.Logger)
		projectHandler.RegisterRoutes(r, d.Project, d.Organization, d.OrgMemberOrg, d.Logger)
		dashboardHandler.RegisterRoutes(r, d.Dashboard, d.DashboardQuery, d.DashboardTemplate, d.Logger)
		annotationHandler.RegisterRoutes(r, d.AnnotationQueue, d.AnnotationItem, d.AnnotationAssignment, d.Logger)
		billingHandler.RegisterRoutes(r, d.BillingUsage, d.BillingBudget, d.BillingContract, d.BillingPricing, d.Logger)
		organizationHandler.RegisterRoutes(r, d.Organization, d.OrgMemberOrg, d.Invitation, d.OrgSettings, d.Logger)
		promptHandler.RegisterRoutes(r, d.Prompt, d.PromptCompiler, d.Logger)
		rbacHandler.RegisterRoutes(r, d.Role, d.Permission, d.OrgMember, d.Scope, d.Logger)
		playgroundHandler.RegisterRoutes(r, d.Playground, d.Project, d.Logger)
		observabilityHandler.RegisterRoutes(
			r,
			d.Observability.TraceService,
			d.Observability.ScoreService,
			d.Observability.ScoreAnalyticsService,
			d.Observability.FilterPresetService,
			d.Logger,
		)
		evaluationHandler.RegisterRoutes(
			r,
			d.EvalScoreConfig, d.EvalDataset, d.EvalDatasetItem, d.EvalDatasetVersion,
			d.EvalExperiment, d.EvalExperimentItem, d.EvalExperimentWizard,
			d.EvalEvaluator, d.EvalEvaluatorExecution,
			d.Logger,
		)
	})
}
