package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	httpTransport "brokle/internal/transport/http"
)

// App represents the main application
type App struct {
	config     *config.Config
	logger     *logrus.Logger
	providers  *ProviderContainer
	httpServer *httpTransport.Server
	mode       DeploymentMode
}

// NewServer creates a new API server application (HTTP only, no workers)
func NewServer(cfg *config.Config) (*App, error) {
	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Initialize core infrastructure
	core, err := ProvideCore(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize core: %w", err)
	}

	// Create ALL services for server
	core.Services = ProvideServerServices(core)
	core.Enterprise = ProvideEnterpriseServices(cfg, logger)

	// Initialize HTTP server
	server, err := ProvideServer(core)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	return &App{
		mode:       ModeServer,
		config:     cfg,
		logger:     logger,
		httpServer: server.HTTPServer,
		providers: &ProviderContainer{
			Core:    core,
			Server:  server,
			Workers: nil, // No workers in server mode
			Mode:    ModeServer,
		},
	}, nil
}

// NewWorker creates a new worker application (background workers only, no HTTP)
func NewWorker(cfg *config.Config) (*App, error) {
	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Initialize core infrastructure
	core, err := ProvideCore(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize core: %w", err)
	}

	// Create ONLY worker services (minimal - no auth)
	core.Services = ProvideWorkerServices(core)
	core.Enterprise = nil // Worker doesn't need enterprise

	// Initialize workers
	workers, err := ProvideWorkers(core)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize workers: %w", err)
	}

	return &App{
		mode:       ModeWorker,
		config:     cfg,
		logger:     logger,
		httpServer: nil, // No HTTP server in worker mode
		providers: &ProviderContainer{
			Core:    core,
			Server:  nil, // No HTTP in worker mode
			Workers: workers,
			Mode:    ModeWorker,
		},
	}, nil
}

// Start starts the application and returns immediately without blocking on shutdown signals
func (a *App) Start() error {
	a.logger.WithField("mode", a.mode).Info("Starting Brokle Platform...")

	switch a.mode {
	case ModeServer:
		// Start HTTP server
		go func() {
			if err := a.httpServer.Start(); err != nil {
				// http.ErrServerClosed is expected during graceful shutdown
				if err != http.ErrServerClosed {
					a.logger.WithError(err).Error("HTTP server failed")
				}
			}
		}()
		a.logger.WithField("port", a.config.Server.Port).Info("HTTP server started")

		// Start gRPC OTLP server (always enabled)
		go func() {
			if err := a.providers.Server.GRPCServer.Start(); err != nil {
				a.logger.WithError(err).Error("gRPC server failed")
			}
		}()
		a.logger.WithField("port", a.config.GRPC.Port).Info("gRPC OTLP server started")

		a.logger.Info("Brokle Platform started successfully")

	case ModeWorker:
		// Start telemetry stream consumer
		if err := a.providers.Workers.TelemetryConsumer.Start(context.Background()); err != nil {
			a.logger.WithError(err).Error("Failed to start telemetry stream consumer")
			return err
		}
		a.logger.Info("Telemetry stream consumer started")
	}

	return nil
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.WithField("mode", a.mode).Info("Shutting down Brokle Platform...")

	var wg sync.WaitGroup

	switch a.mode {
	case ModeServer:
		// Shutdown gRPC server first
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.providers.Server.GRPCServer.Shutdown(ctx); err != nil {
				a.logger.WithError(err).Error("Failed to shutdown gRPC server")
			}
		}()

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

	case ModeWorker:
		// Shutdown workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			if a.providers.Workers != nil {
				if a.providers.Workers.TelemetryConsumer != nil {
					a.providers.Workers.TelemetryConsumer.Stop()
				}
			}
		}()
	}

	// Shutdown all providers (databases)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if a.providers != nil {
			if err := a.providers.Shutdown(); err != nil {
				a.logger.WithError(err).Error("Failed to shutdown providers")
			}
		}
	}()

	// Wait for all shutdowns
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

// Health returns the health status of all components using providers
func (a *App) Health() map[string]string {
	if a.providers != nil {
		return a.providers.HealthCheck()
	}

	return map[string]string{
		"status": "providers not initialized",
	}
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
	if a.providers == nil || a.providers.Core == nil {
		return nil
	}
	return a.providers.Core.Databases
}
