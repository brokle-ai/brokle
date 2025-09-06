// Package auth provides the authentication and authorization domain model.
//
// The auth domain handles JWT tokens, sessions, API keys, roles, permissions,
// and role-based access control (RBAC) across the platform.
package auth

import (
	"context"
	"time"

	"brokle/pkg/ulid"
	"gorm.io/gorm"
)

// UserSession represents an active user session.
type UserSession struct {
	ID               ulid.ULID   `json:"id" gorm:"type:char(26);primaryKey"`
	UserID           ulid.ULID   `json:"user_id" gorm:"type:char(26);not null"`
	Token            string      `json:"token" gorm:"size:255;not null;uniqueIndex"`
	RefreshToken     string      `json:"refresh_token,omitempty" gorm:"size:255;not null;uniqueIndex"`
	ExpiresAt        time.Time   `json:"expires_at"`
	RefreshExpiresAt time.Time   `json:"refresh_expires_at"`
	IPAddress        *string     `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent        *string     `json:"user_agent,omitempty" gorm:"type:text"`
	DeviceInfo       interface{} `json:"device_info,omitempty" gorm:"type:jsonb"` // Device information JSON
	IsActive         bool        `json:"is_active" gorm:"default:true"`
	LastUsedAt       *time.Time  `json:"last_used_at,omitempty"`
	RevokedAt        *time.Time  `json:"revoked_at,omitempty"` // Replaced deleted_at with revoked_at
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
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

// Role represents a role with permissions (supports custom roles).
type Role struct {
	ID             ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty" gorm:"type:char(26)"` // NULL for system roles
	Name           string     `json:"name" gorm:"size:50;not null"`
	DisplayName    string     `json:"display_name" gorm:"size:100;not null"`
	Description    string     `json:"description" gorm:"type:text"`
	IsSystemRole   bool       `json:"is_system_role" gorm:"default:false"` // Cannot be deleted
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Permissions     []Permission     `json:"permissions,omitempty" gorm:"many2many:role_permissions"`
	RolePermissions []RolePermission `json:"role_permissions,omitempty" gorm:"foreignKey:RoleID"`
}

// Permission represents a specific permission in the system.
type Permission struct {
	ID          ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	Name        string    `json:"name" gorm:"size:255;not null;uniqueIndex"` // users.create, projects.read
	DisplayName string    `json:"display_name" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Category    string    `json:"category" gorm:"size:100"` // users, projects, billing, etc.

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relations
	Roles []Role `json:"roles,omitempty" gorm:"many2many:role_permissions"`
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
func NewUserSession(userID ulid.ULID, token, refreshToken string, expiresAt, refreshExpiresAt time.Time, ipAddress, userAgent *string, deviceInfo interface{}) *UserSession {
	return &UserSession{
		ID:               ulid.New(),
		UserID:           userID,
		Token:            token,
		RefreshToken:     refreshToken,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		DeviceInfo:       deviceInfo,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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

func NewPermission(name, displayName, description, category string) *Permission {
	return &Permission{
		ID:          ulid.New(),
		Name:        name,
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

// Table name methods for GORM
func (UserSession) TableName() string   { return "user_sessions" }
func (APIKey) TableName() string        { return "api_keys" }
func (Role) TableName() string          { return "roles" }
func (Permission) TableName() string    { return "permissions" }
func (RolePermission) TableName() string { return "role_permissions" }
func (AuditLog) TableName() string      { return "audit_logs" }