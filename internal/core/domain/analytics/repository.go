package analytics

import (
	"context"
	"time"

	"brokle/pkg/ulid"

	"github.com/shopspring/decimal"
)

// ============================================================================
// AI Provider Pricing Repositories
// ============================================================================
// Purpose: Manage AI provider pricing (OpenAI, Anthropic, Google) for cost analytics
// NOT FOR: User billing
// ============================================================================

// ProviderModelRepository handles AI provider model and pricing data access
type ProviderModelRepository interface {
	// Provider Model CRUD
	CreateProviderModel(ctx context.Context, model *ProviderModel) error
	GetProviderModel(ctx context.Context, modelID ulid.ULID) (*ProviderModel, error)
	GetProviderModelByName(ctx context.Context, projectID *ulid.ULID, modelName string) (*ProviderModel, error)
	GetProviderModelAtTime(ctx context.Context, projectID *ulid.ULID, modelName string, atTime time.Time) (*ProviderModel, error)
	ListProviderModels(ctx context.Context, projectID *ulid.ULID) ([]*ProviderModel, error)
	ListByProviders(ctx context.Context, providers []string) ([]*ProviderModel, error)
	UpdateProviderModel(ctx context.Context, modelID ulid.ULID, model *ProviderModel) error
	DeleteProviderModel(ctx context.Context, modelID ulid.ULID) error

	// Provider Price CRUD
	CreateProviderPrice(ctx context.Context, price *ProviderPrice) error
	GetProviderPrices(ctx context.Context, modelID ulid.ULID, projectID *ulid.ULID) ([]*ProviderPrice, error)
	UpdateProviderPrice(ctx context.Context, priceID ulid.ULID, price *ProviderPrice) error
	DeleteProviderPrice(ctx context.Context, priceID ulid.ULID) error
}

// ProviderPricingService handles provider pricing lookups and cost calculations
// Calculates user spending with AI providers (NOT billing users)
type ProviderPricingService interface {
	// Get provider pricing snapshot at specific time (with 5-min cache)
	// Returns OpenAI/Anthropic rates used to calculate cost visibility
	GetProviderPricingSnapshot(ctx context.Context, projectID *ulid.ULID, modelName string, atTime time.Time) (*ProviderPricingSnapshot, error)

	// Calculate provider costs from usage and pricing
	// Returns what user spent with provider (e.g., OpenAI charged them $0.005)
	CalculateProviderCost(usage map[string]uint64, pricing *ProviderPricingSnapshot) map[string]decimal.Decimal
}
