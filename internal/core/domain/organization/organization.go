// Package organization provides the organizational structure domain model.
//
// The organization domain handles multi-tenancy, organizational hierarchy,
// project management, and team membership across the platform.
package organization

import (
	"encoding/json"
	"time"

	"brokle/pkg/ulid"
	"gorm.io/gorm"
)

// Organization represents a tenant organization with full multi-tenancy.
type Organization struct {
	ID   ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	Name string    `json:"name" gorm:"size:255;not null"`
	Slug string    `json:"slug" gorm:"size:255;not null;uniqueIndex"`

	// Business fields
	BillingEmail       string     `json:"billing_email,omitempty" gorm:"size:255"`
	Plan               string     `json:"plan" gorm:"size:50;default:'free'"`
	SubscriptionStatus string     `json:"subscription_status" gorm:"size:50;default:'active'"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations (using ULIDs to avoid circular imports)
	Projects    []Project               `json:"projects,omitempty" gorm:"foreignKey:OrganizationID"`
	Members     []Member                `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Invitations []Invitation            `json:"invitations,omitempty" gorm:"foreignKey:OrganizationID"`
	Settings    []OrganizationSettings  `json:"settings,omitempty" gorm:"foreignKey:OrganizationID"`
}

// Member represents the many-to-many relationship between users and organizations.
type Member struct {
	OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null;primaryKey;priority:1"`
	UserID         ulid.ULID `json:"user_id" gorm:"type:char(26);not null;primaryKey;priority:2"` // References user.User.ID
	RoleID         ulid.ULID `json:"role_id" gorm:"type:char(26);not null"` // References auth.Role.ID
	Status         string    `json:"status" gorm:"size:20;default:'active'"` // "active", "suspended", "invited"
	JoinedAt       time.Time `json:"joined_at"`
	InvitedBy      *ulid.ULID `json:"invited_by,omitempty" gorm:"type:char(26)"` // References user.User.ID
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	// User and Role will be loaded separately to avoid circular imports
}

// Project represents a project within an organization.
type Project struct {
	ID             ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null"`
	Name           string    `json:"name" gorm:"size:255;not null"`
	Slug           string    `json:"slug" gorm:"size:255;not null"`
	Description    string    `json:"description,omitempty" gorm:"text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}


// Invitation represents an invitation to join an organization.
type Invitation struct {
	ID             ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null"`
	RoleID         ulid.ULID `json:"role_id" gorm:"type:char(26);not null"`      // References auth.Role.ID
	InvitedByID    ulid.ULID `json:"invited_by_id" gorm:"type:char(26);not null"` // References user.User.ID
	Email          string           `json:"email" gorm:"size:255;not null"`
	Token          string           `json:"token" gorm:"size:255;not null;uniqueIndex"`
	Status         InvitationStatus `json:"status" gorm:"size:50;default:'pending'"`
	ExpiresAt      time.Time        `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	// Role and InvitedBy will be loaded separately to avoid circular imports
}

// Request/Response DTOs
type CreateOrganizationRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	Slug         string `json:"slug" validate:"required,min=1,max=50,slug"`
	BillingEmail string `json:"billing_email" validate:"email"`
}

type UpdateOrganizationRequest struct {
	Name         *string `json:"name,omitempty"`
	BillingEmail *string `json:"billing_email,omitempty"`
	Plan         *string `json:"plan,omitempty"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"required,min=1,max=50,slug"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}


type InviteUserRequest struct {
	Email  string    `json:"email" validate:"required,email"`
	RoleID ulid.ULID `json:"role_id" validate:"required"`
}

// InvitationStatus represents the status of an organization invitation
type InvitationStatus string

// Invitation statuses
const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

// Constructor functions
func NewOrganization(name, slug string) *Organization {
	return &Organization{
		ID:                 ulid.New(),
		Name:               name,
		Slug:               slug,
		Plan:               "free",
		SubscriptionStatus: "active",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func NewProject(orgID ulid.ULID, name, slug, description string) *Project {
	return &Project{
		ID:             ulid.New(),
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Description:    description,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}


func NewMember(orgID, userID, roleID ulid.ULID) *Member {
	now := time.Now()
	return &Member{
		OrganizationID: orgID,
		UserID:         userID,
		RoleID:         roleID,
		Status:         "active",
		JoinedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func NewInvitation(orgID, roleID, invitedByID ulid.ULID, email, token string, expiresAt time.Time) *Invitation {
	return &Invitation{
		ID:             ulid.New(),
		OrganizationID: orgID,
		RoleID:         roleID,
		InvitedByID:    invitedByID,
		Email:          email,
		Token:          token,
		Status:         InvitationStatusPending,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Utility methods
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

func (i *Invitation) IsValid() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

func (i *Invitation) Accept() {
	now := time.Now()
	i.Status = InvitationStatusAccepted
	i.AcceptedAt = &now
	i.UpdatedAt = now
}

func (i *Invitation) Revoke() {
	i.Status = InvitationStatusRevoked
	i.UpdatedAt = time.Now()
}

// OrganizationSettings represents key-value settings for an organization.
type OrganizationSettings struct {
	ID             ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null"`
	Key            string    `json:"key" gorm:"size:255;not null"`
	Value          string    `json:"value" gorm:"type:jsonb;not null"` // JSONB for flexible value storage
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// Settings-related DTOs
type CreateOrganizationSettingRequest struct {
	Key   string      `json:"key" validate:"required,min=1,max=255"`
	Value interface{} `json:"value" validate:"required"`
}

type UpdateOrganizationSettingRequest struct {
	Value interface{} `json:"value" validate:"required"`
}

type GetOrganizationSettingsResponse struct {
	Settings map[string]interface{} `json:"settings"`
}

// OrganizationSetting utility methods
func NewOrganizationSettings(orgID ulid.ULID, key string, value interface{}) (*OrganizationSettings, error) {
	// Convert value to JSON string for storage
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return &OrganizationSettings{
		ID:             ulid.New(),
		OrganizationID: orgID,
		Key:            key,
		Value:          string(valueBytes),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (os *OrganizationSettings) GetValue() (interface{}, error) {
	var value interface{}
	err := json.Unmarshal([]byte(os.Value), &value)
	return value, err
}

func (os *OrganizationSettings) SetValue(value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	os.Value = string(valueBytes)
	os.UpdatedAt = time.Now()
	return nil
}

// Table name methods for GORM
func (Organization) TableName() string         { return "organizations" }
func (Member) TableName() string               { return "organization_members" }
func (Project) TableName() string              { return "projects" }
func (Invitation) TableName() string           { return "user_invitations" }
func (OrganizationSettings) TableName() string { return "organization_settings" }