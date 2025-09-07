// Package auth provides the authentication and authorization domain model.
//
// The auth domain handles JWT tokens, sessions, API keys, roles, permissions,
// and role-based access control (RBAC) across the platform.
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"brokle/pkg/ulid"
	"gorm.io/gorm"
)

// UserSession represents an active user session with secure token management.
// SECURITY: Access tokens are NOT stored - only session metadata and hashed refresh tokens.
type UserSession struct {
	ID                   ulid.ULID   `json:"id" gorm:"type:char(26);primaryKey"`
	UserID               ulid.ULID   `json:"user_id" gorm:"type:char(26);not null;index:idx_user_sessions_user_active,priority:1"`
	
	// Secure Token Management (NO ACCESS TOKENS STORED)
	RefreshTokenHash     string      `json:"-" gorm:"type:char(64);not null;uniqueIndex"`             // SHA-256 hash = 64 hex chars
	RefreshTokenVersion  int         `json:"refresh_token_version" gorm:"default:1;not null"`         // For rotation tracking
	CurrentJTI           string      `json:"-" gorm:"type:char(26);not null;index"`                  // Current access token JTI for blacklisting
	
	// Session Metadata
	ExpiresAt            time.Time   `json:"expires_at" gorm:"not null;index"`                        // Access token expiry
	RefreshExpiresAt     time.Time   `json:"refresh_expires_at" gorm:"not null;index"`               // Refresh token expiry
	IPAddress            *string     `json:"ip_address,omitempty" gorm:"type:inet;index"`             // PostgreSQL inet type
	UserAgent            *string     `json:"user_agent,omitempty" gorm:"type:text"`
	DeviceInfo           interface{} `json:"device_info,omitempty" gorm:"type:jsonb"`                 // Device information JSON
	
	// Session State
	IsActive             bool        `json:"is_active" gorm:"default:true;not null;index:idx_user_sessions_user_active,priority:2"`
	LastUsedAt           *time.Time  `json:"last_used_at,omitempty" gorm:"index"`
	RevokedAt            *time.Time  `json:"revoked_at,omitempty" gorm:"index"`
	
	CreatedAt            time.Time   `json:"created_at" gorm:"not null"`
	UpdatedAt            time.Time   `json:"updated_at" gorm:"not null"`
}

// BlacklistedToken represents a revoked access token for immediate revocation capability.
type BlacklistedToken struct {
	JTI       string    `json:"jti" gorm:"type:char(26);primaryKey"`                     // JWT ID (ULID format)
	UserID    ulid.ULID `json:"user_id" gorm:"type:char(26);not null;index"`            // Owner user
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`                       // Token expiry for cleanup
	RevokedAt time.Time `json:"revoked_at" gorm:"not null;default:CURRENT_TIMESTAMP"`   // When revoked
	Reason    string    `json:"reason" gorm:"type:varchar(100);not null"`               // logout, suspicious_activity, etc.
	
	// New fields for user-wide timestamp blacklisting (GDPR/SOC2 compliance)
	TokenType          string `json:"token_type" gorm:"type:varchar(50);not null;default:'individual';index"`  // individual, user_wide_timestamp
	BlacklistTimestamp *int64 `json:"blacklist_timestamp,omitempty" gorm:"index"`                             // Unix timestamp for user-wide blacklisting
	
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
}

// SessionStats represents session statistics and metrics
type SessionStats struct {
	ActiveSessions   int64 `json:"active_sessions"`
	ExpiredSessions  int64 `json:"expired_sessions"`
	TotalSessions    int64 `json:"total_sessions"`
	SessionsToday    int64 `json:"sessions_today"`
	SessionsThisWeek int64 `json:"sessions_this_week"`
	AvgSessionLength int64 `json:"avg_session_length_minutes"`
}

// External repository interfaces to avoid circular imports
// These will be implemented by the actual user and organization repositories

// UserRepository defines the interface for user data access needed by auth services
type UserRepository interface {
	GetByID(ctx context.Context, id ulid.ULID) (interface{}, error)
	GetByEmail(ctx context.Context, email string) (interface{}, error)
	UpdateLastLogin(ctx context.Context, id ulid.ULID) error
}

// OrganizationRepository defines the interface for organization data access needed by auth services  
type OrganizationRepository interface {
	GetByID(ctx context.Context, id ulid.ULID) (interface{}, error)
	IsMember(ctx context.Context, userID, orgID ulid.ULID) (bool, error)
}

// APIKey represents an API key for programmatic access with full scoping.
type APIKey struct {
	ID             ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	UserID         ulid.ULID  `json:"user_id" gorm:"type:char(26);not null"`
	OrganizationID ulid.ULID  `json:"organization_id" gorm:"type:char(26);not null"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty" gorm:"type:char(26)"`
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty" gorm:"type:char(26)"`
	Name           string     `json:"name" gorm:"size:255;not null"`
	KeyPrefix      string     `json:"key_prefix" gorm:"size:8;not null"`       // First 8 chars for display
	KeyHash        string      `json:"-" gorm:"size:255;not null"`                  // Hashed key for storage
	Scopes         []string    `json:"scopes" gorm:"type:json"`               // JSON array of permissions
	RateLimitRPM   int         `json:"rate_limit_rpm" gorm:"default:1000"` // Requests per minute
	Metadata       interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`   // Flexible metadata storage
	IsActive       bool        `json:"is_active" gorm:"default:true"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Role represents a role with permissions (supports both global system roles and organization-specific custom roles).
type Role struct {
	ID             ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty" gorm:"type:char(26)"` // NULL for global system roles
	Name           string     `json:"name" gorm:"size:50;not null"`
	DisplayName    string     `json:"display_name" gorm:"size:100;not null"`
	Description    string     `json:"description" gorm:"type:text"`
	IsSystemRole   bool       `json:"is_system_role" gorm:"default:false"` // true for global system roles (cannot be deleted)
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Permissions     []Permission     `json:"permissions,omitempty" gorm:"many2many:role_permissions"`
	RolePermissions []RolePermission `json:"role_permissions,omitempty" gorm:"foreignKey:RoleID"`
}

// IsGlobalSystemRole returns true if this is a global system role (organization_id is NULL)
func (r *Role) IsGlobalSystemRole() bool {
	return r.IsSystemRole && r.OrganizationID == nil
}

// IsOrganizationRole returns true if this is an organization-specific role
func (r *Role) IsOrganizationRole() bool {
	return r.OrganizationID != nil
}

// CanBeDeleted returns true if this role can be deleted (not a system role)
func (r *Role) CanBeDeleted() bool {
	return !r.IsSystemRole
}

// GetScope returns the scope of the role ("global" or organization ID)
func (r *Role) GetScope() string {
	if r.IsGlobalSystemRole() {
		return "global"
	}
	if r.OrganizationID != nil {
		return r.OrganizationID.String()
	}
	return "unknown"
}

// Permission represents a specific permission in the system using resource:action format.
type Permission struct {
	ID          ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	Name        string    `json:"name" gorm:"size:255;not null;uniqueIndex"` // Legacy: users.create, projects.read (kept for compatibility)
	Resource    string    `json:"resource" gorm:"size:50;not null;index"` // users, projects, billing, etc.
	Action      string    `json:"action" gorm:"size:50;not null;index"` // create, read, update, delete, admin, etc.
	DisplayName string    `json:"display_name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Category    string    `json:"category" gorm:"size:100"` // users, projects, billing, etc.

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Roles []Role `json:"roles,omitempty" gorm:"many2many:role_permissions"`
}

// GetResourceAction returns the resource:action format string
func (p *Permission) GetResourceAction() string {
	return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}

// IsWildcardPermission returns true if this is a wildcard permission (*:* or resource:*)
func (p *Permission) IsWildcardPermission() bool {
	return p.Resource == "*" || p.Action == "*"
}

// MatchesResourceAction checks if this permission matches the given resource:action
func (p *Permission) MatchesResourceAction(resource, action string) bool {
	// Exact match
	if p.Resource == resource && p.Action == action {
		return true
	}
	// Wildcard resource match
	if p.Resource == "*" && p.Action == action {
		return true
	}
	// Wildcard action match  
	if p.Resource == resource && p.Action == "*" {
		return true
	}
	// Full wildcard match
	if p.Resource == "*" && p.Action == "*" {
		return true
	}
	return false
}

// ParseResourceAction parses a resource:action string into resource and action components
func ParseResourceAction(resourceAction string) (resource, action string, err error) {
	parts := strings.Split(resourceAction, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource:action format: %s", resourceAction)
	}
	return parts[0], parts[1], nil
}

// ValidateResourceAction validates a resource:action string format
func ValidateResourceAction(resourceAction string) error {
	_, _, err := ParseResourceAction(resourceAction)
	return err
}

// RolePermission represents the many-to-many relationship between roles and permissions.
type RolePermission struct {
	RoleID       ulid.ULID `json:"role_id" gorm:"type:char(26);not null;primaryKey"`
	PermissionID ulid.ULID `json:"permission_id" gorm:"type:char(26);not null;primaryKey"`
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Role       Role       `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

// AuditLog represents an audit log entry for compliance.
type AuditLog struct {
	ID             ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	UserID         *ulid.ULID `json:"user_id,omitempty" gorm:"type:char(26)"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty" gorm:"type:char(26)"`
	Action         string     `json:"action" gorm:"size:255;not null"`
	Resource       string     `json:"resource" gorm:"size:255"`
	ResourceID     string     `json:"resource_id" gorm:"size:255"`
	Metadata       string     `json:"metadata" gorm:"type:jsonb"`
	IPAddress      string     `json:"ip_address" gorm:"size:45"`
	UserAgent      string     `json:"user_agent" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
}

// Request/Response DTOs
type LoginRequest struct {
	Email      string                 `json:"email" validate:"required,email"`
	Password   string                 `json:"password" validate:"required"`
	Remember   bool                   `json:"remember"` // Extend session duration
	DeviceInfo map[string]interface{} `json:"device_info,omitempty"` // Device information for session tracking
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"` // Always "Bearer"
	ExpiresIn    int64  `json:"expires_in"` // Seconds until expiration
}

type AuthUser struct {
	ID                    ulid.ULID  `json:"id"`
	Email                 string     `json:"email"`
	Name                  string     `json:"name"`
	AvatarURL             *string    `json:"avatar_url,omitempty"`
	IsEmailVerified       bool       `json:"is_email_verified"`
	OnboardingCompleted   bool       `json:"onboarding_completed"`
	DefaultOrganizationID *ulid.ULID `json:"default_organization_id,omitempty"`
}

type CreateAPIKeyRequest struct {
	Name           string     `json:"name" validate:"required,min=1,max=100"`
	OrganizationID ulid.ULID  `json:"organization_id" validate:"required"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty"`
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty"`
	Scopes         []string   `json:"scopes" validate:"required,min=1"`
	RateLimitRPM   int        `json:"rate_limit_rpm" validate:"min=1,max=10000"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID        ulid.ULID  `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`        // Full key - only returned once
	KeyPrefix string     `json:"key_prefix"` // For display purposes
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}



// AuthContext represents the authenticated context for a request.
type AuthContext struct {
	UserID         ulid.ULID  `json:"user_id"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	Role           *string    `json:"role,omitempty"`
	Permissions    []string   `json:"permissions"`
	APIKeyID       *ulid.ULID `json:"api_key_id,omitempty"` // Set if authenticated via API key
	SessionID      *ulid.ULID `json:"session_id,omitempty"` // Set if authenticated via session
}

// Standard permission scopes for the platform
var StandardPermissions = []string{
	// User management
	"users.read",
	"users.write",
	"users.delete",
	"users.admin",

	// Organization management
	"organizations.read",
	"organizations.write",
	"organizations.delete",
	"organizations.admin",

	// Project management
	"projects.read",
	"projects.write",
	"projects.delete",
	"projects.admin",

	// Environment management
	"environments.read",
	"environments.write",
	"environments.delete",
	"environments.admin",

	// API key management
	"api_keys.read",
	"api_keys.write",
	"api_keys.delete",
	"api_keys.admin",

	// Role management
	"roles.read",
	"roles.write",
	"roles.delete",
	"roles.admin",

	// Billing management
	"billing.read",
	"billing.write",
	"billing.admin",

	// System administration
	"system.admin",
	"audit_logs.read",
}

// Blacklisted token types
const (
	TokenTypeIndividual      = "individual"        // Individual JTI-based blacklisting (default)
	TokenTypeUserTimestamp   = "user_wide_timestamp" // User-wide timestamp blacklisting (GDPR/SOC2)
)

// System roles that are pre-defined
var SystemRoles = map[string][]string{
	"owner": {
		"users.admin", "organizations.admin", "projects.admin", 
		"environments.admin", "api_keys.admin", "roles.admin", 
		"billing.admin", "audit_logs.read",
	},
	"admin": {
		"users.read", "users.write", "organizations.read", "organizations.write",
		"projects.admin", "environments.admin", "api_keys.admin", 
		"roles.read", "roles.write", "billing.read",
	},
	"developer": {
		"projects.read", "projects.write", "environments.read", "environments.write",
		"api_keys.read", "api_keys.write",
	},
	"viewer": {
		"projects.read", "environments.read", "api_keys.read",
	},
}

// Constructor functions
func NewUserSession(userID ulid.ULID, refreshTokenHash string, currentJTI string, expiresAt, refreshExpiresAt time.Time, ipAddress, userAgent *string, deviceInfo interface{}) *UserSession {
	return &UserSession{
		ID:                   ulid.New(),
		UserID:               userID,
		RefreshTokenHash:     refreshTokenHash,
		RefreshTokenVersion:  1,
		CurrentJTI:           currentJTI,
		ExpiresAt:            expiresAt,
		RefreshExpiresAt:     refreshExpiresAt,
		IPAddress:            ipAddress,
		UserAgent:            userAgent,
		DeviceInfo:           deviceInfo,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func NewBlacklistedToken(jti string, userID ulid.ULID, expiresAt time.Time, reason string) *BlacklistedToken {
	return &BlacklistedToken{
		JTI:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
		Reason:    reason,
		TokenType: TokenTypeIndividual, // Default to individual JTI blacklisting
		CreatedAt: time.Now(),
	}
}

// NewUserTimestampBlacklistedToken creates a user-wide timestamp blacklist entry for GDPR/SOC2 compliance
func NewUserTimestampBlacklistedToken(userID ulid.ULID, blacklistTimestamp int64, reason string) *BlacklistedToken {
	// Generate a proper ULID for this user-wide blacklist entry
	userWideJTI := ulid.New()
	
	// Set expiry far in the future to cover all possible access token lifetimes
	// We use the blacklist timestamp + reasonable buffer (24 hours) to ensure cleanup
	farFutureExpiry := time.Unix(blacklistTimestamp, 0).Add(24 * time.Hour)
	
	return &BlacklistedToken{
		JTI:                userWideJTI.String(),
		UserID:             userID,
		ExpiresAt:          farFutureExpiry,
		RevokedAt:          time.Now(),
		Reason:             reason,
		TokenType:          TokenTypeUserTimestamp,
		BlacklistTimestamp: &blacklistTimestamp,
		CreatedAt:          time.Now(),
	}
}

func NewAPIKey(userID, orgID ulid.ULID, name, keyPrefix, keyHash string, scopes []string, rateLimitRPM int, expiresAt *time.Time) *APIKey {
	return &APIKey{
		ID:             ulid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		Name:           name,
		KeyPrefix:      keyPrefix,
		KeyHash:        keyHash,
		Scopes:         scopes,
		RateLimitRPM:   rateLimitRPM,
		IsActive:       true,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func NewRole(orgID *ulid.ULID, name, displayName, description string, isSystemRole bool) *Role {
	return &Role{
		ID:             ulid.New(),
		OrganizationID: orgID,
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		IsSystemRole:   isSystemRole,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func NewPermission(resource, action, displayName, description, category string) *Permission {
	// Generate legacy name for compatibility
	name := fmt.Sprintf("%s:%s", resource, action)
	
	return &Permission{
		ID:          ulid.New(),
		Name:        name,        // Legacy format for compatibility
		Resource:    resource,    // New resource field
		Action:      action,      // New action field
		DisplayName: displayName,
		Description: description,
		Category:    category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewAuditLog(userID, orgID *ulid.ULID, action, resource, resourceID, metadata, ipAddress, userAgent string) *AuditLog {
	return &AuditLog{
		ID:             ulid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		Action:         action,
		Resource:       resource,
		ResourceID:     resourceID,
		Metadata:       metadata,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
	}
}

// Utility methods
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *UserSession) IsRefreshExpired() bool {
	return time.Now().After(s.RefreshExpiresAt)
}

func (s *UserSession) IsValid() bool {
	return s.IsActive && !s.IsExpired() && s.RevokedAt == nil
}

func (s *UserSession) MarkAsUsed() {
	now := time.Now()
	s.LastUsedAt = &now
	s.UpdatedAt = now
}

func (s *UserSession) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
	s.IsActive = false
	s.UpdatedAt = now
}

func (s *UserSession) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

func (k *APIKey) IsExpired() bool {
	return k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt)
}

func (k *APIKey) IsValid() bool {
	return k.IsActive && !k.IsExpired()
}

func (k *APIKey) MarkAsUsed() {
	now := time.Now()
	k.LastUsedAt = &now
	k.UpdatedAt = now
}

func (k *APIKey) Deactivate() {
	k.IsActive = false
	k.UpdatedAt = time.Now()
}

func (r *Role) AddPermission(permissionID ulid.ULID) *RolePermission {
	return &RolePermission{
		RoleID:       r.ID,
		PermissionID: permissionID,
		CreatedAt:    time.Now(),
	}
}

// RBAC Request/Response DTOs

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	OrganizationID  *ulid.ULID   `json:"organization_id,omitempty"` // NULL for global system roles
	Name            string       `json:"name" validate:"required,min=1,max=50"`
	DisplayName     string       `json:"display_name" validate:"required,min=1,max=100"`
	Description     string       `json:"description,omitempty"`
	IsSystemRole    bool         `json:"is_system_role,omitempty"` // Only allowed for admin users
	PermissionIDs   []ulid.ULID  `json:"permission_ids,omitempty"`
}

// UpdateRoleRequest represents a request to update an existing role
type UpdateRoleRequest struct {
	DisplayName   *string      `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	Description   *string      `json:"description,omitempty"`
	PermissionIDs []ulid.ULID  `json:"permission_ids,omitempty"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Resource    string `json:"resource" validate:"required,min=1,max=50"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category" validate:"required,min=1,max=100"`
}

// UpdatePermissionRequest represents a request to update an existing permission
type UpdatePermissionRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty" validate:"omitempty,min=1,max=100"`
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	RoleID ulid.ULID `json:"role_id" validate:"required"`
}

// RoleListResponse represents a list of roles with metadata
type RoleListResponse struct {
	Roles      []*Role `json:"roles"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page,omitempty"`
	PageSize   int     `json:"page_size,omitempty"`
}

// PermissionListResponse represents a list of permissions with metadata
type PermissionListResponse struct {
	Permissions []*Permission `json:"permissions"`
	TotalCount  int           `json:"total_count"`
	Page        int           `json:"page,omitempty"`
	PageSize    int           `json:"page_size,omitempty"`
}

// UserPermissionsResponse represents a user's effective permissions in an organization
type UserPermissionsResponse struct {
	UserID         ulid.ULID     `json:"user_id"`
	OrganizationID ulid.ULID     `json:"organization_id"`
	Role           *Role         `json:"role,omitempty"`
	Permissions    []*Permission `json:"permissions"`
	ResourceActions []string     `json:"resource_actions"` // ["users:read", "projects:write", etc.]
}

// CheckPermissionsRequest represents a request to check multiple permissions
type CheckPermissionsRequest struct {
	ResourceActions []string `json:"resource_actions" validate:"required,min=1"`
}

// CheckPermissionsResponse represents the result of checking multiple permissions
type CheckPermissionsResponse struct {
	Results map[string]bool `json:"results"` // resource:action -> has_permission
}

// RoleStatistics represents statistics about roles in an organization
type RoleStatistics struct {
	OrganizationID     ulid.ULID `json:"organization_id"`
	TotalRoles         int       `json:"total_roles"`
	SystemRoles        int       `json:"system_roles"`
	CustomRoles        int       `json:"custom_roles"`
	TotalMembers       int       `json:"total_members"`
	RoleDistribution   map[string]int `json:"role_distribution"` // role_name -> member_count
	PermissionCount    int       `json:"permission_count"`
	LastUpdated        time.Time `json:"last_updated"`
}


// Table name methods for GORM
func (UserSession) TableName() string     { return "user_sessions" }
func (BlacklistedToken) TableName() string { return "blacklisted_tokens" }
func (APIKey) TableName() string          { return "api_keys" }
func (Role) TableName() string            { return "roles" }
func (Permission) TableName() string      { return "permissions" }
func (RolePermission) TableName() string  { return "role_permissions" }
func (AuditLog) TableName() string        { return "audit_logs" }