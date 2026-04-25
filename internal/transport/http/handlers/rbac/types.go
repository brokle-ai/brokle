package rbac

import (
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/pkg/response"
)

// ---- list-roles ------------------------------------------------------

type listRolesResponse struct {
	Roles      []*authDomain.Role `json:"roles"`
	TotalCount int                `json:"total_count"`
}

// ---- list-permissions (canonical {data, pagination} envelope) --------

type listPermissionsBody struct {
	Data       []*authDomain.Permission `json:"data"`
	Pagination *response.Pagination     `json:"pagination"`
}

// ---- custom-role lifecycle -------------------------------------------

type createCustomRoleBody struct {
	Name          string      `json:"name"                    validate:"required,min=1,max=100"`
	Description   string      `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}

type listCustomRolesResponse struct {
	Roles      []*authDomain.Role `json:"roles"`
	TotalCount int                `json:"total_count"`
}

type updateCustomRoleBody struct {
	Description   *string     `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}

// ---- user memberships -------------------------------------------------

type getUserRolesResponse struct {
	Memberships []*authDomain.OrganizationMember `json:"memberships"`
	TotalCount  int                              `json:"total_count"`
}

type getUserPermissionsResponse struct {
	Permissions []string `json:"permissions"`
	TotalCount  int      `json:"total_count"`
}

type assignOrgRoleBody struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type checkUserPermissionsBody struct {
	ResourceActions []string `json:"resource_actions" validate:"required,min=1,dive,required"`
}

type checkUserPermissionsResponse struct {
	Results map[string]bool `json:"results"`
}

// ---- permission discovery --------------------------------------------

type getAvailableResourcesResponse struct {
	Resources  []string `json:"resources"`
	TotalCount int      `json:"total_count"`
}

type getActionsForResourceResponse struct {
	Resource   string   `json:"resource"`
	Actions    []string `json:"actions"`
	TotalCount int      `json:"total_count"`
}

// ---- scopes -----------------------------------------------------------

type checkUserScopesBody struct {
	OrganizationID *string  `json:"organization_id,omitempty"`
	ProjectID      *string  `json:"project_id,omitempty"`
	Scopes         []string `json:"scopes"                    validate:"required,min=1,dive,required"`
}

type checkUserScopesResponse struct {
	Results map[string]bool `json:"results"`
}

type getScopeCategoriesResponse struct {
	Categories []authDomain.ScopeCategory `json:"categories"`
	TotalCount int                        `json:"total_count"`
}

type getAvailableScopesResponse struct {
	Level      string   `json:"level,omitempty"`
	Scopes     []string `json:"scopes"`
	TotalCount int      `json:"total_count"`
}
