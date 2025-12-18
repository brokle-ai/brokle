package credentials

import (
	"time"

	"brokle/pkg/ulid"
)

// LLMProvider represents supported LLM providers
type LLMProvider string

const (
	ProviderOpenAI    LLMProvider = "openai"
	ProviderAnthropic LLMProvider = "anthropic"
)

// ValidProviders returns all valid provider values
func ValidProviders() []LLMProvider {
	return []LLMProvider{ProviderOpenAI, ProviderAnthropic}
}

// IsValid checks if the provider is a valid value
func (p LLMProvider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (p LLMProvider) String() string {
	return string(p)
}

// LLMProviderCredential represents an encrypted LLM API key for a project.
// The actual API key is stored encrypted with AES-256-GCM and is never
// returned to the frontend.
type LLMProviderCredential struct {
	// Primary key
	ID ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`

	// Project scope (one key per provider per project)
	ProjectID ulid.ULID `json:"project_id" gorm:"type:char(26);not null;index"`

	// Provider type (openai, anthropic)
	Provider LLMProvider `json:"provider" gorm:"type:llm_provider;not null"`

	// Encrypted API key (AES-256-GCM: nonce + ciphertext + tag, base64 encoded)
	// Never serialized to JSON, never returned to frontend
	EncryptedKey string `json:"-" gorm:"column:encrypted_key;not null"`

	// Masked preview for safe display (e.g., "sk-***abcd")
	KeyPreview string `json:"key_preview" gorm:"size:20;not null"`

	// Optional custom base URL for Azure OpenAI, proxies, etc.
	BaseURL *string `json:"base_url,omitempty" gorm:"type:text"`

	// Audit fields
	CreatedBy *ulid.ULID `json:"created_by,omitempty" gorm:"type:char(26)"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName returns the database table name
func (LLMProviderCredential) TableName() string {
	return "llm_provider_credentials"
}

// LLMProviderCredentialResponse is the safe response DTO (no encrypted data).
// This is what gets returned to the frontend.
type LLMProviderCredentialResponse struct {
	ID         ulid.ULID   `json:"id"`
	ProjectID  ulid.ULID   `json:"project_id"`
	Provider   LLMProvider `json:"provider"`
	KeyPreview string      `json:"key_preview"` // Masked: "sk-***abcd"
	BaseURL    *string     `json:"base_url,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// DecryptedKeyConfig holds the decrypted key configuration for LLM execution.
// This is ONLY used internally during prompt execution and is NEVER persisted
// or returned via API.
type DecryptedKeyConfig struct {
	Provider LLMProvider
	APIKey   string // Decrypted API key - handle with care
	BaseURL  string // Custom base URL (if configured)
}

// ToResponse converts an entity to a safe response DTO
func (c *LLMProviderCredential) ToResponse() *LLMProviderCredentialResponse {
	return &LLMProviderCredentialResponse{
		ID:         c.ID,
		ProjectID:  c.ProjectID,
		Provider:   c.Provider,
		KeyPreview: c.KeyPreview,
		BaseURL:    c.BaseURL,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// MaskAPIKey creates a masked preview of an API key.
// Format: first 3 chars + "***" + last 4 chars (e.g., "sk-***abcd")
// For short keys (< 8 chars), returns "***" for security.
func MaskAPIKey(key string) string {
	if len(key) < 8 {
		return "***"
	}
	return key[:3] + "***" + key[len(key)-4:]
}

// DetectProvider attempts to detect the provider from a model name.
// Returns empty string if unknown.
func DetectProvider(model string) LLMProvider {
	// OpenAI models
	openAIPrefixes := []string{"gpt-", "o1", "text-", "davinci", "curie", "babbage", "ada"}
	for _, prefix := range openAIPrefixes {
		if len(model) >= len(prefix) && model[:len(prefix)] == prefix {
			return ProviderOpenAI
		}
	}

	// Anthropic models
	anthropicPrefixes := []string{"claude-"}
	for _, prefix := range anthropicPrefixes {
		if len(model) >= len(prefix) && model[:len(prefix)] == prefix {
			return ProviderAnthropic
		}
	}

	return ""
}
