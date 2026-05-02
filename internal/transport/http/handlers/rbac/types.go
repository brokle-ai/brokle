package rbac

import (
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/pkg/response"
)

// ---- list-roles ------------------------------------------------------

type listRolesResponse struct {
	Data []*authDomain.Role `json:"data"`
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
	Data []*authDomain.Role `json:"data"`
}

type updateCustomRoleBody struct {
	Description   *string     `json:"description,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}

// ---- user memberships -------------------------------------------------

type getUserRolesResponse struct {
	Data []*authDomain.OrganizationMember `json:"data"`
}

type assignOrgRoleBody struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// ---- permission discovery --------------------------------------------

type getAvailableResourcesResponse struct {
	Data []string `json:"data"`
}

type getActionsForResourceResponse struct {
	Data []string `json:"data"`
}
