package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerfiles "github.com/swaggo/files"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/handlers"
	"brokle/internal/transport/http/middleware"
	"github.com/redis/go-redis/v9"
)

// Server represents the HTTP server
type Server struct {
	config              *config.Config
	logger              *logrus.Logger
	server              *http.Server
	handlers            *handlers.Handlers
	engine              *gin.Engine
	authMiddleware      *middleware.AuthMiddleware
	sdkAuthMiddleware   *middleware.SDKAuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
}

// NewServer creates a new HTTP server instance
func NewServer(
	cfg *config.Config,
	logger *logrus.Logger,
	handlers *handlers.Handlers,
	jwtService auth.JWTService,
	blacklistedTokens auth.BlacklistedTokenService,
	orgMemberService auth.OrganizationMemberService,
	apiKeyService auth.APIKeyService,
	redisClient *redis.Client,
) *Server {
	// Create stateless auth middleware
	authMiddleware := middleware.NewAuthMiddleware(
		jwtService,
		blacklistedTokens,
		orgMemberService,
		logger,
	)

	// Create SDK auth middleware for API key authentication
	sdkAuthMiddleware := middleware.NewSDKAuthMiddleware(
		apiKeyService,
		logger,
	)

	// Create rate limiting middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		redisClient,
		&cfg.Auth,
		logger,
	)

	return &Server{
		config:              cfg,
		logger:              logger,
		handlers:            handlers,
		authMiddleware:      authMiddleware,
		sdkAuthMiddleware:   sdkAuthMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Setup Gin mode
	if s.config.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	s.engine = gin.New()

	// Setup CORS
	corsConfig := cors.DefaultConfig()
	
	// Handle wildcard origins gracefully
	if len(s.config.Server.CORSAllowedOrigins) == 1 && s.config.Server.CORSAllowedOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = s.config.Server.CORSAllowedOrigins
	}
	
	corsConfig.AllowMethods = s.config.Server.CORSAllowedMethods
	corsConfig.AllowHeaders = s.config.Server.CORSAllowedHeaders
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 5 * time.Minute
	s.engine.Use(cors.New(corsConfig))

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      s.engine,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		s.logger.WithField("port", s.config.Server.Port).Info("Starting HTTP server")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal
	return s.waitForShutdown()
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Global middleware
	s.engine.Use(middleware.RequestID())
	s.engine.Use(middleware.Logger(s.logger))
	s.engine.Use(middleware.Recovery(s.logger))
	s.engine.Use(middleware.Metrics())

	// Health check (no auth required)
	s.engine.GET("/health", s.handlers.Health.Check)
	s.engine.GET("/health/ready", s.handlers.Health.Ready)
	s.engine.GET("/health/live", s.handlers.Health.Live)

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

// setupDashboardRoutes configures dashboard routes (/api/v1/*)
func (s *Server) setupDashboardRoutes(router *gin.RouterGroup) {
	// Apply IP-based rate limiting to all API routes
	router.Use(s.rateLimitMiddleware.RateLimitByIP())

	// Auth routes (no auth required but rate limited)
	auth := router.Group("/auth")
	{
		auth.POST("/login", s.handlers.Auth.Login)
		auth.POST("/signup", s.handlers.Auth.Signup)
		auth.POST("/refresh", s.handlers.Auth.RefreshToken)
		auth.POST("/forgot-password", s.handlers.Auth.ForgotPassword)
		auth.POST("/reset-password", s.handlers.Auth.ResetPassword)
		// Note: validate-key moved to SDK routes (/v1/auth/validate-key)
	}

	// Protected routes (require JWT auth)
	protected := router.Group("")
	protected.Use(s.authMiddleware.RequireAuth())
	protected.Use(s.rateLimitMiddleware.RateLimitByUser()) // User-based rate limiting after auth

	// User routes
	users := protected.Group("/users")
	{
		users.GET("/me", s.handlers.User.GetProfile)
		users.PUT("/me", s.handlers.User.UpdateProfile)
	}
	
	// Onboarding routes
	onboarding := protected.Group("/onboarding")
	{
		onboarding.GET("/questions", s.handlers.User.Onboarding.GetQuestions)
		onboarding.POST("/responses", s.handlers.User.Onboarding.SubmitResponses)
		onboarding.POST("/skip/:id", s.handlers.User.Onboarding.SkipQuestion)
		onboarding.GET("/status", s.handlers.User.Onboarding.GetStatus)
	}
	
	// Auth session management routes (protected)
	authSessions := protected.Group("/auth")
	{
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
		orgs.POST("", s.authMiddleware.RequirePermission("organizations:create"), s.handlers.Organization.Create)
		orgs.GET("/:orgId", s.authMiddleware.RequirePermission("organizations:read"), s.handlers.Organization.Get)
		orgs.PUT("/:orgId", s.authMiddleware.RequirePermission("organizations:update"), s.handlers.Organization.Update)
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
		orgs.POST("/:orgId/roles", s.authMiddleware.RequirePermission("roles:create"), s.handlers.RBAC.CreateCustomRole)
		orgs.GET("/:orgId/roles/:roleId", s.authMiddleware.RequirePermission("roles:read"), s.handlers.RBAC.GetCustomRole)
		orgs.PUT("/:orgId/roles/:roleId", s.authMiddleware.RequirePermission("roles:update"), s.handlers.RBAC.UpdateCustomRole)
		orgs.DELETE("/:orgId/roles/:roleId", s.authMiddleware.RequirePermission("roles:delete"), s.handlers.RBAC.DeleteCustomRole)
	}

	// Project routes with clean RBAC permissions
	projects := protected.Group("/projects")
	{
		projects.GET("", s.handlers.Project.List) // Lists projects for authenticated user
		projects.POST("", s.authMiddleware.RequirePermission("projects:create"), s.handlers.Project.Create)
		projects.GET("/:projectId", s.authMiddleware.RequirePermission("projects:read"), s.handlers.Project.Get)
		projects.PUT("/:projectId", s.authMiddleware.RequirePermission("projects:update"), s.handlers.Project.Update)
		projects.DELETE("/:projectId", s.authMiddleware.RequirePermission("projects:delete"), s.handlers.Project.Delete)
		projects.GET("/:projectId/api-keys", s.authMiddleware.RequirePermission("api-keys:read"), s.handlers.APIKey.List)
		projects.POST("/:projectId/api-keys", s.authMiddleware.RequirePermission("api-keys:create"), s.handlers.APIKey.Create)
		projects.DELETE("/:projectId/api-keys/:keyId", s.authMiddleware.RequirePermission("api-keys:delete"), s.handlers.APIKey.Delete)
	}


	// Analytics routes
	analytics := protected.Group("/analytics")
	{
		analytics.GET("/overview", s.handlers.Analytics.Overview)
		analytics.GET("/requests", s.handlers.Analytics.Requests)
		analytics.GET("/costs", s.handlers.Analytics.Costs)
		analytics.GET("/providers", s.handlers.Analytics.Providers)
		analytics.GET("/models", s.handlers.Analytics.Models)

		// Observability analytics routes (read-only for dashboard)
		analytics.GET("/traces", s.handlers.Observability.ListTraces)
		analytics.GET("/traces/:id", s.handlers.Observability.GetTrace)
		analytics.GET("/traces/:id/observations", s.handlers.Observability.GetTraceWithObservations)
		analytics.GET("/traces/:id/stats", s.handlers.Observability.GetTraceStats)
		analytics.GET("/observations", s.handlers.Observability.ListObservations)
		analytics.GET("/observations/:id", s.handlers.Observability.GetObservation)
		analytics.GET("/quality-scores", s.handlers.Observability.ListQualityScores)
		analytics.GET("/quality-scores/:id", s.handlers.Observability.GetQualityScore)
		analytics.GET("/traces/:id/quality-scores", s.handlers.Observability.GetQualityScoresByTrace)
		analytics.GET("/observations/:id/quality-scores", s.handlers.Observability.GetQualityScoresByObservation)
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
		
		// Permission checking
		rbac.GET("/users/:userId/organizations/:orgId/permissions", s.handlers.RBAC.GetUserPermissions)
		rbac.POST("/users/:userId/organizations/:orgId/permissions/check", s.handlers.RBAC.CheckUserPermissions)
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

// setupSDKRoutes configures SDK routes (/v1/*)
func (s *Server) setupSDKRoutes(router *gin.RouterGroup) {
	// OpenAI-compatible endpoints
	router.POST("/chat/completions", s.handlers.AI.ChatCompletions)
	router.POST("/completions", s.handlers.AI.Completions)
	router.POST("/embeddings", s.handlers.AI.Embeddings)
	router.GET("/models", s.handlers.AI.ListModels)
	router.GET("/models/:model", s.handlers.AI.GetModel)

	// AI routing decisions
	router.POST("/route", s.handlers.AI.RouteRequest)

	// High-performance unified telemetry batch system with ULID-based deduplication
	telemetry := router.Group("/telemetry")
	{
		telemetry.POST("/batch", s.handlers.Observability.ProcessTelemetryBatch)     // Batch processing (main endpoint)
		telemetry.GET("/health", s.handlers.Observability.GetTelemetryHealth)       // Health monitoring
		telemetry.GET("/metrics", s.handlers.Observability.GetTelemetryMetrics)     // Performance metrics
		telemetry.GET("/performance", s.handlers.Observability.GetTelemetryPerformanceStats) // Performance stats
		telemetry.GET("/batch/:batch_id", s.handlers.Observability.GetBatchStatus)  // Batch status tracking
		telemetry.POST("/validate", s.handlers.Observability.ValidateEvent)         // Event validation
	}

	// Cache management endpoints
	cache := router.Group("/cache")
	{
		cache.GET("/status", s.handlers.AI.CacheStatus)         // Cache health
		cache.POST("/invalidate", s.handlers.AI.InvalidateCache) // Cache management
	}
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the server
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down HTTP server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Server forced to shutdown")
		return err
	}

	s.logger.Info("HTTP server stopped gracefully")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}