package analytics

import (
	"time"

	"brokle/pkg/ulid"

	"github.com/shopspring/decimal"
)

// ============================================================================
// AI Provider Pricing Entities
// ============================================================================
// Purpose: Track AI provider pricing (OpenAI, Anthropic, Google) for cost analytics
// NOT FOR: User billing - Brokle doesn't charge based on these prices
// FOR: Cost visibility - "You spent $50 with OpenAI this month"
// ============================================================================

// ProviderModel represents an AI provider's LLM model definition (OpenAI, Anthropic, Google)
// Used to track provider pricing for cost analytics, NOT for billing users
type ProviderModel struct {
	ID              ulid.ULID              `json:"id" db:"id"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	ProjectID       *ulid.ULID             `json:"project_id,omitempty" db:"project_id"`
	ModelName       string                 `json:"model_name" db:"model_name"`
	MatchPattern    string                 `json:"match_pattern" db:"match_pattern"`
	StartDate       time.Time              `json:"start_date" db:"start_date"`
	Unit            string                 `json:"unit" db:"unit"`
	TokenizerID     *string                `json:"tokenizer_id,omitempty" db:"tokenizer_id"`
	TokenizerConfig map[string]interface{} `json:"tokenizer_config,omitempty" db:"tokenizer_config"`
}

// ProviderPrice represents AI provider pricing per usage type
// Examples: OpenAI charges $2.50/1M input tokens, Anthropic charges $3.00/1M
// Supports: input, output, cache_read_input_tokens, audio_input, batch_input, etc.
type ProviderPrice struct {
	ID              ulid.ULID       `json:"id" db:"id"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	ProviderModelID ulid.ULID       `json:"provider_model_id" db:"provider_model_id"`
	ProjectID       *ulid.ULID      `json:"project_id,omitempty" db:"project_id"`
	UsageType       string          `json:"usage_type" db:"usage_type"`
	Price           decimal.Decimal `json:"price" db:"price"`
}

// ProviderPricingSnapshot represents provider pricing snapshot captured at ingestion time
// Purpose: Audit trail for "What was OpenAI's pricing on Nov 22, 2025?"
// Used for historical cost analysis and billing dispute resolution
type ProviderPricingSnapshot struct {
	ModelName    string
	Pricing      map[string]decimal.Decimal // usage_type → provider_price_per_million
	SnapshotTime time.Time
}
