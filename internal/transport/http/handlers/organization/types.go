package organization

import (
	"time"

	"github.com/google/uuid"

	"brokle/internal/core/domain/organization"
	"brokle/pkg/response"
)

// ---- response DTOs ---------------------------------------------------

type organizationResponse struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Plan               string    `json:"plan"`
	SubscriptionStatus string    `json:"subscription_status"`
	BillingEmail       *string   `json:"billing_email,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type memberResponse struct {
	UserID    uuid.UUID  `json:"user_id"`
	RoleID    uuid.UUID  `json:"role_id"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`
	JoinedAt  time.Time  `json:"joined_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type invitationResponse struct {
	ID             uuid.UUID                `json:"id"`
	OrganizationID uuid.UUID                `json:"organization_id"`
	Email          string                   `json:"email"`
	Status         string                   `json:"status"`
	TokenPreview   *string                  `json:"token_preview,omitempty"`
	RoleID         uuid.UUID                `json:"role_id"`
	Role           *organization.RoleRef    `json:"role,omitempty"`
	Inviter        *organization.InviterRef `json:"inviter,omitempty"`
	InvitedByID    *uuid.UUID               `json:"invited_by_id,omitempty"`
	Message        *string                  `json:"message,omitempty"`
	ResentCount    int                      `json:"resent_count"`
	ResentAt       *time.Time               `json:"resent_at,omitempty"`
	ExpiresAt      time.Time                `json:"expires_at"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

type userInvitationResponse struct {
	ID               uuid.UUID                `json:"id"`
	Email            string                   `json:"email"`
	Status           string                   `json:"status"`
	OrganizationID   uuid.UUID                `json:"organization_id"`
	OrganizationName string                   `json:"organization_name"`
	Role             *organization.RoleRef    `json:"role,omitempty"`
	Inviter          *organization.InviterRef `json:"inviter,omitempty"`
	Message          *string                  `json:"message,omitempty"`
	ExpiresAt        time.Time                `json:"expires_at"`
	CreatedAt        time.Time                `json:"created_at"`
}

type invitationDetailsBody struct {
	OrganizationID   uuid.UUID `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	Email            string    `json:"email"`
	RoleName         string    `json:"role_name"`
	InviterName      string    `json:"inviter_name"`
	ExpiresAt        time.Time `json:"expires_at"`
	IsExpired        bool      `json:"is_expired"`
}

type settingResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Key            string    `json:"key"`
	Value          any       `json:"value"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ---- list-response bodies -------------------------------------------

type listOrganizationsBody struct {
	Data       []organizationResponse `json:"data"`
	Pagination *response.Pagination   `json:"pagination"`
}

type listMembersBody struct {
	Data []memberResponse `json:"data"`
}

type listPendingInvitationsBody struct {
	Data []invitationResponse `json:"data"`
}

type listUserInvitationsBody struct {
	Data []userInvitationResponse `json:"data"`
}

type listSettingsBody struct {
	Settings map[string]any `json:"settings"`
}

// ---- request bodies -------------------------------------------------

type createOrganizationBody struct {
	Name        string `json:"name"                  validate:"required,min=2,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

type updateOrganizationBody struct {
	Name         *string `json:"name,omitempty"          validate:"omitempty,min=2,max=100"`
	BillingEmail *string `json:"billing_email,omitempty" validate:"omitempty,email"`
	Description  *string `json:"description,omitempty"   validate:"omitempty,max=500"`
}

type createInvitationBody struct {
	Email   string    `json:"email"             validate:"required,email"`
	RoleID  uuid.UUID `json:"role_id"           validate:"required"`
	Message *string   `json:"message,omitempty" validate:"omitempty,max=500"`
}

type acceptInvitationBody struct {
	Token string `json:"token" validate:"required,min=1"`
}

type declineInvitationBody struct {
	Token string `json:"token" validate:"required,min=1"`
}

type createSettingBody struct {
	Key   string `json:"key"   validate:"required,min=1,max=255"`
	Value any    `json:"value"`
}

type updateSettingBody struct {
	Value any `json:"value"`
}
