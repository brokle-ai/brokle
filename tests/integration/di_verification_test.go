//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/app"
	"brokle/internal/config"
)

// TestDIContainer_ServerMode verifies server-only mode dependencies
func TestDIContainer_ServerMode(t *testing.T) {
	// Set server mode for config validation
	os.Setenv("APP_MODE", "server")
	defer os.Unsetenv("APP_MODE")

	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err, "Should be able to load configuration")

	// Override configuration for testing
	cfg.Database.Database = cfg.Database.Database + "_test"
	cfg.ClickHouse.Database = cfg.ClickHouse.Database + "_test"
	cfg.Server.Port = 0 // Use random port for testing

	// Create server application (HTTP only, no workers)
	application, err := app.NewServer(cfg)
	require.NoError(t, err, "Should be able to create server application")
	require.NotNil(t, application, "Application should not be nil")

	// Verify application starts correctly
	err = application.Start()
	require.NoError(t, err, "Should be able to start server")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = application.Shutdown(ctx)
	}()

	// Verify provider container is initialized
	providers := application.GetProviders()
	require.NotNil(t, providers, "Providers should be initialized")

	t.Run("Server Mode Verification", func(t *testing.T) {
		// Verify mode
		require.Equal(t, app.ModeServer, providers.Mode, "Should be in server mode")

		// Verify core infrastructure
		require.NotNil(t, providers.Core, "Core should be initialized")
		require.NotNil(t, providers.Core.Databases, "Core databases should exist")
		require.NotNil(t, providers.Core.Services, "Core services should exist")

		// Verify HTTP server exists
		require.NotNil(t, providers.Server, "Server should be initialized")
		require.NotNil(t, providers.Server.HTTPServer, "HTTP server should exist")

		// Verify workers are nil in server mode
		require.Nil(t, providers.Workers, "Workers should be nil in server mode")
	})

	t.Run("Database Dependencies", func(t *testing.T) {
		// Verify database connections
		databases := application.GetDatabases()
		require.NotNil(t, databases, "Databases should be initialized")
		assert.NotNil(t, databases.Postgres, "PostgreSQL connection should exist")
		assert.NotNil(t, databases.Redis, "Redis connection should exist")
		assert.NotNil(t, databases.ClickHouse, "ClickHouse connection should exist")

		// Verify database health
		health := application.Health()
		assert.Contains(t, health, "postgres", "Postgres health should be reported")
		assert.Contains(t, health, "redis", "Redis health should be reported")
		assert.Contains(t, health, "clickhouse", "ClickHouse health should be reported")
		assert.Contains(t, health, "mode", "Mode should be reported")
		assert.Equal(t, "server", health["mode"], "Mode should be server")
	})

	t.Run("Repository Dependencies", func(t *testing.T) {
		// Verify repositories are initialized
		repos := providers.Core.Repos
		require.NotNil(t, repos, "Repository container should be initialized")

		// Verify observability repositories
		assert.NotNil(t, repos.Observability, "Observability repositories should exist")

		// Verify other repositories
		assert.NotNil(t, repos.User, "User repositories should exist")
		assert.NotNil(t, repos.Auth, "Auth repositories should exist")
		assert.NotNil(t, repos.Organization, "Organization repositories should exist")
		assert.NotNil(t, repos.Billing, "Billing repositories should exist")
	})

	t.Run("Service Dependencies", func(t *testing.T) {
		// Verify services are initialized
		services := providers.Core.Services
		require.NotNil(t, services, "Service container should be initialized")

		// Verify other services
		assert.NotNil(t, services.User, "User services should exist")
		assert.NotNil(t, services.Auth, "Auth services should exist")
		assert.NotNil(t, services.OrganizationService, "Organization service should exist")
		assert.NotNil(t, services.Observability, "Observability service should exist")
		assert.NotNil(t, services.Billing, "Billing services should exist")

		// Verify auth services have all required components
		assert.NotNil(t, services.Auth.Auth, "Auth service should exist")
		assert.NotNil(t, services.Auth.JWT, "JWT service should exist")
		assert.NotNil(t, services.Auth.APIKey, "API Key service should exist")
		assert.NotNil(t, services.Auth.Role, "Role service should exist")
		assert.NotNil(t, services.Auth.Permission, "Permission service should exist")
	})

	t.Run("Server Has All Services - Complete ServiceContainer", func(t *testing.T) {
		services := providers.Core.Services
		require.NotNil(t, services, "Services should exist")

		// Server mode should have ALL services (no nils)
		assert.NotNil(t, services.Auth, "Server needs Auth services")
		assert.NotNil(t, services.User, "Server needs User services")
		assert.NotNil(t, services.OrganizationService, "Server needs Org service")
		assert.NotNil(t, services.MemberService, "Server needs Member service")
		assert.NotNil(t, services.ProjectService, "Server needs Project service")
		assert.NotNil(t, services.InvitationService, "Server needs Invitation service")
		assert.NotNil(t, services.SettingsService, "Server needs Settings service")
		assert.NotNil(t, services.Observability, "Server needs Observability")
		assert.NotNil(t, services.Billing, "Server needs Billing")

		// Server should have enterprise services
		assert.NotNil(t, providers.Core.Enterprise, "Server should have Enterprise services")
	})
}

// TestDIContainer_WorkerMode verifies worker-only mode dependencies
func TestDIContainer_WorkerMode(t *testing.T) {
	// Set worker mode for config validation
	os.Setenv("APP_MODE", "worker")
	defer os.Unsetenv("APP_MODE")

	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err, "Should be able to load configuration")

	// Override configuration for testing
	cfg.Database.Database = cfg.Database.Database + "_test"
	cfg.ClickHouse.Database = cfg.ClickHouse.Database + "_test"

	// Create worker application (workers only, no HTTP)
	application, err := app.NewWorker(cfg)
	require.NoError(t, err, "Should be able to create worker application")
	require.NotNil(t, application, "Application should not be nil")

	// Verify application starts correctly
	err = application.Start()
	require.NoError(t, err, "Should be able to start workers")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = application.Shutdown(ctx)
	}()

	// Verify provider container is initialized
	providers := application.GetProviders()
	require.NotNil(t, providers, "Providers should be initialized")

	t.Run("Worker Mode Verification", func(t *testing.T) {
		// Verify mode
		require.Equal(t, app.ModeWorker, providers.Mode, "Should be in worker mode")

		// Verify core infrastructure
		require.NotNil(t, providers.Core, "Core should be initialized")
		require.NotNil(t, providers.Core.Databases, "Core databases should exist")
		require.NotNil(t, providers.Core.Services, "Core services should exist")

		// Verify workers exist
		require.NotNil(t, providers.Workers, "Workers should be initialized")
		require.NotNil(t, providers.Workers.TelemetryConsumer, "Telemetry consumer should exist")

		// Verify HTTP server is nil in worker mode
		require.Nil(t, providers.Server, "Server should be nil in worker mode")
	})

	t.Run("Worker Health", func(t *testing.T) {
		// Verify worker health reporting
		health := application.Health()
		assert.Contains(t, health, "mode", "Mode should be reported")
		assert.Equal(t, "worker", health["mode"], "Mode should be worker")

		// Verify workers are healthy
		workers := application.GetWorkers()
		require.NotNil(t, workers, "Workers should be accessible")
	})

	t.Run("Shared Core Infrastructure", func(t *testing.T) {
		// Verify workers share the same core infrastructure
		databases := application.GetDatabases()
		require.NotNil(t, databases, "Databases should be accessible")
		assert.NotNil(t, databases.Postgres, "PostgreSQL connection should exist")
		assert.NotNil(t, databases.Redis, "Redis connection should exist")
		assert.NotNil(t, databases.ClickHouse, "ClickHouse connection should exist")

		// Verify services are accessible (workers use these for processing)
		services := providers.Core.Services
		require.NotNil(t, services, "Services should be accessible")
		assert.NotNil(t, services.Observability, "Observability service should exist")
		assert.NotNil(t, services.Billing, "Billing service should exist")
	})

	t.Run("Worker Service Separation - No Auth Services", func(t *testing.T) {
		services := providers.Core.Services
		require.NotNil(t, services, "Services should exist")

		// Worker should NOT have auth/user/org services (clean separation)
		assert.Nil(t, services.Auth, "Worker should not initialize Auth services")
		assert.Nil(t, services.User, "Worker should not initialize User services")
		assert.Nil(t, services.OrganizationService, "Worker should not initialize Org service")
		assert.Nil(t, services.MemberService, "Worker should not initialize Member service")
		assert.Nil(t, services.ProjectService, "Worker should not initialize Project service")
		assert.Nil(t, services.InvitationService, "Worker should not initialize Invitation service")
		assert.Nil(t, services.SettingsService, "Worker should not initialize Settings service")

		// Worker SHOULD have processing services
		assert.NotNil(t, services.Observability, "Worker needs Observability services")
		assert.NotNil(t, services.Billing, "Worker needs Billing services for analytics")

		// Worker should not have enterprise services
		assert.Nil(t, providers.Core.Enterprise, "Worker should not have Enterprise services")
	})
}

// TestHealthCheckIntegration_ServerMode verifies server health check
func TestHealthCheckIntegration_ServerMode(t *testing.T) {
	os.Setenv("APP_MODE", "server")
	defer os.Unsetenv("APP_MODE")

	cfg, err := config.Load()
	require.NoError(t, err)

	cfg.Database.Database = cfg.Database.Database + "_test"
	cfg.ClickHouse.Database = cfg.ClickHouse.Database + "_test"
	cfg.Server.Port = 0

	// Create and start server
	application, err := app.NewServer(cfg)
	require.NoError(t, err)

	err = application.Start()
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = application.Shutdown(ctx)
	}()

	// Test health check
	health := application.Health()
	require.NotNil(t, health, "Health check should return results")

	// Verify expected health components (server mode)
	expectedComponents := []string{
		"postgres",
		"redis",
		"clickhouse",
		"mode",
	}

	for _, component := range expectedComponents {
		assert.Contains(t, health, component, "Health check should include %s", component)
	}

	// Verify mode
	assert.Equal(t, "server", health["mode"], "Mode should be server")

	// Worker health should NOT be present in server mode
	assert.NotContains(t, health, "telemetry_stream_consumer", "Worker health should not be in server mode")
}

// TestHealthCheckIntegration_WorkerMode verifies worker health check
func TestHealthCheckIntegration_WorkerMode(t *testing.T) {
	os.Setenv("APP_MODE", "worker")
	defer os.Unsetenv("APP_MODE")

	cfg, err := config.Load()
	require.NoError(t, err)

	cfg.Database.Database = cfg.Database.Database + "_test"
	cfg.ClickHouse.Database = cfg.ClickHouse.Database + "_test"

	// Create and start worker
	application, err := app.NewWorker(cfg)
	require.NoError(t, err)

	err = application.Start()
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = application.Shutdown(ctx)
	}()

	// Test health check
	health := application.Health()
	require.NotNil(t, health, "Health check should return results")

	// Verify expected health components (worker mode)
	expectedComponents := []string{
		"postgres",
		"redis",
		"clickhouse",
		"mode",
	}

	for _, component := range expectedComponents {
		assert.Contains(t, health, component, "Health check should include %s", component)
	}

	// Verify mode
	assert.Equal(t, "worker", health["mode"], "Mode should be worker")
}
