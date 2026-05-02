package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	"brokle/pkg/pagination"
)

type Repository interface {
	// User aggregate
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByEmailWithPassword(ctx context.Context, email string) (*User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters *ListFilters) ([]*User, int, error)

	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	SetDefaultOrganization(ctx context.Context, userID, orgID uuid.UUID) error

	// Profile sub-aggregate
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	CreateProfile(ctx context.Context, profile *UserProfile) error
	UpdateProfile(ctx context.Context, profile *UserProfile) error
}

// Filter is a generic alias kept for compatibility.
type Filter = ListFilters

// ListFilters defines filters for listing users.
type ListFilters struct {
	IsActive        *bool      `json:"is_active,omitempty"`
	IsVerified      *bool      `json:"is_verified,omitempty"`
	IsEmailVerified *bool      `json:"is_email_verified,omitempty"`
	CreatedAfter    *time.Time `json:"created_after,omitempty"`
	CreatedBefore   *time.Time `json:"created_before,omitempty"`
	LastLoginAfter  *time.Time `json:"last_login_after,omitempty"`
	Search          string     `json:"search,omitempty"`
	HasDefaultOrg   *bool      `json:"has_default_org,omitempty"`

	pagination.Params `json:",inline"`
}

// UserFilters is an alias for ListFilters for backward compatibility.
type UserFilters = ListFilters
