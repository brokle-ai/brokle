package http

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/handlers"
	"brokle/internal/transport/http/middleware"

	"github.com/redis/go-redis/v9"
)

type Server struct {
	config              *config.Config
	logger              *slog.Logger
	server              *http.Server
	listener            net.Listener
	handlers            *handlers.Handlers
	engine              *gin.Engine
	authMiddleware      *middleware.AuthMiddleware
	sdkAuthMiddleware   *middleware.SDKAuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
	csrfMiddleware      *middleware.CSRFMiddleware
	serveErr            chan error
}

func NewServer(
	cfg *config.Config,
	logger *slog.Logger,
	handlers *handlers.Handlers,
	jwtService auth.JWTService,
	blacklistedTokens auth.BlacklistedTokenService,
	orgMemberService auth.OrganizationMemberService,
	apiKeyService auth.APIKeyService,
	redisClient *redis.Client,
) *Server {
	authMiddleware := middleware.NewAuthMiddleware(
		jwtService,
		blacklistedTokens,
		orgMemberService,
		logger,
	)

	sdkAuthMiddleware := middleware.NewSDKAuthMiddleware(
		apiKeyService,
		logger,
	)

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		redisClient,
		&cfg.Auth,
		logger,
	)

	csrfMiddleware := middleware.NewCSRFMiddleware(logger)

	return &Server{
		config:              cfg,
		logger:              logger,
		handlers:            handlers,
		authMiddleware:      authMiddleware,
		sdkAuthMiddleware:   sdkAuthMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
		csrfMiddleware:      csrfMiddleware,
	}
}

func (s *Server) Start() error {
	if s.config.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	s.engine = gin.New()

	corsConfig := cors.DefaultConfig()

	// Validate wildcard incompatibility with credentials
	if len(s.config.Server.CORSAllowedOrigins) == 1 && s.config.Server.CORSAllowedOrigins[0] == "*" {
		// CRITICAL: Wildcard incompatible with AllowCredentials (cookies won't work)
		s.logger.Error("CORS misconfiguration: cannot use wildcard (*) origins with AllowCredentials (httpOnly cookies require specific origins). " +
			"Set specific origins in CORS_ALLOWED_ORIGINS environment variable.")
		os.Exit(1)
	}

	// Configure specific origins (only reached if not wildcard)
	corsConfig.AllowOrigins = s.config.Server.CORSAllowedOrigins

	// Validate at least one origin is configured
	if len(s.config.Server.CORSAllowedOrigins) == 0 {
		s.logger.Error("CORS misconfiguration: AllowCredentials requires specific AllowedOrigins. " +
			"Set CORS_ALLOWED_ORIGINS environment variable.")
		os.Exit(1)
	}

	corsConfig.AllowMethods = s.config.Server.CORSAllowedMethods

	// Ensure X-CSRF-Token is always allowed (required for CSRF protection)
	allowedHeaders := s.config.Server.CORSAllowedHeaders
	corsConfig.AllowHeaders = append(allowedHeaders, "X-CSRF-Token")

	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 5 * time.Minute
	s.engine.Use(cors.New(corsConfig))

	s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      s.engine,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	s.serveErr = make(chan error, 1)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to bind HTTP server: %w", err)
	}
	s.listener = lis

	s.logger.Info("HTTP server listening", "port", s.config.Server.Port)

	go func() {
		if err := s.server.Serve(lis); err != nil && err != http.ErrServerClosed {
			s.serveErr <- err
		}
	}()

	return nil
}

func (s *Server) ServeErr() <-chan error {
	return s.serveErr
}

func (s *Server) setupRoutes() {
	s.engine.Use(middleware.RequestID())
	s.engine.Use(middleware.Logger(s.logger))
	s.engine.Use(middleware.Recovery(s.logger))
	s.engine.Use(middleware.Metrics())

	// Health check (no auth required, support both GET and HEAD for Docker)
	s.engine.GET("/health", s.handlers.Health.Check)
	s.engine.HEAD("/health", s.handlers.Health.Check)
	s.engine.GET("/health/ready", s.handlers.Health.Ready)
	s.engine.HEAD("/health/ready", s.handlers.Health.Ready)
	s.engine.GET("/health/live", s.handlers.Health.Live)
	s.engine.HEAD("/health/live", s.handlers.Health.Live)

	// Metrics endpoint (restricted)
	s.engine.GET("/metrics", s.handlers.Metrics.Handler)

	// Swagger documentation
	if s.config.Server.Environment == "development" {
		s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	// SDK routes (/v1) - API Key authentication for SDKs
	sdk := s.engine.Group("/v1")

	// Public SDK auth routes (no authentication required)
	sdkAuth := sdk.Group("/auth")
	{
		sdkAuth.POST("/validate-key", s.handlers.Auth.ValidateAPIKeyHandler)
	}

	// Protected SDK routes (require API key authentication)
	sdk.Use(s.sdkAuthMiddleware.RequireSDKAuth())
	sdk.Use(s.rateLimitMiddleware.RateLimitByAPIKey())
	s.setupSDKRoutes(sdk)

	// Dashboard routes (/api/v1) - Bearer token authentication for dashboard
	dashboard := s.engine.Group("/api/v1")
	s.setupDashboardRoutes(dashboard)

	// WebSocket endpoint
	s.engine.GET("/ws", s.handlers.WebSocket.Handle)
}

func (s *Server) setupDashboardRoutes(router *gin.RouterGroup) {
	// Apply IP-based rate limiting to all API routes
	router.Use(s.rateLimitMiddleware.RateLimitByIP())

	// Auth routes (no auth required but rate limited)
	auth := router.Group("/auth")
	{
		auth.POST("/login", s.handlers.Auth.Login)
		auth.POST("/signup", s.handlers.Auth.Signup)
		auth.POST("/complete-oauth-signup", s.handlers.Auth.CompleteOAuthSignup)         // OAuth Step 2
		auth.POST("/exchange-session/:session_id", s.handlers.Auth.ExchangeLoginSession) // OAuth token exchange
		auth.POST("/refresh", s.handlers.Auth.RefreshToken)
		auth.POST("/forgot-password", s.handlers.Auth.ForgotPassword)
		auth.POST("/reset-password", s.handlers.Auth.ResetPassword)

		// OAuth routes (Google/GitHub signup)
		auth.GET("/google", s.handlers.Auth.InitiateGoogleOAuth)
		auth.GET("/google/callback", s.handlers.Auth.GoogleCallback)
		auth.GET("/github", s.handlers.Auth.InitiateGitHubOAuth)
		auth.GET("/github/callback", s.handlers.Auth.GitHubCallback)

		// Note: validate-key moved to SDK routes (/v1/auth/validate-key)
	}

	// Public invitation validation (no auth required, rate limited)
	router.GET("/invitations/validate/:token", s.handlers.Organization.ValidateInvitationToken)

	// Protected routes (require JWT auth + CSRF validation)
	protected := router.Group("")
	protected.Use(s.authMiddleware.RequireAuth())          // Step 1: Validate JWT from cookie
	protected.Use(s.csrfMiddleware.ValidateCSRF())         // Step 2: Validate CSRF for mutations
	protected.Use(s.rateLimitMiddleware.RateLimitByUser()) // Step 3: User-based rate limiting

	// User routes
	users := protected.Group("/users")
	{
		users.GET("/me", s.handlers.User.GetProfile)
		users.PUT("/me", s.handlers.User.UpdateProfile)
		users.PUT("/me/default-organization", s.handlers.User.SetDefaultOrganization)
	}

	// Auth session management routes (protected)
	authSessions := protected.Group("/auth")
	{
		authSessions.GET("/me", s.handlers.Auth.GetCurrentUser) // Get current user with token expiry metadata
		authSessions.POST("/logout", s.handlers.Auth.Logout)
		authSessions.GET("/sessions", s.handlers.Auth.ListSessions)
		authSessions.GET("/sessions/:session_id", s.handlers.Auth.GetSession)
		authSessions.POST("/sessions/:session_id/revoke", s.handlers.Auth.RevokeSession)
		authSessions.POST("/sessions/revoke-all", s.handlers.Auth.RevokeAllSessions)
	}

	// Organization routes with clean RBAC permissions
	orgs := protected.Group("/organizations")
	{
		orgs.GET("", s.handlers.Organization.List) // No org context required for listing user's orgs
		orgs.POST("", s.handlers.Organization.Create)
		orgs.GET("/:orgId", s.handlers.Organization.Get)
		orgs.PATCH("/:orgId", s.authMiddleware.RequirePermission("organizations:write"), s.handlers.Organization.Update)
		orgs.DELETE("/:orgId", s.authMiddleware.RequirePermission("organizations:delete"), s.handlers.Organization.Delete)
		orgs.GET("/:orgId/members", s.authMiddleware.RequirePermission("members:read"), s.handlers.Organization.ListMembers)
		orgs.POST("/:orgId/members", s.authMiddleware.RequirePermission("members:invite"), s.handlers.Organization.InviteMember)
		orgs.DELETE("/:orgId/members/:userId", s.authMiddleware.RequirePermission("members:remove"), s.handlers.Organization.RemoveMember)

		// Organization settings routes with permission middleware
		orgs.GET("/:orgId/settings", s.authMiddleware.RequirePermission("settings:read"), s.handlers.Organization.GetSettings)
		orgs.POST("/:orgId/settings", s.authMiddleware.RequirePermission("settings:write"), s.handlers.Organization.CreateSetting)
		orgs.GET("/:orgId/settings/:key", s.authMiddleware.RequirePermission("settings:read"), s.handlers.Organization.GetSetting)
		orgs.PUT("/:orgId/settings/:key", s.authMiddleware.RequirePermission("settings:write"), s.handlers.Organization.UpdateSetting)
		orgs.DELETE("/:orgId/settings/:key", s.authMiddleware.RequirePermission("settings:write"), s.handlers.Organization.DeleteSetting)
		orgs.POST("/:orgId/settings/bulk", s.authMiddleware.RequirePermission("settings:write"), s.handlers.Organization.BulkCreateSettings)
		orgs.GET("/:orgId/settings/export", s.authMiddleware.RequirePermission("settings:export"), s.handlers.Organization.ExportSettings)
		orgs.POST("/:orgId/settings/import", s.authMiddleware.RequireAllPermissions([]string{"settings:write", "settings:import"}), s.handlers.Organization.ImportSettings)
		orgs.POST("/:orgId/settings/reset", s.authMiddleware.RequireAnyPermission([]string{"settings:admin", "organizations:admin"}), s.handlers.Organization.ResetToDefaults)

		// Custom role management routes for organizations
		orgs.GET("/:orgId/roles", s.authMiddleware.RequirePermission("roles:read"), s.handlers.RBAC.GetCustomRoles)
		orgs.POST("/:orgId/roles", s.authMiddleware.RequirePermission("roles:write"), s.handlers.RBAC.CreateCustomRole)
		orgs.GET("/:orgId/roles/:roleId", s.authMiddleware.RequirePermission("roles:read"), s.handlers.RBAC.GetCustomRole)
		orgs.PUT("/:orgId/roles/:roleId", s.authMiddleware.RequirePermission("roles:write"), s.handlers.RBAC.UpdateCustomRole)
		orgs.DELETE("/:orgId/roles/:roleId", s.authMiddleware.RequirePermission("roles:delete"), s.handlers.RBAC.DeleteCustomRole)
	}

	// Project routes (top-level with optional org filtering)
	projects := protected.Group("/projects")
	{
		projects.GET("", s.handlers.Project.List) // Supports ?organization_id= filter
		projects.POST("", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Project.Create)
		projects.GET("/:projectId", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Project.Get)
		projects.PUT("/:projectId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Project.Update)
		projects.POST("/:projectId/archive", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Project.Archive)
		projects.POST("/:projectId/unarchive", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Project.Unarchive)
		projects.DELETE("/:projectId", s.authMiddleware.RequirePermission("projects:delete"), s.handlers.Project.Delete)

		// API key routes nested under projects (double-nesting only)
		projects.GET("/:projectId/api-keys", s.authMiddleware.RequirePermission("api-keys:read"), s.handlers.APIKey.List)
		projects.POST("/:projectId/api-keys", s.authMiddleware.RequirePermission("api-keys:create"), s.handlers.APIKey.Create)
		projects.DELETE("/:projectId/api-keys/:keyId", s.authMiddleware.RequirePermission("api-keys:delete"), s.handlers.APIKey.Delete)

		// Prompt Management routes nested under projects
		prompts := projects.Group("/:projectId/prompts")
		{
			// Protected labels settings (must come before :promptId routes)
			prompts.GET("/settings/protected-labels", s.authMiddleware.RequirePermission("prompts:read"), s.handlers.Prompt.GetProtectedLabels)
			prompts.PUT("/settings/protected-labels", s.authMiddleware.RequirePermission("prompts:update"), s.handlers.Prompt.SetProtectedLabels)

			// Prompt CRUD
			prompts.GET("", s.authMiddleware.RequirePermission("prompts:read"), s.handlers.Prompt.ListPrompts)
			prompts.POST("", s.authMiddleware.RequirePermission("prompts:create"), s.handlers.Prompt.CreatePrompt)
			prompts.GET("/:promptId", s.authMiddleware.RequirePermission("prompts:read"), s.handlers.Prompt.GetPrompt)
			prompts.PUT("/:promptId", s.authMiddleware.RequirePermission("prompts:update"), s.handlers.Prompt.UpdatePrompt)
			prompts.DELETE("/:promptId", s.authMiddleware.RequirePermission("prompts:delete"), s.handlers.Prompt.DeletePrompt)

			// Version management
			prompts.GET("/:promptId/versions", s.authMiddleware.RequirePermission("prompts:read"), s.handlers.Prompt.ListVersions)
			prompts.POST("/:promptId/versions", s.authMiddleware.RequirePermission("prompts:create"), s.handlers.Prompt.CreateVersion)
			prompts.GET("/:promptId/versions/:versionId", s.authMiddleware.RequirePermission("prompts:read"), s.handlers.Prompt.GetVersion)
			prompts.PATCH("/:promptId/versions/:versionId/labels", s.authMiddleware.RequirePermission("prompts:update"), s.handlers.Prompt.SetLabels)

			// Version diff
			prompts.GET("/:promptId/diff", s.authMiddleware.RequirePermission("prompts:read"), s.handlers.Prompt.GetVersionDiff)
		}

		// Playground routes (execution - global endpoints)
		playground := protected.Group("/playground")
		{
			playground.POST("/execute", s.handlers.Playground.Execute)
			playground.POST("/stream", s.handlers.Playground.Stream)
		}

		// Project-scoped playground routes (session management)
		projectPlayground := projects.Group("/:projectId/playground")
		{
			// Session management
			sessions := projectPlayground.Group("/sessions")
			{
				sessions.GET("", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Playground.ListSessions)
				sessions.POST("", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Playground.CreateSession)
				sessions.GET("/:sessionId", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Playground.GetSession)
				sessions.PUT("/:sessionId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Playground.UpdateSession)
				sessions.POST("/:sessionId/update", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Playground.UpdateSession) // For sendBeacon (POST only)
				sessions.DELETE("/:sessionId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Playground.DeleteSession)
			}
		}

		// Credentials routes (AI provider API key management)
		credentials := projects.Group("/:projectId/credentials")
		{
			// AI provider credentials - supports multiple configurations per adapter
			aiCreds := credentials.Group("/ai")
			{
				aiCreds.POST("", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Credentials.Create)
				aiCreds.POST("/test", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Credentials.TestConnection)
				aiCreds.GET("", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Credentials.List)
				aiCreds.GET("/models", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Credentials.GetAvailableModels)
				aiCreds.GET("/:credentialId", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Credentials.Get)
				aiCreds.PATCH("/:credentialId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Credentials.Update)
				aiCreds.DELETE("/:credentialId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Credentials.Delete)
			}
		}

		// Score Config routes (Evaluation domain - score definitions)
		scoreConfigs := projects.Group("/:projectId/score-configs")
		{
			scoreConfigs.GET("", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Evaluation.List)
			scoreConfigs.POST("", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Evaluation.Create)
			scoreConfigs.GET("/:configId", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Evaluation.Get)
			scoreConfigs.PUT("/:configId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Evaluation.Update)
			scoreConfigs.DELETE("/:configId", s.authMiddleware.RequirePermission("projects:write"), s.handlers.Evaluation.Delete)
		}
	}

	// Analytics routes
	analytics := protected.Group("/analytics")
	{
		analytics.GET("/overview", s.handlers.Analytics.Overview)
		analytics.GET("/requests", s.handlers.Analytics.Requests)
		analytics.GET("/costs", s.handlers.Analytics.Costs)
		analytics.GET("/providers", s.handlers.Analytics.Providers)
		analytics.GET("/models", s.handlers.Analytics.Models)
	}

	// Traces routes
	traces := protected.Group("/traces")
	{
		// Read operations
		traces.GET("", s.handlers.Observability.ListTraces)
		traces.GET("/filter-options", s.handlers.Observability.GetTraceFilterOptions) // Must be before /:id
		traces.GET("/:id", s.handlers.Observability.GetTrace)
		traces.GET("/:id/spans", s.handlers.Observability.GetTraceWithSpans)
		traces.GET("/:id/scores", s.handlers.Observability.GetTraceWithScores)
		// Delete operations (admin data cleanup)
		traces.DELETE("/:id", s.handlers.Observability.DeleteTrace)
	}

	// Spans routes
	spans := protected.Group("/spans")
	{
		// Read operations
		spans.GET("", s.handlers.Observability.ListSpans)
		spans.GET("/:id", s.handlers.Observability.GetSpan)
		// Delete operations (admin data cleanup)
		spans.DELETE("/:id", s.handlers.Observability.DeleteSpan)
	}

	// Quality Scores routes - observability data (ClickHouse)
	scores := protected.Group("/scores")
	{
		// Read operations
		scores.GET("", s.handlers.Observability.ListScores)
		scores.GET("/:id", s.handlers.Observability.GetScore)
		// Write operations (corrections/enrichment via dashboard)
		scores.PUT("/:id", s.handlers.Observability.UpdateScore)
	}

	// Logs routes
	logs := protected.Group("/logs")
	{
		logs.GET("/requests", s.handlers.Logs.ListRequests)
		logs.GET("/requests/:requestId", s.handlers.Logs.GetRequest)
		logs.GET("/export", s.handlers.Logs.Export)
	}

	// Billing routes
	billing := protected.Group("/billing")
	{
		billing.GET("/:orgId/usage", s.handlers.Billing.GetUsage)
		billing.GET("/:orgId/invoices", s.handlers.Billing.ListInvoices)
		billing.GET("/:orgId/subscription", s.handlers.Billing.GetSubscription)
		billing.POST("/:orgId/subscription", s.handlers.Billing.UpdateSubscription)
	}

	// RBAC routes (require authentication)
	rbac := protected.Group("/rbac")
	{
		// Role management
		rbac.GET("/roles", s.handlers.RBAC.ListRoles)
		rbac.POST("/roles", s.handlers.RBAC.CreateRole)
		rbac.GET("/roles/:roleId", s.handlers.RBAC.GetRole)
		rbac.PUT("/roles/:roleId", s.handlers.RBAC.UpdateRole)
		rbac.DELETE("/roles/:roleId", s.handlers.RBAC.DeleteRole)
		rbac.GET("/roles/statistics", s.handlers.RBAC.GetRoleStatistics)

		// Permission management
		rbac.GET("/permissions", s.handlers.RBAC.ListPermissions)
		rbac.POST("/permissions", s.handlers.RBAC.CreatePermission)
		rbac.GET("/permissions/:permissionId", s.handlers.RBAC.GetPermission)
		rbac.GET("/permissions/resources", s.handlers.RBAC.GetAvailableResources)
		rbac.GET("/permissions/resources/:resource/actions", s.handlers.RBAC.GetActionsForResource)

		// User role assignment
		rbac.GET("/users/:userId/organizations/:orgId/role", s.handlers.RBAC.GetUserRole)
		rbac.POST("/users/:userId/organizations/:orgId/role", s.handlers.RBAC.AssignOrganizationRole)

		// Permission checking (legacy)
		rbac.GET("/users/:userId/organizations/:orgId/permissions", s.handlers.RBAC.GetUserPermissions)
		rbac.POST("/users/:userId/organizations/:orgId/permissions/check", s.handlers.RBAC.CheckUserPermissions)

		// Scope checking
		rbac.POST("/users/:userId/scopes/check", s.handlers.RBAC.CheckUserScopes)
		rbac.GET("/users/:userId/scopes", s.handlers.RBAC.GetUserScopes)

		// Scope metadata
		rbac.GET("/scopes", s.handlers.RBAC.GetAvailableScopes)
		rbac.GET("/scopes/categories", s.handlers.RBAC.GetScopeCategories)
	}

	// Admin routes (require admin permissions)
	adminRoutes := protected.Group("/admin")
	adminRoutes.Use(s.authMiddleware.RequirePermission("admin:manage")) // Admin permission middleware
	{
		// Token management endpoints
		adminRoutes.POST("/tokens/revoke", s.handlers.Admin.RevokeToken)
		adminRoutes.POST("/users/:userID/tokens/revoke", s.handlers.Admin.RevokeUserTokens)
		adminRoutes.GET("/tokens/blacklisted", s.handlers.Admin.ListBlacklistedTokens)
		adminRoutes.GET("/tokens/stats", s.handlers.Admin.GetTokenStats)
	}
}

func (s *Server) setupSDKRoutes(router *gin.RouterGroup) {
	// OTLP (OpenTelemetry Protocol) ingestion - 100% spec compliant
	// POST /v1/traces - OTLP standard endpoint for trace ingestion
	// Supports: Protobuf + JSON formats, gzip compression
	// Compatible with: OpenTelemetry Collector, OTLP SDKs, direct integrations
	router.POST("/traces", s.handlers.OTLP.HandleTraces)

	// OTLP metrics and logs endpoints (OpenTelemetry specification)
	router.POST("/metrics", s.handlers.OTLPMetrics.HandleMetrics) // POST /v1/metrics - OTLP metrics ingestion
	router.POST("/logs", s.handlers.OTLPLogs.HandleLogs)          // POST /v1/logs - OTLP logs ingestion

	// Prompt Management SDK routes
	prompts := router.Group("/prompts")
	{
		prompts.GET("", s.handlers.Prompt.ListPromptsSDK)        // List prompts
		prompts.POST("", s.handlers.Prompt.UpsertPrompt)         // Create or update prompt (upsert)
		prompts.GET("/:name", s.handlers.Prompt.GetPromptByName) // Get prompt by name with label/version
	}

	// Score SDK routes (with optional ScoreConfig validation)
	scores := router.Group("/scores")
	{
		scores.POST("", s.handlers.SDKScore.Create)            // POST /v1/scores - Create single score
		scores.POST("/batch", s.handlers.SDKScore.CreateBatch) // POST /v1/scores/batch - Batch create scores
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
