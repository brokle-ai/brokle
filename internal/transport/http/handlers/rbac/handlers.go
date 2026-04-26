// Package rbac is the dashboard-plane RBAC handler domain.
//
// Operation surface:
//   - Role discovery + statistics
//   - Org-scoped custom-role lifecycle (CRUD on roles owned by an
//     organization; distinct from system roles)
//   - User membership + role assignment
//   - Permission discovery
//   - Scope introspection
package rbac

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	authService "brokle/internal/core/services/auth"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	roleSvc      *authService.RoleService
	permSvc      *authService.PermissionService
	orgMemberSvc *authService.OrganizationMemberService
	scopeSvc     *authService.ScopeService
	logger       *slog.Logger
}

// New constructs a Handler with all required services.
func New(
	roleSvc *authService.RoleService,
	permSvc *authService.PermissionService,
	orgMemberSvc *authService.OrganizationMemberService,
	scopeSvc *authService.ScopeService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		roleSvc:      roleSvc,
		permSvc:      permSvc,
		orgMemberSvc: orgMemberSvc,
		scopeSvc:     scopeSvc,
		logger:       logger,
	}
}

// ---- role discovery --------------------------------------------------

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	scopeType := r.URL.Query().Get("scope_type")
	if scopeType == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"Scope type is required", "scope_type parameter cannot be empty",
			appErrors.WithParam("scope_type"),
		))
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	roles, err := h.roleSvc.GetRolesByScopeType(r.Context(), scopeType)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: list roles failed",
			"user_id", userID, "scope_type", scopeType, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, listRolesResponse{Roles: roles, TotalCount: len(roles)})
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := request.URLParamUUID(r, "roleId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	role, err := h.roleSvc.GetRoleByID(r.Context(), roleID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get role failed",
			"role_id", roleID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, role)
}

func (h *Handler) GetRoleStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.roleSvc.GetRoleStatistics(r.Context())
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get role statistics failed",
			"error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, stats)
}

// ---- custom-role lifecycle -------------------------------------------

func (h *Handler) CreateCustomRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body createCustomRoleBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	req := &authDomain.CreateRoleRequest{
		ScopeType:     authDomain.ScopeOrganization,
		Name:          body.Name,
		Description:   body.Description,
		PermissionIDs: body.PermissionIDs,
	}
	role, err := h.roleSvc.CreateCustomRole(r.Context(), authDomain.ScopeOrganization, orgID, req)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: create custom role failed",
			"user_id", userID, "org_id", orgID, "role_name", req.Name, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "rbac: custom role created",
		"user_id", userID, "org_id", orgID, "role_id", role.ID)
	response.Created(w, role)
}

func (h *Handler) ListCustomRoles(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	roles, err := h.roleSvc.GetCustomRolesByOrganization(r.Context(), orgID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: list custom roles failed",
			"org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, listCustomRolesResponse{Roles: roles, TotalCount: len(roles)})
}

func (h *Handler) GetCustomRole(w http.ResponseWriter, r *http.Request) {
	if _, err := request.URLParamUUID(r, "orgId"); err != nil {
		response.WriteError(w, err)
		return
	}
	roleID, err := request.URLParamUUID(r, "roleId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	role, err := h.roleSvc.GetRoleByID(r.Context(), roleID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get custom role failed",
			"role_id", roleID, "error", err)
		response.WriteError(w, err)
		return
	}
	if role.IsSystemRole() {
		response.WriteError(w, appErrors.NewValidationError(
			"Cannot access system role through custom-role endpoint",
			"use /api/v1/rbac/roles/{roleId} for system roles",
		))
		return
	}
	response.Success(w, role)
}

func (h *Handler) UpdateCustomRole(w http.ResponseWriter, r *http.Request) {
	if _, err := request.URLParamUUID(r, "orgId"); err != nil {
		response.WriteError(w, err)
		return
	}
	roleID, err := request.URLParamUUID(r, "roleId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body updateCustomRoleBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	req := &authDomain.UpdateRoleRequest{
		Description:   body.Description,
		PermissionIDs: body.PermissionIDs,
	}
	role, err := h.roleSvc.UpdateCustomRole(r.Context(), roleID, req)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: update custom role failed",
			"user_id", userID, "role_id", roleID, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "rbac: custom role updated",
		"user_id", userID, "role_id", role.ID)
	response.Success(w, role)
}

func (h *Handler) DeleteCustomRole(w http.ResponseWriter, r *http.Request) {
	if _, err := request.URLParamUUID(r, "orgId"); err != nil {
		response.WriteError(w, err)
		return
	}
	roleID, err := request.URLParamUUID(r, "roleId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.roleSvc.DeleteCustomRole(r.Context(), roleID); err != nil {
		h.logger.WarnContext(r.Context(), "rbac: delete custom role failed",
			"user_id", userID, "role_id", roleID, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "rbac: custom role deleted",
		"user_id", userID, "role_id", roleID)
	response.NoContent(w)
}

// ---- user memberships -------------------------------------------------

func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	memberships, err := h.orgMemberSvc.GetUserMemberships(r.Context(), userID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get user memberships failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, getUserRolesResponse{Memberships: memberships, TotalCount: len(memberships)})
}

func (h *Handler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	perms, err := h.orgMemberSvc.GetUserEffectivePermissions(r.Context(), userID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get user permissions failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, getUserPermissionsResponse{Permissions: perms, TotalCount: len(perms)})
}

func (h *Handler) AssignOrganizationRole(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body assignOrgRoleBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.RoleID == uuid.Nil {
		response.WriteError(w, appErrors.NewValidationError(
			"Invalid role ID", "role_id is required",
			appErrors.WithParam("role_id"),
		))
		return
	}
	inviter := httpctx.MustGetUserID(r.Context())

	member, err := h.orgMemberSvc.AddMember(r.Context(), userID, orgID, body.RoleID, nil)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: assign org role failed",
			"inviter_id", inviter, "user_id", userID, "org_id", orgID,
			"role_id", body.RoleID, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "rbac: org role assigned",
		"inviter_id", inviter, "user_id", userID, "org_id", orgID,
		"role_id", body.RoleID)
	response.Created(w, member)
}

func (h *Handler) RemoveOrganizationMember(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	orgID, err := request.URLParamUUID(r, "orgId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	actor := httpctx.MustGetUserID(r.Context())

	if err := h.orgMemberSvc.RemoveMember(r.Context(), userID, orgID); err != nil {
		h.logger.WarnContext(r.Context(), "rbac: remove org member failed",
			"actor_id", actor, "user_id", userID, "org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "rbac: org member removed",
		"actor_id", actor, "user_id", userID, "org_id", orgID)
	response.NoContent(w)
}

func (h *Handler) CheckUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body checkUserPermissionsBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if len(body.ResourceActions) == 0 {
		response.WriteError(w, appErrors.NewValidationError(
			"Resource actions are required",
			"resource_actions must contain at least one entry",
			appErrors.WithParam("resource_actions"),
		))
		return
	}

	results, err := h.orgMemberSvc.CheckUserPermissions(r.Context(), userID, body.ResourceActions)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: check user permissions failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, checkUserPermissionsResponse{Results: results})
}

// ---- permission discovery --------------------------------------------

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	limit, err := request.QueryInt(r, "limit", 50)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset, err := request.QueryInt(r, "offset", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := h.permSvc.ListPermissions(r.Context(), limit, offset)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: list permissions failed", "error", err)
		response.WriteError(w, err)
		return
	}
	page := offset/limit + 1
	response.Success(w, listPermissionsBody{
		Data:       items,
		Pagination: response.BuildPagination(page, limit, total),
	})
}

func (h *Handler) GetPermission(w http.ResponseWriter, r *http.Request) {
	permissionID, err := request.URLParamUUID(r, "permissionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	perm, err := h.permSvc.GetPermission(r.Context(), permissionID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get permission failed",
			"permission_id", permissionID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, perm)
}

func (h *Handler) GetAvailableResources(w http.ResponseWriter, r *http.Request) {
	resources, err := h.permSvc.GetAvailableResources(r.Context())
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get available resources failed",
			"error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, getAvailableResourcesResponse{Resources: resources, TotalCount: len(resources)})
}

func (h *Handler) GetActionsForResource(w http.ResponseWriter, r *http.Request) {
	resource := chi.URLParam(r, "resource")
	if resource == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"Resource parameter is required", "resource parameter cannot be empty",
			appErrors.WithParam("resource"),
		))
		return
	}
	actions, err := h.permSvc.GetActionsForResource(r.Context(), resource)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get actions for resource failed",
			"resource", resource, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, getActionsForResourceResponse{
		Resource: resource, Actions: actions, TotalCount: len(actions),
	})
}

// ---- scopes ----------------------------------------------------------

func (h *Handler) CheckUserScopes(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body checkUserScopesBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if len(body.Scopes) == 0 {
		response.WriteError(w, appErrors.NewValidationError(
			"Scopes are required", "scopes must contain at least one entry",
			appErrors.WithParam("scopes"),
		))
		return
	}

	var orgID *uuid.UUID
	if body.OrganizationID != nil && *body.OrganizationID != "" {
		parsed, err := uuid.Parse(*body.OrganizationID)
		if err != nil {
			response.WriteError(w, appErrors.NewValidationError(
				"Invalid organization ID",
				"organization_id must be a valid UUID",
				appErrors.WithParam("organization_id"),
			))
			return
		}
		orgID = &parsed
	}
	var projectID *uuid.UUID
	if body.ProjectID != nil && *body.ProjectID != "" {
		parsed, err := uuid.Parse(*body.ProjectID)
		if err != nil {
			response.WriteError(w, appErrors.NewValidationError(
				"Invalid project ID",
				"project_id must be a valid UUID",
				appErrors.WithParam("project_id"),
			))
			return
		}
		projectID = &parsed
	}

	results := make(map[string]bool, len(body.Scopes))
	for _, scope := range body.Scopes {
		has, err := h.scopeSvc.HasScope(r.Context(), userID, scope, orgID, projectID)
		if err != nil {
			h.logger.WarnContext(r.Context(), "rbac: has-scope failed",
				"user_id", userID, "scope", scope, "error", err)
			results[scope] = false
			continue
		}
		results[scope] = has
	}
	response.Success(w, checkUserScopesResponse{Results: results})
}

func (h *Handler) GetUserScopes(w http.ResponseWriter, r *http.Request) {
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	orgIDPtr, err := request.QueryOptionalUUID(r, "organization_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	projectIDPtr, err := request.QueryOptionalUUID(r, "project_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	resolution, err := h.scopeSvc.GetUserScopes(r.Context(), userID, orgIDPtr, projectIDPtr)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get user scopes failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, resolution)
}

func (h *Handler) GetScopeCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.scopeSvc.GetScopesByCategory(r.Context())
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get scope categories failed",
			"error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, getScopeCategoriesResponse{
		Categories: categories, TotalCount: len(categories),
	})
}

func (h *Handler) GetAvailableScopes(w http.ResponseWriter, r *http.Request) {
	var level authDomain.ScopeLevel
	if raw := r.URL.Query().Get("level"); raw != "" {
		level = authDomain.ScopeLevel(raw)
		if level != authDomain.ScopeLevelOrganization &&
			level != authDomain.ScopeLevelProject &&
			level != authDomain.ScopeLevelGlobal {
			response.WriteError(w, appErrors.NewValidationError(
				"Invalid scope level",
				"level must be 'organization', 'project', or 'global'",
				appErrors.WithParam("level"),
			))
			return
		}
	}
	scopes, err := h.scopeSvc.GetAvailableScopes(r.Context(), level)
	if err != nil {
		h.logger.WarnContext(r.Context(), "rbac: get available scopes failed",
			"level", level, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, getAvailableScopesResponse{
		Level: string(level), Scopes: scopes, TotalCount: len(scopes),
	})
}
