// Package user provides the user domain model and core business logic.
//
// The user domain handles user account management, profiles, and preferences.
// It maintains user identity and authentication state across the platform.
package user

import (
	"time"

	"brokle/pkg/ulid"
	"gorm.io/gorm"
)

// User represents a platform user with full authentication support.
type User struct {
	ID        ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	Email     string    `json:"email" gorm:"size:255;not null;uniqueIndex"`
	FirstName string    `json:"first_name" gorm:"size:255;not null"`
	LastName  string    `json:"last_name" gorm:"size:255;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`

	// Profile fields
	AvatarURL string `json:"avatar_url,omitempty" gorm:"size:500"`
	Phone     string `json:"phone,omitempty" gorm:"size:50"`
	Timezone  string `json:"timezone" gorm:"size:50;default:'UTC'"`
	Language  string `json:"language" gorm:"size:10;default:'en'"`

	// Email verification
	IsEmailVerified bool       `json:"is_email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`

	// Authentication tracking
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LoginCount  int        `json:"login_count" gorm:"default:0"`

	// Onboarding
	OnboardingCompleted bool `json:"onboarding_completed" gorm:"default:false"`

	// Default Organization
	DefaultOrganizationID *ulid.ULID `json:"default_organization_id,omitempty" gorm:"type:char(26)"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// UserProfile represents extended user profile information.
type UserProfile struct {
	UserID      ulid.ULID `json:"user_id" db:"user_id"`
	Bio         *string   `json:"bio,omitempty" db:"bio"`
	Location    *string   `json:"location,omitempty" db:"location"`
	Website     *string   `json:"website,omitempty" db:"website"`
	TwitterURL  *string   `json:"twitter_url,omitempty" db:"twitter_url"`
	LinkedInURL *string   `json:"linkedin_url,omitempty" db:"linkedin_url"`
	GitHubURL   *string   `json:"github_url,omitempty" db:"github_url"`
	Timezone    string    `json:"timezone" db:"timezone"`
	Language    string    `json:"language" db:"language"`
	Theme       string    `json:"theme" db:"theme"` // light, dark, auto
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserPreferences represents user application preferences.
type UserPreferences struct {
	UserID                ulid.ULID `json:"user_id" db:"user_id"`
	EmailNotifications    bool      `json:"email_notifications" db:"email_notifications"`
	PushNotifications     bool      `json:"push_notifications" db:"push_notifications"`
	MarketingEmails       bool      `json:"marketing_emails" db:"marketing_emails"`
	WeeklyReports         bool      `json:"weekly_reports" db:"weekly_reports"`
	MonthlyReports        bool      `json:"monthly_reports" db:"monthly_reports"`
	SecurityAlerts        bool      `json:"security_alerts" db:"security_alerts"`
	BillingAlerts         bool      `json:"billing_alerts" db:"billing_alerts"`
	UsageThresholdPercent int       `json:"usage_threshold_percent" db:"usage_threshold_percent"` // 0-100
	Theme                 string    `json:"theme" db:"theme"`                                     // light, dark, system
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents the data needed to create a new user.
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Password  string `json:"password" validate:"required,min=8"`
	Timezone  string `json:"timezone,omitempty" validate:"omitempty"`
	Language  string `json:"language,omitempty" validate:"omitempty,len=2"`
}

// UpdateUserRequest represents the data that can be updated for a user.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Timezone  *string `json:"timezone,omitempty" validate:"omitempty"`
	Language  *string `json:"language,omitempty" validate:"omitempty,len=2"`
}

// UpdateProfileRequest represents the data that can be updated for a user profile.
type UpdateProfileRequest struct {
	Bio         *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Location    *string `json:"location,omitempty" validate:"omitempty,max=100"`
	Website     *string `json:"website,omitempty" validate:"omitempty,url"`
	TwitterURL  *string `json:"twitter_url,omitempty" validate:"omitempty,url"`
	LinkedInURL *string `json:"linkedin_url,omitempty" validate:"omitempty,url"`
	GitHubURL   *string `json:"github_url,omitempty" validate:"omitempty,url"`
	Timezone    *string `json:"timezone,omitempty" validate:"omitempty"`
	Language    *string `json:"language,omitempty" validate:"omitempty,len=2"`
	Theme       *string `json:"theme,omitempty" validate:"omitempty,oneof=light dark auto"`
}

// UpdatePreferencesRequest represents the data that can be updated for user preferences.
type UpdatePreferencesRequest struct {
	EmailNotifications    *bool `json:"email_notifications,omitempty"`
	PushNotifications     *bool `json:"push_notifications,omitempty"`
	MarketingEmails       *bool `json:"marketing_emails,omitempty"`
	WeeklyReports         *bool `json:"weekly_reports,omitempty"`
	MonthlyReports        *bool `json:"monthly_reports,omitempty"`
	SecurityAlerts        *bool `json:"security_alerts,omitempty"`
	BillingAlerts         *bool `json:"billing_alerts,omitempty"`
	UsageThresholdPercent *int  `json:"usage_threshold_percent,omitempty" validate:"omitempty,min=0,max=100"`
}

// PublicUser represents a user without sensitive information.
type PublicUser struct {
	ID                 ulid.ULID  `json:"id"`
	Name               string     `json:"name"`
	AvatarURL          *string    `json:"avatar_url,omitempty"`
	IsEmailVerified    bool       `json:"is_email_verified"`
	OnboardingCompleted bool      `json:"onboarding_completed"`
	CreatedAt          time.Time  `json:"created_at"`
}

// ToPublic converts a User to PublicUser, removing sensitive information.
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:        u.ID,
		Name:      u.GetFullName(),
		AvatarURL: &u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}

// GetFullName returns the user's full name.
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsEmailVerified checks if user's email is verified.
func (u *User) IsVerified() bool {
	return u.IsEmailVerified
}

// MarkEmailAsVerified marks the user's email as verified.
func (u *User) MarkEmailAsVerified() {
	now := time.Now()
	u.IsEmailVerified = true
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

// UpdateLastLogin updates the user's last login timestamp and count.
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.LoginCount++
	u.UpdatedAt = now
}

// SetPassword sets the user's password hash.
func (u *User) SetPassword(hashedPassword string) {
	u.Password = hashedPassword
	u.UpdatedAt = time.Now()
}

// CompleteOnboarding marks the user's onboarding as completed.
func (u *User) CompleteOnboarding() {
	u.OnboardingCompleted = true
	u.UpdatedAt = time.Now()
}

// SetDefaultOrganization sets the user's default organization.
func (u *User) SetDefaultOrganization(orgID ulid.ULID) {
	u.DefaultOrganizationID = &orgID
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user account.
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Reactivate reactivates the user account.
func (u *User) Reactivate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// NewUser creates a new user with default values.
func NewUser(email, firstName, lastName string) *User {
	return &User{
		ID:                     ulid.New(),
		Email:                  email,
		FirstName:              firstName,
		LastName:               lastName,
		IsActive:               true,
		IsEmailVerified:        false,
		OnboardingCompleted:    false,
		Timezone:               "UTC",
		Language:               "en",
		LoginCount:             0,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}

// NewUserProfile creates a new user profile with default values.
func NewUserProfile(userID ulid.ULID) *UserProfile {
	return &UserProfile{
		UserID:    userID,
		Timezone:  "UTC",
		Language:  "en",
		Theme:     "light",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewUserPreferences creates new user preferences with default values.
func NewUserPreferences(userID ulid.ULID) *UserPreferences {
	return &UserPreferences{
		UserID:                userID,
		EmailNotifications:    true,
		PushNotifications:     true,
		MarketingEmails:       false,
		WeeklyReports:         true,
		MonthlyReports:        true,
		SecurityAlerts:        true,
		BillingAlerts:         true,
		UsageThresholdPercent: 80,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
}

// Table name method for GORM
func (User) TableName() string { return "users" }