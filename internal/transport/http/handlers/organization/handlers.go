// Package organization exposes /api/v1/organizations/* and
// /api/v1/invitations/* operations on the dashboard plane (apiAdmin).
// Covers organization CRUD, member listing/removal, invitation
// lifecycle (create / list pending / resend / revoke / accept /
// decline / user's own), and organization-settings CRUD.
//
// Every route requires RequireAuth — httpctx.MustGetUserID is safe.
package organization

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"brokle/internal/core/domain/organization"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
)

type handler struct {
	orgSvc        organization.OrganizationService
	memberSvc     organization.MemberService
	invitationSvc organization.InvitationService
	settingsSvc   organization.OrganizationSettingsService
	logger        *slog.Logger
}

// RegisterRoutes registers every organization / invitation / settings
// operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	orgSvc organization.OrganizationService,
	memberSvc organization.MemberService,
	invitationSvc organization.InvitationService,
	settingsSvc organization.OrganizationSettingsService,
	logger *slog.Logger,
) {
	h := &handler{
		orgSvc:        orgSvc,
		memberSvc:     memberSvc,
		invitationSvc: invitationSvc,
		settingsSvc:   settingsSvc,
		logger:        logger,
	}

	// ----- organizations -----

	huma.Register(api, huma.Operation{
		OperationID: "list-organizations",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations",
		Tags:        []string{"organizations"},
		Summary:     "List organizations the authenticated user belongs to",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listOrganizations)

	huma.Register(api, huma.Operation{
		OperationID:   "create-organization",
		Method:        http.MethodPost,
		Path:          "/api/v1/organizations",
		Tags:          []string{"organizations"},
		Summary:       "Create an organization; caller becomes owner",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createOrganization)

	huma.Register(api, huma.Operation{
		OperationID: "get-organization",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}",
		Tags:        []string{"organizations"},
		Summary:     "Get organization details",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getOrganization)

	huma.Register(api, huma.Operation{
		OperationID: "update-organization",
		Method:      http.MethodPatch,
		Path:        "/api/v1/organizations/{orgId}",
		Tags:        []string{"organizations"},
		Summary:     "Partially update organization",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateOrganization)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-organization",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}",
		Tags:          []string{"organizations"},
		Summary:       "Delete an organization",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteOrganization)

	// ----- members -----

	huma.Register(api, huma.Operation{
		OperationID: "list-organization-members",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/members",
		Tags:        []string{"organizations"},
		Summary:     "List members of an organization",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listMembers)

	huma.Register(api, huma.Operation{
		OperationID:   "remove-organization-member",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}/members/{userId}",
		Tags:          []string{"organizations"},
		Summary:       "Remove a member from an organization",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.removeMember)

	// ----- invitations (org-scoped) -----

	huma.Register(api, huma.Operation{
		OperationID:   "create-invitation",
		Method:        http.MethodPost,
		Path:          "/api/v1/organizations/{orgId}/invitations",
		Tags:          []string{"invitations"},
		Summary:       "Invite a user to join the organization",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createInvitation)

	huma.Register(api, huma.Operation{
		OperationID: "list-pending-invitations",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/invitations",
		Tags:        []string{"invitations"},
		Summary:     "List pending invitations for an organization",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listPendingInvitations)

	huma.Register(api, huma.Operation{
		OperationID: "resend-invitation",
		Method:      http.MethodPost,
		Path:        "/api/v1/organizations/{orgId}/invitations/{invitationId}/resend",
		Tags:        []string{"invitations"},
		Summary:     "Resend an invitation with a new token",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.resendInvitation)

	huma.Register(api, huma.Operation{
		OperationID:   "revoke-invitation",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}/invitations/{invitationId}",
		Tags:          []string{"invitations"},
		Summary:       "Revoke a pending invitation",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.revokeInvitation)

	// ----- invitations (user-scoped) -----

	huma.Register(api, huma.Operation{
		OperationID: "list-user-invitations",
		Method:      http.MethodGet,
		Path:        "/api/v1/invitations",
		Tags:        []string{"invitations"},
		Summary:     "List the authenticated user's pending invitations",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listUserInvitations)

	huma.Register(api, huma.Operation{
		OperationID: "validate-invitation-token",
		Method:      http.MethodGet,
		Path:        "/api/v1/invitations/validate/{token}",
		Tags:        []string{"invitations"},
		Summary:     "Validate an invitation token and return display details",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.validateInvitationToken)

	huma.Register(api, huma.Operation{
		OperationID:   "accept-invitation",
		Method:        http.MethodPost,
		Path:          "/api/v1/invitations/accept",
		Tags:          []string{"invitations"},
		Summary:       "Accept an invitation",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.acceptInvitation)

	huma.Register(api, huma.Operation{
		OperationID:   "decline-invitation",
		Method:        http.MethodPost,
		Path:          "/api/v1/invitations/decline",
		Tags:          []string{"invitations"},
		Summary:       "Decline an invitation",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.declineInvitation)

	// ----- settings -----

	huma.Register(api, huma.Operation{
		OperationID: "list-organization-settings",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/settings",
		Tags:        []string{"organization-settings"},
		Summary:     "Get all organization settings as key-value pairs",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listSettings)

	huma.Register(api, huma.Operation{
		OperationID:   "create-organization-setting",
		Method:        http.MethodPost,
		Path:          "/api/v1/organizations/{orgId}/settings",
		Tags:          []string{"organization-settings"},
		Summary:       "Create an organization setting",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createSetting)

	huma.Register(api, huma.Operation{
		OperationID: "get-organization-setting",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/settings/{key}",
		Tags:        []string{"organization-settings"},
		Summary:     "Get a single organization setting by key",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getSetting)

	huma.Register(api, huma.Operation{
		OperationID: "update-organization-setting",
		Method:      http.MethodPut,
		Path:        "/api/v1/organizations/{orgId}/settings/{key}",
		Tags:        []string{"organization-settings"},
		Summary:     "Update an organization setting",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateSetting)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-organization-setting",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}/settings/{key}",
		Tags:          []string{"organization-settings"},
		Summary:       "Delete an organization setting",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteSetting)
}

// ----- shared helpers --------------------------------------------------

func parseOrg(orgIDStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	return id, nil
}

func parseInvitation(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid invitation ID", "invitationId must be a valid UUID")
	}
	return id, nil
}

func parseUser(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	return id, nil
}

// ----- response DTOs ---------------------------------------------------

func toOrganizationResponse(o *organization.Organization) organizationResponse {
	return organizationResponse{
		ID:                 o.ID,
		Name:               o.Name,
		Plan:               o.Plan,
		SubscriptionStatus: o.SubscriptionStatus,
		BillingEmail:       o.BillingEmail,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	}
}

func toMemberResponse(m *organization.Member) memberResponse {
	return memberResponse{
		UserID:    m.UserID,
		RoleID:    m.RoleID,
		Status:    m.Status,
		InvitedBy: m.InvitedBy,
		JoinedAt:  m.JoinedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toInvitationResponse(inv *organization.Invitation) invitationResponse {
	return invitationResponse{
		ID:             inv.ID,
		OrganizationID: inv.OrganizationID,
		Email:          inv.Email,
		Status:         string(inv.Status),
		TokenPreview:   inv.TokenPreview,
		RoleID:         inv.RoleID,
		Role:           inv.Role,
		Inviter:        inv.Inviter,
		InvitedByID:    inv.InvitedByID,
		Message:        inv.Message,
		ResentCount:    inv.ResentCount,
		ResentAt:       inv.ResentAt,
		ExpiresAt:      inv.ExpiresAt,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,
	}
}

// ----- list-organizations ----------------------------------------------

func (h *handler) listOrganizations(ctx context.Context, in *ListOrganizationsInput) (*ListOrganizationsOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	orgs, err := h.orgSvc.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	page := in.Page
	if page < 1 {
		page = 1
	}
	limit := in.Limit
	if limit < 1 {
		limit = 20
	}
	sortDir := in.SortDir
	if sortDir == "" {
		sortDir = "desc"
	}

	filtered := make([]organizationResponse, 0, len(orgs))
	for _, o := range orgs {
		if in.Search != "" && !strings.Contains(strings.ToLower(o.Name), strings.ToLower(in.Search)) {
			continue
		}
		filtered = append(filtered, toOrganizationResponse(o))
	}

	sort.Slice(filtered, func(i, j int) bool {
		if sortDir == "asc" {
			return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)
	offset := (page - 1) * limit
	end := offset + limit
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}
	filtered = filtered[offset:end]

	return &ListOrganizationsOutput{Body: listOrganizationsBody{
		Data:       filtered,
		Pagination: response.BuildPagination(page, limit, int64(total)),
	}}, nil
}

// ----- create-organization ---------------------------------------------

func (h *handler) createOrganization(ctx context.Context, in *CreateOrganizationInput) (*CreateOrganizationOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	org, err := h.orgSvc.CreateOrganization(ctx, userID, &organization.CreateOrganizationRequest{
		Name: in.Body.Name,
	})
	if err != nil {
		return nil, err
	}
	return &CreateOrganizationOutput{Body: toOrganizationResponse(org)}, nil
}

// ----- get-organization ------------------------------------------------

func (h *handler) getOrganization(ctx context.Context, in *GetOrganizationInput) (*GetOrganizationOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	canAccess, err := h.memberSvc.CanUserAccessOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewForbiddenError("Insufficient permissions to access this organization")
	}

	org, err := h.orgSvc.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &GetOrganizationOutput{Body: toOrganizationResponse(org)}, nil
}

// ----- update-organization ---------------------------------------------

func (h *handler) updateOrganization(ctx context.Context, in *UpdateOrganizationInput) (*UpdateOrganizationOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	canAccess, err := h.memberSvc.CanUserAccessOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewForbiddenError("Insufficient permissions to update this organization")
	}

	if err := h.orgSvc.UpdateOrganization(ctx, orgID, &organization.UpdateOrganizationRequest{
		Name:         in.Body.Name,
		BillingEmail: in.Body.BillingEmail,
	}); err != nil {
		return nil, err
	}

	org, err := h.orgSvc.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &UpdateOrganizationOutput{Body: toOrganizationResponse(org)}, nil
}

// ----- delete-organization ---------------------------------------------

func (h *handler) deleteOrganization(ctx context.Context, in *DeleteOrganizationInput) (*DeleteOrganizationOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	canAccess, err := h.memberSvc.CanUserAccessOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewForbiddenError("Insufficient permissions to delete this organization")
	}
	if err := h.orgSvc.DeleteOrganization(ctx, orgID); err != nil {
		return nil, err
	}
	return &DeleteOrganizationOutput{}, nil
}

// ----- list-members ----------------------------------------------------

func (h *handler) listMembers(ctx context.Context, in *ListMembersInput) (*ListMembersOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	canAccess, err := h.memberSvc.CanUserAccessOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewForbiddenError("Insufficient permissions to view organization members")
	}

	members, err := h.memberSvc.GetMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	out := make([]memberResponse, 0, len(members))
	for _, m := range members {
		if in.Status != "" && m.Status != in.Status {
			continue
		}
		out = append(out, toMemberResponse(m))
	}
	return &ListMembersOutput{Body: listMembersBody{Members: out, Total: len(out)}}, nil
}

// ----- remove-member ---------------------------------------------------

func (h *handler) removeMember(ctx context.Context, in *RemoveMemberInput) (*RemoveMemberOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	targetUserID, err := parseUser(in.UserID)
	if err != nil {
		return nil, err
	}
	callerID := httpctx.MustGetUserID(ctx)

	canAccess, err := h.memberSvc.CanUserAccessOrganization(ctx, callerID, orgID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, appErrors.NewForbiddenError("Insufficient permissions to remove members from this organization")
	}
	if err := h.memberSvc.RemoveMember(ctx, orgID, targetUserID, callerID); err != nil {
		return nil, err
	}
	return &RemoveMemberOutput{}, nil
}

// ----- create-invitation -----------------------------------------------

func (h *handler) createInvitation(ctx context.Context, in *CreateInvitationInput) (*CreateInvitationOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	invitation, err := h.invitationSvc.InviteUser(ctx, orgID, userID, &organization.InviteUserRequest{
		Email:   strings.ToLower(strings.TrimSpace(in.Body.Email)),
		RoleID:  in.Body.RoleID,
		Message: in.Body.Message,
	})
	if err != nil {
		return nil, err
	}
	return &CreateInvitationOutput{Body: toInvitationResponse(invitation)}, nil
}

// ----- list-pending-invitations ----------------------------------------

func (h *handler) listPendingInvitations(ctx context.Context, in *ListPendingInvitationsInput) (*ListPendingInvitationsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	isMember, err := h.memberSvc.IsMember(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, appErrors.NewForbiddenError("You are not authorized to view this organization's invitations")
	}

	// inv.Inviter and inv.Role are pre-loaded via LEFT JOIN in the
	// repository — no per-row user/role lookups here.
	invitations, err := h.invitationSvc.GetPendingInvitations(ctx, orgID)
	if err != nil {
		return nil, err
	}
	resp := make([]invitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		resp = append(resp, toInvitationResponse(inv))
	}
	return &ListPendingInvitationsOutput{Body: listPendingInvitationsBody{
		Invitations: resp,
		Total:       len(resp),
	}}, nil
}

// ----- resend-invitation -----------------------------------------------

func (h *handler) resendInvitation(ctx context.Context, in *ResendInvitationInput) (*ResendInvitationOutput, error) {
	if _, err := parseOrg(in.OrgID); err != nil {
		return nil, err
	}
	invitationID, err := parseInvitation(in.InvitationID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	invitation, err := h.invitationSvc.ResendInvitation(ctx, invitationID, userID)
	if err != nil {
		return nil, err
	}
	return &ResendInvitationOutput{Body: toInvitationResponse(invitation)}, nil
}

// ----- revoke-invitation -----------------------------------------------

func (h *handler) revokeInvitation(ctx context.Context, in *RevokeInvitationInput) (*RevokeInvitationOutput, error) {
	if _, err := parseOrg(in.OrgID); err != nil {
		return nil, err
	}
	invitationID, err := parseInvitation(in.InvitationID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.invitationSvc.RevokeInvitation(ctx, invitationID, userID); err != nil {
		return nil, err
	}
	return &RevokeInvitationOutput{}, nil
}

// ----- list-user-invitations -------------------------------------------

func (h *handler) listUserInvitations(ctx context.Context, in *ListUserInvitationsInput) (*ListUserInvitationsOutput, error) {
	claims := httpctx.MustGetTokenClaims(ctx)

	// inv.Role and inv.Inviter are pre-loaded via LEFT JOIN.
	invitations, err := h.invitationSvc.GetUserInvitations(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	out := make([]userInvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		if inv.Status != organization.InvitationStatusPending {
			continue
		}
		if now.After(inv.ExpiresAt) {
			continue
		}

		orgName := ""
		if org, err := h.orgSvc.GetOrganization(ctx, inv.OrganizationID); err == nil {
			orgName = org.Name
		}

		out = append(out, userInvitationResponse{
			ID:               inv.ID,
			Email:            inv.Email,
			Status:           string(inv.Status),
			OrganizationID:   inv.OrganizationID,
			OrganizationName: orgName,
			Role:             inv.Role,
			Inviter:          inv.Inviter,
			Message:          inv.Message,
			ExpiresAt:        inv.ExpiresAt,
			CreatedAt:        inv.CreatedAt,
		})
	}
	return &ListUserInvitationsOutput{Body: listUserInvitationsBody{
		Invitations: out,
		Total:       len(out),
	}}, nil
}

// ----- validate-invitation-token ---------------------------------------

func (h *handler) validateInvitationToken(ctx context.Context, in *ValidateInvitationTokenInput) (*ValidateInvitationTokenOutput, error) {
	invitation, err := h.invitationSvc.GetInvitationByToken(ctx, in.Token)
	if err != nil {
		return nil, err
	}

	isExpired := time.Now().After(invitation.ExpiresAt) ||
		invitation.Status != organization.InvitationStatusPending
	if isExpired {
		// Closest AppError type for 410 Gone semantics; surfaces as 409.
		return nil, appErrors.NewConflictError("Invitation has expired or is no longer valid", appErrors.WithCode("invitation_expired"))
	}

	org, err := h.orgSvc.GetOrganization(ctx, invitation.OrganizationID)
	if err != nil {
		return nil, err
	}

	// Inviter and Role are hydration-only fields: single-row reads
	// (GetInvitationByToken) leave them nil. Fall through to sentinels
	// when absent — no per-request user/role service lookups.
	roleName := "Member"
	if invitation.Role != nil {
		roleName = invitation.Role.Name
	}
	inviterName := "Unknown"
	if invitation.Inviter != nil {
		inviterName = invitation.Inviter.FirstName
	}

	return &ValidateInvitationTokenOutput{Body: invitationDetailsBody{
		OrganizationID:   org.ID,
		OrganizationName: org.Name,
		Email:            invitation.Email,
		RoleName:         roleName,
		InviterName:      inviterName,
		ExpiresAt:        invitation.ExpiresAt,
		IsExpired:        isExpired,
	}}, nil
}

// ----- accept-invitation -----------------------------------------------

func (h *handler) acceptInvitation(ctx context.Context, in *AcceptInvitationInput) (*AcceptInvitationOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	if _, err := h.invitationSvc.AcceptInvitation(ctx, in.Body.Token, userID); err != nil {
		return nil, err
	}
	return &AcceptInvitationOutput{}, nil
}

// ----- decline-invitation ----------------------------------------------

func (h *handler) declineInvitation(ctx context.Context, in *DeclineInvitationInput) (*DeclineInvitationOutput, error) {
	if err := h.invitationSvc.DeclineInvitation(ctx, in.Body.Token); err != nil {
		return nil, err
	}
	return &DeclineInvitationOutput{}, nil
}

// ----- settings --------------------------------------------------------

func toSettingResponse(s *organization.OrganizationSettings) settingResponse {
	value, _ := s.GetValue()
	return settingResponse{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Key:            s.Key,
		Value:          value,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

// ----- list-settings ---------------------------------------------------

func (h *handler) listSettings(ctx context.Context, in *ListSettingsInput) (*ListSettingsOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	settings, err := h.settingsSvc.GetAllSettings(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &ListSettingsOutput{Body: listSettingsBody{Settings: settings}}, nil
}

// ----- create-setting --------------------------------------------------

func (h *handler) createSetting(ctx context.Context, in *CreateSettingInput) (*CreateSettingOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	setting, err := h.settingsSvc.CreateSetting(ctx, orgID, userID, &organization.CreateOrganizationSettingRequest{
		Key:   in.Body.Key,
		Value: in.Body.Value,
	})
	if err != nil {
		return nil, err
	}
	return &CreateSettingOutput{Body: toSettingResponse(setting)}, nil
}

// ----- get-setting -----------------------------------------------------

func (h *handler) getSetting(ctx context.Context, in *GetSettingInput) (*GetSettingOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	setting, err := h.settingsSvc.GetSetting(ctx, orgID, in.Key)
	if err != nil {
		return nil, err
	}
	return &GetSettingOutput{Body: toSettingResponse(setting)}, nil
}

// ----- update-setting --------------------------------------------------

func (h *handler) updateSetting(ctx context.Context, in *UpdateSettingInput) (*UpdateSettingOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	setting, err := h.settingsSvc.UpdateSetting(ctx, orgID, in.Key, userID, &organization.UpdateOrganizationSettingRequest{
		Value: in.Body.Value,
	})
	if err != nil {
		return nil, err
	}
	return &UpdateSettingOutput{Body: toSettingResponse(setting)}, nil
}

// ----- delete-setting --------------------------------------------------

func (h *handler) deleteSetting(ctx context.Context, in *DeleteSettingInput) (*DeleteSettingOutput, error) {
	orgID, err := parseOrg(in.OrgID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.settingsSvc.DeleteSetting(ctx, orgID, in.Key, userID); err != nil {
		return nil, err
	}
	return &DeleteSettingOutput{}, nil
}
