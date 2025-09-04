// Package organization provides the organizational structure domain model.
//
// The organization domain handles multi-tenancy, organizational hierarchy,
// project management, and team membership across the platform.
package organization

import (
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
	Projects    []Project    `json:"projects,omitempty" gorm:"foreignKey:OrganizationID"`
	Members     []Member     `json:"members,omitempty" gorm:"foreignKey:OrganizationID"`
	Invitations []Invitation `json:"invitations,omitempty" gorm:"foreignKey:OrganizationID"`
}

// Member represents the many-to-many relationship between users and organizations.
type Member struct {
	ID             ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);not null"`
	UserID         ulid.ULID `json:"user_id" gorm:"type:char(26);not null"` // References user.User.ID
	RoleID         ulid.ULID `json:"role_id" gorm:"type:char(26);not null"` // References auth.Role.ID
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

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
	Organization Organization  `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Environments []Environment `json:"environments,omitempty" gorm:"foreignKey:ProjectID"`
}

// Environment represents an environment within a project (dev, staging, prod).
type Environment struct {
	ID        ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	ProjectID ulid.ULID `json:"project_id" gorm:"type:char(26);not null"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Slug      string    `json:"slug" gorm:"size:255;not null"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
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

type CreateEnvironmentRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	Slug string `json:"slug" validate:"required,min=1,max=50,slug"`
}

type UpdateEnvironmentRequest struct {
	Name *string `json:"name,omitempty"`
	Slug *string `json:"slug,omitempty"`
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

func NewEnvironment(projectID ulid.ULID, name, slug string) *Environment {
	return &Environment{
		ID:        ulid.New(),
		ProjectID: projectID,
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewMember(orgID, userID, roleID ulid.ULID) *Member {
	return &Member{
		ID:             ulid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		RoleID:         roleID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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

// Table name methods for GORM
func (Organization) TableName() string { return "organizations" }
func (Member) TableName() string       { return "organization_members" }
func (Project) TableName() string      { return "projects" }
func (Environment) TableName() string  { return "environments" }
func (Invitation) TableName() string   { return "user_invitations" }