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
	UpdatedAt          time.Time              `json:"updated_at"`
	CreatedAt          time.Time              `json:"created_at"`
	TrialEndsAt        *time.Time             `json:"trial_ends_at,omitempty"`
	DeletedAt          gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
	Plan               string                 `json:"plan" gorm:"size:50;default:'free'"`
	SubscriptionStatus string                 `json:"subscription_status" gorm:"size:50;default:'active'"`
	BillingEmail       string                 `json:"billing_email,omitempty" gorm:"size:255"`
	Name               string                 `json:"name" gorm:"size:255;not null"`
	Projects           []Project              `json:"projects,omitempty" gorm:"foreignKey:OrganizationID"`
	Members            []Member               `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Invitations        []Invitation           `json:"invitations,omitempty" gorm:"foreignKey:OrganizationID"`
	Settings           []OrganizationSettings `json:"settings,omitempty" gorm:"foreignKey:OrganizationID"`
	ID                 ulid.ULID              `json:"id" gorm:"type:char(26);primaryKey"`
}

// Member represents the many-to-many relationship between users and organizations.
type Member struct {
	JoinedAt       time.Time      `json:"joined_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	InvitedBy      *ulid.ULID     `json:"invited_by,omitempty" gorm:"type:char(26)"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Status         string         `json:"status" gorm:"size:20;default:'active'"`
	Organization   Organization   `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	OrganizationID ulid.ULID      `json:"organization_id" gorm:"type:char(26);not null;primaryKey;priority:1"`
	UserID         ulid.ULID      `json:"user_id" gorm:"type:char(26);not null;primaryKey;priority:2"`
	RoleID         ulid.ULID      `json:"role_id" gorm:"type:char(26);not null"`
}

// Project represents a project within an organization.
type Project struct {
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Name           string         `json:"name" gorm:"size:255;not null"`
	Description    string         `json:"description,omitempty" gorm:"text"`
	Status         string         `json:"status" gorm:"size:20;not null;default:active;check:status IN ('active','archived')"`
	Organization   Organization   `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	ID             ulid.ULID      `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID      `json:"organization_id" gorm:"type:char(26);not null"`
}

// Invitation represents an invitation to join an organization.
type Invitation struct {
	ExpiresAt      time.Time        `json:"expires_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	CreatedAt      time.Time        `json:"created_at"`
	AcceptedAt     *time.Time       `json:"accepted_at,omitempty"`
	DeletedAt      gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index"`
	Status         InvitationStatus `json:"status" gorm:"size:50;default:'pending'"`
	Token          string           `json:"token" gorm:"size:255;not null;uniqueIndex"`
	Email          string           `json:"email" gorm:"size:255;not null"`
	Organization   Organization     `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	ID             ulid.ULID        `json:"id" gorm:"type:char(26);primaryKey"`
	InvitedByID    ulid.ULID        `json:"invited_by_id" gorm:"type:char(26);not null"`
	RoleID         ulid.ULID        `json:"role_id" gorm:"type:char(26);not null"`
	OrganizationID ulid.ULID        `json:"organization_id" gorm:"type:char(26);not null"`
}

// Request/Response DTOs
type CreateOrganizationRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=100"`
	BillingEmail string `json:"billing_email" validate:"email"`
}

type UpdateOrganizationRequest struct {
	Name         *string `json:"name,omitempty"`
	BillingEmail *string `json:"billing_email,omitempty"`
	Plan         *string `json:"plan,omitempty"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
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
func NewOrganization(name string) *Organization {
	return &Organization{
		ID:                 ulid.New(),
		Name:               name,
		Plan:               "free",
		SubscriptionStatus: "active",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func NewProject(orgID ulid.ULID, name, description string) *Project {
	return &Project{
		ID:             ulid.New(),
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		Status:         "active",
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

// Project utility methods
func (p *Project) IsActive() bool {
	return p.Status == "active"
}

func (p *Project) IsArchived() bool {
	return p.Status == "archived"
}

func (p *Project) Archive() {
	p.Status = "archived"
	p.UpdatedAt = time.Now()
}

func (p *Project) Unarchive() {
	p.Status = "active"
	p.UpdatedAt = time.Now()
}

// OrganizationSettings represents key-value settings for an organization.
type OrganizationSettings struct {
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Key            string       `json:"key" gorm:"size:255;not null"`
	Value          string       `json:"value" gorm:"type:jsonb;not null"`
	Organization   Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	ID             ulid.ULID    `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID    `json:"organization_id" gorm:"type:char(26);not null"`
}

// Settings-related DTOs
type CreateOrganizationSettingRequest struct {
	Value interface{} `json:"value" validate:"required"`
	Key   string      `json:"key" validate:"required,min=1,max=255"`
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

// OrganizationWithProjectsAndRole represents an organization with its projects and the user's role
type OrganizationWithProjectsAndRole struct {
	Organization *Organization
	RoleName     string
	Projects     []*Project
}

// Table name methods for GORM
func (Organization) TableName() string         { return "organizations" }
func (Member) TableName() string               { return "organization_members" }
func (Project) TableName() string              { return "projects" }
func (Invitation) TableName() string           { return "user_invitations" }
func (OrganizationSettings) TableName() string { return "organization_settings" }
