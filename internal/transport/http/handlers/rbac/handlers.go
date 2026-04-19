// Package rbac is the dashboard-plane RBAC handler domain.
//
// Operation inventory (post-CLAUDE.md#14 audit):
//
//   - Platform-admin operations were removed. CreateRole (generic),
//     UpdateRole, DeleteRole, CreatePermission (generic), and the
//     legacy GetUserRole shim had no reachable caller — every role in
//     seeds/roles.yaml is organization-scoped and there is no
//     platform-admin role that could satisfy the guard. Kept in git
//     history if a real need ever surfaces.
//
//   - Custom-role lifecycle (B): org-scoped CRUD on roles owned by an
//     organization. Distinct from system roles.
//
//   - Read-only discovery + membership + user queries (C): role/permission
//     introspection, user-membership assignment, scope resolution.
package rbac

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	roleSvc      authDomain.RoleService
	permSvc      authDomain.PermissionService
	orgMemberSvc authDomain.OrganizationMemberService
	scopeSvc     authDomain.ScopeService
	logger       *slog.Logger
}

// RegisterRoutes registers every RBAC operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	roleSvc authDomain.RoleService,
	permSvc authDomain.PermissionService,
	orgMemberSvc authDomain.OrganizationMemberService,
	scopeSvc authDomain.ScopeService,
	logger *slog.Logger,
) {
	h := &handler{
		roleSvc:      roleSvc,
		permSvc:      permSvc,
		orgMemberSvc: orgMemberSvc,
		scopeSvc:     scopeSvc,
		logger:       logger,
	}

	// ----- role discovery ------------------------------------------------

	huma.Register(api, huma.Operation{
		OperationID: "list-roles",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/roles",
		Tags:        []string{"rbac"},
		Summary:     "List roles by scope type",
		Description: "Lists system-template roles for the given scope_type (system|organization|project|environment).",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listRoles)

	huma.Register(api, huma.Operation{
		OperationID: "get-role-statistics",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/roles/statistics",
		Tags:        []string{"rbac"},
		Summary:     "Aggregate role statistics across all scopes",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getRoleStatistics)

	huma.Register(api, huma.Operation{
		OperationID: "get-role",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/roles/{roleId}",
		Tags:        []string{"rbac"},
		Summary:     "Get a role by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getRole)

	// ----- org-scoped custom-role lifecycle ------------------------------

	huma.Register(api, huma.Operation{
		OperationID:   "create-custom-role",
		Method:        http.MethodPost,
		Path:          "/api/v1/organizations/{orgId}/roles",
		Tags:          []string{"rbac"},
		Summary:       "Create an organization-scoped custom role",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createCustomRole)

	huma.Register(api, huma.Operation{
		OperationID: "list-custom-roles",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/roles",
		Tags:        []string{"rbac"},
		Summary:     "List custom roles for an organization",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listCustomRoles)

	huma.Register(api, huma.Operation{
		OperationID: "get-custom-role",
		Method:      http.MethodGet,
		Path:        "/api/v1/organizations/{orgId}/roles/{roleId}",
		Tags:        []string{"rbac"},
		Summary:     "Get a custom role by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getCustomRole)

	huma.Register(api, huma.Operation{
		OperationID: "update-custom-role",
		Method:      http.MethodPatch,
		Path:        "/api/v1/organizations/{orgId}/roles/{roleId}",
		Tags:        []string{"rbac"},
		Summary:     "Update a custom role",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateCustomRole)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-custom-role",
		Method:        http.MethodDelete,
		Path:          "/api/v1/organizations/{orgId}/roles/{roleId}",
		Tags:          []string{"rbac"},
		Summary:       "Delete a custom role",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteCustomRole)

	// ----- user memberships / role assignment ----------------------------

	huma.Register(api, huma.Operation{
		OperationID: "get-user-roles",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/users/{userId}/roles",
		Tags:        []string{"rbac"},
		Summary:     "Get a user's organization memberships",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getUserRoles)

	huma.Register(api, huma.Operation{
		OperationID: "get-user-permissions",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/users/{userId}/permissions",
		Tags:        []string{"rbac"},
		Summary:     "Get a user's effective permissions across all scopes",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getUserPermissions)

	huma.Register(api, huma.Operation{
		OperationID:   "assign-organization-role",
		Method:        http.MethodPost,
		Path:          "/api/v1/rbac/users/{userId}/organizations/{orgId}/roles",
		Tags:          []string{"rbac"},
		Summary:       "Assign a role to a user in an organization",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.assignOrganizationRole)

	huma.Register(api, huma.Operation{
		OperationID:   "remove-organization-member",
		Method:        http.MethodDelete,
		Path:          "/api/v1/rbac/users/{userId}/organizations/{orgId}",
		Tags:          []string{"rbac"},
		Summary:       "Remove a user from an organization",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.removeOrganizationMember)

	huma.Register(api, huma.Operation{
		OperationID: "check-user-permissions",
		Method:      http.MethodPost,
		Path:        "/api/v1/rbac/users/{userId}/permissions/check",
		Tags:        []string{"rbac"},
		Summary:     "Check whether a user holds a set of resource:action permissions",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.checkUserPermissions)

	// ----- permission discovery ------------------------------------------

	huma.Register(api, huma.Operation{
		OperationID: "list-permissions",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/permissions",
		Tags:        []string{"rbac"},
		Summary:     "List permissions (paginated)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listPermissions)

	huma.Register(api, huma.Operation{
		OperationID: "get-available-resources",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/permissions/resources",
		Tags:        []string{"rbac"},
		Summary:     "List resources that can have permissions assigned",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getAvailableResources)

	huma.Register(api, huma.Operation{
		OperationID: "get-actions-for-resource",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/permissions/resources/{resource}/actions",
		Tags:        []string{"rbac"},
		Summary:     "List available actions for a given resource",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getActionsForResource)

	huma.Register(api, huma.Operation{
		OperationID: "get-permission",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/permissions/{permissionId}",
		Tags:        []string{"rbac"},
		Summary:     "Get a permission by ID",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getPermission)

	// ----- scopes --------------------------------------------------------

	huma.Register(api, huma.Operation{
		OperationID: "check-user-scopes",
		Method:      http.MethodPost,
		Path:        "/api/v1/rbac/users/{userId}/scopes/check",
		Tags:        []string{"rbac"},
		Summary:     "Check whether a user holds a set of scopes in the given org/project context",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.checkUserScopes)

	huma.Register(api, huma.Operation{
		OperationID: "get-user-scopes",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/users/{userId}/scopes",
		Tags:        []string{"rbac"},
		Summary:     "Resolve a user's effective scopes in the given org/project context",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getUserScopes)

	huma.Register(api, huma.Operation{
		OperationID: "get-scope-categories",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/scopes/categories",
		Tags:        []string{"rbac"},
		Summary:     "List scope categories (for UI grouping)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getScopeCategories)

	huma.Register(api, huma.Operation{
		OperationID: "get-available-scopes",
		Method:      http.MethodGet,
		Path:        "/api/v1/rbac/scopes",
		Tags:        []string{"rbac"},
		Summary:     "List available scopes (optionally filtered by level)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getAvailableScopes)
}

// ============================================================================
// Role discovery (C)
// ============================================================================

type ListRolesInput struct {
	ScopeType string `query:"scope_type" required:"true" enum:"system,organization,project,environment" doc:"Scope-type filter"`
}
type ListRolesOutput struct {
	Body listRolesResponse
}
type listRolesResponse struct {
	Roles      []*authDomain.Role `json:"roles"`
	TotalCount int                `json:"total_count"`
}

func (h *handler) listRoles(ctx context.Context, in *ListRolesInput) (*ListRolesOutput, error) {
	if in.ScopeType == "" {
		return nil, appErrors.NewValidationError("Scope type is required", "scope_type parameter cannot be empty")
	}
	userID := httpctx.MustGetUserID(ctx)
	roles, err := h.roleSvc.GetRolesByScopeType(ctx, in.ScopeType)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: list roles failed", "user_id", userID, "scope_type", in.ScopeType, "error", err)
		return nil, err
	}
	return &ListRolesOutput{Body: listRolesResponse{Roles: roles, TotalCount: len(roles)}}, nil
}

type GetRoleInput struct {
	RoleID string `path:"roleId" format:"uuid" doc:"Role ID"`
}
type GetRoleOutput struct {
	Body *authDomain.Role
}

func (h *handler) getRole(ctx context.Context, in *GetRoleInput) (*GetRoleOutput, error) {
	roleID, err := uuid.Parse(in.RoleID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid role ID", "roleId must be a valid UUID")
	}
	role, err := h.roleSvc.GetRoleByID(ctx, roleID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get role failed", "role_id", roleID, "error", err)
		return nil, err
	}
	return &GetRoleOutput{Body: role}, nil
}

type GetRoleStatisticsInput struct{}
type GetRoleStatisticsOutput struct {
	Body *authDomain.RoleStatistics
}

func (h *handler) getRoleStatistics(ctx context.Context, _ *GetRoleStatisticsInput) (*GetRoleStatisticsOutput, error) {
	stats, err := h.roleSvc.GetRoleStatistics(ctx)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get role statistics failed", "error", err)
		return nil, err
	}
	return &GetRoleStatisticsOutput{Body: stats}, nil
}

// ============================================================================
// Custom-role lifecycle (B)
// ============================================================================

type createCustomRoleBody struct {
	Name          string      `json:"name" minLength:"1" maxLength:"100"`
	Description   string      `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}
type CreateCustomRoleInput struct {
	OrgID string `path:"orgId" format:"uuid" doc:"Organization that owns the custom role"`
	Body  createCustomRoleBody
}
type CreateCustomRoleOutput struct {
	Body *authDomain.Role
}

func (h *handler) createCustomRole(ctx context.Context, in *CreateCustomRoleInput) (*CreateCustomRoleOutput, error) {
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	req := &authDomain.CreateRoleRequest{
		ScopeType:     authDomain.ScopeOrganization,
		Name:          in.Body.Name,
		Description:   in.Body.Description,
		PermissionIDs: in.Body.PermissionIDs,
	}
	role, err := h.roleSvc.CreateCustomRole(ctx, authDomain.ScopeOrganization, orgID, req)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: create custom role failed", "user_id", userID, "org_id", orgID, "role_name", req.Name, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "rbac: custom role created", "user_id", userID, "org_id", orgID, "role_id", role.ID)
	return &CreateCustomRoleOutput{Body: role}, nil
}

type ListCustomRolesInput struct {
	OrgID string `path:"orgId" format:"uuid" doc:"Organization that owns the custom roles"`
}
type ListCustomRolesOutput struct {
	Body listCustomRolesResponse
}
type listCustomRolesResponse struct {
	Roles      []*authDomain.Role `json:"roles"`
	TotalCount int                `json:"total_count"`
}

func (h *handler) listCustomRoles(ctx context.Context, in *ListCustomRolesInput) (*ListCustomRolesOutput, error) {
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	roles, err := h.roleSvc.GetCustomRolesByOrganization(ctx, orgID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: list custom roles failed", "org_id", orgID, "error", err)
		return nil, err
	}
	return &ListCustomRolesOutput{Body: listCustomRolesResponse{Roles: roles, TotalCount: len(roles)}}, nil
}

type GetCustomRoleInput struct {
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization that owns the custom role"`
	RoleID string `path:"roleId" format:"uuid" doc:"Custom role ID"`
}
type GetCustomRoleOutput struct {
	Body *authDomain.Role
}

func (h *handler) getCustomRole(ctx context.Context, in *GetCustomRoleInput) (*GetCustomRoleOutput, error) {
	if _, err := uuid.Parse(in.OrgID); err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	roleID, err := uuid.Parse(in.RoleID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid role ID", "roleId must be a valid UUID")
	}
	role, err := h.roleSvc.GetRoleByID(ctx, roleID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get custom role failed", "role_id", roleID, "error", err)
		return nil, err
	}
	if role.IsSystemRole() {
		return nil, appErrors.NewValidationError("Cannot access system role through custom-role endpoint", "use /api/v1/rbac/roles/{roleId} for system roles")
	}
	return &GetCustomRoleOutput{Body: role}, nil
}

type updateCustomRoleBody struct {
	Description   *string     `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}
type UpdateCustomRoleInput struct {
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization that owns the custom role"`
	RoleID string `path:"roleId" format:"uuid" doc:"Custom role ID"`
	Body   updateCustomRoleBody
}
type UpdateCustomRoleOutput struct {
	Body *authDomain.Role
}

func (h *handler) updateCustomRole(ctx context.Context, in *UpdateCustomRoleInput) (*UpdateCustomRoleOutput, error) {
	if _, err := uuid.Parse(in.OrgID); err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	roleID, err := uuid.Parse(in.RoleID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid role ID", "roleId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	req := &authDomain.UpdateRoleRequest{
		Description:   in.Body.Description,
		PermissionIDs: in.Body.PermissionIDs,
	}
	role, err := h.roleSvc.UpdateCustomRole(ctx, roleID, req)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: update custom role failed", "user_id", userID, "role_id", roleID, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "rbac: custom role updated", "user_id", userID, "role_id", role.ID)
	return &UpdateCustomRoleOutput{Body: role}, nil
}

type DeleteCustomRoleInput struct {
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization that owns the custom role"`
	RoleID string `path:"roleId" format:"uuid" doc:"Custom role ID"`
}
type DeleteCustomRoleOutput struct{}

func (h *handler) deleteCustomRole(ctx context.Context, in *DeleteCustomRoleInput) (*DeleteCustomRoleOutput, error) {
	if _, err := uuid.Parse(in.OrgID); err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	roleID, err := uuid.Parse(in.RoleID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid role ID", "roleId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.roleSvc.DeleteCustomRole(ctx, roleID); err != nil {
		h.logger.WarnContext(ctx, "rbac: delete custom role failed", "user_id", userID, "role_id", roleID, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "rbac: custom role deleted", "user_id", userID, "role_id", roleID)
	return &DeleteCustomRoleOutput{}, nil
}

// ============================================================================
// User memberships / role assignment (C)
// ============================================================================

type GetUserRolesInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User ID"`
}
type GetUserRolesOutput struct {
	Body getUserRolesResponse
}
type getUserRolesResponse struct {
	Memberships []*authDomain.OrganizationMember `json:"memberships"`
	TotalCount  int                              `json:"total_count"`
}

func (h *handler) getUserRoles(ctx context.Context, in *GetUserRolesInput) (*GetUserRolesOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	memberships, err := h.orgMemberSvc.GetUserMemberships(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get user memberships failed", "user_id", userID, "error", err)
		return nil, err
	}
	return &GetUserRolesOutput{Body: getUserRolesResponse{Memberships: memberships, TotalCount: len(memberships)}}, nil
}

type GetUserPermissionsInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User ID"`
}
type GetUserPermissionsOutput struct {
	Body getUserPermissionsResponse
}
type getUserPermissionsResponse struct {
	Permissions []string `json:"permissions"`
	TotalCount  int      `json:"total_count"`
}

func (h *handler) getUserPermissions(ctx context.Context, in *GetUserPermissionsInput) (*GetUserPermissionsOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	perms, err := h.orgMemberSvc.GetUserEffectivePermissions(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get user permissions failed", "user_id", userID, "error", err)
		return nil, err
	}
	return &GetUserPermissionsOutput{Body: getUserPermissionsResponse{Permissions: perms, TotalCount: len(perms)}}, nil
}

type assignOrgRoleBody struct {
	RoleID uuid.UUID `json:"role_id"`
}
type AssignOrganizationRoleInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User receiving the role"`
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization the role applies within"`
	Body   assignOrgRoleBody
}
type AssignOrganizationRoleOutput struct {
	Body *authDomain.OrganizationMember
}

func (h *handler) assignOrganizationRole(ctx context.Context, in *AssignOrganizationRoleInput) (*AssignOrganizationRoleOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	if in.Body.RoleID == uuid.Nil {
		return nil, appErrors.NewValidationError("Invalid role ID", "role_id is required")
	}
	inviter := httpctx.MustGetUserID(ctx)

	member, err := h.orgMemberSvc.AddMember(ctx, userID, orgID, in.Body.RoleID, nil)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: assign org role failed", "inviter_id", inviter, "user_id", userID, "org_id", orgID, "role_id", in.Body.RoleID, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "rbac: org role assigned", "inviter_id", inviter, "user_id", userID, "org_id", orgID, "role_id", in.Body.RoleID)
	return &AssignOrganizationRoleOutput{Body: member}, nil
}

type RemoveOrganizationMemberInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User to remove"`
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization to remove from"`
}
type RemoveOrganizationMemberOutput struct{}

func (h *handler) removeOrganizationMember(ctx context.Context, in *RemoveOrganizationMemberInput) (*RemoveOrganizationMemberOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "orgId must be a valid UUID")
	}
	actor := httpctx.MustGetUserID(ctx)

	if err := h.orgMemberSvc.RemoveMember(ctx, userID, orgID); err != nil {
		h.logger.WarnContext(ctx, "rbac: remove org member failed", "actor_id", actor, "user_id", userID, "org_id", orgID, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "rbac: org member removed", "actor_id", actor, "user_id", userID, "org_id", orgID)
	return &RemoveOrganizationMemberOutput{}, nil
}

type checkUserPermissionsBody struct {
	ResourceActions []string `json:"resource_actions" minItems:"1" doc:"List of resource:action strings to check"`
}
type CheckUserPermissionsInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User ID"`
	Body   checkUserPermissionsBody
}
type CheckUserPermissionsOutput struct {
	Body checkUserPermissionsResponse
}
type checkUserPermissionsResponse struct {
	Results map[string]bool `json:"results"`
}

func (h *handler) checkUserPermissions(ctx context.Context, in *CheckUserPermissionsInput) (*CheckUserPermissionsOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	if len(in.Body.ResourceActions) == 0 {
		return nil, appErrors.NewValidationError("Resource actions are required", "resource_actions must contain at least one entry")
	}
	results, err := h.orgMemberSvc.CheckUserPermissions(ctx, userID, in.Body.ResourceActions)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: check user permissions failed", "user_id", userID, "error", err)
		return nil, err
	}
	return &CheckUserPermissionsOutput{Body: checkUserPermissionsResponse{Results: results}}, nil
}

// ============================================================================
// Permission discovery (C)
// ============================================================================

type ListPermissionsInput struct {
	Limit  int `query:"limit" required:"false" minimum:"1" maximum:"100" doc:"Page size (default 50, max 100)"`
	Offset int `query:"offset" required:"false" minimum:"0" doc:"Offset, 0-indexed"`
}
type ListPermissionsOutput struct {
	Body *authDomain.PermissionListResponse
}

func (h *handler) listPermissions(ctx context.Context, in *ListPermissionsInput) (*ListPermissionsOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}
	resp, err := h.permSvc.ListPermissions(ctx, limit, offset)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: list permissions failed", "error", err)
		return nil, err
	}
	return &ListPermissionsOutput{Body: resp}, nil
}

type GetPermissionInput struct {
	PermissionID string `path:"permissionId" format:"uuid" doc:"Permission ID"`
}
type GetPermissionOutput struct {
	Body *authDomain.Permission
}

func (h *handler) getPermission(ctx context.Context, in *GetPermissionInput) (*GetPermissionOutput, error) {
	permissionID, err := uuid.Parse(in.PermissionID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid permission ID", "permissionId must be a valid UUID")
	}
	perm, err := h.permSvc.GetPermission(ctx, permissionID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get permission failed", "permission_id", permissionID, "error", err)
		return nil, err
	}
	return &GetPermissionOutput{Body: perm}, nil
}

type GetAvailableResourcesInput struct{}
type GetAvailableResourcesOutput struct {
	Body getAvailableResourcesResponse
}
type getAvailableResourcesResponse struct {
	Resources  []string `json:"resources"`
	TotalCount int      `json:"total_count"`
}

func (h *handler) getAvailableResources(ctx context.Context, _ *GetAvailableResourcesInput) (*GetAvailableResourcesOutput, error) {
	resources, err := h.permSvc.GetAvailableResources(ctx)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get available resources failed", "error", err)
		return nil, err
	}
	return &GetAvailableResourcesOutput{Body: getAvailableResourcesResponse{Resources: resources, TotalCount: len(resources)}}, nil
}

type GetActionsForResourceInput struct {
	Resource string `path:"resource" minLength:"1" doc:"Resource name"`
}
type GetActionsForResourceOutput struct {
	Body getActionsForResourceResponse
}
type getActionsForResourceResponse struct {
	Resource   string   `json:"resource"`
	Actions    []string `json:"actions"`
	TotalCount int      `json:"total_count"`
}

func (h *handler) getActionsForResource(ctx context.Context, in *GetActionsForResourceInput) (*GetActionsForResourceOutput, error) {
	if in.Resource == "" {
		return nil, appErrors.NewValidationError("Resource parameter is required", "resource parameter cannot be empty")
	}
	actions, err := h.permSvc.GetActionsForResource(ctx, in.Resource)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get actions for resource failed", "resource", in.Resource, "error", err)
		return nil, err
	}
	return &GetActionsForResourceOutput{Body: getActionsForResourceResponse{Resource: in.Resource, Actions: actions, TotalCount: len(actions)}}, nil
}

// ============================================================================
// Scopes (C)
// ============================================================================

type checkUserScopesBody struct {
	OrganizationID *string  `json:"organization_id,omitempty" format:"uuid"`
	ProjectID      *string  `json:"project_id,omitempty" format:"uuid"`
	Scopes         []string `json:"scopes" minItems:"1" doc:"Scope strings to check"`
}
type CheckUserScopesInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User ID"`
	Body   checkUserScopesBody
}
type CheckUserScopesOutput struct {
	Body checkUserScopesResponse
}
type checkUserScopesResponse struct {
	Results map[string]bool `json:"results"`
}

func (h *handler) checkUserScopes(ctx context.Context, in *CheckUserScopesInput) (*CheckUserScopesOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	if len(in.Body.Scopes) == 0 {
		return nil, appErrors.NewValidationError("Scopes are required", "scopes must contain at least one entry")
	}

	var orgID *uuid.UUID
	if in.Body.OrganizationID != nil && *in.Body.OrganizationID != "" {
		parsed, err := uuid.Parse(*in.Body.OrganizationID)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid organization ID", "organization_id must be a valid UUID")
		}
		orgID = &parsed
	}
	var projectID *uuid.UUID
	if in.Body.ProjectID != nil && *in.Body.ProjectID != "" {
		parsed, err := uuid.Parse(*in.Body.ProjectID)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid project ID", "project_id must be a valid UUID")
		}
		projectID = &parsed
	}

	results := make(map[string]bool, len(in.Body.Scopes))
	for _, scope := range in.Body.Scopes {
		has, err := h.scopeSvc.HasScope(ctx, userID, scope, orgID, projectID)
		if err != nil {
			h.logger.WarnContext(ctx, "rbac: has-scope failed", "user_id", userID, "scope", scope, "error", err)
			results[scope] = false
			continue
		}
		results[scope] = has
	}
	return &CheckUserScopesOutput{Body: checkUserScopesResponse{Results: results}}, nil
}

type GetUserScopesInput struct {
	UserID         string `path:"userId" format:"uuid" doc:"User ID"`
	OrganizationID string `query:"organization_id" required:"false" format:"uuid"`
	ProjectID      string `query:"project_id" required:"false" format:"uuid"`
}
type GetUserScopesOutput struct {
	Body *authDomain.ScopeResolution
}

func (h *handler) getUserScopes(ctx context.Context, in *GetUserScopesInput) (*GetUserScopesOutput, error) {
	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	var orgID *uuid.UUID
	if in.OrganizationID != "" {
		parsed, err := uuid.Parse(in.OrganizationID)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid organization ID", "organization_id must be a valid UUID")
		}
		orgID = &parsed
	}
	var projectID *uuid.UUID
	if in.ProjectID != "" {
		parsed, err := uuid.Parse(in.ProjectID)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid project ID", "project_id must be a valid UUID")
		}
		projectID = &parsed
	}
	resolution, err := h.scopeSvc.GetUserScopes(ctx, userID, orgID, projectID)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get user scopes failed", "user_id", userID, "error", err)
		return nil, err
	}
	return &GetUserScopesOutput{Body: resolution}, nil
}

type GetScopeCategoriesInput struct{}
type GetScopeCategoriesOutput struct {
	Body getScopeCategoriesResponse
}
type getScopeCategoriesResponse struct {
	Categories []authDomain.ScopeCategory `json:"categories"`
	TotalCount int                        `json:"total_count"`
}

func (h *handler) getScopeCategories(ctx context.Context, _ *GetScopeCategoriesInput) (*GetScopeCategoriesOutput, error) {
	categories, err := h.scopeSvc.GetScopesByCategory(ctx)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get scope categories failed", "error", err)
		return nil, err
	}
	return &GetScopeCategoriesOutput{Body: getScopeCategoriesResponse{Categories: categories, TotalCount: len(categories)}}, nil
}

type GetAvailableScopesInput struct {
	Level string `query:"level" required:"false" enum:"organization,project,global" doc:"Scope level filter"`
}
type GetAvailableScopesOutput struct {
	Body getAvailableScopesResponse
}
type getAvailableScopesResponse struct {
	Level      string   `json:"level,omitempty"`
	Scopes     []string `json:"scopes"`
	TotalCount int      `json:"total_count"`
}

func (h *handler) getAvailableScopes(ctx context.Context, in *GetAvailableScopesInput) (*GetAvailableScopesOutput, error) {
	var level authDomain.ScopeLevel
	if in.Level != "" {
		level = authDomain.ScopeLevel(in.Level)
		if level != authDomain.ScopeLevelOrganization && level != authDomain.ScopeLevelProject && level != authDomain.ScopeLevelGlobal {
			return nil, appErrors.NewValidationError("Invalid scope level", "level must be 'organization', 'project', or 'global'")
		}
	}
	scopes, err := h.scopeSvc.GetAvailableScopes(ctx, level)
	if err != nil {
		h.logger.WarnContext(ctx, "rbac: get available scopes failed", "level", level, "error", err)
		return nil, err
	}
	return &GetAvailableScopesOutput{Body: getAvailableScopesResponse{Level: string(level), Scopes: scopes, TotalCount: len(scopes)}}, nil
}
