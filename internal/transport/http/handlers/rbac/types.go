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

type assignOrgRoleBody struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
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
