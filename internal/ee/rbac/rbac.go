package rbac

import (
	"context"
	"errors"
)

// RBACManager interface for enterprise role-based access control
type RBACManager interface {
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, roleID string, role *Role) error
	DeleteRole(ctx context.Context, roleID string) error
	GetRole(ctx context.Context, roleID string) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	CheckPermission(ctx context.Context, userID, resource, action string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

// Role represents an RBAC role
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
	Scopes      []string `json:"scopes"` // org, project, environment
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// Permission represents a permission with resource and action
type Permission struct {
	Resource string `json:"resource"` // e.g., "project", "environment", "api_key"
	Action   string `json:"action"`   // e.g., "read", "write", "delete", "admin"
	Scope    string `json:"scope"`    // e.g., "org:123", "project:456"
}

// StubRBAC provides stub implementation for OSS version
type StubRBAC struct{}

// New returns the RBAC manager implementation (stub or real based on build tags)
func New() RBACManager {
	return &StubRBAC{}
}

func (s *StubRBAC) CreateRole(ctx context.Context, role *Role) error {
	return errors.New("custom roles require Enterprise license")
}

func (s *StubRBAC) UpdateRole(ctx context.Context, roleID string, role *Role) error {
	return errors.New("custom roles require Enterprise license")
}

func (s *StubRBAC) DeleteRole(ctx context.Context, roleID string) error {
	return errors.New("custom roles require Enterprise license")
}

func (s *StubRBAC) GetRole(ctx context.Context, roleID string) (*Role, error) {
	return nil, errors.New("custom roles require Enterprise license")
}

func (s *StubRBAC) ListRoles(ctx context.Context) ([]*Role, error) {
	// Return basic OSS roles
	return []*Role{
		{ID: "owner", Name: "Owner", Description: "Full access to organization"},
		{ID: "admin", Name: "Admin", Description: "Administrative access"},
		{ID: "developer", Name: "Developer", Description: "Development access"},
		{ID: "viewer", Name: "Viewer", Description: "Read-only access"},
	}, nil
}

func (s *StubRBAC) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	// Basic role assignment works in OSS (owner, admin, developer, viewer)
	basicRoles := []string{"owner", "admin", "developer", "viewer"}
	for _, role := range basicRoles {
		if role == roleID {
			return nil // Allow basic role assignment
		}
	}
	return errors.New("custom roles require Enterprise license")
}

func (s *StubRBAC) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return nil // Allow role removal in OSS
}

func (s *StubRBAC) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Basic permission checking in OSS - always allow for simplicity
	return true, nil
}

func (s *StubRBAC) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// Return basic permissions
	return []string{"read", "write"}, nil
}