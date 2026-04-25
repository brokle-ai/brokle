package auth

import (
	"github.com/google/uuid"

	"brokle/pkg/pagination"
)

// TokenClaims represents JWT token claims.
type TokenClaims struct {
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	Email          string     `json:"email"`
	TokenType      string     `json:"token_type"`
	Issuer         string     `json:"iss"`
	Subject        string     `json:"sub"`
	Scopes         []string   `json:"scopes,omitempty"`
	IssuedAt       int64      `json:"iat"`
	ExpiresAt      int64      `json:"exp"`
	NotBefore      int64      `json:"nbf"`
	UserID         uuid.UUID  `json:"user_id"`
}

// Request/Response DTOs
type UpdateAuthProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty" validate:"omitempty,len=2"`
}

// Filter types
type APIKeyFilters struct {
	// Domain filters
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	IsExpired      *bool      `json:"is_expired,omitempty"` // Filter by expiration status

	// Pagination (embedded for DRY)
	pagination.Params `json:",inline"`
}
