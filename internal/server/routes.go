package server

import (
	"github.com/go-chi/chi/v5"

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
//	  │     └── h.AuthSDK.ValidateAPIKey
//	  └── sdkAuth    chi.Group  → RequireSDKAuth + LimitByAPIKey
//	        ├── observability OTLP (raw protobuf)
//	        ├── annotation, prompt, playground SDK
//	        ├── observability SDK (span query)
//	        └── evaluation SDK
//
//	/api/v1 (dashboard plane, cookie+JWT auth)
//	  ├── dashPublic chi.Group  → LimitByIP
//	  │     ├── h.Auth.* (login, signup, OAuth, refresh)
//	  │     └── h.Website.SubmitContactForm
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
	h := d.Handlers
	rateLimitD := d.rateLimitMiddlewareDeps()
	sdkAuthD := d.sdkAuthMiddlewareDeps()
	authD := d.authMiddlewareDeps()

	// =================================================================
	// /v1 SDK plane
	// =================================================================

	// SDK pre-auth: validate-key (dual rate-limit defence — IP for
	// flood, key-prefix for credential brute force).
	r.Group(func(r chi.Router) {
		r.Use(middleware.LimitByIP(rateLimitD))
		r.Use(middleware.LimitByKeyPrefix(rateLimitD))

		r.Post("/v1/auth/validate-key", h.AuthSDK.ValidateAPIKey)
	})

	// SDK authed surface — RequireSDKAuth + LimitByAPIKey.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireSDKAuth(sdkAuthD))
		r.Use(middleware.LimitByAPIKey(rateLimitD))

		// OTLP protobuf ingestion.
		r.Post("/v1/traces", h.Observability.OTLP.HandleTraces)
		r.Post("/v1/logs", h.Observability.OTLP.HandleLogs)
		r.Post("/v1/metrics", h.Observability.OTLP.HandleMetrics)

		// Span query.
		r.Post("/v1/spans/query", h.Observability.SDK.QuerySpans)
		r.Post("/v1/spans/query/validate", h.Observability.SDK.ValidateFilter)

		// Playground execution.
		r.Post("/v1/playground/execute", h.Playground.SdkExecute)

		// Prompts.
		r.Post("/v1/prompts", h.Prompt.UpsertPrompt)
		r.Get("/v1/prompts", h.Prompt.ListPromptsSDK)
		r.Get("/v1/prompts/{name}", h.Prompt.GetPromptByName)

		// Annotation queue items.
		r.Post("/v1/annotation-queues/{queueId}/items", h.Annotation.SdkAddItems)
		r.Get("/v1/annotation-queues/{queueId}/items", h.Annotation.SdkListItems)

		// Evaluation — datasets.
		r.Post("/v1/datasets", h.Evaluation.SdkCreateDataset)
		r.Get("/v1/datasets", h.Evaluation.SdkListDatasets)
		r.Get("/v1/datasets/{datasetId}", h.Evaluation.SdkGetDataset)
		r.Patch("/v1/datasets/{datasetId}", h.Evaluation.SdkUpdateDataset)
		r.Delete("/v1/datasets/{datasetId}", h.Evaluation.SdkDeleteDataset)
		r.Get("/v1/datasets/{datasetId}/info", h.Evaluation.SdkGetDatasetWithVersionInfo)
		r.Post("/v1/datasets/{datasetId}/pin", h.Evaluation.SdkPinDatasetVersion)
		r.Post("/v1/datasets/{datasetId}/items", h.Evaluation.SdkBatchCreateDatasetItems)
		r.Get("/v1/datasets/{datasetId}/items", h.Evaluation.SdkListDatasetItems)
		r.Get("/v1/datasets/{datasetId}/items/export", h.Evaluation.SdkExportDatasetItems)
		r.Post("/v1/datasets/{datasetId}/items/import-json", h.Evaluation.SdkImportItemsJSON)
		r.Post("/v1/datasets/{datasetId}/items/import-csv", h.Evaluation.SdkImportItemsCSV)
		r.Post("/v1/datasets/{datasetId}/items/from-traces", h.Evaluation.SdkItemsFromTraces)
		r.Post("/v1/datasets/{datasetId}/items/from-spans", h.Evaluation.SdkItemsFromSpans)
		r.Post("/v1/datasets/{datasetId}/versions", h.Evaluation.SdkCreateDatasetVersion)
		r.Get("/v1/datasets/{datasetId}/versions", h.Evaluation.SdkListDatasetVersions)
		r.Get("/v1/datasets/{datasetId}/versions/{versionId}", h.Evaluation.SdkGetDatasetVersion)
		r.Get("/v1/datasets/{datasetId}/versions/{versionId}/items", h.Evaluation.SdkGetDatasetVersionItems)

		// Evaluation — experiments.
		r.Post("/v1/experiments", h.Evaluation.SdkCreateExperiment)
		r.Get("/v1/experiments", h.Evaluation.SdkListExperiments)
		r.Post("/v1/experiments/compare", h.Evaluation.SdkCompareExperiments)
		r.Get("/v1/experiments/{experimentId}", h.Evaluation.SdkGetExperiment)
		r.Patch("/v1/experiments/{experimentId}", h.Evaluation.SdkUpdateExperiment)
		r.Post("/v1/experiments/{experimentId}/rerun", h.Evaluation.SdkRerunExperiment)
		r.Post("/v1/experiments/{experimentId}/items", h.Evaluation.SdkBatchCreateExperimentItems)

		// Evaluation — scores.
		r.Post("/v1/scores", h.Evaluation.SdkCreateScore)
		r.Post("/v1/scores/batch", h.Evaluation.SdkCreateScoreBatch)
	})

	// =================================================================
	// /api/v1 dashboard plane
	// =================================================================

	// Dashboard pre-auth: login, signup, password reset, OAuth,
	// token refresh, website contact form. LimitByIP only.
	r.Group(func(r chi.Router) {
		r.Use(middleware.LimitByIP(rateLimitD))

		r.Post("/api/v1/auth/login", h.Auth.Login)
		r.Post("/api/v1/auth/signup", h.Auth.Signup)
		r.Post("/api/v1/auth/refresh", h.Auth.Refresh)
		r.Post("/api/v1/auth/forgot-password", h.Auth.ForgotPassword)
		r.Post("/api/v1/auth/reset-password", h.Auth.ResetPassword)
		r.Get("/api/v1/auth/google", h.Auth.InitiateGoogleOAuth)
		r.Get("/api/v1/auth/google/callback", h.Auth.GoogleOAuthCallback)
		r.Get("/api/v1/auth/github", h.Auth.InitiateGithubOAuth)
		r.Get("/api/v1/auth/github/callback", h.Auth.GithubOAuthCallback)
		r.Post("/api/v1/auth/complete-oauth-signup", h.Auth.CompleteOAuthSignup)
		r.Post("/api/v1/auth/exchange-session/{session_id}", h.Auth.ExchangeLoginSession)

		r.Post("/api/v1/website/contact", h.Website.SubmitContact)
	})

	// Dashboard authed surface — three scopes: profile/global,
	// org-scoped, project-scoped. Each tenant scope gets its own
	// subgroup with the matching membership middleware.
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authD))
		r.Use(middleware.LimitByUser(rateLimitD))

		// -----------------------------------------------------------
		// Profile + global (no tenant scope)
		// -----------------------------------------------------------

		// Auth (protected) — me, logout, password change, profile,
		// session management.
		r.Get("/api/v1/auth/me", h.Auth.GetCurrentUser)
		r.Post("/api/v1/auth/logout", h.Auth.Logout)
		r.Post("/api/v1/auth/change-password", h.Auth.ChangePassword)
		r.Get("/api/v1/auth/profile", h.Auth.GetProfile)
		r.Patch("/api/v1/auth/profile", h.Auth.UpdateProfile)
		r.Get("/api/v1/auth/sessions", h.Auth.ListSessions)
		r.Get("/api/v1/auth/sessions/{session_id}", h.Auth.GetSession)
		r.Post("/api/v1/auth/sessions/{session_id}/revoke", h.Auth.RevokeSession)
		r.Post("/api/v1/auth/sessions/revoke-all", h.Auth.RevokeAllSessions)

		// User (current-user profile — distinct from auth/profile;
		// returns the org-hierarchy-aware shape the dashboard sidebar
		// reads on every page load). Convention matches GitHub /user,
		// Sentry /users/me/, PostHog /api/users/@me/.
		r.Get("/api/v1/users/me", h.User.GetProfile)
		r.Patch("/api/v1/users/me", h.User.UpdateProfile)
		r.Put("/api/v1/users/me/default-organization", h.User.SetDefaultOrganization)

		// Organization list/create + invitation flow (pre-membership;
		// the recipient accepts before becoming an org member).
		r.Get("/api/v1/organizations", h.Organization.ListOrganizations)
		r.Post("/api/v1/organizations", h.Organization.CreateOrganization)
		r.Get("/api/v1/invitations", h.Organization.ListUserInvitations)
		r.Get("/api/v1/invitations/validate/{token}", h.Organization.ValidateInvitationToken)
		r.Post("/api/v1/invitations/accept", h.Organization.AcceptInvitation)
		r.Post("/api/v1/invitations/decline", h.Organization.DeclineInvitation)

		// Dashboard global catalogs — view definitions + templates.
		r.Get("/api/v1/dashboards/view-definitions", h.Dashboard.ViewDefinitions)
		r.Get("/api/v1/dashboard-templates", h.Dashboard.ListTemplates)
		r.Get("/api/v1/dashboard-templates/{templateId}", h.Dashboard.GetTemplate)

		// Annotation — user's own assignments (cross-project view).
		r.Get("/api/v1/annotation-queues/my-assignments", h.Annotation.MyAssignments)

		// Billing — contracts (enterprise pricing) + per-org
		// effective-pricing lookup. List/CRUD on contracts itself
		// is global; usage + budgets are org-scoped (below).
		r.Post("/api/v1/billing/contracts", h.Billing.CreateContract)
		r.Get("/api/v1/billing/organizations/{orgId}/contracts", h.Billing.ListContracts)
		r.Get("/api/v1/billing/organizations/{orgId}/effective-pricing", h.Billing.GetEffectivePricing)
		r.Get("/api/v1/billing/contracts/{contractId}", h.Billing.GetContract)
		r.Put("/api/v1/billing/contracts/{contractId}", h.Billing.UpdateContract)
		r.Delete("/api/v1/billing/contracts/{contractId}", h.Billing.CancelContract)
		r.Put("/api/v1/billing/contracts/{contractId}/activate", h.Billing.ActivateContract)
		r.Put("/api/v1/billing/contracts/{contractId}/tiers", h.Billing.UpdateContractTiers)
		r.Get("/api/v1/billing/contracts/{contractId}/history", h.Billing.GetContractHistory)

		// RBAC — role/permission catalog + per-user role/permission
		// lookups. Global catalog (not org-scoped).
		r.Get("/api/v1/rbac/roles", h.RBAC.ListRoles)
		r.Get("/api/v1/rbac/roles/statistics", h.RBAC.GetRoleStatistics)
		r.Get("/api/v1/rbac/roles/{roleId}", h.RBAC.GetRole)
		r.Get("/api/v1/rbac/users/{userId}/roles", h.RBAC.GetUserRoles)
		r.Get("/api/v1/rbac/users/{userId}/permissions", h.RBAC.GetUserPermissions)
		r.Post("/api/v1/rbac/users/{userId}/permissions/check", h.RBAC.CheckUserPermissions)
		r.Post("/api/v1/rbac/users/{userId}/organizations/{orgId}/roles", h.RBAC.AssignOrganizationRole)
		r.Delete("/api/v1/rbac/users/{userId}/organizations/{orgId}", h.RBAC.RemoveOrganizationMember)
		r.Post("/api/v1/rbac/users/{userId}/scopes/check", h.RBAC.CheckUserScopes)
		r.Get("/api/v1/rbac/users/{userId}/scopes", h.RBAC.GetUserScopes)
		r.Get("/api/v1/rbac/permissions", h.RBAC.ListPermissions)
		r.Get("/api/v1/rbac/permissions/resources", h.RBAC.GetAvailableResources)
		r.Get("/api/v1/rbac/permissions/resources/{resource}/actions", h.RBAC.GetActionsForResource)
		r.Get("/api/v1/rbac/permissions/{permissionId}", h.RBAC.GetPermission)
		r.Get("/api/v1/rbac/scopes", h.RBAC.GetAvailableScopes)
		r.Get("/api/v1/rbac/scopes/categories", h.RBAC.GetScopeCategories)

		// -----------------------------------------------------------
		// Org-scoped: /api/v1/organizations/{orgId}/...
		// RequireOrganizationAccess validates membership pre-handler
		// and pins orgID via httpctx.WithOrganizationID.
		// -----------------------------------------------------------
		r.Route("/api/v1/organizations/{orgId}", func(r chi.Router) {
			r.Use(middleware.RequireOrganizationAccess(authD))

			// Organization detail + members + invitations + settings.
			r.With(middleware.RequirePermission(authD, "organizations:read")).Get("/", h.Organization.GetOrganization)
			r.With(middleware.RequirePermission(authD, "organizations:write")).Patch("/", h.Organization.UpdateOrganization)
			r.With(middleware.RequirePermission(authD, "organizations:delete")).Delete("/", h.Organization.DeleteOrganization)
			r.With(middleware.RequirePermission(authD, "members:read")).Get("/members", h.Organization.ListMembers)
			r.With(middleware.RequirePermission(authD, "members:remove")).Delete("/members/{userId}", h.Organization.RemoveMember)
			r.With(middleware.RequirePermission(authD, "members:invite")).Post("/invitations", h.Organization.CreateInvitation)
			r.With(middleware.RequirePermission(authD, "members:read")).Get("/invitations", h.Organization.ListPendingInvitations)
			r.With(middleware.RequirePermission(authD, "members:invite")).Post("/invitations/{invitationId}/resend", h.Organization.ResendInvitation)
			r.With(middleware.RequirePermission(authD, "members:invite")).Delete("/invitations/{invitationId}", h.Organization.RevokeInvitation)
			r.With(middleware.RequirePermission(authD, "settings:read")).Get("/settings", h.Organization.ListSettings)
			r.With(middleware.RequirePermission(authD, "settings:write")).Post("/settings", h.Organization.CreateSetting)
			r.With(middleware.RequirePermission(authD, "settings:read")).Get("/settings/{key}", h.Organization.GetSetting)
			r.With(middleware.RequirePermission(authD, "settings:write")).Put("/settings/{key}", h.Organization.UpdateSetting)
			r.With(middleware.RequirePermission(authD, "settings:write")).Delete("/settings/{key}", h.Organization.DeleteSetting)

			// Billing — usage + budgets (org-scoped).
			r.With(middleware.RequirePermission(authD, "billing:read")).Get("/usage/overview", h.Billing.GetUsageOverview)
			r.With(middleware.RequirePermission(authD, "billing:read")).Get("/usage/timeseries", h.Billing.GetUsageTimeSeries)
			r.With(middleware.RequirePermission(authD, "billing:read")).Get("/usage/by-project", h.Billing.GetUsageByProject)
			r.With(middleware.RequirePermission(authD, "billing:export")).Get("/usage/export", h.Billing.ExportUsage)
			r.With(middleware.RequirePermission(authD, "billing:read")).Get("/budgets", h.Billing.ListBudgets)
			r.With(middleware.RequirePermission(authD, "billing:manage")).Post("/budgets", h.Billing.CreateBudget)
			r.With(middleware.RequirePermission(authD, "billing:read")).Get("/budgets/alerts", h.Billing.GetBudgetAlerts)
			r.With(middleware.RequirePermission(authD, "billing:manage")).Post("/budgets/alerts/{alertId}/acknowledge", h.Billing.AcknowledgeBudgetAlert)
			r.With(middleware.RequirePermission(authD, "billing:read")).Get("/budgets/{budgetId}", h.Billing.GetBudget)
			r.With(middleware.RequirePermission(authD, "billing:manage")).Put("/budgets/{budgetId}", h.Billing.UpdateBudget)
			r.With(middleware.RequirePermission(authD, "billing:manage")).Delete("/budgets/{budgetId}", h.Billing.DeleteBudget)

			// AI provider credentials (org-level).
			r.With(middleware.RequirePermission(authD, "credentials:write")).Post("/credentials/ai", h.Credentials.Create)
			r.With(middleware.RequirePermission(authD, "credentials:read")).Get("/credentials/ai", h.Credentials.List)
			r.With(middleware.RequirePermission(authD, "credentials:read")).Get("/credentials/ai/{credentialId}", h.Credentials.Get)
			r.With(middleware.RequirePermission(authD, "credentials:write")).Patch("/credentials/ai/{credentialId}", h.Credentials.Update)
			r.With(middleware.RequirePermission(authD, "credentials:write")).Delete("/credentials/ai/{credentialId}", h.Credentials.Delete)
			r.With(middleware.RequirePermission(authD, "credentials:read")).Post("/credentials/ai/test", h.Credentials.TestConnection)
			r.With(middleware.RequirePermission(authD, "credentials:read")).Get("/credentials/ai/models", h.Credentials.GetAvailableModels)

			// Custom roles (RBAC) — org-scoped role lifecycle.
			r.With(middleware.RequirePermission(authD, "roles:write")).Post("/roles", h.RBAC.CreateCustomRole)
			r.With(middleware.RequirePermission(authD, "roles:read")).Get("/roles", h.RBAC.ListCustomRoles)
			r.With(middleware.RequirePermission(authD, "roles:read")).Get("/roles/{roleId}", h.RBAC.GetCustomRole)
			r.With(middleware.RequirePermission(authD, "roles:write")).Patch("/roles/{roleId}", h.RBAC.UpdateCustomRole)
			r.With(middleware.RequirePermission(authD, "roles:delete")).Delete("/roles/{roleId}", h.RBAC.DeleteCustomRole)

			// Project list + create (mounted under org subgroup
			// because both operations need only org context).
			r.With(middleware.RequirePermission(authD, "projects:read")).Get("/projects", h.Project.List)
			r.With(middleware.RequirePermission(authD, "projects:write")).Post("/projects", h.Project.Create)
		})

		// -----------------------------------------------------------
		// Project-scoped: /api/v1/projects/{projectId}/...
		// RequireProjectAccess loads the project, verifies the caller
		// is a member of the project's organization, then pins both
		// projectID and orgID into ctx. The pinned orgID is consumed
		// by RequirePermission's scope-aware resolver (Langfuse MAX
		// semantics across org+project roles).
		// -----------------------------------------------------------
		r.Route("/api/v1/projects/{projectId}", func(r chi.Router) {
			r.Use(middleware.RequireProjectAccess(authD))

			// Project detail.
			r.With(middleware.RequirePermission(authD, "projects:read")).Get("/", h.Project.Get)
			r.With(middleware.RequirePermission(authD, "projects:write")).Put("/", h.Project.Update)
			r.With(middleware.RequirePermission(authD, "projects:delete")).Delete("/", h.Project.Delete)
			r.With(middleware.RequirePermission(authD, "projects:write")).Post("/archive", h.Project.Archive)
			r.With(middleware.RequirePermission(authD, "projects:write")).Post("/unarchive", h.Project.Unarchive)

			// API keys.
			r.With(middleware.RequirePermission(authD, "api-keys:read")).Get("/api-keys", h.APIKey.List)
			r.With(middleware.RequirePermission(authD, "api-keys:create")).Post("/api-keys", h.APIKey.Create)
			r.With(middleware.RequirePermission(authD, "api-keys:delete")).Delete("/api-keys/{keyId}", h.APIKey.Delete)

			// Project-level role overrides (Langfuse two-tier RBAC).
			// members:read to view, members:update to assign/change/remove.
			// The user being assigned must already be an org member (the
			// service enforces this).
			r.With(middleware.RequirePermission(authD, "members:read")).Get("/members", h.ProjectMember.List)
			r.With(middleware.RequirePermission(authD, "members:update")).Post("/members", h.ProjectMember.Add)
			r.With(middleware.RequirePermission(authD, "members:update")).Patch("/members/{userId}", h.ProjectMember.UpdateRole)
			r.With(middleware.RequirePermission(authD, "members:update")).Delete("/members/{userId}", h.ProjectMember.Remove)

			// Overview.
			r.With(middleware.RequirePermission(authD, "analytics:read")).Get("/overview", h.Overview.GetOverview)

			// Dashboards (project-scoped).
			r.With(middleware.RequirePermission(authD, "dashboards:read")).Get("/dashboards", h.Dashboard.List)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Post("/dashboards", h.Dashboard.Create)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Post("/dashboards/import", h.Dashboard.ImportDashboard)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Post("/dashboards/from-template", h.Dashboard.CreateFromTemplate)
			r.With(middleware.RequirePermission(authD, "dashboards:read")).Get("/dashboards/variable-options", h.Dashboard.VariableOptions)
			r.With(middleware.RequirePermission(authD, "dashboards:read")).Get("/dashboards/{dashboardId}", h.Dashboard.Get)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Put("/dashboards/{dashboardId}", h.Dashboard.Update)
			r.With(middleware.RequirePermission(authD, "dashboards:delete")).Delete("/dashboards/{dashboardId}", h.Dashboard.Delete)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Post("/dashboards/{dashboardId}/duplicate", h.Dashboard.Duplicate)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Post("/dashboards/{dashboardId}/lock", h.Dashboard.Lock)
			r.With(middleware.RequirePermission(authD, "dashboards:write")).Post("/dashboards/{dashboardId}/unlock", h.Dashboard.Unlock)
			r.With(middleware.RequirePermission(authD, "dashboards:read")).Get("/dashboards/{dashboardId}/export", h.Dashboard.Export)
			r.With(middleware.RequirePermission(authD, "dashboards:execute")).Post("/dashboards/{dashboardId}/execute", h.Dashboard.ExecuteDashboard)
			r.With(middleware.RequirePermission(authD, "dashboards:execute")).Post("/dashboards/{dashboardId}/widgets/{widgetId}/execute", h.Dashboard.ExecuteWidget)

			// Annotation queues (project-scoped).
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues", h.Annotation.CreateQueue)
			r.With(middleware.RequirePermission(authD, "annotation-queues:read")).Get("/annotation-queues", h.Annotation.ListQueues)
			r.With(middleware.RequirePermission(authD, "annotation-queues:read")).Get("/annotation-queues/{queueId}", h.Annotation.GetQueue)
			r.With(middleware.RequirePermission(authD, "annotation-queues:read")).Get("/annotation-queues/{queueId}/stats", h.Annotation.GetQueueWithStats)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Put("/annotation-queues/{queueId}", h.Annotation.UpdateQueue)
			r.With(middleware.RequirePermission(authD, "annotation-queues:delete")).Delete("/annotation-queues/{queueId}", h.Annotation.DeleteQueue)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues/{queueId}/items", h.Annotation.AddItems)
			r.With(middleware.RequirePermission(authD, "annotation-queues:read")).Get("/annotation-queues/{queueId}/items", h.Annotation.ListItems)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues/{queueId}/items/claim", h.Annotation.ClaimNext)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues/{queueId}/items/{itemId}/complete", h.Annotation.CompleteItem)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues/{queueId}/items/{itemId}/skip", h.Annotation.SkipItem)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues/{queueId}/items/{itemId}/release", h.Annotation.ReleaseLock)
			r.With(middleware.RequirePermission(authD, "annotation-queues:delete")).Delete("/annotation-queues/{queueId}/items/{itemId}", h.Annotation.DeleteItem)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Post("/annotation-queues/{queueId}/assignments", h.Annotation.AssignUser)
			r.With(middleware.RequirePermission(authD, "annotation-queues:read")).Get("/annotation-queues/{queueId}/assignments", h.Annotation.ListAssignments)
			r.With(middleware.RequirePermission(authD, "annotation-queues:write")).Delete("/annotation-queues/{queueId}/assignments/{userId}", h.Annotation.UnassignUser)

			// Prompts.
			r.With(middleware.RequirePermission(authD, "prompts:read")).Get("/prompts", h.Prompt.ListPrompts)
			r.With(middleware.RequirePermission(authD, "prompts:create")).Post("/prompts", h.Prompt.CreatePrompt)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Get("/prompts/settings/protected-labels", h.Prompt.GetProtectedLabels)
			r.With(middleware.RequirePermission(authD, "prompts:update")).Put("/prompts/settings/protected-labels", h.Prompt.SetProtectedLabels)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Post("/prompts/validate-template", h.Prompt.ValidateTemplate)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Post("/prompts/preview-template", h.Prompt.PreviewTemplate)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Post("/prompts/detect-dialect", h.Prompt.DetectDialect)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Get("/prompts/{promptId}", h.Prompt.GetPrompt)
			r.With(middleware.RequirePermission(authD, "prompts:update")).Put("/prompts/{promptId}", h.Prompt.UpdatePrompt)
			r.With(middleware.RequirePermission(authD, "prompts:delete")).Delete("/prompts/{promptId}", h.Prompt.DeletePrompt)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Get("/prompts/{promptId}/diff", h.Prompt.GetVersionDiff)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Get("/prompts/{promptId}/versions", h.Prompt.ListVersions)
			r.With(middleware.RequirePermission(authD, "prompts:create")).Post("/prompts/{promptId}/versions", h.Prompt.CreateVersion)
			r.With(middleware.RequirePermission(authD, "prompts:read")).Get("/prompts/{promptId}/versions/{versionId}", h.Prompt.GetVersion)
			r.With(middleware.RequirePermission(authD, "prompts:update")).Patch("/prompts/{promptId}/versions/{versionId}/labels", h.Prompt.SetLabels)

			// Playground (execute + stream + sessions).
			r.With(middleware.RequirePermission(authD, "playground:execute")).Post("/playground/execute", h.Playground.Execute)
			r.With(middleware.RequirePermission(authD, "playground:execute")).Post("/playground/stream", h.Playground.Stream)
			r.With(middleware.RequirePermission(authD, "playground:execute")).Post("/playground/sessions", h.Playground.CreateSession)
			r.With(middleware.RequirePermission(authD, "playground:read")).Get("/playground/sessions", h.Playground.ListSessions)
			r.With(middleware.RequirePermission(authD, "playground:read")).Get("/playground/sessions/{sessionId}", h.Playground.GetSession)
			r.With(middleware.RequirePermission(authD, "playground:execute")).Put("/playground/sessions/{sessionId}", h.Playground.UpdateSession)
			r.With(middleware.RequirePermission(authD, "playground:execute")).Delete("/playground/sessions/{sessionId}", h.Playground.DeleteSession)

			// Comments (trace-attached threads).
			r.With(middleware.RequirePermission(authD, "comments:write")).Post("/traces/{id}/comments", h.Comment.Create)
			r.With(middleware.RequirePermission(authD, "comments:read")).Get("/traces/{id}/comments", h.Comment.List)
			r.With(middleware.RequirePermission(authD, "comments:read")).Get("/traces/{id}/comments/count", h.Comment.Count)
			r.With(middleware.RequirePermission(authD, "comments:write")).Put("/traces/{id}/comments/{comment_id}", h.Comment.Update)
			r.With(middleware.RequirePermission(authD, "comments:delete")).Delete("/traces/{id}/comments/{comment_id}", h.Comment.Delete)
			r.With(middleware.RequirePermission(authD, "comments:write")).Post("/traces/{id}/comments/{comment_id}/reactions", h.Comment.ToggleReaction)
			r.With(middleware.RequirePermission(authD, "comments:write")).Post("/traces/{id}/comments/{comment_id}/replies", h.Comment.CreateReply)

			// Observability — traces.
			obs := h.Observability.Dashboard
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/traces", obs.ListTraces)
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/traces/filter-options", obs.GetTraceFilterOptions)
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/traces/attributes", obs.DiscoverAttributes)
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/traces/{id}", obs.GetTrace)
			r.With(middleware.RequirePermission(authD, "traces:delete")).Delete("/traces/{id}", obs.DeleteTrace)
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/traces/{id}/spans", obs.GetTraceSpans)
			r.With(middleware.RequirePermission(authD, "scores:read")).Get("/traces/{id}/scores", obs.GetTraceScores)
			r.With(middleware.RequirePermission(authD, "scores:write")).Post("/traces/{id}/scores", obs.CreateTraceScore)
			r.With(middleware.RequirePermission(authD, "scores:delete")).Delete("/traces/{id}/scores/{scoreId}", obs.DeleteTraceScore)
			r.With(middleware.RequirePermission(authD, "traces:read")).Put("/traces/{id}/tags", obs.UpdateTraceTags)
			r.With(middleware.RequirePermission(authD, "traces:read")).Put("/traces/{id}/bookmark", obs.UpdateTraceBookmark)

			// Observability — spans.
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/spans", obs.ListSpans)
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/spans/{id}", obs.GetSpan)
			r.With(middleware.RequirePermission(authD, "traces:delete")).Delete("/spans/{id}", obs.DeleteSpan)

			// Observability — scores.
			r.With(middleware.RequirePermission(authD, "scores:read")).Get("/scores", obs.ListScores)
			r.With(middleware.RequirePermission(authD, "scores:read")).Get("/scores/{id}", obs.GetScore)
			r.With(middleware.RequirePermission(authD, "scores:write")).Put("/scores/{id}", obs.UpdateScore)
			r.With(middleware.RequirePermission(authD, "scores:read")).Get("/scores/analytics", obs.GetScoreAnalytics)
			r.With(middleware.RequirePermission(authD, "scores:read")).Get("/scores/names", obs.GetScoreNames)

			// Observability — sessions.
			r.With(middleware.RequirePermission(authD, "traces:read")).Get("/sessions", obs.ListSessions)

			// Observability — filter presets.
			r.With(middleware.RequirePermission(authD, "filter-presets:write")).Post("/filter-presets", obs.CreateFilterPreset)
			r.With(middleware.RequirePermission(authD, "filter-presets:read")).Get("/filter-presets", obs.ListFilterPresets)
			r.With(middleware.RequirePermission(authD, "filter-presets:read")).Get("/filter-presets/{id}", obs.GetFilterPreset)
			r.With(middleware.RequirePermission(authD, "filter-presets:write")).Patch("/filter-presets/{id}", obs.UpdateFilterPreset)
			r.With(middleware.RequirePermission(authD, "filter-presets:delete")).Delete("/filter-presets/{id}", obs.DeleteFilterPreset)

			// Evaluation — score configs.
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/score-configs", h.Evaluation.CreateScoreConfig)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/score-configs", h.Evaluation.ListScoreConfigs)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/score-configs/{configId}", h.Evaluation.GetScoreConfig)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Put("/score-configs/{configId}", h.Evaluation.UpdateScoreConfig)
			r.With(middleware.RequirePermission(authD, "evaluations:delete")).Delete("/score-configs/{configId}", h.Evaluation.DeleteScoreConfig)

			// Evaluation — datasets.
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets", h.Evaluation.DashCreateDataset)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets", h.Evaluation.DashListDatasets)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}", h.Evaluation.DashGetDataset)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Put("/datasets/{datasetId}", h.Evaluation.DashUpdateDataset)
			r.With(middleware.RequirePermission(authD, "evaluations:delete")).Delete("/datasets/{datasetId}", h.Evaluation.DashDeleteDataset)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/info", h.Evaluation.DashGetDatasetWithVersionInfo)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/pin", h.Evaluation.DashPinDatasetVersion)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/items", h.Evaluation.DashListDatasetItems)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/items", h.Evaluation.DashCreateDatasetItem)
			r.With(middleware.RequirePermission(authD, "evaluations:delete")).Delete("/datasets/{datasetId}/items/{itemId}", h.Evaluation.DashDeleteDatasetItem)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/items/export", h.Evaluation.DashExportDatasetItems)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/items/import-json", h.Evaluation.DashImportItemsJSON)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/items/import-csv", h.Evaluation.DashImportItemsCSV)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/items/from-traces", h.Evaluation.DashItemsFromTraces)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/items/from-spans", h.Evaluation.DashItemsFromSpans)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/datasets/{datasetId}/versions", h.Evaluation.DashCreateDatasetVersion)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/versions", h.Evaluation.DashListDatasetVersions)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/versions/{versionId}", h.Evaluation.DashGetDatasetVersion)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/versions/{versionId}/items", h.Evaluation.DashGetDatasetVersionItems)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/datasets/{datasetId}/fields", h.Evaluation.WizardDatasetFields)

			// Evaluation — experiments.
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/experiments", h.Evaluation.DashCreateExperiment)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/experiments", h.Evaluation.DashListExperiments)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Post("/experiments/compare", h.Evaluation.DashCompareExperiments)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/experiments/{experimentId}", h.Evaluation.DashGetExperiment)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Put("/experiments/{experimentId}", h.Evaluation.DashUpdateExperiment)
			r.With(middleware.RequirePermission(authD, "evaluations:delete")).Delete("/experiments/{experimentId}", h.Evaluation.DashDeleteExperiment)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/experiments/{experimentId}/rerun", h.Evaluation.DashRerunExperiment)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/experiments/{experimentId}/progress", h.Evaluation.DashGetExperimentProgress)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/experiments/{experimentId}/metrics", h.Evaluation.DashGetExperimentMetrics)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/experiments/{experimentId}/items", h.Evaluation.DashListExperimentItems)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/experiments/{experimentId}/config", h.Evaluation.WizardGetConfig)

			// Evaluation — experiment wizard.
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/experiments/wizard", h.Evaluation.WizardCreate)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Post("/experiments/wizard/validate", h.Evaluation.WizardValidate)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Post("/experiments/wizard/estimate", h.Evaluation.WizardEstimate)

			// Evaluation — evaluators.
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/evaluators", h.Evaluation.CreateEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators", h.Evaluation.ListEvaluators)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators/{evaluatorId}", h.Evaluation.GetEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Put("/evaluators/{evaluatorId}", h.Evaluation.UpdateEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:delete")).Delete("/evaluators/{evaluatorId}", h.Evaluation.DeleteEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/evaluators/{evaluatorId}/activate", h.Evaluation.ActivateEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/evaluators/{evaluatorId}/deactivate", h.Evaluation.DeactivateEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/evaluators/{evaluatorId}/trigger", h.Evaluation.TriggerEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:write")).Post("/evaluators/{evaluatorId}/test", h.Evaluation.TestEvaluator)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators/{evaluatorId}/analytics", h.Evaluation.GetEvaluatorAnalytics)

			// Evaluation — evaluator executions.
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators/{evaluatorId}/executions", h.Evaluation.ListExecutions)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators/{evaluatorId}/executions/latest", h.Evaluation.GetLatestExecution)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators/{evaluatorId}/executions/{executionId}", h.Evaluation.GetExecution)
			r.With(middleware.RequirePermission(authD, "evaluations:read")).Get("/evaluators/{evaluatorId}/executions/{executionId}/detail", h.Evaluation.GetExecutionDetail)
		})
	})
}
