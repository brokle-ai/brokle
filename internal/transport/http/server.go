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
	config             *config.Config
	logger             *logrus.Logger
	server             *http.Server
	handlers           *handlers.Handlers
	engine             *gin.Engine
	authMiddleware     *middleware.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
}

// NewServer creates a new HTTP server instance
func NewServer(
	cfg *config.Config, 
	logger *logrus.Logger, 
	handlers *handlers.Handlers,
	jwtService auth.JWTService,
	blacklistedTokens auth.BlacklistedTokenService,
	redisClient *redis.Client,
) *Server {
	// Create stateless auth middleware
	authMiddleware := middleware.NewAuthMiddleware(
		jwtService,
		blacklistedTokens,
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

	// API v1 routes
	v1 := s.engine.Group("/api/v1")
	s.setupV1Routes(v1)

	// WebSocket endpoint
	s.engine.GET("/ws", s.handlers.WebSocket.Handle)

	// OpenAI-compatible routes (with auth)
	openai := s.engine.Group("/v1")
	openai.Use(middleware.APIKeyAuth(s.handlers.Auth))
	s.setupOpenAIRoutes(openai)
}

// setupV1Routes configures API v1 routes
func (s *Server) setupV1Routes(router *gin.RouterGroup) {
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
		users.GET("/me/preferences", s.handlers.User.GetPreferences)
		users.PUT("/me/preferences", s.handlers.User.UpdatePreferences)
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

	// Organization routes
	orgs := protected.Group("/organizations")
	{
		orgs.GET("", s.handlers.Organization.List)
		orgs.POST("", s.handlers.Organization.Create)
		orgs.GET("/:orgId", s.handlers.Organization.Get)
		orgs.PUT("/:orgId", s.handlers.Organization.Update)
		orgs.DELETE("/:orgId", s.handlers.Organization.Delete)
		orgs.GET("/:orgId/members", s.handlers.Organization.ListMembers)
		orgs.POST("/:orgId/members", s.handlers.Organization.InviteMember)
		orgs.DELETE("/:orgId/members/:userId", s.handlers.Organization.RemoveMember)
		
		// Organization settings routes
		orgs.GET("/:orgId/settings", s.handlers.Organization.GetSettings)
		orgs.POST("/:orgId/settings", s.handlers.Organization.CreateSetting)
		orgs.GET("/:orgId/settings/:key", s.handlers.Organization.GetSetting)
		orgs.PUT("/:orgId/settings/:key", s.handlers.Organization.UpdateSetting)
		orgs.DELETE("/:orgId/settings/:key", s.handlers.Organization.DeleteSetting)
		orgs.POST("/:orgId/settings/bulk", s.handlers.Organization.BulkCreateSettings)
		orgs.GET("/:orgId/settings/export", s.handlers.Organization.ExportSettings)
		orgs.POST("/:orgId/settings/import", s.handlers.Organization.ImportSettings)
		orgs.POST("/:orgId/settings/reset", s.handlers.Organization.ResetToDefaults)
	}

	// Project routes
	projects := protected.Group("/projects")
	{
		projects.GET("", s.handlers.Project.List)
		projects.POST("", s.handlers.Project.Create)
		projects.GET("/:projectId", s.handlers.Project.Get)
		projects.PUT("/:projectId", s.handlers.Project.Update)
		projects.DELETE("/:projectId", s.handlers.Project.Delete)
		projects.GET("/:projectId/environments", s.handlers.Environment.List)
	}

	// Environment routes
	envs := protected.Group("/environments")
	{
		envs.POST("", s.handlers.Environment.Create)
		envs.GET("/:envId", s.handlers.Environment.Get)
		envs.PUT("/:envId", s.handlers.Environment.Update)
		envs.DELETE("/:envId", s.handlers.Environment.Delete)
		envs.GET("/:envId/api-keys", s.handlers.APIKey.List)
		envs.POST("/:envId/api-keys", s.handlers.APIKey.Create)
		envs.DELETE("/:envId/api-keys/:keyId", s.handlers.APIKey.Delete)
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

	// Admin routes (require admin role)
	adminRoutes := protected.Group("/admin")
	adminRoutes.Use(middleware.RequireRole("admin")) // Admin role middleware
	{
		// Token management endpoints
		adminRoutes.POST("/tokens/revoke", s.handlers.Admin.RevokeToken)
		adminRoutes.POST("/users/:userID/tokens/revoke", s.handlers.Admin.RevokeUserTokens)
		adminRoutes.GET("/tokens/blacklisted", s.handlers.Admin.ListBlacklistedTokens)
		adminRoutes.GET("/tokens/stats", s.handlers.Admin.GetTokenStats)
	}
}

// setupOpenAIRoutes configures OpenAI-compatible routes
func (s *Server) setupOpenAIRoutes(router *gin.RouterGroup) {
	// Apply API key-based rate limiting for OpenAI routes
	router.Use(s.rateLimitMiddleware.RateLimitByAPIKey())

	// Chat completions
	router.POST("/chat/completions", s.handlers.AI.ChatCompletions)
	
	// Completions
	router.POST("/completions", s.handlers.AI.Completions)
	
	// Embeddings
	router.POST("/embeddings", s.handlers.AI.Embeddings)
	
	// Models
	router.GET("/models", s.handlers.AI.ListModels)
	router.GET("/models/:model", s.handlers.AI.GetModel)
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