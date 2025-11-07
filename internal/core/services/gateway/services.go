package gateway

import (
	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/gateway"
	"brokle/internal/infrastructure/providers"
	gatewayRepo "brokle/internal/infrastructure/repository/gateway"
	"gorm.io/gorm"
)

// Services aggregates all gateway-related services
type Services struct {
	Gateway gateway.GatewayService
	Routing gateway.RoutingService
	Cost    gateway.CostService
}

// Dependencies holds the dependencies required to create gateway services
type Dependencies struct {
	DB              *gorm.DB
	ProviderFactory providers.ProviderFactory
	Logger          *logrus.Logger
}

// NewServices creates a new instance of all gateway services with proper dependency injection
func NewServices(deps *Dependencies) *Services {
	// Initialize repositories
	repos := gatewayRepo.NewRepositories(deps.DB)

	// Initialize services with dependencies
	costService := NewCostService(
		repos.Model,
		repos.Provider,
		repos.ProviderConfig,
		deps.Logger,
	)

	routingService := NewRoutingService(
		repos.Provider,
		repos.Model,
		repos.ProviderConfig,
		costService,
		deps.Logger,
	)

	gatewayService := NewGatewayService(
		repos.Provider,
		repos.Model,
		repos.ProviderConfig,
		routingService,
		costService,
		deps.ProviderFactory,
		deps.Logger,
	)

	return &Services{
		Gateway: gatewayService,
		Routing: routingService,
		Cost:    costService,
	}
}

// NewGatewayServices creates gateway-specific services for dependency injection
func NewGatewayServices(
	db *gorm.DB,
	providerFactory providers.ProviderFactory,
	logger *logrus.Logger,
) (
	gateway.GatewayService,
	gateway.RoutingService,
	gateway.CostService,
) {
	deps := &Dependencies{
		DB:              db,
		ProviderFactory: providerFactory,
		Logger:          logger,
	}

	services := NewServices(deps)
	return services.Gateway, services.Routing, services.Cost
}

// ServiceConfig holds configuration for gateway services
type ServiceConfig struct {
	// Cost calculation settings
	DefaultCurrency            string
	TokenEstimationMultiplier  float64
	EnableCostOptimization     bool
	EnableProviderComparison   bool
	
	// Routing settings
	DefaultRoutingStrategy     gateway.RoutingStrategy
	EnableFallbackRouting      bool
	EnableLoadBalancing        bool
	EnableABTesting            bool
	MaxConcurrentProviders     int
	
	// Health check settings
	HealthCheckInterval        int // seconds
	ProviderTimeoutSeconds     int
	EnableHealthChecks         bool
	
	// Analytics settings
	EnableUsageTracking        bool
	EnableMetricsCollection    bool
	BatchSizeForAnalytics      int
}

// DefaultServiceConfig returns default configuration for gateway services
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		// Cost settings
		DefaultCurrency:            "USD",
		TokenEstimationMultiplier:  1.0,
		EnableCostOptimization:     true,
		EnableProviderComparison:   true,
		
		// Routing settings
		DefaultRoutingStrategy:     gateway.RoutingStrategyCostOptimized,
		EnableFallbackRouting:      true,
		EnableLoadBalancing:        true,
		EnableABTesting:            false,
		MaxConcurrentProviders:     3,
		
		// Health check settings
		HealthCheckInterval:        60, // 1 minute
		ProviderTimeoutSeconds:     30,
		EnableHealthChecks:         true,
		
		// Analytics settings
		EnableUsageTracking:        true,
		EnableMetricsCollection:    true,
		BatchSizeForAnalytics:      100,
	}
}

// ConfigurableServices creates services with custom configuration
func ConfigurableServices(deps *Dependencies, config *ServiceConfig) *Services {
	if config == nil {
		config = DefaultServiceConfig()
	}

	// Initialize repositories
	repos := gatewayRepo.NewRepositories(deps.DB)

	// Initialize services with configuration
	costService := &CostService{
		modelRepo:          repos.Model,
		providerRepo:       repos.Provider,
		providerConfigRepo: repos.ProviderConfig,
		logger:             deps.Logger,
		// TODO: Add configuration options to service structs
	}

	routingService := &RoutingService{
		providerRepo:       repos.Provider,
		modelRepo:          repos.Model,
		providerConfigRepo: repos.ProviderConfig,
		costService:        costService,
		logger:             deps.Logger,
		// TODO: Add configuration options to service structs
	}

	gatewayService := &GatewayService{
		providerRepo:       repos.Provider,
		modelRepo:          repos.Model,
		providerConfigRepo: repos.ProviderConfig,
		routingService:     routingService,
		costService:        costService,
		providerFactory:    deps.ProviderFactory,
		logger:             deps.Logger,
		// TODO: Add configuration options to service structs
	}

	return &Services{
		Gateway: gatewayService,
		Routing: routingService,
		Cost:    costService,
	}
}

// HealthChecker provides health check functionality for gateway services
type HealthChecker struct {
	services *Services
	logger   *logrus.Logger
}

// NewHealthChecker creates a new health checker for gateway services
func NewHealthChecker(services *Services, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		services: services,
		logger:   logger,
	}
}

// CheckHealth performs a comprehensive health check of all gateway services
func (h *HealthChecker) CheckHealth() map[string]interface{} {
	healthStatus := map[string]interface{}{
		"gateway_service": "healthy",
		"routing_service": "healthy", 
		"cost_service":    "healthy",
		"overall":         "healthy",
		"timestamp":       "now", // Would use actual timestamp
	}

	// TODO: Implement actual health checks for each service
	// This would typically check:
	// - Database connectivity
	// - External provider connectivity
	// - Cache connectivity
	// - Service-specific health metrics

	return healthStatus
}

// MetricsCollector collects metrics from gateway services
type MetricsCollector struct {
	services *Services
	logger   *logrus.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(services *Services, logger *logrus.Logger) *MetricsCollector {
	return &MetricsCollector{
		services: services,
		logger:   logger,
	}
}

// CollectMetrics gathers metrics from all gateway services
func (m *MetricsCollector) CollectMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"requests_processed":     0,
		"total_cost_calculated":  0.0,
		"providers_routed":       0,
		"errors_occurred":        0,
		"average_response_time":  0.0,
		"cache_hit_ratio":        0.0,
	}

	// TODO: Implement actual metrics collection
	// This would gather real-time metrics from each service

	return metrics
}

// ServiceInitializer handles initialization and startup of gateway services
type ServiceInitializer struct {
	config *ServiceConfig
	logger *logrus.Logger
}

// NewServiceInitializer creates a new service initializer
func NewServiceInitializer(config *ServiceConfig, logger *logrus.Logger) *ServiceInitializer {
	if config == nil {
		config = DefaultServiceConfig()
	}
	
	return &ServiceInitializer{
		config: config,
		logger: logger,
	}
}

// Initialize performs startup initialization for gateway services
func (si *ServiceInitializer) Initialize(services *Services) error {
	logger := si.logger.WithField("component", "service_initializer")
	logger.Info("Initializing gateway services")

	// TODO: Implement service initialization logic:
	// - Validate configuration
	// - Initialize provider connections
	// - Load initial data (models, providers, configurations)
	// - Start background workers
	// - Initialize health checks
	// - Set up metrics collection

	logger.Info("Gateway services initialized successfully")
	return nil
}

// Shutdown gracefully shuts down gateway services
func (si *ServiceInitializer) Shutdown(services *Services) error {
	logger := si.logger.WithField("component", "service_initializer")
	logger.Info("Shutting down gateway services")

	// TODO: Implement graceful shutdown logic:
	// - Stop accepting new requests
	// - Complete in-flight requests
	// - Close provider connections
	// - Stop background workers
	// - Flush metrics and logs

	logger.Info("Gateway services shutdown completed")
	return nil
}