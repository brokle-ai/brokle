# KeyPair Database Schema Design

## Overview
Replacing the existing `api_keys` table with a new `key_pairs` table that supports public+secret key authentication model as required by the refactored SDK.

## Key Format Standards
- **Public Key Format**: `pk_{projectId}_{randomString}` (e.g., `pk_01JA5X2B3C4D5E6F7G8H9J0K1L_abc123def456`)
- **Secret Key Format**: `sk_{randomString}` (e.g., `sk_xyz789uvw456rst123`)
- **Public Key Prefix**: Always `pk_` followed by project ID and random string
- **Secret Key Prefix**: Always `sk_` followed by random string

## New Schema: key_pairs Table

```sql
CREATE TABLE key_pairs (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id CHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id CHAR(26) REFERENCES environments(id) ON DELETE CASCADE,

    -- Key pair identification
    name VARCHAR(255) NOT NULL,

    -- Public key (pk_projectId_random) - stored in plain text
    public_key VARCHAR(255) NOT NULL UNIQUE,

    -- Secret key hash (sk_random hashed) - never store plain text
    secret_key_hash VARCHAR(255) NOT NULL UNIQUE,
    secret_key_prefix VARCHAR(8) NOT NULL, -- 'sk_' prefix for validation

    -- Scoping and permissions
    scopes JSON, -- ['gateway:read', 'analytics:read', etc.]

    -- Rate limiting and usage controls
    rate_limit_rpm INTEGER DEFAULT 1000,

    -- Status and lifecycle
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,

    -- Metadata for enterprise features
    metadata JSONB,

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

## Indexes for Performance

```sql
-- Primary lookup indexes
CREATE INDEX idx_key_pairs_public_key ON key_pairs(public_key);
CREATE INDEX idx_key_pairs_secret_key_hash ON key_pairs(secret_key_hash);
CREATE INDEX idx_key_pairs_secret_key_prefix ON key_pairs(secret_key_prefix);

-- Relationship indexes
CREATE INDEX idx_key_pairs_user_id ON key_pairs(user_id);
CREATE INDEX idx_key_pairs_organization_id ON key_pairs(organization_id);
CREATE INDEX idx_key_pairs_project_id ON key_pairs(project_id);
CREATE INDEX idx_key_pairs_environment_id ON key_pairs(environment_id);

-- Status and lifecycle indexes
CREATE INDEX idx_key_pairs_is_active ON key_pairs(is_active);
CREATE INDEX idx_key_pairs_deleted_at ON key_pairs(deleted_at);
CREATE INDEX idx_key_pairs_expires_at ON key_pairs(expires_at);
CREATE INDEX idx_key_pairs_last_used_at ON key_pairs(last_used_at);
```

## Key Differences from api_keys Table

| Aspect | Old api_keys | New key_pairs |
|--------|-------------|---------------|
| Authentication | Single API key | Public + Secret key pair |
| Public Key | Not exposed | `pk_projectId_random` format |
| Secret Key | Hashed single key | `sk_random` format, hashed |
| Project Scope | Optional project_id | Required project_id (derived from public key) |
| Environment | Optional environment_id | Optional environment_id |
| Key Format | Custom prefix | Standardized pk_/sk_ prefixes |
| Validation | Single key lookup | Two-factor key validation |

## Authentication Flow

1. **Client Request**: Includes both `X-Public-Key` and `X-Secret-Key` headers
2. **Public Key Validation**:
   - Validate format: `pk_{projectId}_{random}`
   - Extract project ID from public key
   - Look up key pair record by public_key
3. **Secret Key Validation**:
   - Hash the provided secret key
   - Compare with stored secret_key_hash
   - Validate secret key format: `sk_{random}`
4. **Authorization**:
   - Check key pair is active and not expired
   - Validate scopes for requested operation
   - Update last_used_at timestamp
5. **Context**: Set user/org/project context from key pair record

## Domain Entity (Go)

```go
type KeyPair struct {
    ID             ulid.ULID  `json:"id" gorm:"type:char(26);primaryKey"`
    UserID         ulid.ULID  `json:"user_id" gorm:"type:char(26);not null"`
    OrganizationID ulid.ULID  `json:"organization_id" gorm:"type:char(26);not null"`
    ProjectID      ulid.ULID  `json:"project_id" gorm:"type:char(26);not null"`
    EnvironmentID  *ulid.ULID `json:"environment_id,omitempty" gorm:"type:char(26)"`

    Name              string     `json:"name" gorm:"size:255;not null"`
    PublicKey         string     `json:"public_key" gorm:"size:255;not null;uniqueIndex"`
    SecretKeyHash     string     `json:"-" gorm:"size:255;not null;uniqueIndex"`
    SecretKeyPrefix   string     `json:"secret_key_prefix" gorm:"size:8;not null"`

    Scopes         []string   `json:"scopes" gorm:"type:json"`
    RateLimitRPM   int        `json:"rate_limit_rpm" gorm:"default:1000"`
    Metadata       interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`

    IsActive       bool       `json:"is_active" gorm:"default:true"`
    ExpiresAt      *time.Time `json:"expires_at,omitempty"`
    LastUsedAt     *time.Time `json:"last_used_at,omitempty"`

    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
    DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

    // Relationships
    User         User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
    Project      Project      `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
    Environment  *Environment `json:"environment,omitempty" gorm:"foreignKey:EnvironmentID"`
}
```

## Security Considerations

1. **Secret Key Storage**: Always hash secret keys, never store in plain text
2. **Public Key Exposure**: Public keys are safe to expose in logs/responses
3. **Project ID Extraction**: Extract project context directly from public key format
4. **Rate Limiting**: Apply per-key-pair rate limiting based on rate_limit_rpm
5. **Scoping**: Enforce scope-based access control for different API operations
6. **Audit Trail**: Log all key usage in audit_logs table
7. **Key Rotation**: Support key expiration and rotation through expires_at

## Migration Strategy

1. **Create new key_pairs table** alongside existing api_keys table
2. **Implement new authentication middleware** that supports both systems temporarily
3. **Add key pair management endpoints** for creating/managing key pairs
4. **Migrate existing data** (if any) or start fresh for early development
5. **Update all handlers** to use new authentication context
6. **Remove api_keys table** and old authentication code
7. **Update OpenAPI specifications** to reflect new authentication scheme

## SDK Integration

The new schema directly supports the SDK requirements:
- Public key format includes project ID for automatic project scoping
- Secret key provides secure authentication
- Headers: `X-Public-Key` and `X-Secret-Key`
- Clean separation between identification (public) and authentication (secret)
- Backward compatibility not required as per user request