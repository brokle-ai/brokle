package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// PermissionService implements auth.PermissionService interface
type PermissionService struct {
	permissionRepo authDomain.PermissionRepository
	rolePermRepo   authDomain.RolePermissionRepository
}

// NewPermissionService creates a new permission service instance
func NewPermissionService(
	permissionRepo authDomain.PermissionRepository,
	rolePermRepo authDomain.RolePermissionRepository,
) *PermissionService {
	return &PermissionService{
		permissionRepo: permissionRepo,
		rolePermRepo:   rolePermRepo,
	}
}

// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(ctx context.Context, req *authDomain.CreatePermissionRequest) (*authDomain.Permission, error) {
	// Validate permission doesn't already exist
	existing, err := s.permissionRepo.GetByResourceAction(ctx, req.Resource, req.Action)
	if err == nil && existing != nil {
		return nil, appErrors.Conflict("", "permission " + req.Resource + ":" + req.Action + " already exists")
	}

	// Create permission
	permission := authDomain.NewPermission(req.Resource, req.Action, req.Description)
	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, appErrors.Internal("failed to create permission", err)
	}

	return permission, nil
}

// GetPermission retrieves a permission by ID
func (s *PermissionService) GetPermission(ctx context.Context, permissionID uuid.UUID) (*authDomain.Permission, error) {
	return s.permissionRepo.GetByID(ctx, permissionID)
}

// GetPermissionByName retrieves a permission by legacy name
func (s *PermissionService) GetPermissionByName(ctx context.Context, name string) (*authDomain.Permission, error) {
	return s.permissionRepo.GetByName(ctx, name)
}

// GetPermissionByResourceAction retrieves a permission by resource:action
func (s *PermissionService) GetPermissionByResourceAction(ctx context.Context, resource, action string) (*authDomain.Permission, error) {
	return s.permissionRepo.GetByResourceAction(ctx, resource, action)
}

// UpdatePermission updates a permission
func (s *PermissionService) UpdatePermission(ctx context.Context, permissionID uuid.UUID, req *authDomain.UpdatePermissionRequest) error {
	// Get existing permission
	permission, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return appErrors.NotFound("permission")
	}

	// Update fields
	if req.Description != nil {
		permission.Description = req.Description
	}

	return s.permissionRepo.Update(ctx, permission)
}

// DeletePermission deletes a permission
func (s *PermissionService) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	// Check if permission is in use
	// This would need to be implemented properly in a production system
	return s.permissionRepo.Delete(ctx, permissionID)
}

// ListPermissions returns a paginated slice of permissions plus the total
// count. Handler-layer code wraps the result into the canonical
// {data, pagination} envelope.
func (s *PermissionService) ListPermissions(ctx context.Context, limit, offset int) ([]*authDomain.Permission, int64, error) {
	permissions, err := s.permissionRepo.GetAllPermissions(ctx)
	if err != nil {
		return nil, 0, appErrors.Internal("failed to list permissions", err)
	}

	total := int64(len(permissions))

	// In-memory pagination — fine for the RBAC permission set (small, bounded).
	start := offset
	end := offset + limit
	if start > len(permissions) {
		start = len(permissions)
	}
	if end > len(permissions) {
		end = len(permissions)
	}
	return permissions[start:end], total, nil
}

// GetAllPermissions returns all permissions
func (s *PermissionService) GetAllPermissions(ctx context.Context) ([]*authDomain.Permission, error) {
	return s.permissionRepo.GetAllPermissions(ctx)
}

// GetPermissionsByResource returns all permissions for a resource
func (s *PermissionService) GetPermissionsByResource(ctx context.Context, resource string) ([]*authDomain.Permission, error) {
	return s.permissionRepo.GetByResource(ctx, resource)
}

// GetPermissionsByNames returns permissions by legacy names
func (s *PermissionService) GetPermissionsByNames(ctx context.Context, names []string) ([]*authDomain.Permission, error) {
	permissions := make([]*authDomain.Permission, 0, len(names))
	for _, name := range names {
		perm, err := s.permissionRepo.GetByName(ctx, name)
		if err != nil {
			return nil, appErrors.NotFound("permission", appErrors.WithMessage("permission " + name + " not found"))
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// GetPermissionsByResourceActions returns permissions by resource:action format
func (s *PermissionService) GetPermissionsByResourceActions(ctx context.Context, resourceActions []string) ([]*authDomain.Permission, error) {
	permissions := make([]*authDomain.Permission, 0, len(resourceActions))
	for _, resourceAction := range resourceActions {
		resource, action, err := s.ParseResourceAction(resourceAction)
		if err != nil {
			return nil, appErrors.InvalidParam("resource_action", "Invalid resource:action format: "+resourceAction)
		}

		perm, err := s.permissionRepo.GetByResourceAction(ctx, resource, action)
		if err != nil {
			return nil, appErrors.NotFound("permission", appErrors.WithMessage("permission " + resourceAction + " not found"))
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

// SearchPermissions returns a filtered + paginated slice of permissions
// plus the total matching count. Handler-layer code wraps the result
// into the canonical {data, pagination} envelope.
func (s *PermissionService) SearchPermissions(ctx context.Context, query string, limit, offset int) ([]*authDomain.Permission, int64, error) {
	// Basic search implementation - in production this would be done at DB level
	allPermissions, err := s.permissionRepo.GetAllPermissions(ctx)
	if err != nil {
		return nil, 0, appErrors.Internal("failed to search permissions", err)
	}

	// Filter permissions by query
	filtered := make([]*authDomain.Permission, 0)
	q := strings.ToLower(query)
	for _, perm := range allPermissions {
		desc := ""
		if perm.Description != nil {
			desc = *perm.Description
		}
		if strings.Contains(strings.ToLower(perm.Name), q) ||
			strings.Contains(strings.ToLower(desc), q) ||
			strings.Contains(strings.ToLower(perm.Resource), q) ||
			strings.Contains(strings.ToLower(perm.Action), q) {
			filtered = append(filtered, perm)
		}
	}

	total := int64(len(filtered))
	start := offset
	end := offset + limit
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

// GetAvailableResources returns all distinct resources
func (s *PermissionService) GetAvailableResources(ctx context.Context) ([]string, error) {
	return s.permissionRepo.GetAvailableResources(ctx)
}

// GetActionsForResource returns all actions for a resource
func (s *PermissionService) GetActionsForResource(ctx context.Context, resource string) ([]string, error) {
	return s.permissionRepo.GetActionsForResource(ctx, resource)
}

// ValidatePermissionName validates legacy permission name
func (s *PermissionService) ValidatePermissionName(ctx context.Context, name string) error {
	if !strings.Contains(name, ".") {
		return appErrors.InvalidParam("name", "Invalid permission name format: "+name+" (must contain dot)")
	}
	return nil
}

// ValidateResourceAction validates resource:action format
func (s *PermissionService) ValidateResourceAction(ctx context.Context, resource, action string) error {
	if resource == "" || action == "" {
		return appErrors.InvalidParam("resource_action", "Resource and action cannot be empty")
	}
	return nil
}

// PermissionExists checks if a resource:action permission exists
func (s *PermissionService) PermissionExists(ctx context.Context, resource, action string) (bool, error) {
	_, err := s.permissionRepo.GetByResourceAction(ctx, resource, action)
	if err != nil {
		if appErrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BulkPermissionExists checks if multiple resource:action permissions exist
func (s *PermissionService) BulkPermissionExists(ctx context.Context, resourceActions []string) (map[string]bool, error) {
	results := make(map[string]bool)
	for _, resourceAction := range resourceActions {
		resource, action, err := s.ParseResourceAction(resourceAction)
		if err != nil {
			results[resourceAction] = false
			continue
		}

		exists, err := s.PermissionExists(ctx, resource, action)
		if err != nil {
			results[resourceAction] = false
			continue
		}
		results[resourceAction] = exists
	}
	return results, nil
}

// ParseResourceAction parses resource:action format
func (s *PermissionService) ParseResourceAction(resourceAction string) (resource, action string, err error) {
	parts := strings.Split(resourceAction, ":")
	if len(parts) != 2 {
		return "", "", appErrors.InvalidParam("resource_action", "Invalid resource:action format: "+resourceAction)
	}
	return parts[0], parts[1], nil
}

// FormatResourceAction formats resource and action into resource:action
func (s *PermissionService) FormatResourceAction(resource, action string) string {
	return fmt.Sprintf("%s:%s", resource, action)
}

// IsValidResourceActionFormat checks if string is valid resource:action format
func (s *PermissionService) IsValidResourceActionFormat(resourceAction string) bool {
	parts := strings.Split(resourceAction, ":")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}
