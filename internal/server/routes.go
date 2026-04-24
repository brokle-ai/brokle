package server

import (
	"github.com/danielgtaylor/huma/v2"
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
	"brokle/internal/transport/http/middleware/humawrap"
)

// addRoutes wires the full HTTP surface area onto a chi router. Run
// once at server start; never per-request. The function is a single
// authoritative dependency map for the service — when a route 404s,
// this is the file to grep.
//
// Architecture:
//
//   - Cross-cutting middleware (CORS, CSRF) is applied at the chi mux
//     level by installGlobalMiddleware, scoped to /api/v1 via
//     pathPrefix. It runs for every dashboard-plane request, including
//     OPTIONS preflight and OpenAPI docs paths, before chi even routes
//     the request.
//
//   - Auth + rate limiting are applied at the Huma layer via
//     huma.NewGroup(api).UseMiddleware, with the existing
//     func(http.Handler) http.Handler middleware adapted via
//     humawrap.Wrap. This is the load-bearing fix for the original
//     bug where chi r.Group middleware never reached Huma operations
//     (humachi binds its adapter to the captured router reference at
//     construction time; routes registered via huma.Register land on
//     the mux, not on any subrouter built later via r.Route / r.Group).
//
//   - Rate-limit scoping follows the industry split (GitHub, Stripe,
//     OpenAI, Anthropic, Cloudflare): IP buckets for pre-auth routes
//     only, principal buckets (user ID / API key ID) for authed
//     routes. NEVER layer IP on top of principal — shared-egress
//     clients (NAT, CGNAT, BFF/SSR pods, serverless) collapse all
//     users into one IP bucket and cross-throttle. Concretely:
//
//       * dashPublic → LimitByIP (login, signup, OAuth, refresh,
//         contact form — the pre-auth dashboard surface).
//       * dashAuth   → LimitByUser only (principal = user ID).
//       * sdkPublic  → LimitByIP + LimitByKeyPrefix (validate-key;
//         dual defense: IP for distributed floods, key-prefix for
//         credential brute force).
//       * sdkAuth    → LimitByAPIKey only (principal = API key ID).
//       * OTLP raw   → LimitByAPIKey only (same principal as sdkAuth).
//
//   - Raw chi routes (OTLP protobuf ingestion) keep their native chi
//     middleware chain because Huma cannot model protobuf streams.
//     They register on a chi r.Group with RequireSDKAuth + LimitByAPIKey
//     and are unaffected by the Huma binding issue.
//
// Plane layout:
//
//	/v1 (apiPublic, SDK plane, X-API-Key auth)
//	  ├── sdkPublic     huma.Group  → LimitByIP + LimitByKeyPrefix
//	  │     └── auth.RegisterSDKRoutes (validate-key)
//	  ├── sdkAuth       huma.Group  → RequireSDKAuth + LimitByAPIKey
//	  │     ├── annotation, prompt, playground SDK
//	  │     ├── observability SDK
//	  │     └── evaluation SDK
//	  └── chi r.Group → RequireSDKAuth + LimitByAPIKey
//	        └── observability OTLP (raw protobuf)
//
//	/api/v1 (apiAdmin, dashboard plane, cookie+JWT auth)
//	  ├── dashPublic    huma.Group  → LimitByIP
//	  │     ├── auth.RegisterPublicRoutes (login, signup, OAuth, refresh)
//	  │     └── website.RegisterRoutes (contact form)
//	  └── dashAuth      huma.Group  → RequireAuth + LimitByUser
//	        └── auth (protected), user, apikey, comment, overview,
//	            credentials, project, dashboard, annotation, billing,
//	            organization, prompt, rbac, playground,
//	            observability, evaluation
//
// NEVER call r.Use(...) on the top-level chi.Mux here — that belongs
// in installGlobalMiddleware. NEVER attach auth or rate-limit
// middleware via r.Route / r.Group expecting it to apply to Huma
// routes — it won't. Use huma.NewGroup + humawrap.Wrap.
func addRoutes(r chi.Router, apiPublic, apiAdmin huma.API, d Deps) {
	rateLimitD := d.rateLimitMiddlewareDeps()
	sdkAuthD := d.sdkAuthMiddlewareDeps()

	// -------------------- /v1 SDK plane --------------------

	// validate-key — public, unauthenticated. Dual rate-limit defense:
	// LimitByIP catches distributed IP floods, LimitByKeyPrefix catches
	// credential brute-force where the attacker rotates source IPs.
	sdkPublic := huma.NewGroup(apiPublic)
	sdkPublic.UseMiddleware(humawrap.WrapMany(
		middleware.LimitByIP(rateLimitD),
		middleware.LimitByKeyPrefix(rateLimitD),
	)...)
	authHandler.RegisterSDKRoutes(sdkPublic, d.APIKey, d.Logger)

	// All other SDK Huma routes — RequireSDKAuth + LimitByAPIKey.
	sdkAuth := huma.NewGroup(apiPublic)
	sdkAuth.UseMiddleware(humawrap.WrapMany(
		middleware.RequireSDKAuth(sdkAuthD),
		middleware.LimitByAPIKey(rateLimitD),
	)...)
	annotationHandler.RegisterSDKRoutes(sdkAuth, d.AnnotationItem, d.Logger)
	promptHandler.RegisterSDKRoutes(sdkAuth, d.Prompt, d.Logger)
	playgroundHandler.RegisterSDKRoutes(sdkAuth, d.Playground, d.Logger)
	observabilityHandler.RegisterSDKRoutes(sdkAuth, d.Observability.SpanQueryService, d.Logger)
	evaluationHandler.RegisterSDKRoutes(
		sdkAuth,
		d.EvalScoreConfig, d.EvalDataset, d.EvalDatasetItem, d.EvalDatasetVersion,
		d.EvalExperiment, d.EvalExperimentItem,
		d.Observability.ScoreService,
		d.Logger,
	)

	// OTLP ingestion — HUMA-EXEMPT raw chi routes. Register on a chi
	// r.Group with the same auth + rate-limit chain; this works
	// natively because the routes register on the chi subrouter, not
	// via humachi.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireSDKAuth(sdkAuthD))
		r.Use(middleware.LimitByAPIKey(rateLimitD))
		observabilityHandler.RegisterOTLPChiRoutes(r, observabilityHandler.OTLPDeps{
			StreamProducer:       d.Observability.StreamProducer,
			DeduplicationService: d.Observability.DeduplicationService,
			OTLPConverter:        d.Observability.OTLPConverterService,
			LogsConverter:        d.Observability.OTLPLogsConverterService,
			EventsConverter:      d.Observability.OTLPEventsConverterService,
			MetricsConverter:     d.Observability.OTLPMetricsConverterService,
			Logger:               d.Logger,
		})
	})

	// -------------------- /api/v1 dashboard plane --------------------
	// Mux-level middleware (CORS, CSRF) is already in place via
	// installGlobalMiddleware. Rate limiting lives on the Huma groups:
	// LimitByIP on the pre-auth surface, LimitByUser on the authed
	// surface. See the block comment above for the rationale.

	// Public dashboard routes — login, signup, password reset, OAuth,
	// token refresh, website contact form. No auth required.
	// LimitByIP defends the pre-auth surface against brute-force
	// credential stuffing and unauthenticated flood.
	dashPublic := huma.NewGroup(apiAdmin)
	dashPublic.UseMiddleware(humawrap.Wrap(middleware.LimitByIP(rateLimitD)))
	authHandler.RegisterPublicRoutes(dashPublic, authHandler.PublicDeps{
		Auth:          d.Auth,
		User:          d.User,
		Registration:  d.Registration,
		Session:       d.Session,
		OAuthProvider: d.OAuthProvider,
		Config:        d.Config,
		Logger:        d.Logger,
	})
	websiteHandler.RegisterRoutes(dashPublic, d.Website, d.Logger)

	// Authed dashboard routes — RequireAuth + LimitByUser.
	dashAuth := huma.NewGroup(apiAdmin)
	dashAuth.UseMiddleware(humawrap.WrapMany(
		middleware.RequireAuth(d.authMiddlewareDeps()),
		middleware.LimitByUser(rateLimitD),
	)...)

	// Chi-native sibling of dashAuth. Hosts handlers that have moved
	// off Huma during the chi-only migration. Shares the same
	// middleware chain (RequireAuth + LimitByUser) and mounts at the
	// mux root — converted handlers own the full `/api/v1/...` path
	// inside their own RegisterRoutes. Coexists with the Huma group
	// until the migration completes; at that point Phase 3 collapses
	// both groups into a single chi topology and this bridge
	// disappears.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(d.authMiddlewareDeps()))
		r.Use(middleware.LimitByUser(rateLimitD))
		credentialsHandler.RegisterRoutes(r, d.Credential, d.CredentialModelCatalog, d.Logger)
	})
	authHandler.RegisterProtectedRoutes(dashAuth, authHandler.ProtectedDeps{
		Auth:          d.Auth,
		User:          d.User,
		Profile:       d.Profile,
		Registration:  d.Registration,
		Session:       d.Session,
		OAuthProvider: d.OAuthProvider,
		Config:        d.Config,
		Logger:        d.Logger,
	})
	userHandler.RegisterRoutes(dashAuth, d.User, d.Profile, d.Organization, d.Logger)
	apikeyHandler.RegisterRoutes(dashAuth, d.APIKey, d.Logger)
	commentHandler.RegisterRoutes(dashAuth, d.Comment, d.Logger)
	overviewHandler.RegisterRoutes(dashAuth, d.Overview, d.Logger)
	// credentials — migrated to chi; mounted on the chi bridge group above.
	projectHandler.RegisterRoutes(dashAuth, d.Project, d.Organization, d.OrgMemberOrg, d.Logger)
	dashboardHandler.RegisterRoutes(dashAuth, d.Dashboard, d.DashboardQuery, d.DashboardTemplate, d.Logger)
	annotationHandler.RegisterRoutes(dashAuth, d.AnnotationQueue, d.AnnotationItem, d.AnnotationAssignment, d.Logger)
	billingHandler.RegisterRoutes(dashAuth, d.BillingUsage, d.BillingBudget, d.BillingContract, d.BillingPricing, d.Logger)
	organizationHandler.RegisterRoutes(dashAuth, d.Organization, d.OrgMemberOrg, d.Invitation, d.OrgSettings, d.Logger)
	promptHandler.RegisterRoutes(dashAuth, d.Prompt, d.PromptCompiler, d.Logger)
	rbacHandler.RegisterRoutes(dashAuth, d.Role, d.Permission, d.OrgMember, d.Scope, d.Logger)
	playgroundHandler.RegisterRoutes(dashAuth, d.Playground, d.Project, d.Logger)
	observabilityHandler.RegisterRoutes(
		dashAuth,
		d.Observability.TraceService,
		d.Observability.ScoreService,
		d.Observability.ScoreAnalyticsService,
		d.Observability.FilterPresetService,
		d.Logger,
	)
	evaluationHandler.RegisterRoutes(
		dashAuth,
		d.EvalScoreConfig, d.EvalDataset, d.EvalDatasetItem, d.EvalDatasetVersion,
		d.EvalExperiment, d.EvalExperimentItem, d.EvalExperimentWizard,
		d.EvalEvaluator, d.EvalEvaluatorExecution,
		d.Logger,
	)
}
