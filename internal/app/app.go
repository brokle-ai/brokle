package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/transport/http"
	"brokle/internal/transport/http/handlers"
)

// App represents the main application
type App struct {
	config     *config.Config
	logger     *logrus.Logger
	providers  *ProviderContainer
	httpServer *http.Server
}

// Application is an alias for App for compatibility
type Application = App

// New creates a new application instance
func New(cfg *config.Config) (*App, error) {

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return &App{
		config: cfg,
		logger: logger,
	}, nil
}

// Start starts the application and returns immediately without blocking on shutdown signals
func (a *App) Start() error {
	a.logger.Info("Starting Brokle Platform...")

	// Initialize all providers using modular DI
	providers, err := ProvideAll(a.config, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}
	a.providers = providers

	// Initialize HTTP handlers with direct service injection
	httpHandlers := handlers.NewHandlers(
		a.config,
		a.logger,
		providers.Services.Auth.Auth,                   // Auth service from modular DI
		providers.Services.Auth.APIKey,                 // API key service for authentication
		providers.Services.Auth.BlacklistedTokens,     // Blacklisted tokens service
		providers.Services.User.User,                   // User service from modular DI
		providers.Services.User.Profile,                // Profile service from modular DI
		providers.Services.User.Onboarding,             // Onboarding service from modular DI
		providers.Services.OrganizationService,         // Direct organization service
		providers.Services.MemberService,               // Direct member service
		providers.Services.ProjectService,              // Direct project service
		providers.Services.InvitationService,           // Direct invitation service
		providers.Services.SettingsService,             // Direct settings service
		providers.Services.Auth.Role,                   // Role service for RBAC
		providers.Services.Auth.Permission,             // Permission service for RBAC
		providers.Services.Auth.OrganizationMembers,    // Organization member service for normalized RBAC
		providers.Services.Observability,               // Observability service registry
		providers.Services.Gateway,                     // Gateway service for AI API endpoints
		// All enterprise services available through providers.Enterprise
	)

	// Initialize HTTP server with auth services
	a.httpServer = http.NewServer(
		a.config,
		a.logger,
		httpHandlers,
		providers.Services.Auth.JWT,
		providers.Services.Auth.BlacklistedTokens,
		providers.Services.Auth.OrganizationMembers,
		providers.Services.Auth.APIKey,
		providers.Databases.Redis.Client,
	)

	// Start HTTP server in goroutine
	go func() {
		if err := a.httpServer.Start(); err != nil {
			a.logger.WithError(err).Fatal("HTTP server failed")
		}
	}()

	a.logger.WithField("port", a.config.Server.Port).Info("Brokle Platform started successfully")

	return nil
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down Brokle Platform...")

	// Shutdown components gracefully
	var wg sync.WaitGroup

	// Shutdown HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if a.httpServer != nil {
			if err := a.httpServer.Shutdown(ctx); err != nil {
				a.logger.WithError(err).Error("Failed to shutdown HTTP server")
			}
		}
	}()

	// Shutdown all providers
	wg.Add(1)
	go func() {
		defer wg.Done()
		if a.providers != nil {
			if err := a.providers.Shutdown(); err != nil {
				a.logger.WithError(err).Error("Failed to shutdown providers")
			}
		}
	}()

	// Wait for all shutdowns to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("Brokle Platform shutdown completed")
		return nil
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout exceeded, forcing shutdown")
		return ctx.Err()
	}
}

// GetProviders returns the provider container for access to all services and dependencies
func (a *App) GetProviders() *ProviderContainer {
	return a.providers
}

// GetServices returns all services for backward compatibility
func (a *App) GetServices() *Services {
	if a.providers == nil {
		return nil
	}
	return a.providers.GetAllServices()
}

// GetRepositories returns all repositories for backward compatibility
func (a *App) GetRepositories() *Repositories {
	if a.providers == nil {
		return nil
	}
	return a.providers.GetAllRepositories()
}

// Health returns the health status of all components using providers
func (a *App) Health() map[string]string {
	if a.providers != nil {
		return a.providers.HealthCheck()
	}
	
	return map[string]string{
		"status": "providers not initialized",
	}
}

// GetGatewayServices returns the gateway services for AI API operations
func (a *App) GetGatewayServices() *GatewayServices {
	if a.providers == nil || a.providers.Services == nil {
		return nil
	}
	return a.providers.Services.Gateway
}

// GetWorkers returns the worker container for background processing
func (a *App) GetWorkers() *WorkerContainer {
	if a.providers == nil {
		return nil
	}
	return a.providers.Workers
}

// GetLogger returns the application logger
func (a *App) GetLogger() *logrus.Logger {
	return a.logger
}

// GetConfig returns the application configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}

// GetDatabases returns the database connections
func (a *App) GetDatabases() *DatabaseContainer {
	if a.providers == nil {
		return nil
	}
	return a.providers.Databases
}

// Services is an alias for GetServices for compatibility
func (a *App) Services() *Services {
	return a.GetServices()
}

// Logger is an alias for GetLogger for compatibility  
func (a *App) Logger() *logrus.Logger {
	return a.GetLogger()
}
