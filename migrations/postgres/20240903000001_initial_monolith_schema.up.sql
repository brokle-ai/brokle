-- ===================================
-- BROKLE MONOLITH INITIAL SCHEMA
-- ===================================

-- Create users table with ULID (USER DOMAIN)
CREATE TABLE users (
    id CHAR(26) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    avatar_url VARCHAR(500),
    phone VARCHAR(50),
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_activity_at TIMESTAMP WITH TIME ZONE,
    onboarding_completed BOOLEAN DEFAULT FALSE,
    onboarding_completed_at TIMESTAMP WITH TIME ZONE,
    default_organization_id CHAR(26),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create organizations table with ULID (ORGANIZATION DOMAIN)
CREATE TABLE organizations (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    billing_email VARCHAR(255),
    plan VARCHAR(50) DEFAULT 'free',
    subscription_status VARCHAR(50) DEFAULT 'active',
    trial_ends_at TIMESTAMP WITH TIME ZONE,
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Add foreign key constraint for user default organization
ALTER TABLE users ADD CONSTRAINT fk_users_default_organization_id 
    FOREIGN KEY (default_organization_id) REFERENCES organizations(id) ON DELETE SET NULL;

-- Create organization_members table with ULID (ORGANIZATION DOMAIN)
CREATE TABLE organization_members (
    id CHAR(26) PRIMARY KEY,
    organization_id CHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id CHAR(26) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(organization_id, user_id)
);

-- Create projects table with ULID (ORGANIZATION DOMAIN)
CREATE TABLE projects (
    id CHAR(26) PRIMARY KEY,
    organization_id CHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(organization_id, slug)
);

-- Create environments table with ULID (ORGANIZATION DOMAIN)
CREATE TABLE environments (
    id CHAR(26) PRIMARY KEY,
    project_id CHAR(26) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(project_id, slug)
);

-- Create invitations table (ORGANIZATION DOMAIN)
CREATE TABLE invitations (
    id CHAR(26) PRIMARY KEY,
    organization_id CHAR(26) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_id CHAR(26) NOT NULL,
    email VARCHAR(255) NOT NULL,
    user_id CHAR(26) REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create sessions table (AUTH DOMAIN)
CREATE TABLE sessions (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL,
    refresh_token VARCHAR(500) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create api_keys table with ULID (AUTH DOMAIN)
CREATE TABLE api_keys (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id CHAR(26) REFERENCES organizations(id) ON DELETE CASCADE,
    project_id CHAR(26) REFERENCES projects(id) ON DELETE CASCADE,
    environment_id CHAR(26) REFERENCES environments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    rate_limit_rpm INTEGER DEFAULT 1000,
    allowed_ips INET[],
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create roles table (AUTH DOMAIN)
CREATE TABLE roles (
    id CHAR(26) PRIMARY KEY,
    organization_id CHAR(26) REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(organization_id, name)
);

-- Add foreign key constraint for organization members role
ALTER TABLE organization_members ADD CONSTRAINT fk_organization_members_role_id 
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT;

-- Add foreign key constraint for invitations role
ALTER TABLE invitations ADD CONSTRAINT fk_invitations_role_id 
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT;

-- Create permissions table (AUTH DOMAIN)
CREATE TABLE permissions (
    id CHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create role_permissions table (AUTH DOMAIN)
CREATE TABLE role_permissions (
    id CHAR(26) PRIMARY KEY,
    role_id CHAR(26) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id CHAR(26) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Create audit_logs table (AUTH DOMAIN)
CREATE TABLE audit_logs (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) REFERENCES users(id) ON DELETE SET NULL,
    organization_id CHAR(26) REFERENCES organizations(id) ON DELETE SET NULL,
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100),
    metadata TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create password_reset_tokens table (AUTH DOMAIN)
CREATE TABLE password_reset_tokens (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create email_verification_tokens table (AUTH DOMAIN)
CREATE TABLE email_verification_tokens (
    id CHAR(26) PRIMARY KEY,
    user_id CHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ===================================
-- INDEXES FOR PERFORMANCE
-- ===================================

-- Users table indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_default_organization_id ON users(default_organization_id);

-- Organizations table indexes
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at);

-- Organization members table indexes
CREATE INDEX idx_organization_members_organization_id ON organization_members(organization_id);
CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX idx_organization_members_role_id ON organization_members(role_id);
CREATE INDEX idx_organization_members_deleted_at ON organization_members(deleted_at);

-- Projects table indexes
CREATE INDEX idx_projects_organization_id ON projects(organization_id);
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at);

-- Environments table indexes
CREATE INDEX idx_environments_project_id ON environments(project_id);
CREATE INDEX idx_environments_slug ON environments(slug);
CREATE INDEX idx_environments_deleted_at ON environments(deleted_at);

-- Invitations table indexes
CREATE INDEX idx_invitations_organization_id ON invitations(organization_id);
CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_user_id ON invitations(user_id);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_status ON invitations(status);
CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);

-- Sessions table indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_is_active ON sessions(is_active);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- API keys table indexes
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX idx_api_keys_project_id ON api_keys(project_id);
CREATE INDEX idx_api_keys_environment_id ON api_keys(environment_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_deleted_at ON api_keys(deleted_at);

-- Roles table indexes
CREATE INDEX idx_roles_organization_id ON roles(organization_id);
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- Permissions table indexes
CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_category ON permissions(category);

-- Role permissions table indexes
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Audit logs table indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- Password reset tokens table indexes
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- Email verification tokens table indexes
CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens(token);
CREATE INDEX idx_email_verification_tokens_expires_at ON email_verification_tokens(expires_at);

-- ===================================
-- TRIGGERS FOR UPDATED_AT
-- ===================================

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create updated_at triggers for all tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organization_members_updated_at BEFORE UPDATE ON organization_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_environments_updated_at BEFORE UPDATE ON environments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_invitations_updated_at BEFORE UPDATE ON invitations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ===================================
-- SEED DEFAULT PERMISSIONS
-- ===================================

-- Insert system permissions
INSERT INTO permissions (id, name, display_name, description, category) VALUES
-- User Management
('01HZQV8XYZA9ABCDEF123456', 'users.read', 'Read Users', 'View user profiles and information', 'user_management'),
('01HZQV8XYZB9ABCDEF123457', 'users.write', 'Write Users', 'Create and update user profiles', 'user_management'),
('01HZQV8XYZC9ABCDEF123458', 'users.delete', 'Delete Users', 'Delete user accounts', 'user_management'),
('01HZQV8XYZD9ABCDEF123459', 'users.invite', 'Invite Users', 'Send invitations to new users', 'user_management'),

-- Organization Management
('01HZQV8XYZE9ABCDEF12345A', 'organizations.read', 'Read Organizations', 'View organization details', 'organization_management'),
('01HZQV8XYZF9ABCDEF12345B', 'organizations.write', 'Write Organizations', 'Create and update organizations', 'organization_management'),
('01HZQV8XYZG9ABCDEF12345C', 'organizations.delete', 'Delete Organizations', 'Delete organizations', 'organization_management'),
('01HZQV8XYZH9ABCDEF12345D', 'organizations.members', 'Manage Members', 'Add, remove, and modify organization members', 'organization_management'),

-- Project Management
('01HZQV8XYZI9ABCDEF12345E', 'projects.read', 'Read Projects', 'View project details', 'project_management'),
('01HZQV8XYZJ9ABCDEF12345F', 'projects.write', 'Write Projects', 'Create and update projects', 'project_management'),
('01HZQV8XYZK9ABCDEF123460', 'projects.delete', 'Delete Projects', 'Delete projects', 'project_management'),

-- Environment Management
('01HZQV8XYZL9ABCDEF123461', 'environments.read', 'Read Environments', 'View environment details', 'environment_management'),
('01HZQV8XYZM9ABCDEF123462', 'environments.write', 'Write Environments', 'Create and update environments', 'environment_management'),
('01HZQV8XYZN9ABCDEF123463', 'environments.delete', 'Delete Environments', 'Delete environments', 'environment_management'),

-- API Key Management
('01HZQV8XYZO9ABCDEF123464', 'api_keys.read', 'Read API Keys', 'View API keys', 'api_key_management'),
('01HZQV8XYZP9ABCDEF123465', 'api_keys.write', 'Write API Keys', 'Create and update API keys', 'api_key_management'),
('01HZQV8XYZQ9ABCDEF123466', 'api_keys.delete', 'Delete API Keys', 'Delete API keys', 'api_key_management'),

-- Role & Permission Management
('01HZQV8XYZR9ABCDEF123467', 'roles.read', 'Read Roles', 'View roles and permissions', 'rbac_management'),
('01HZQV8XYZS9ABCDEF123468', 'roles.write', 'Write Roles', 'Create and update roles', 'rbac_management'),
('01HZQV8XYZT9ABCDEF123469', 'roles.delete', 'Delete Roles', 'Delete roles', 'rbac_management'),
('01HZQV8XYZU9ABCDEF12346A', 'permissions.manage', 'Manage Permissions', 'Assign permissions to roles', 'rbac_management'),

-- Audit & Security
('01HZQV8XYZV9ABCDEF12346B', 'audit.read', 'Read Audit Logs', 'View audit logs and security events', 'security'),
('01HZQV8XYZW9ABCDEF12346C', 'security.manage', 'Manage Security', 'Manage security settings and configurations', 'security');

-- ===================================
-- SEED DEFAULT SYSTEM ROLES
-- ===================================

-- Insert system roles (organization_id is NULL for system roles)
INSERT INTO roles (id, organization_id, name, display_name, description, is_system) VALUES
('01HZQV8XYZX9ABCDEF12346D', NULL, 'super_admin', 'Super Administrator', 'System-wide administrator with all permissions', true),
('01HZQV8XYZY9ABCDEF12346E', NULL, 'owner', 'Owner', 'Organization owner with full control', true),
('01HZQV8XYZZ9ABCDEF12346F', NULL, 'admin', 'Administrator', 'Organization administrator with most permissions', true),
('01HZQV8XYZ19ABCDEF123470', NULL, 'developer', 'Developer', 'Developer with project and environment access', true),
('01HZQV8XYZ29ABCDEF123471', NULL, 'viewer', 'Viewer', 'Read-only access to organization resources', true);

-- ===================================
-- ASSIGN PERMISSIONS TO SYSTEM ROLES
-- ===================================

-- Super Admin gets all permissions
INSERT INTO role_permissions (id, role_id, permission_id)
SELECT 
    '01HZQV8XYZ3' || LPAD((ROW_NUMBER() OVER())::text, 13, '0'),
    '01HZQV8XYZX9ABCDEF12346D',
    id
FROM permissions;

-- Owner gets most permissions (excluding super admin specific ones)
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01HZQV8XYZ49ABCDEF123480', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZA9ABCDEF123456'), -- users.read
('01HZQV8XYZ59ABCDEF123481', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZB9ABCDEF123457'), -- users.write
('01HZQV8XYZ69ABCDEF123482', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZC9ABCDEF123458'), -- users.delete
('01HZQV8XYZ79ABCDEF123483', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZD9ABCDEF123459'), -- users.invite
('01HZQV8XYZ89ABCDEF123484', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZE9ABCDEF12345A'), -- organizations.read
('01HZQV8XYZ99ABCDEF123485', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZF9ABCDEF12345B'), -- organizations.write
('01HZQV8XYZA9ABCDEF123486', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZG9ABCDEF12345C'), -- organizations.delete
('01HZQV8XYZB9ABCDEF123487', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZH9ABCDEF12345D'), -- organizations.members
('01HZQV8XYZC9ABCDEF123488', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZI9ABCDEF12345E'), -- projects.read
('01HZQV8XYZD9ABCDEF123489', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZJ9ABCDEF12345F'), -- projects.write
('01HZQV8XYZE9ABCDEF12348A', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZK9ABCDEF123460'), -- projects.delete
('01HZQV8XYZF9ABCDEF12348B', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZL9ABCDEF123461'), -- environments.read
('01HZQV8XYZG9ABCDEF12348C', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZM9ABCDEF123462'), -- environments.write
('01HZQV8XYZH9ABCDEF12348D', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZN9ABCDEF123463'), -- environments.delete
('01HZQV8XYZI9ABCDEF12348E', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZO9ABCDEF123464'), -- api_keys.read
('01HZQV8XYZJ9ABCDEF12348F', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZP9ABCDEF123465'), -- api_keys.write
('01HZQV8XYZK9ABCDEF123490', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZQ9ABCDEF123466'), -- api_keys.delete
('01HZQV8XYZL9ABCDEF123491', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZR9ABCDEF123467'), -- roles.read
('01HZQV8XYZM9ABCDEF123492', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZS9ABCDEF123468'), -- roles.write
('01HZQV8XYZN9ABCDEF123493', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZT9ABCDEF123469'), -- roles.delete
('01HZQV8XYZO9ABCDEF123494', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZU9ABCDEF12346A'), -- permissions.manage
('01HZQV8XYZP9ABCDEF123495', '01HZQV8XYZY9ABCDEF12346E', '01HZQV8XYZV9ABCDEF12346B'); -- audit.read

-- Admin gets management permissions (excluding delete organization and security.manage)
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01HZQV8XYZQ9ABCDEF123496', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZA9ABCDEF123456'), -- users.read
('01HZQV8XYZR9ABCDEF123497', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZB9ABCDEF123457'), -- users.write
('01HZQV8XYZS9ABCDEF123498', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZD9ABCDEF123459'), -- users.invite
('01HZQV8XYZT9ABCDEF123499', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZE9ABCDEF12345A'), -- organizations.read
('01HZQV8XYZU9ABCDEF12349A', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZF9ABCDEF12345B'), -- organizations.write
('01HZQV8XYZV9ABCDEF12349B', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZH9ABCDEF12345D'), -- organizations.members
('01HZQV8XYZW9ABCDEF12349C', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZI9ABCDEF12345E'), -- projects.read
('01HZQV8XYZX9ABCDEF12349D', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZJ9ABCDEF12345F'), -- projects.write
('01HZQV8XYZY9ABCDEF12349E', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZK9ABCDEF123460'), -- projects.delete
('01HZQV8XYZZ9ABCDEF12349F', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZL9ABCDEF123461'), -- environments.read
('01HZQV8XYZ19ABCDEF1234A0', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZM9ABCDEF123462'), -- environments.write
('01HZQV8XYZ29ABCDEF1234A1', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZN9ABCDEF123463'), -- environments.delete
('01HZQV8XYZ39ABCDEF1234A2', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZO9ABCDEF123464'), -- api_keys.read
('01HZQV8XYZ49ABCDEF1234A3', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZP9ABCDEF123465'), -- api_keys.write
('01HZQV8XYZ59ABCDEF1234A4', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZQ9ABCDEF123466'), -- api_keys.delete
('01HZQV8XYZ69ABCDEF1234A5', '01HZQV8XYZZ9ABCDEF12346F', '01HZQV8XYZV9ABCDEF12346B'); -- audit.read

-- Developer gets project and environment permissions
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01HZQV8XYZ79ABCDEF1234A6', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZA9ABCDEF123456'), -- users.read
('01HZQV8XYZ89ABCDEF1234A7', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZE9ABCDEF12345A'), -- organizations.read
('01HZQV8XYZ99ABCDEF1234A8', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZI9ABCDEF12345E'), -- projects.read
('01HZQV8XYZA9ABCDEF1234A9', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZJ9ABCDEF12345F'), -- projects.write
('01HZQV8XYZB9ABCDEF1234AA', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZL9ABCDEF123461'), -- environments.read
('01HZQV8XYZC9ABCDEF1234AB', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZM9ABCDEF123462'), -- environments.write
('01HZQV8XYZD9ABCDEF1234AC', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZO9ABCDEF123464'), -- api_keys.read
('01HZQV8XYZE9ABCDEF1234AD', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZP9ABCDEF123465'), -- api_keys.write
('01HZQV8XYZF9ABCDEF1234AE', '01HZQV8XYZ19ABCDEF123470', '01HZQV8XYZQ9ABCDEF123466'); -- api_keys.delete

-- Viewer gets read-only permissions
INSERT INTO role_permissions (id, role_id, permission_id) VALUES
('01HZQV8XYZG9ABCDEF1234AF', '01HZQV8XYZ29ABCDEF123471', '01HZQV8XYZA9ABCDEF123456'), -- users.read
('01HZQV8XYZH9ABCDEF1234B0', '01HZQV8XYZ29ABCDEF123471', '01HZQV8XYZE9ABCDEF12345A'), -- organizations.read
('01HZQV8XYZI9ABCDEF1234B1', '01HZQV8XYZ29ABCDEF123471', '01HZQV8XYZI9ABCDEF12345E'), -- projects.read
('01HZQV8XYZJ9ABCDEF1234B2', '01HZQV8XYZ29ABCDEF123471', '01HZQV8XYZL9ABCDEF123461'), -- environments.read
('01HZQV8XYZK9ABCDEF1234B3', '01HZQV8XYZ29ABCDEF123471', '01HZQV8XYZO9ABCDEF123464'); -- api_keys.read