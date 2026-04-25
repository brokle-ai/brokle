package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	GetProviderModel(ctx context.Context, modelID uuid.UUID) (*ProviderModel, error)
	GetProviderModelByName(ctx context.Context, projectID *uuid.UUID, modelName string) (*ProviderModel, error)
	GetProviderModelAtTime(ctx context.Context, projectID *uuid.UUID, modelName string, atTime time.Time) (*ProviderModel, error)
	ListProviderModels(ctx context.Context, projectID *uuid.UUID) ([]*ProviderModel, error)
	ListByProviders(ctx context.Context, providers []string) ([]*ProviderModel, error)
	UpdateProviderModel(ctx context.Context, modelID uuid.UUID, model *ProviderModel) error
	DeleteProviderModel(ctx context.Context, modelID uuid.UUID) error

	// Provider Price CRUD
	CreateProviderPrice(ctx context.Context, price *ProviderPrice) error
	GetProviderPrices(ctx context.Context, modelID uuid.UUID, projectID *uuid.UUID) ([]*ProviderPrice, error)
	UpdateProviderPrice(ctx context.Context, priceID uuid.UUID, price *ProviderPrice) error
	DeleteProviderPrice(ctx context.Context, priceID uuid.UUID) error
}
