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

	// Basic user settings (kept in users table)
	Timezone string `json:"timezone" gorm:"size:50;default:'UTC'"`
	Language string `json:"language" gorm:"size:10;default:'en'"`

	// Email verification
	IsEmailVerified bool       `json:"is_email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`

	// Authentication tracking
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LoginCount  int        `json:"login_count" gorm:"default:0"`

	// Default Organization
	DefaultOrganizationID *ulid.ULID `json:"default_organization_id,omitempty" gorm:"type:char(26)"`

	// Signup information
	Role           string  `json:"role" gorm:"size:100;not null"`
	ReferralSource *string `json:"referral_source,omitempty" gorm:"size:100"`

	// Authentication method tracking
	AuthMethod      string  `json:"auth_method" gorm:"column:auth_method;size:20;default:'password'"` // password | oauth
	OAuthProvider   *string `json:"oauth_provider,omitempty" gorm:"column:oauth_provider;size:50"`    // google | github | etc
	OAuthProviderID *string `json:"-" gorm:"column:oauth_provider_id;size:255"`                       // Provider's unique user ID (hidden from JSON)

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// UserProfile represents extended user profile information and preferences.
type UserProfile struct {
	UserID ulid.ULID `json:"user_id" gorm:"type:char(26);primaryKey"`

	// Profile information
	Bio         *string `json:"bio,omitempty" gorm:"type:text"`
	Location    *string `json:"location,omitempty" gorm:"size:100"`
	Website     *string `json:"website,omitempty" gorm:"size:500"`
	TwitterURL  *string `json:"twitter_url,omitempty" gorm:"column:twitter_url;size:500"`
	LinkedInURL *string `json:"linkedin_url,omitempty" gorm:"column:linkedin_url;size:500"`
	GitHubURL   *string `json:"github_url,omitempty" gorm:"column:github_url;size:500"`

	// Contact information (moved from users table)
	AvatarURL *string `json:"avatar_url,omitempty" gorm:"column:avatar_url;size:500"`
	Phone     *string `json:"phone,omitempty" gorm:"size:50"`

	// Display preferences
	Timezone string `json:"timezone" gorm:"size:50;default:'UTC'"`
	Language string `json:"language" gorm:"size:10;default:'en'"`
	Theme    string `json:"theme" gorm:"size:20;default:'light'"` // light, dark, auto

	// Notification preferences (moved from user_preferences table)
	EmailNotifications    bool `json:"email_notifications" gorm:"default:true"`
	PushNotifications     bool `json:"push_notifications" gorm:"default:true"`
	MarketingEmails       bool `json:"marketing_emails" gorm:"default:false"`
	WeeklyReports         bool `json:"weekly_reports" gorm:"default:true"`
	MonthlyReports        bool `json:"monthly_reports" gorm:"default:true"`
	SecurityAlerts        bool `json:"security_alerts" gorm:"default:true"`
	BillingAlerts         bool `json:"billing_alerts" gorm:"default:true"`
	UsageThresholdPercent int  `json:"usage_threshold_percent" gorm:"default:80"` // 0-100

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	Timezone  *string `json:"timezone,omitempty" validate:"omitempty"`
	Language  *string `json:"language,omitempty" validate:"omitempty,len=2"`
}

// UpdateProfileRequest represents the data that can be updated for a user profile.
type UpdateProfileRequest struct {
	// Profile information
	Bio         *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Location    *string `json:"location,omitempty" validate:"omitempty,max=100"`
	Website     *string `json:"website,omitempty" validate:"omitempty,url"`
	TwitterURL  *string `json:"twitter_url,omitempty" validate:"omitempty,url"`
	LinkedInURL *string `json:"linkedin_url,omitempty" validate:"omitempty,url"`
	GitHubURL   *string `json:"github_url,omitempty" validate:"omitempty,url"`

	// Contact information
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=50"`

	// Display preferences
	Timezone *string `json:"timezone,omitempty" validate:"omitempty"`
	Language *string `json:"language,omitempty" validate:"omitempty,len=2"`
	Theme    *string `json:"theme,omitempty" validate:"omitempty,oneof=light dark auto"`

	// Notification preferences
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
	ID              ulid.ULID `json:"id"`
	Name            string    `json:"name"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToPublic converts a User to PublicUser, removing sensitive information.
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:              u.ID,
		Name:            u.GetFullName(),
		IsEmailVerified: u.IsEmailVerified,
		CreatedAt:       u.CreatedAt,
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
func NewUser(email, firstName, lastName, role string) *User {
	return &User{
		ID:              ulid.New(),
		Email:           email,
		FirstName:       firstName,
		LastName:        lastName,
		Role:            role,
		IsActive:        true,
		IsEmailVerified: false,
		Timezone:        "UTC",
		Language:        "en",
		LoginCount:      0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// NewUserProfile creates a new user profile with default values.
func NewUserProfile(userID ulid.ULID) *UserProfile {
	return &UserProfile{
		UserID:   userID,
		Timezone: "UTC",
		Language: "en",
		Theme:    "light",

		// Default notification preferences
		EmailNotifications:    true,
		PushNotifications:     true,
		MarketingEmails:       false,
		WeeklyReports:         true,
		MonthlyReports:        true,
		SecurityAlerts:        true,
		BillingAlerts:         true,
		UsageThresholdPercent: 80,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Table name methods for GORM
func (User) TableName() string        { return "users" }
func (UserProfile) TableName() string { return "user_profiles" }
