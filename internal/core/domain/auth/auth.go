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

// KeyPair represents a public+secret key pair for API authentication with project scoping.
// Replaces the old APIKey system with a more secure two-key authentication model.
// Format: Public key = pk_projectId_random, Secret key = sk_random (hashed)
type KeyPair struct {
	ID             ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	UserID         ulid.ULID  `json:"user_id" gorm:"type:char(26);not null"`
	OrganizationID ulid.ULID  `json:"organization_id" gorm:"type:char(26);not null"`
	ProjectID      ulid.ULID  `json:"project_id" gorm:"type:char(26);not null"`  // Required - derived from public key
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty" gorm:"type:char(26)"`

	// Key pair identification
	Name           string     `json:"name" gorm:"size:255;not null"`

	// Public key (pk_projectId_random) - stored in plain text for lookup
	PublicKey      string     `json:"public_key" gorm:"size:255;not null;uniqueIndex"`

	// Secret key hash (sk_random hashed) - never store plain text
	SecretKeyHash  string     `json:"-" gorm:"size:255;not null;uniqueIndex"`
	SecretKeyPrefix string    `json:"secret_key_prefix" gorm:"size:8;not null;default:'sk_'"` // Always 'sk_'

	// Scoping and permissions
	Scopes         []string   `json:"scopes" gorm:"type:json"`               // JSON array of permissions

	// Rate limiting and usage controls
	RateLimitRPM   int        `json:"rate_limit_rpm" gorm:"default:1000"`    // Requests per minute

	// Status and lifecycle
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`

	// Metadata for enterprise features
	Metadata       interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`  // Flexible metadata storage

	// Audit fields
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// KeyPairScope represents the scopes/permissions for key pair access control
type KeyPairScope string

const (
	// Gateway scopes for AI routing and proxy functionality
	ScopeGatewayRead   KeyPairScope = "gateway:read"
	ScopeGatewayWrite  KeyPairScope = "gateway:write"

	// Analytics scopes for metrics and reporting
	ScopeAnalyticsRead KeyPairScope = "analytics:read"

	// Config scopes for configuration management
	ScopeConfigRead    KeyPairScope = "config:read"
	ScopeConfigWrite   KeyPairScope = "config:write"

	// Admin scope for full access (enterprise)
	ScopeAdmin         KeyPairScope = "admin"
)

// ValidatePublicKeyFormat validates the public key format: pk_projectId_random
func (kp *KeyPair) ValidatePublicKeyFormat() error {
	if !strings.HasPrefix(kp.PublicKey, "pk_") {
		return fmt.Errorf("public key must start with 'pk_', got: %s", kp.PublicKey)
	}

	parts := strings.Split(kp.PublicKey, "_")
	if len(parts) < 3 {
		return fmt.Errorf("public key must be in format pk_projectId_random, got: %s", kp.PublicKey)
	}

	projectIDPart := parts[1]
	if len(projectIDPart) != 26 {
		return fmt.Errorf("project ID in public key must be 26 characters (ULID), got: %d characters", len(projectIDPart))
	}

	return nil
}

// ExtractProjectIDFromPublicKey extracts the project ID from the public key format
func (kp *KeyPair) ExtractProjectIDFromPublicKey() (ulid.ULID, error) {
	if err := kp.ValidatePublicKeyFormat(); err != nil {
		return ulid.ULID{}, err
	}

	parts := strings.Split(kp.PublicKey, "_")
	projectIDStr := parts[1]

	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("invalid project ID in public key: %w", err)
	}

	return projectID, nil
}

// ValidateSecretKeyPrefix validates that the secret key prefix is 'sk_'
func (kp *KeyPair) ValidateSecretKeyPrefix() error {
	if kp.SecretKeyPrefix != "sk_" {
		return fmt.Errorf("secret key prefix must be 'sk_', got: %s", kp.SecretKeyPrefix)
	}
	return nil
}

// HasScope checks if the key pair has a specific scope
func (kp *KeyPair) HasScope(scope KeyPairScope) bool {
	scopeStr := string(scope)
	for _, s := range kp.Scopes {
		if s == scopeStr {
			return true
		}
		// Check for admin scope which grants all permissions
		if s == string(ScopeAdmin) {
			return true
		}
	}
	return false
}

// IsExpired checks if the key pair has expired
func (kp *KeyPair) IsExpired() bool {
	if kp.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*kp.ExpiresAt)
}

// IsValid checks if the key pair is valid for use (active and not expired)
func (kp *KeyPair) IsValid() bool {
	return kp.IsActive && !kp.IsExpired()
}

// Role represents both system template roles and custom scoped roles
type Role struct {
	ID          ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	Name        string     `json:"name" gorm:"size:50;not null"`
	ScopeType   string     `json:"scope_type" gorm:"size:20;not null"`
	ScopeID     *ulid.ULID `json:"scope_id,omitempty" gorm:"type:char(26);index"`
	Description string     `json:"description" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	Permissions     []Permission     `json:"permissions,omitempty" gorm:"many2many:role_permissions"`
	RolePermissions []RolePermission `json:"role_permissions,omitempty" gorm:"foreignKey:RoleID"`
}

// OrganizationMember represents user membership in an organization with a single role
type OrganizationMember struct {
	UserID         ulid.ULID `json:"user_id" gorm:"type:char(26);primaryKey"`
	OrganizationID ulid.ULID `json:"organization_id" gorm:"type:char(26);primaryKey"`
	RoleID         ulid.ULID `json:"role_id" gorm:"type:char(26);not null"`
	Status         string    `json:"status" gorm:"size:20;default:active"`
	JoinedAt       time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	InvitedBy      *ulid.ULID `json:"invited_by,omitempty" gorm:"type:char(26)"`

	// Relations
	Role *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

// ProjectMember represents user membership in a project with a single role (future)
type ProjectMember struct {
	UserID    ulid.ULID `json:"user_id" gorm:"type:char(26);primaryKey"`
	ProjectID ulid.ULID `json:"project_id" gorm:"type:char(26);primaryKey"`
	RoleID    ulid.ULID `json:"role_id" gorm:"type:char(26);not null"`
	Status    string    `json:"status" gorm:"size:20;default:active"`
	JoinedAt  time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relations
	Role *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

// EnvironmentMember represents user membership in an environment with a single role (future)
type EnvironmentMember struct {
	UserID        ulid.ULID `json:"user_id" gorm:"type:char(26);primaryKey"`
	EnvironmentID ulid.ULID `json:"environment_id" gorm:"type:char(26);primaryKey"`
	RoleID        ulid.ULID `json:"role_id" gorm:"type:char(26);not null"`
	Status        string    `json:"status" gorm:"size:20;default:active"`
	JoinedAt      time.Time `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relations
	Role *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

// Scope constants for roles
const (
	ScopeSystem       = "system"       // System template roles
	ScopeOrganization = "organization" // Organization-specific roles
	ScopeProject      = "project"      // Project-specific roles  
	ScopeEnvironment  = "environment"  // Environment-specific roles
)

// Membership status constants
const (
	MemberStatusActive    = "active"
	MemberStatusInvited   = "invited"
	MemberStatusSuspended = "suspended"
)

// Helper methods for scoped roles
func (r *Role) IsSystemRole() bool {
	return r.ScopeType == ScopeSystem && r.ScopeID == nil
}

func (r *Role) IsCustomRole() bool {
	return r.ScopeType != ScopeSystem && r.ScopeID != nil
}

func (r *Role) IsOrganizationRole() bool {
	return r.ScopeType == ScopeOrganization
}

func (r *Role) IsProjectRole() bool {
	return r.ScopeType == ScopeProject
}

func (r *Role) IsEnvironmentRole() bool {
	return r.ScopeType == ScopeEnvironment
}

func (r *Role) GetScopeDisplay() string {
	switch r.ScopeType {
	case ScopeSystem:
		return "System"
	case ScopeOrganization:
		if r.ScopeID == nil {
			return "Organization Template"
		}
		return "Organization Custom"
	case ScopeProject:
		return "Project"
	case ScopeEnvironment:
		return "Environment"
	default:
		return "Unknown"
	}
}

// Helper methods for organization membership
func (m *OrganizationMember) IsActive() bool {
	return m.Status == MemberStatusActive
}

func (m *OrganizationMember) IsInvited() bool {
	return m.Status == MemberStatusInvited
}

func (m *OrganizationMember) IsSuspended() bool {
	return m.Status == MemberStatusSuspended
}

func (m *OrganizationMember) Activate() {
	m.Status = MemberStatusActive
}

func (m *OrganizationMember) Suspend() {
	m.Status = MemberStatusSuspended
}

// Helper methods for project membership
func (m *ProjectMember) IsActive() bool {
	return m.Status == MemberStatusActive
}

func (m *ProjectMember) Activate() {
	m.Status = MemberStatusActive
}

// Helper methods for environment membership
func (m *EnvironmentMember) IsActive() bool {
	return m.Status == MemberStatusActive
}

func (m *EnvironmentMember) Activate() {
	m.Status = MemberStatusActive
}

// Permission represents a normalized permission using resource:action format
type Permission struct {
	ID          ulid.ULID `json:"id" gorm:"type:char(26);primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Resource    string    `json:"resource" gorm:"size:50;not null;index"`
	Action      string    `json:"action" gorm:"size:50;not null;index"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`

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

// RolePermission represents the many-to-many relationship between template roles and permissions
type RolePermission struct {
	RoleID       ulid.ULID  `json:"role_id" gorm:"type:char(26);not null;primaryKey"`
	PermissionID ulid.ULID  `json:"permission_id" gorm:"type:char(26);not null;primaryKey"`
	GrantedAt    time.Time  `json:"granted_at" gorm:"default:CURRENT_TIMESTAMP"`
	GrantedBy    *ulid.ULID `json:"granted_by,omitempty" gorm:"type:char(26)"`

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

type CreateKeyPairRequest struct {
	Name           string     `json:"name" validate:"required,min=1,max=100"`
	OrganizationID ulid.ULID  `json:"organization_id" validate:"required"`
	ProjectID      ulid.ULID  `json:"project_id" validate:"required"`  // Required for key pair generation
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty"`
	Scopes         []string   `json:"scopes" validate:"required,min=1"`
	RateLimitRPM   int        `json:"rate_limit_rpm" validate:"min=1,max=10000"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type CreateKeyPairResponse struct {
	ID           ulid.ULID  `json:"id"`
	Name         string     `json:"name"`
	PublicKey    string     `json:"public_key"`    // pk_projectId_random format
	SecretKey    string     `json:"secret_key"`    // sk_random format - only returned once
	Scopes       []string   `json:"scopes"`
	ProjectID    ulid.ULID  `json:"project_id"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// UpdateKeyPairRequest represents the request to update a key pair.
type UpdateKeyPairRequest struct {
	Name         *string    `json:"name,omitempty"`
	Scopes       []string   `json:"scopes,omitempty"`
	RateLimitRPM *int       `json:"rate_limit_rpm,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// KeyPairFilters represents filters for querying key pairs.
type KeyPairFilters struct {
	UserID         *ulid.ULID `json:"user_id,omitempty"`
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"`
	ProjectID      *ulid.ULID `json:"project_id,omitempty"`
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	HasScopes      []string   `json:"has_scopes,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}



// AuthContext represents the authenticated context for a request.
// AuthContext represents clean user identity context (permissions resolved dynamically)
type AuthContext struct {
	UserID       ulid.ULID  `json:"user_id"`
	KeyPairID    *ulid.ULID `json:"key_pair_id,omitempty"`    // Set if authenticated via key pair
	SessionID    *ulid.ULID `json:"session_id,omitempty"`     // Set if authenticated via session

	// Additional context for key pair authentication
	OrganizationID *ulid.ULID `json:"organization_id,omitempty"` // From key pair
	ProjectID      *ulid.ULID `json:"project_id,omitempty"`      // From key pair
	EnvironmentID  *ulid.ULID `json:"environment_id,omitempty"`  // From key pair
	Scopes         []string   `json:"scopes,omitempty"`          // From key pair
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

func NewKeyPair(userID, orgID, projectID ulid.ULID, name, publicKey, secretKeyHash string, scopes []string, rateLimitRPM int, expiresAt *time.Time) *KeyPair {
	return &KeyPair{
		ID:              ulid.New(),
		UserID:          userID,
		OrganizationID:  orgID,
		ProjectID:       projectID,
		Name:            name,
		PublicKey:       publicKey,
		SecretKeyHash:   secretKeyHash,
		SecretKeyPrefix: "sk_",
		Scopes:          scopes,
		RateLimitRPM:    rateLimitRPM,
		IsActive:        true,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func NewRole(name, scopeType, description string) *Role {
	return &Role{
		ID:          ulid.New(),
		Name:        name,
		ScopeType:   scopeType,
		ScopeID:     nil, // System/template role
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewCustomRole creates a custom role scoped to a specific organization/project/environment
func NewCustomRole(name, scopeType, description string, scopeID ulid.ULID) *Role {
	return &Role{
		ID:          ulid.New(),
		Name:        name,
		ScopeType:   scopeType,
		ScopeID:     &scopeID,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewOrganizationMember(userID, organizationID, roleID ulid.ULID, invitedBy *ulid.ULID) *OrganizationMember {
	return &OrganizationMember{
		UserID:         userID,
		OrganizationID: organizationID,
		RoleID:         roleID,
		Status:         MemberStatusActive,
		JoinedAt:       time.Now(),
		InvitedBy:      invitedBy,
	}
}

func NewProjectMember(userID, projectID, roleID ulid.ULID) *ProjectMember {
	return &ProjectMember{
		UserID:    userID,
		ProjectID: projectID,
		RoleID:    roleID,
		Status:    MemberStatusActive,
		JoinedAt:  time.Now(),
	}
}

func NewEnvironmentMember(userID, environmentID, roleID ulid.ULID) *EnvironmentMember {
	return &EnvironmentMember{
		UserID:        userID,
		EnvironmentID: environmentID,
		RoleID:        roleID,
		Status:        MemberStatusActive,
		JoinedAt:      time.Now(),
	}
}

func NewPermission(resource, action, description string) *Permission {
	name := fmt.Sprintf("%s:%s", resource, action)
	
	return &Permission{
		ID:          ulid.New(),
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
	UserID    ulid.ULID  `json:"user_id" gorm:"type:char(26);not null;index"`
	Token     string     `json:"-" gorm:"size:255;not null;uniqueIndex"`
	Used      bool       `json:"used" gorm:"default:false;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null;index"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
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

func NewPasswordResetToken(userID ulid.ULID, token string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		ID:        ulid.New(),
		UserID:    userID,
		Token:     token,
		Used:      false,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

// KeyPair methods moved to the KeyPair struct definition above
// (IsExpired, IsValid, etc. are already defined there)

func (r *Role) AddPermission(permissionID ulid.ULID, grantedBy *ulid.ULID) *RolePermission {
	return &RolePermission{
		RoleID:       r.ID,
		PermissionID: permissionID,
		GrantedAt:    time.Now(),
		GrantedBy:    grantedBy,
	}
}

// RBAC Request/Response DTOs

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	ScopeType     string      `json:"scope_type" validate:"required,oneof=system organization project environment"`
	Name          string      `json:"name" validate:"required,min=1,max=100"`
	Description   string      `json:"description,omitempty"`
	PermissionIDs []ulid.ULID `json:"permission_ids,omitempty"`
}

// UpdateRoleRequest represents a request to update an existing role
type UpdateRoleRequest struct {
	Description   *string     `json:"description,omitempty"`
	PermissionIDs []ulid.ULID `json:"permission_ids,omitempty"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Resource    string `json:"resource" validate:"required,min=1,max=50"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
	Description string `json:"description,omitempty"`
}

// UpdatePermissionRequest represents a request to update an existing permission
type UpdatePermissionRequest struct {
	Description *string `json:"description,omitempty"`
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

// UserPermissionsResponse represents a user's effective permissions across all scopes
type UserPermissionsResponse struct {
	UserID          ulid.ULID     `json:"user_id"`
	Roles           []*Role       `json:"roles"`
	Permissions     []*Permission `json:"permissions"`
	ResourceActions []string      `json:"resource_actions"` // ["users:read", "projects:write", etc.]
}

// CheckPermissionsRequest represents a request to check multiple permissions
type CheckPermissionsRequest struct {
	ResourceActions []string `json:"resource_actions" validate:"required,min=1"`
}

// CheckPermissionsResponse represents the result of checking multiple permissions
type CheckPermissionsResponse struct {
	Results map[string]bool `json:"results"` // resource:action -> has_permission
}

// RoleStatistics represents statistics about roles across all scopes
type RoleStatistics struct {
	TotalRoles         int               `json:"total_roles"`
	SystemRoles        int               `json:"system_roles"`
	OrganizationRoles  int               `json:"organization_roles"`
	ProjectRoles       int               `json:"project_roles"`
	ScopeDistribution  map[string]int    `json:"scope_distribution"` // scope_type -> role_count
	RoleDistribution   map[string]int    `json:"role_distribution"`  // role_name -> member_count
	PermissionCount    int               `json:"permission_count"`
	LastUpdated        time.Time         `json:"last_updated"`
}


// Table name methods for GORM
func (UserSession) TableName() string         { return "user_sessions" }
func (BlacklistedToken) TableName() string    { return "blacklisted_tokens" }
func (KeyPair) TableName() string             { return "key_pairs" }
func (Role) TableName() string                { return "roles" }
func (OrganizationMember) TableName() string  { return "organization_members" }
func (ProjectMember) TableName() string       { return "project_members" }
func (EnvironmentMember) TableName() string   { return "environment_members" }
func (Permission) TableName() string          { return "permissions" }
func (RolePermission) TableName() string      { return "role_permissions" }
func (AuditLog) TableName() string            { return "audit_logs" }
func (PasswordResetToken) TableName() string  { return "password_reset_tokens" }