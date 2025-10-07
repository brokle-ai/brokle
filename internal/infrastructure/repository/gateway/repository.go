package gateway

import (
	"gorm.io/gorm"

	"brokle/internal/core/domain/gateway"
)

// Repositories aggregates all gateway-related repositories
type Repositories struct {
	Provider       gateway.ProviderRepository
	Model          gateway.ModelRepository
	ProviderConfig gateway.ProviderConfigRepository
	// TODO: Add other repositories as they are implemented:
	// HealthMetrics  gateway.HealthMetricsRepository
	// RoutingRule    gateway.RoutingRuleRepository
	// Cache          gateway.CacheRepository
}

// NewRepositories creates a new instance of gateway repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Provider:       NewProviderRepository(db),
		Model:          NewModelRepository(db),
		ProviderConfig: NewProviderConfigRepository(db),
	}
}

// NewProviderRepositories creates provider-specific repositories
func NewProviderRepositories(db *gorm.DB) (
	gateway.ProviderRepository,
	gateway.ModelRepository,
	gateway.ProviderConfigRepository,
) {
	return NewProviderRepository(db),
		NewModelRepository(db),
		NewProviderConfigRepository(db)
}