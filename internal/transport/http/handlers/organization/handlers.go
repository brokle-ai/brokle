// Package organization exposes /api/v1/organizations/* and
// /api/v1/invitations/* operations on the dashboard plane. Covers
// organization CRUD, member listing/removal, invitation lifecycle
// (create / list pending / resend / revoke / accept / decline /
// user's own), and organization-settings CRUD.
//
// Every route requires RequireAuth.
package organization

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"brokle/internal/core/domain/organization"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// OrganizationService / MemberService / InvitationService / SettingsService
// describe the narrow method sets the handler consumes (Go idiom: accept
// interfaces). Concrete *organizationService.* types from prod wiring
// satisfy these; tests inject fakes.
type OrganizationService interface {
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*organization.Organization, error)
	CreateOrganization(ctx context.Context, userID uuid.UUID, req *organization.CreateOrganizationRequest) (*organization.Organization, error)
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*organization.Organization, error)
	UpdateOrganization(ctx context.Context, orgID uuid.UUID, req *organization.UpdateOrganizationRequest) error
	DeleteOrganization(ctx context.Context, orgID uuid.UUID) error
}

type MemberService interface {
	CanUserAccessOrganization(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
	GetMembers(ctx context.Context, orgID uuid.UUID) ([]*organization.Member, error)
	RemoveMember(ctx context.Context, orgID, targetUserID, callerID uuid.UUID) error
	IsMember(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
}

type InvitationService interface {
	InviteUser(ctx context.Context, orgID, userID uuid.UUID, req *organization.InviteUserRequest) (*organization.Invitation, error)
	GetPendingInvitations(ctx context.Context, orgID uuid.UUID) ([]*organization.Invitation, error)
	ResendInvitation(ctx context.Context, invitationID, userID uuid.UUID) (*organization.Invitation, error)
	RevokeInvitation(ctx context.Context, invitationID, userID uuid.UUID) error
	GetUserInvitations(ctx context.Context, email string) ([]*organization.Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*organization.Invitation, error)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*organization.AcceptInvitationResult, error)
	DeclineInvitation(ctx context.Context, token string) error
}

type SettingsService interface {
	ListSettings(ctx context.Context, orgID uuid.UUID) (map[string]any, error)
	CreateSetting(ctx context.Context, orgID, userID uuid.UUID, req *organization.CreateOrganizationSettingRequest) (*organization.OrganizationSettings, error)
	GetSetting(ctx context.Context, orgID uuid.UUID, key string) (*organization.OrganizationSettings, error)
	UpdateSetting(ctx context.Context, orgID uuid.UUID, key string, userID uuid.UUID, req *organization.UpdateOrganizationSettingRequest) (*organization.OrganizationSettings, error)
	DeleteSetting(ctx context.Context, orgID uuid.UUID, key string, userID uuid.UUID) error
}

type handler struct {
	orgSvc        OrganizationService
	memberSvc     MemberService
	invitationSvc InvitationService
	settingsSvc   SettingsService
	logger        *slog.Logger
}

// RegisterRoutes mounts the organization / invitation / settings
// routes on r. Expected mount context: the authed dashboard chi group.
func RegisterRoutes(
	r chi.Router,
	orgSvc OrganizationService,
	memberSvc MemberService,
	invitationSvc InvitationService,
	settingsSvc SettingsService,
	logger *slog.Logger,
) {
	h := &handler{
		orgSvc:        orgSvc,
		memberSvc:     memberSvc,
		invitationSvc: invitationSvc,
		settingsSvc:   settingsSvc,
		logger:        logger,
	}

	r.Route("/api/v1/organizations", func(r chi.Router) {
		r.Get("/", h.listOrganizations)
		r.Post("/", h.createOrganization)
		r.Route("/{orgId}", func(r chi.Router) {
			r.Get("/", h.getOrganization)
			r.Patch("/", h.updateOrganization)
			r.Delete("/", h.deleteOrganization)
			r.Get("/members", h.listMembers)
			r.Delete("/members/{userId}", h.removeMember)
			r.Post("/invitations", h.createInvitation)
			r.Get("/invitations", h.listPendingInvitations)
			r.Post("/invitations/{invitationId}/resend", h.resendInvitation)
			r.Delete("/invitations/{invitationId}", h.revokeInvitation)
			r.Route("/settings", func(r chi.Router) {
				r.Get("/", h.listSettings)
				r.Post("/", h.createSetting)
				r.Get("/{key}", h.getSetting)
				r.Put("/{key}", h.updateSetting)
				r.Delete("/{key}", h.deleteSetting)
			})
		})
	})

	r.Route("/api/v1/invitations", func(r chi.Router) {
		r.Get("/", h.listUserInvitations)
		r.Get("/validate/{token}", h.validateInvitationToken)
		r.Post("/accept", h.acceptInvitation)
		r.Post("/decline", h.declineInvitation)
	})
}

// ----- response helpers -----------------------------------------------

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

// ----- list-organizations ---------------------------------------------

func (h *handler) listOrganizations(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	orgs, err := h.orgSvc.GetUserOrganizations(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	page, err := request.QueryInt(r, "page", 1)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if page < 1 {
		page = 1
	}
	limit, err := request.QueryInt(r, "limit", 20)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit < 1 {
		limit = 20
	}
	sortDir := r.URL.Query().Get("sort_dir")
	if sortDir == "" {
		sortDir = "desc"
	}
	search := r.URL.Query().Get("search")

	filtered := make([]organizationResponse, 0, len(orgs))
	for _, o := range orgs {
		if search != "" && !strings.Contains(strings.ToLower(o.Name), strings.ToLower(search)) {
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

	response.Success(w, listOrganizationsBody{
		Data:       filtered,
		Pagination: response.BuildPagination(page, limit, int64(total)),
	})
}

// ----- create-organization --------------------------------------------

func (h *handler) createOrganization(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body createOrganizationBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	org, err := h.orgSvc.CreateOrganization(r.Context(), userID, &organization.CreateOrganizationRequest{
		Name: body.Name,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, toOrganizationResponse(org))
}

// ----- get-organization -----------------------------------------------

func (h *handler) getOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	canAccess, err := h.memberSvc.CanUserAccessOrganization(r.Context(), userID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if !canAccess {
		response.WriteError(w, appErrors.NewForbiddenError("Insufficient permissions to access this organization"))
		return
	}

	org, err := h.orgSvc.GetOrganization(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toOrganizationResponse(org))
}

// ----- update-organization --------------------------------------------

func (h *handler) updateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	canAccess, err := h.memberSvc.CanUserAccessOrganization(r.Context(), userID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if !canAccess {
		response.WriteError(w, appErrors.NewForbiddenError("Insufficient permissions to update this organization"))
		return
	}

	var body updateOrganizationBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.orgSvc.UpdateOrganization(r.Context(), orgID, &organization.UpdateOrganizationRequest{
		Name:         body.Name,
		BillingEmail: body.BillingEmail,
	}); err != nil {
		response.WriteError(w, err)
		return
	}

	org, err := h.orgSvc.GetOrganization(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toOrganizationResponse(org))
}

// ----- delete-organization --------------------------------------------

func (h *handler) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	canAccess, err := h.memberSvc.CanUserAccessOrganization(r.Context(), userID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if !canAccess {
		response.WriteError(w, appErrors.NewForbiddenError("Insufficient permissions to delete this organization"))
		return
	}
	if err := h.orgSvc.DeleteOrganization(r.Context(), orgID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ----- list-members ---------------------------------------------------

func (h *handler) listMembers(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	canAccess, err := h.memberSvc.CanUserAccessOrganization(r.Context(), userID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if !canAccess {
		response.WriteError(w, appErrors.NewForbiddenError("Insufficient permissions to view organization members"))
		return
	}

	members, err := h.memberSvc.GetMembers(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	status := r.URL.Query().Get("status")
	out := make([]memberResponse, 0, len(members))
	for _, m := range members {
		if status != "" && m.Status != status {
			continue
		}
		out = append(out, toMemberResponse(m))
	}
	response.Success(w, listMembersBody{Members: out, Total: len(out)})
}

// ----- remove-member --------------------------------------------------

func (h *handler) removeMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	targetUserID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	callerID := httpctx.MustGetUserID(r.Context())

	canAccess, err := h.memberSvc.CanUserAccessOrganization(r.Context(), callerID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if !canAccess {
		response.WriteError(w, appErrors.NewForbiddenError("Insufficient permissions to remove members from this organization"))
		return
	}
	if err := h.memberSvc.RemoveMember(r.Context(), orgID, targetUserID, callerID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ----- create-invitation ----------------------------------------------

func (h *handler) createInvitation(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body createInvitationBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	invitation, err := h.invitationSvc.InviteUser(r.Context(), orgID, userID, &organization.InviteUserRequest{
		Email:   strings.ToLower(strings.TrimSpace(body.Email)),
		RoleID:  body.RoleID,
		Message: body.Message,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, toInvitationResponse(invitation))
}

// ----- list-pending-invitations ---------------------------------------

func (h *handler) listPendingInvitations(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	isMember, err := h.memberSvc.IsMember(r.Context(), userID, orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if !isMember {
		response.WriteError(w, appErrors.NewForbiddenError("You are not authorized to view this organization's invitations"))
		return
	}

	invitations, err := h.invitationSvc.GetPendingInvitations(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	resp := make([]invitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		resp = append(resp, toInvitationResponse(inv))
	}
	response.Success(w, listPendingInvitationsBody{
		Invitations: resp,
		Total:       len(resp),
	})
}

// ----- resend-invitation ----------------------------------------------

func (h *handler) resendInvitation(w http.ResponseWriter, r *http.Request) {
	if _, err := request.URLParamUUID(r, "orgId"); err != nil {
		response.WriteError(w, err)
		return
	}
	invitationID, err := request.URLParamUUID(r, "invitationId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	invitation, err := h.invitationSvc.ResendInvitation(r.Context(), invitationID, userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toInvitationResponse(invitation))
}

// ----- revoke-invitation ----------------------------------------------

func (h *handler) revokeInvitation(w http.ResponseWriter, r *http.Request) {
	if _, err := request.URLParamUUID(r, "orgId"); err != nil {
		response.WriteError(w, err)
		return
	}
	invitationID, err := request.URLParamUUID(r, "invitationId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.invitationSvc.RevokeInvitation(r.Context(), invitationID, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ----- list-user-invitations ------------------------------------------

func (h *handler) listUserInvitations(w http.ResponseWriter, r *http.Request) {
	claims := httpctx.MustGetTokenClaims(r.Context())

	invitations, err := h.invitationSvc.GetUserInvitations(r.Context(), claims.Email)
	if err != nil {
		response.WriteError(w, err)
		return
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
		if org, err := h.orgSvc.GetOrganization(r.Context(), inv.OrganizationID); err == nil {
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
	response.Success(w, listUserInvitationsBody{
		Invitations: out,
		Total:       len(out),
	})
}

// ----- validate-invitation-token --------------------------------------

func (h *handler) validateInvitationToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	invitation, err := h.invitationSvc.GetInvitationByToken(r.Context(), token)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	isExpired := time.Now().After(invitation.ExpiresAt) ||
		invitation.Status != organization.InvitationStatusPending
	if isExpired {
		response.WriteError(w, appErrors.NewConflictError(
			"Invitation has expired or is no longer valid",
			appErrors.WithCode("invitation_expired"),
		))
		return
	}

	org, err := h.orgSvc.GetOrganization(r.Context(), invitation.OrganizationID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	roleName := "Member"
	if invitation.Role != nil {
		roleName = invitation.Role.Name
	}
	inviterName := "Unknown"
	if invitation.Inviter != nil {
		inviterName = invitation.Inviter.FirstName
	}

	response.Success(w, invitationDetailsBody{
		OrganizationID:   org.ID,
		OrganizationName: org.Name,
		Email:            invitation.Email,
		RoleName:         roleName,
		InviterName:      inviterName,
		ExpiresAt:        invitation.ExpiresAt,
		IsExpired:        isExpired,
	})
}

// ----- accept-invitation ----------------------------------------------

func (h *handler) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	var body acceptInvitationBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if _, err := h.invitationSvc.AcceptInvitation(r.Context(), body.Token, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ----- decline-invitation ---------------------------------------------

func (h *handler) declineInvitation(w http.ResponseWriter, r *http.Request) {
	var body declineInvitationBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.invitationSvc.DeclineInvitation(r.Context(), body.Token); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ----- list-settings --------------------------------------------------

func (h *handler) listSettings(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	settings, err := h.settingsSvc.ListSettings(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, listSettingsBody{Settings: settings})
}

// ----- create-setting -------------------------------------------------

func (h *handler) createSetting(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body createSettingBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	setting, err := h.settingsSvc.CreateSetting(r.Context(), orgID, userID, &organization.CreateOrganizationSettingRequest{
		Key:   body.Key,
		Value: body.Value,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, toSettingResponse(setting))
}

// ----- get-setting ----------------------------------------------------

func (h *handler) getSetting(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	key := chi.URLParam(r, "key")
	setting, err := h.settingsSvc.GetSetting(r.Context(), orgID, key)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toSettingResponse(setting))
}

// ----- update-setting -------------------------------------------------

func (h *handler) updateSetting(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	key := chi.URLParam(r, "key")
	userID := httpctx.MustGetUserID(r.Context())

	var body updateSettingBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	setting, err := h.settingsSvc.UpdateSetting(r.Context(), orgID, key, userID, &organization.UpdateOrganizationSettingRequest{
		Value: body.Value,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toSettingResponse(setting))
}

// ----- delete-setting -------------------------------------------------

func (h *handler) deleteSetting(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	key := chi.URLParam(r, "key")
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.settingsSvc.DeleteSetting(r.Context(), orgID, key, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}
