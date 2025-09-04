package user

import (
	"context"

	"brokle/pkg/ulid"
)

// Service defines the interface for user business logic.
// This interface encapsulates all user-related operations and business rules,
// providing a clean API for the transport layer while maintaining separation
// of concerns between business logic and data access.
type Service interface {
	// User management
	Register(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, userID ulid.ULID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByEmailWithPassword(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, userID ulid.ULID, req *UpdateUserRequest) (*User, error)
	DeactivateUser(ctx context.Context, userID ulid.ULID) error
	ReactivateUser(ctx context.Context, userID ulid.ULID) error
	DeleteUser(ctx context.Context, userID ulid.ULID) error
	
	// Profile management
	GetProfile(ctx context.Context, userID ulid.ULID) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID ulid.ULID, req *UpdateProfileRequest) (*UserProfile, error)
	
	// Preferences management
	GetPreferences(ctx context.Context, userID ulid.ULID) (*UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID ulid.ULID, req *UpdatePreferencesRequest) (*UserPreferences, error)
	
	// User listing and search
	ListUsers(ctx context.Context, filters *ListFilters) ([]*User, int, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, int, error)
	
	// User verification and activation
	VerifyEmail(ctx context.Context, userID ulid.ULID, token string) error
	MarkEmailAsVerified(ctx context.Context, userID ulid.ULID) error
	SendVerificationEmail(ctx context.Context, userID ulid.ULID) error
	
	// Password management (separate from auth domain)
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error
	
	// User activity tracking
	UpdateLastLogin(ctx context.Context, userID ulid.ULID) error
	GetUserActivity(ctx context.Context, userID ulid.ULID) (*UserActivity, error)
	
	// Organization management
	SetDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error
	GetDefaultOrganization(ctx context.Context, userID ulid.ULID) (*ulid.ULID, error)
	
	// Onboarding
	CompleteOnboarding(ctx context.Context, userID ulid.ULID) error
	IsOnboardingCompleted(ctx context.Context, userID ulid.ULID) (bool, error)
	
	// User statistics
	GetUserStats(ctx context.Context) (*UserStats, error)
	
	// Batch operations
	GetUsersByIDs(ctx context.Context, userIDs []ulid.ULID) ([]*User, error)
	GetPublicUsers(ctx context.Context, userIDs []ulid.ULID) ([]*PublicUser, error)
}

// UserActivity represents user activity and engagement metrics.
type UserActivity struct {
	UserID           ulid.ULID `json:"user_id"`
	LastLoginAt      *string   `json:"last_login_at,omitempty"`
	TotalLogins      int64     `json:"total_logins"`
	DashboardViews   int64     `json:"dashboard_views"`
	APIRequestsCount int64     `json:"api_requests_count"`
	LastAPIRequestAt *string   `json:"last_api_request_at,omitempty"`
	CreatedProjects  int64     `json:"created_projects"`
	JoinedOrgs       int64     `json:"joined_orgs"`
}

// UserStats represents aggregate user statistics.
type UserStats struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	VerifiedUsers     int64 `json:"verified_users"`
	NewUsersToday     int64 `json:"new_users_today"`
	NewUsersThisWeek  int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`
}