package rbac

import (
	authDomain "brokle/internal/core/domain/auth"

	"github.com/google/uuid"
)

// Huma operation types for the rbac package.

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

type GetRoleInput struct {
	RoleID string `path:"roleId" format:"uuid" doc:"Role ID"`
}

type GetRoleOutput struct {
	Body *authDomain.Role
}

type GetRoleStatisticsInput struct{}

type GetRoleStatisticsOutput struct {
	Body *authDomain.RoleStatistics
}

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

type GetCustomRoleInput struct {
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization that owns the custom role"`
	RoleID string `path:"roleId" format:"uuid" doc:"Custom role ID"`
}

type GetCustomRoleOutput struct {
	Body *authDomain.Role
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

type DeleteCustomRoleInput struct {
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization that owns the custom role"`
	RoleID string `path:"roleId" format:"uuid" doc:"Custom role ID"`
}

type DeleteCustomRoleOutput struct{}

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

type RemoveOrganizationMemberInput struct {
	UserID string `path:"userId" format:"uuid" doc:"User to remove"`
	OrgID  string `path:"orgId" format:"uuid" doc:"Organization to remove from"`
}

type RemoveOrganizationMemberOutput struct{}

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

type ListPermissionsInput struct {
	Limit  int `query:"limit" required:"false" minimum:"1" maximum:"100" doc:"Page size (default 50, max 100)"`
	Offset int `query:"offset" required:"false" minimum:"0" doc:"Offset, 0-indexed"`
}

type ListPermissionsOutput struct {
	Body *authDomain.PermissionListResponse
}

type GetPermissionInput struct {
	PermissionID string `path:"permissionId" format:"uuid" doc:"Permission ID"`
}

type GetPermissionOutput struct {
	Body *authDomain.Permission
}

type GetAvailableResourcesInput struct{}

type GetAvailableResourcesOutput struct {
	Body getAvailableResourcesResponse
}

type getAvailableResourcesResponse struct {
	Resources  []string `json:"resources"`
	TotalCount int      `json:"total_count"`
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

type GetUserScopesInput struct {
	UserID         string `path:"userId" format:"uuid" doc:"User ID"`
	OrganizationID string `query:"organization_id" required:"false" format:"uuid"`
	ProjectID      string `query:"project_id" required:"false" format:"uuid"`
}

type GetUserScopesOutput struct {
	Body *authDomain.ScopeResolution
}

type GetScopeCategoriesInput struct{}

type GetScopeCategoriesOutput struct {
	Body getScopeCategoriesResponse
}

type getScopeCategoriesResponse struct {
	Categories []authDomain.ScopeCategory `json:"categories"`
	TotalCount int                        `json:"total_count"`
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
