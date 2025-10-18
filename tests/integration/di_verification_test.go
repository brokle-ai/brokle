//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/app"
	"brokle/internal/config"
)

// TestDependencyInjectionSetup verifies that all dependencies are properly initialized
func TestDependencyInjectionSetup(t *testing.T) {
	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err, "Should be able to load configuration")

	// Override configuration for testing
	cfg.Database.Database = cfg.Database.Database + "_test"
	cfg.ClickHouse.Database = cfg.ClickHouse.Database + "_test"
	cfg.Server.Port = 0 // Use random port for testing

	// Create application instance
	application, err := app.New(cfg)
	require.NoError(t, err, "Should be able to create application instance")
	require.NotNil(t, application, "Application should not be nil")

	// Verify application starts correctly
	err = application.Start()
	require.NoError(t, err, "Should be able to start application")
	defer func() {
		_ = application.Shutdown(nil)
	}()

	// Verify provider container is initialized
	providers := application.GetProviders()
	require.NotNil(t, providers, "Providers should be initialized")

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
	})

	t.Run("Repository Dependencies", func(t *testing.T) {
		// Verify repositories are initialized
		repos := providers.Repos
		require.NotNil(t, repos, "Repository container should be initialized")

		// Verify gateway repositories
		assert.NotNil(t, repos.Gateway, "Gateway repositories should exist")
		assert.NotNil(t, repos.Gateway.Provider, "Provider repository should exist")
		assert.NotNil(t, repos.Gateway.Model, "Model repository should exist")
		assert.NotNil(t, repos.Gateway.ProviderConfig, "Provider config repository should exist")
		assert.NotNil(t, repos.Gateway.Analytics, "Gateway analytics repository should exist")

		// Verify observability repositories
		assert.NotNil(t, repos.Observability, "Observability repositories should exist")
		assert.NotNil(t, repos.Observability.TelemetryAnalytics, "Telemetry analytics repository should exist")

		// Verify other repositories
		assert.NotNil(t, repos.User, "User repositories should exist")
		assert.NotNil(t, repos.Auth, "Auth repositories should exist")
		assert.NotNil(t, repos.Organization, "Organization repositories should exist")
		assert.NotNil(t, repos.Billing, "Billing repositories should exist")
	})

	t.Run("Service Dependencies", func(t *testing.T) {
		// Verify services are initialized
		services := providers.Services
		require.NotNil(t, services, "Service container should be initialized")

		// Verify gateway services
		gatewayServices := application.GetGatewayServices()
		require.NotNil(t, gatewayServices, "Gateway services should exist")
		assert.NotNil(t, gatewayServices.Gateway, "Gateway service should exist")
		assert.NotNil(t, gatewayServices.Routing, "Routing service should exist")
		assert.NotNil(t, gatewayServices.Cost, "Cost service should exist")

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

	t.Run("Worker Dependencies", func(t *testing.T) {
		// Verify workers are initialized
		workers := application.GetWorkers()
		require.NotNil(t, workers, "Worker container should be initialized")
		assert.NotNil(t, workers.TelemetryAnalytics, "Telemetry analytics worker should exist")
		assert.NotNil(t, workers.GatewayAnalytics, "Gateway analytics worker should exist")

		// Verify worker health
		health := application.Health()
		assert.Contains(t, health, "telemetry_analytics_worker", "Telemetry worker health should be reported")
		assert.Contains(t, health, "gateway_analytics_worker", "Gateway worker health should be reported")

		// Verify workers are healthy
		assert.True(t, workers.GatewayAnalytics.IsHealthy(), "Gateway analytics worker should be healthy")
	})

	t.Run("Enterprise Dependencies", func(t *testing.T) {
		// Verify enterprise services (stubs in OSS)
		enterprise := providers.Enterprise
		require.NotNil(t, enterprise, "Enterprise container should be initialized")
		assert.NotNil(t, enterprise.SSO, "SSO service should exist (stub in OSS)")
		assert.NotNil(t, enterprise.RBAC, "RBAC service should exist (stub in OSS)")
		assert.NotNil(t, enterprise.Compliance, "Compliance service should exist (stub in OSS)")
		assert.NotNil(t, enterprise.Analytics, "Analytics service should exist (stub in OSS)")
	})

	t.Run("Backward Compatibility", func(t *testing.T) {
		// Verify backward compatibility methods work
		services := application.GetServices()
		require.NotNil(t, services, "GetServices should return services")
		
		servicesAlias := application.Services()
		assert.Equal(t, services, servicesAlias, "Services() alias should return same as GetServices()")

		repositories := application.GetRepositories()
		require.NotNil(t, repositories, "GetRepositories should return repositories")

		logger := application.GetLogger()
		require.NotNil(t, logger, "GetLogger should return logger")
		
		loggerAlias := application.Logger()
		assert.Equal(t, logger, loggerAlias, "Logger() alias should return same as GetLogger()")

		config := application.GetConfig()
		require.NotNil(t, config, "GetConfig should return config")
	})

	t.Run("Gateway Service Integration", func(t *testing.T) {
		// Verify gateway service can be accessed and has proper dependencies
		gatewayServices := application.GetGatewayServices()
		require.NotNil(t, gatewayServices, "Gateway services should be accessible")

		// This tests that the gateway service was properly injected with its dependencies
		gatewayService := gatewayServices.Gateway
		require.NotNil(t, gatewayService, "Gateway service should exist")

		routingService := gatewayServices.Routing
		require.NotNil(t, routingService, "Routing service should exist")

		costService := gatewayServices.Cost
		require.NotNil(t, costService, "Cost service should exist")

		// Verify the services are properly integrated in the main services container
		mainServices := providers.Services
		assert.Equal(t, gatewayServices, mainServices.Gateway, "Gateway services should match in main container")
	})
}

// TestHealthCheckIntegration verifies the health check system works correctly
func TestHealthCheckIntegration(t *testing.T) {
	// Load configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Override configuration for testing
	cfg.Database.Database = cfg.Database.Database + "_test"
	cfg.ClickHouse.Database = cfg.ClickHouse.Database + "_test"
	cfg.Server.Port = 0

	// Create and start application
	application, err := app.New(cfg)
	require.NoError(t, err)

	err = application.Start()
	require.NoError(t, err)
	defer func() {
		_ = application.Shutdown(nil)
	}()

	// Test health check
	health := application.Health()
	require.NotNil(t, health, "Health check should return results")

	// Verify expected health components
	expectedComponents := []string{
		"postgres",
		"redis", 
		"clickhouse",
		"telemetry_analytics_worker",
		"gateway_analytics_worker",
	}

	for _, component := range expectedComponents {
		assert.Contains(t, health, component, "Health check should include %s", component)
		
		// Most components should be healthy in test environment
		if component != "clickhouse" { // ClickHouse might not always be available in test
			assert.NotContains(t, health[component], "unhealthy", "Component %s should be healthy", component)
		}
	}
}