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
	ID              ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
	CreatedAt       time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	ProjectID       *ulid.ULID             `json:"project_id,omitempty" gorm:"type:char(26)"`
	ModelName       string                 `json:"model_name" gorm:"column:model_name;size:255;not null"`
	MatchPattern    string                 `json:"match_pattern" gorm:"column:match_pattern;size:500;not null"`
	StartDate       time.Time              `json:"start_date" gorm:"not null;default:now()"`
	Unit            string                 `json:"unit" gorm:"size:50;not null;default:'TOKENS'"`
	TokenizerID     *string                `json:"tokenizer_id,omitempty" gorm:"size:100"`
	TokenizerConfig map[string]interface{} `json:"tokenizer_config,omitempty" gorm:"type:jsonb;serializer:json"`
}

// TableName returns the table name for GORM
func (ProviderModel) TableName() string { return "provider_models" }

// ProviderPrice represents AI provider pricing per usage type
// Examples: OpenAI charges $2.50/1M input tokens, Anthropic charges $3.00/1M
// Supports: input, output, cache_read_input_tokens, audio_input, batch_input, etc.
type ProviderPrice struct {
	ID              ulid.ULID       `json:"id" gorm:"type:char(26);primaryKey"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	ProviderModelID ulid.ULID       `json:"provider_model_id" gorm:"type:char(26);not null"`
	ProjectID       *ulid.ULID      `json:"project_id,omitempty" gorm:"type:char(26)"`
	UsageType       string          `json:"usage_type" gorm:"size:100;not null"`
	Price           decimal.Decimal `json:"price" gorm:"type:decimal(20,12);not null"`
}

// TableName returns the table name for GORM
func (ProviderPrice) TableName() string { return "provider_prices" }

// ProviderPricingSnapshot represents provider pricing snapshot captured at ingestion time
// Purpose: Audit trail for "What was OpenAI's pricing on Nov 22, 2025?"
// Used for historical cost analysis and billing dispute resolution
type ProviderPricingSnapshot struct {
	ModelName    string
	Pricing      map[string]decimal.Decimal // usage_type → provider_price_per_million
	SnapshotTime time.Time
}
