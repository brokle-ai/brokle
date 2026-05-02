package auth

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/uid"
)

// APIKey represents an industry-standard API key with secure hash
// storage. Format: bk_{40_char_random}. Security: full key is hashed
// with SHA-256 (deterministic, enables O(1) lookup). Organization is
// derived via projects.organization_id (no redundant storage). Status
// is determined by deleted_at (soft delete) and expires_at.
type APIKey struct {
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	KeyHash    string     `json:"-"`
	KeyPreview string     `json:"key_preview"`
	Name       string     `json:"name"`
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	UserID     uuid.UUID  `json:"user_id"`
}

// IsExpired reports whether the key has passed its expires_at.
func (k *APIKey) IsExpired() bool {
	return k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt)
}

// IsValid reports whether the key is currently usable. Soft-deleted
// keys are filtered at the repository layer; this only checks expiry.
func (k *APIKey) IsValid() bool {
	return !k.IsExpired()
}

// MarkAsUsed updates the last-used timestamp.
func (k *APIKey) MarkAsUsed() {
	now := time.Now()
	k.LastUsedAt = &now
	k.UpdatedAt = now
}

// NewAPIKey constructs an API key record. The caller computes the
// SHA-256 hash + key preview before invoking this; the plaintext is
// returned to the user once and never persisted.
func NewAPIKey(userID, projectID uuid.UUID, name, keyHash, keyPreview string, expiresAt *time.Time) *APIKey {
	return &APIKey{
		ID:         uid.New(),
		KeyHash:    keyHash,
		KeyPreview: keyPreview,
		ProjectID:  projectID,
		UserID:     userID,
		Name:       name,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// ----- Wire DTOs for the API key feature ----------------------------

// CreateAPIKeyRequest is the body of POST /api/v1/api-keys.
type CreateAPIKeyRequest struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Name      string     `json:"name" validate:"required,min=1,max=100"`
	ProjectID uuid.UUID  `json:"project_id" validate:"required"`
}

// CreateAPIKeyResponse is the success body of POST /api/v1/api-keys.
// The plaintext Key is shown once at creation time and never returned
// on subsequent reads — clients must store it themselves.
type CreateAPIKeyResponse struct {
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name" example:"Production API Key"`
	Key        string     `json:"key" example:"bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd"`
	KeyPreview string     `json:"key_preview" example:"bk_AbCd...AbCd"`
	ProjectID  uuid.UUID  `json:"project_id"`
}

// ValidateAPIKeyResponse represents the result of API-key validation
// at the SDK-auth middleware boundary — Valid is the gate, the rest
// is hydrated context for downstream handlers.
type ValidateAPIKeyResponse struct {
	APIKey         *APIKey      `json:"api_key"`
	AuthContext    *AuthContext `json:"auth_context,omitempty"`
	ProjectID      uuid.UUID    `json:"project_id"`
	OrganizationID uuid.UUID    `json:"organization_id"`
	Valid          bool         `json:"valid"`
}
