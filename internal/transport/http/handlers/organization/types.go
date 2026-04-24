package organization

import (
	"brokle/internal/core/domain/organization"
	"brokle/pkg/response"
	"time"

	"github.com/google/uuid"
)

// Huma operation types for the organization package.

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
	Status    string     `json:"status"`
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

type ListOrganizationsInput struct {
	Search  string `query:"search" required:"false" doc:"Filter by name substring"`
	Page    int    `query:"page" required:"false" minimum:"1" doc:"Page number (1-indexed); default 1"`
	Limit   int    `query:"limit" required:"false" minimum:"1" maximum:"100" doc:"Items per page; default 20"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc" doc:"Created-at sort direction; default desc"`
}

type listOrganizationsBody struct {
	Data       []organizationResponse `json:"data"`
	Pagination *response.Pagination   `json:"pagination"`
}

type ListOrganizationsOutput struct {
	Body listOrganizationsBody
}

type CreateOrganizationInput struct {
	Body createOrganizationBody
}

type createOrganizationBody struct {
	Name        string `json:"name" minLength:"2" maxLength:"100" doc:"Organization name"`
	Description string `json:"description,omitempty" maxLength:"500"`
}

type CreateOrganizationOutput struct {
	Body organizationResponse
}

type GetOrganizationInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type GetOrganizationOutput struct {
	Body organizationResponse
}

type UpdateOrganizationInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Body  updateOrganizationBody
}

type updateOrganizationBody struct {
	Name         *string `json:"name,omitempty" minLength:"2" maxLength:"100"`
	BillingEmail *string `json:"billing_email,omitempty" format:"email"`
	Description  *string `json:"description,omitempty" maxLength:"500"`
}

type UpdateOrganizationOutput struct {
	Body organizationResponse
}

type DeleteOrganizationInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type DeleteOrganizationOutput struct{}

type ListMembersInput struct {
	OrgID  string `path:"orgId" format:"uuid"`
	Status string `query:"status" required:"false" enum:"active,invited,suspended"`
	Role   string `query:"role" required:"false"`
}

type listMembersBody struct {
	Members []memberResponse `json:"members"`
	Total   int              `json:"total"`
}

type ListMembersOutput struct {
	Body listMembersBody
}

type RemoveMemberInput struct {
	OrgID  string `path:"orgId" format:"uuid"`
	UserID string `path:"userId" format:"uuid"`
}

type RemoveMemberOutput struct{}

type CreateInvitationInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Body  createInvitationBody
}

type createInvitationBody struct {
	Email   string    `json:"email" format:"email" doc:"Email address of user to invite"`
	RoleID  uuid.UUID `json:"role_id" doc:"Role ID to assign"`
	Message *string   `json:"message,omitempty" maxLength:"500" doc:"Optional personal message"`
}

type CreateInvitationOutput struct {
	Body invitationResponse
}

type ListPendingInvitationsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type listPendingInvitationsBody struct {
	Invitations []invitationResponse `json:"invitations"`
	Total       int                  `json:"total"`
}

type ListPendingInvitationsOutput struct {
	Body listPendingInvitationsBody
}

type ResendInvitationInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	InvitationID string `path:"invitationId" format:"uuid"`
}

type ResendInvitationOutput struct {
	Body invitationResponse
}

type RevokeInvitationInput struct {
	OrgID        string `path:"orgId" format:"uuid"`
	InvitationID string `path:"invitationId" format:"uuid"`
}

type RevokeInvitationOutput struct{}

type ListUserInvitationsInput struct{}

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

type listUserInvitationsBody struct {
	Invitations []userInvitationResponse `json:"invitations"`
	Total       int                      `json:"total"`
}

type ListUserInvitationsOutput struct {
	Body listUserInvitationsBody
}

type ValidateInvitationTokenInput struct {
	Token string `path:"token" doc:"Invitation token"`
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

type ValidateInvitationTokenOutput struct {
	Body invitationDetailsBody
}

type AcceptInvitationInput struct {
	Body acceptInvitationBody
}

type acceptInvitationBody struct {
	Token string `json:"token" minLength:"1" doc:"Invitation token"`
}

type AcceptInvitationOutput struct{}

type DeclineInvitationInput struct {
	Body declineInvitationBody
}

type declineInvitationBody struct {
	Token string `json:"token" minLength:"1" doc:"Invitation token"`
}

type DeclineInvitationOutput struct{}

type settingResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Key            string    `json:"key"`
	Value          any       `json:"value"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListSettingsInput struct {
	OrgID string `path:"orgId" format:"uuid"`
}

type listSettingsBody struct {
	Settings map[string]any `json:"settings"`
}

type ListSettingsOutput struct {
	Body listSettingsBody
}

type CreateSettingInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Body  createSettingBody
}

type createSettingBody struct {
	Key   string `json:"key" minLength:"1" maxLength:"255" doc:"Setting key"`
	Value any    `json:"value" doc:"Setting value (any JSON type)"`
}

type CreateSettingOutput struct {
	Body settingResponse
}

type GetSettingInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Key   string `path:"key" minLength:"1"`
}

type GetSettingOutput struct {
	Body settingResponse
}

type UpdateSettingInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Key   string `path:"key" minLength:"1"`
	Body  updateSettingBody
}

type updateSettingBody struct {
	Value any `json:"value" doc:"New setting value (any JSON type)"`
}

type UpdateSettingOutput struct {
	Body settingResponse
}

type DeleteSettingInput struct {
	OrgID string `path:"orgId" format:"uuid"`
	Key   string `path:"key" minLength:"1"`
}

type DeleteSettingOutput struct{}
