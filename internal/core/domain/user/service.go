package user

import (
	"context"

	"brokle/pkg/ulid"
)

// UserService defines the interface for core user management operations.
type UserService interface {
	// User lifecycle management
	GetUser(ctx context.Context, userID ulid.ULID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByEmailWithPassword(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, userID ulid.ULID, req *UpdateUserRequest) (*User, error)
	DeactivateUser(ctx context.Context, userID ulid.ULID) error
	ReactivateUser(ctx context.Context, userID ulid.ULID) error
	DeleteUser(ctx context.Context, userID ulid.ULID) error
	
	// User listing and search
	ListUsers(ctx context.Context, filters *ListFilters) ([]*User, int, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, int, error)
	GetUsersByIDs(ctx context.Context, userIDs []ulid.ULID) ([]*User, error)
	GetPublicUsers(ctx context.Context, userIDs []ulid.ULID) ([]*PublicUser, error)
	
	// Email verification
	VerifyEmail(ctx context.Context, userID ulid.ULID, token string) error
	MarkEmailAsVerified(ctx context.Context, userID ulid.ULID) error
	SendVerificationEmail(ctx context.Context, userID ulid.ULID) error
	
	// Password management
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error
	
	// Activity tracking
	UpdateLastLogin(ctx context.Context, userID ulid.ULID) error
	GetUserActivity(ctx context.Context, userID ulid.ULID) (*UserActivity, error)
	
	// Organization context
	SetDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error
	GetDefaultOrganization(ctx context.Context, userID ulid.ULID) (*ulid.ULID, error)
	
	// User statistics
	GetUserStats(ctx context.Context) (*UserStats, error)
}

// ProfileService defines the interface for user profile management.
type ProfileService interface {
	// Profile management
	GetProfile(ctx context.Context, userID ulid.ULID) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID ulid.ULID, req *UpdateProfileRequest) (*UserProfile, error)
	UploadAvatar(ctx context.Context, userID ulid.ULID, imageData []byte, contentType string) (*UserProfile, error)
	RemoveAvatar(ctx context.Context, userID ulid.ULID) error
	
	// Profile visibility and privacy
	UpdateProfileVisibility(ctx context.Context, userID ulid.ULID, visibility ProfileVisibility) error
	GetPublicProfile(ctx context.Context, userID ulid.ULID) (*PublicProfile, error)
	
	// Profile completeness and validation
	GetProfileCompleteness(ctx context.Context, userID ulid.ULID) (*ProfileCompleteness, error)
	ValidateProfile(ctx context.Context, userID ulid.ULID) (*ProfileValidation, error)
}

// PreferenceService defines the interface for user preferences and settings management.
type PreferenceService interface {
	// General preferences
	GetPreferences(ctx context.Context, userID ulid.ULID) (*UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID ulid.ULID, req *UpdatePreferencesRequest) (*UserPreferences, error)
	ResetPreferences(ctx context.Context, userID ulid.ULID) (*UserPreferences, error)
	
	// Notification preferences
	GetNotificationPreferences(ctx context.Context, userID ulid.ULID) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID ulid.ULID, req *UpdateNotificationPreferencesRequest) (*NotificationPreferences, error)
	
	// Theme and UI preferences
	GetThemePreferences(ctx context.Context, userID ulid.ULID) (*ThemePreferences, error)
	UpdateThemePreferences(ctx context.Context, userID ulid.ULID, req *UpdateThemePreferencesRequest) (*ThemePreferences, error)
	
	// Privacy preferences
	GetPrivacyPreferences(ctx context.Context, userID ulid.ULID) (*PrivacyPreferences, error)
	UpdatePrivacyPreferences(ctx context.Context, userID ulid.ULID, req *UpdatePrivacyPreferencesRequest) (*PrivacyPreferences, error)
}

// OnboardingService defines the interface for user onboarding and initial setup.
type OnboardingService interface {
	// Onboarding flow
	GetOnboardingStatus(ctx context.Context, userID ulid.ULID) (*OnboardingStatus, error)
	CompleteOnboardingStep(ctx context.Context, userID ulid.ULID, step OnboardingStep) error
	CompleteOnboarding(ctx context.Context, userID ulid.ULID) error
	IsOnboardingCompleted(ctx context.Context, userID ulid.ULID) (bool, error)
	RestartOnboarding(ctx context.Context, userID ulid.ULID) error
	
	// Onboarding customization
	GetOnboardingFlow(ctx context.Context, userType UserType) (*OnboardingFlow, error)
	UpdateOnboardingPreferences(ctx context.Context, userID ulid.ULID, req *UpdateOnboardingPreferencesRequest) error
}

// Supporting types for the new service interfaces

// ProfileVisibility represents profile visibility options
type ProfileVisibility string

const (
	ProfileVisibilityPublic   ProfileVisibility = "public"
	ProfileVisibilityPrivate  ProfileVisibility = "private"
	ProfileVisibilityFriends  ProfileVisibility = "friends"
	ProfileVisibilityTeam     ProfileVisibility = "team"
)

// PublicProfile represents a public view of a user profile
type PublicProfile struct {
	UserID    ulid.ULID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Title     string    `json:"title,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	Location  string    `json:"location,omitempty"`
}

// ProfileCompleteness represents profile completion status
type ProfileCompleteness struct {
	OverallScore    int               `json:"overall_score"`    // 0-100
	CompletedFields []string          `json:"completed_fields"`
	MissingFields   []string          `json:"missing_fields"`
	Recommendations []string          `json:"recommendations"`
	Sections        map[string]int    `json:"sections"`         // section -> completion %
}

// ProfileValidation represents profile validation results
type ProfileValidation struct {
	IsValid bool                    `json:"is_valid"`
	Errors  []ProfileValidationError `json:"errors"`
}

type ProfileValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NotificationPreferences represents user notification settings
type NotificationPreferences struct {
	EmailNotifications    bool `json:"email_notifications"`
	PushNotifications     bool `json:"push_notifications"`
	SMSNotifications      bool `json:"sms_notifications"`
	MarketingEmails       bool `json:"marketing_emails"`
	SecurityAlerts        bool `json:"security_alerts"`
	ProductUpdates        bool `json:"product_updates"`
	WeeklyDigest          bool `json:"weekly_digest"`
	InvitationNotifications bool `json:"invitation_notifications"`
}

type UpdateNotificationPreferencesRequest struct {
	EmailNotifications      *bool `json:"email_notifications,omitempty"`
	PushNotifications       *bool `json:"push_notifications,omitempty"`
	SMSNotifications        *bool `json:"sms_notifications,omitempty"`
	MarketingEmails         *bool `json:"marketing_emails,omitempty"`
	SecurityAlerts          *bool `json:"security_alerts,omitempty"`
	ProductUpdates          *bool `json:"product_updates,omitempty"`
	WeeklyDigest            *bool `json:"weekly_digest,omitempty"`
	InvitationNotifications *bool `json:"invitation_notifications,omitempty"`
}

// ThemePreferences represents user UI/theme preferences
type ThemePreferences struct {
	Theme            string `json:"theme"`              // light, dark, auto
	PrimaryColor     string `json:"primary_color"`
	Language         string `json:"language"`
	TimeFormat       string `json:"time_format"`        // 12h, 24h
	DateFormat       string `json:"date_format"`
	Timezone         string `json:"timezone"`
	CompactMode      bool   `json:"compact_mode"`
	ShowAnimations   bool   `json:"show_animations"`
	HighContrast     bool   `json:"high_contrast"`
}

type UpdateThemePreferencesRequest struct {
	Theme          *string `json:"theme,omitempty"`
	PrimaryColor   *string `json:"primary_color,omitempty"`
	Language       *string `json:"language,omitempty"`
	TimeFormat     *string `json:"time_format,omitempty"`
	DateFormat     *string `json:"date_format,omitempty"`
	Timezone       *string `json:"timezone,omitempty"`
	CompactMode    *bool   `json:"compact_mode,omitempty"`
	ShowAnimations *bool   `json:"show_animations,omitempty"`
	HighContrast   *bool   `json:"high_contrast,omitempty"`
}

// PrivacyPreferences represents user privacy settings
type PrivacyPreferences struct {
	ProfileVisibility ProfileVisibility `json:"profile_visibility"`
	ShowEmail         bool             `json:"show_email"`
	ShowLastSeen      bool             `json:"show_last_seen"`
	AllowDirectMessages bool           `json:"allow_direct_messages"`
	DataProcessingConsent bool         `json:"data_processing_consent"`
	AnalyticsConsent   bool            `json:"analytics_consent"`
	ThirdPartyIntegrations bool        `json:"third_party_integrations"`
}

type UpdatePrivacyPreferencesRequest struct {
	ProfileVisibility      *ProfileVisibility `json:"profile_visibility,omitempty"`
	ShowEmail              *bool             `json:"show_email,omitempty"`
	ShowLastSeen           *bool             `json:"show_last_seen,omitempty"`
	AllowDirectMessages    *bool             `json:"allow_direct_messages,omitempty"`
	DataProcessingConsent  *bool             `json:"data_processing_consent,omitempty"`
	AnalyticsConsent       *bool             `json:"analytics_consent,omitempty"`
	ThirdPartyIntegrations *bool             `json:"third_party_integrations,omitempty"`
}

// OnboardingStep represents a step in the onboarding process
type OnboardingStep string

const (
	OnboardingStepProfile      OnboardingStep = "profile"
	OnboardingStepPreferences  OnboardingStep = "preferences"
	OnboardingStepOrganization OnboardingStep = "organization"
	OnboardingStepProject      OnboardingStep = "project"
	OnboardingStepIntegration  OnboardingStep = "integration"
	OnboardingStepComplete     OnboardingStep = "complete"
)

// OnboardingStatus represents the user's onboarding progress
type OnboardingStatus struct {
	UserID           ulid.ULID                    `json:"user_id"`
	IsCompleted      bool                         `json:"is_completed"`
	CompletedSteps   []OnboardingStep             `json:"completed_steps"`
	CurrentStep      OnboardingStep               `json:"current_step"`
	TotalSteps       int                          `json:"total_steps"`
	CompletionRate   int                          `json:"completion_rate"` // 0-100
	StepProgress     map[OnboardingStep]bool      `json:"step_progress"`
	StartedAt        *string                      `json:"started_at,omitempty"`
	CompletedAt      *string                      `json:"completed_at,omitempty"`
}

// OnboardingFlow represents the onboarding flow configuration
type OnboardingFlow struct {
	UserType UserType         `json:"user_type"`
	Steps    []OnboardingStep `json:"steps"`
	Optional []OnboardingStep `json:"optional_steps"`
}

type UpdateOnboardingPreferencesRequest struct {
	SkipOptionalSteps *bool   `json:"skip_optional_steps,omitempty"`
	PreferredFlow     *string `json:"preferred_flow,omitempty"`
}

// UserType represents different types of users
type UserType string

const (
	UserTypeDeveloper UserType = "developer"
	UserTypeManager   UserType = "manager"
	UserTypeAnalyst   UserType = "analyst"
	UserTypeAdmin     UserType = "admin"
)

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