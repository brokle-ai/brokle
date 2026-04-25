package apikey

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/response"
)

// createAPIKeyBody — POST /api/v1/projects/{projectId}/api-keys
type createAPIKeyBody struct {
	Name         string `json:"name"          validate:"required,min=2,max=100"`
	ExpiryOption string `json:"expiry_option" validate:"required,oneof=30days 90days never"`
}

// apiKey is the wire shape returned by list/create.
type apiKey struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty"`
	KeyPreview string     `json:"key_preview"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Status     string     `json:"status"`
	LastUsed   *time.Time `json:"last_used,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedBy  uuid.UUID  `json:"created_by"`
}

// listAPIKeysResponse — GET /api/v1/projects/{projectId}/api-keys
type listAPIKeysResponse struct {
	Data       []apiKey             `json:"data"`
	Pagination *response.Pagination `json:"pagination"`
}
