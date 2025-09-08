package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

// Start starts the application
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
		providers.Services.Auth.Auth,               // Auth service from modular DI
		providers.Services.Auth.BlacklistedTokens, // Blacklisted tokens service
		providers.Services.User.User,               // User service from modular DI
		providers.Services.User.Profile,            // Profile service from modular DI
		providers.Services.User.Onboarding,         // Onboarding service from modular DI
		providers.Services.OrganizationService, // Direct organization service
		providers.Services.MemberService,       // Direct member service
		providers.Services.ProjectService,      // Direct project service
		providers.Services.EnvironmentService,  // Direct environment service
		providers.Services.InvitationService,   // Direct invitation service
		providers.Services.SettingsService,     // Direct settings service
		providers.Services.Auth.Role,           // Role service for RBAC
		providers.Services.Auth.Permission,     // Permission service for RBAC
		// All enterprise services available through providers.Enterprise
	)

	// Initialize HTTP server with auth services
	a.httpServer = http.NewServer(
		a.config, 
		a.logger, 
		httpHandlers,
		providers.Services.Auth.JWT,
		providers.Services.Auth.BlacklistedTokens,
		providers.Services.Auth.Role,
		providers.Databases.Redis.Client,
	)

	// Start HTTP server in goroutine
	go func() {
		if err := a.httpServer.Start(); err != nil {
			a.logger.WithError(err).Fatal("HTTP server failed")
		}
	}()

	a.logger.WithField("port", a.config.Server.Port).Info("Brokle Platform started successfully")

	// Wait for shutdown signal
	return a.waitForShutdown()
}

// Run starts the application without waiting for shutdown (for main.go compatibility)
func (a *App) Run() error {
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
		providers.Services.Auth.Auth,               // Auth service from modular DI
		providers.Services.Auth.BlacklistedTokens, // Blacklisted tokens service
		providers.Services.User.User,               // User service from modular DI
		providers.Services.User.Profile,            // Profile service from modular DI
		providers.Services.User.Onboarding,         // Onboarding service from modular DI
		providers.Services.OrganizationService, // Direct organization service
		providers.Services.MemberService,       // Direct member service
		providers.Services.ProjectService,      // Direct project service
		providers.Services.EnvironmentService,  // Direct environment service
		providers.Services.InvitationService,   // Direct invitation service
		providers.Services.SettingsService,     // Direct settings service
		providers.Services.Auth.Role,           // Role service for RBAC
		providers.Services.Auth.Permission,     // Permission service for RBAC
		// All enterprise services available through providers.Enterprise
	)

	// Initialize HTTP server with auth services
	a.httpServer = http.NewServer(
		a.config, 
		a.logger, 
		httpHandlers,
		providers.Services.Auth.JWT,
		providers.Services.Auth.BlacklistedTokens,
		providers.Services.Auth.Role,
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

// waitForShutdown waits for shutdown signal and gracefully shuts down
func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	a.logger.Info("Shutting down Brokle Platform...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
		a.logger.Info("Brokle Platform shut down gracefully")
		return nil
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout exceeded, forcing exit")
		return ctx.Err()
	}
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
