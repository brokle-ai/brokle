-- ===================================
-- BROKLE MONOLITH SCHEMA ROLLBACK
-- ===================================

-- Drop all tables in reverse dependency order

-- Drop role_permissions table
DROP TABLE IF EXISTS role_permissions CASCADE;

-- Drop permissions table
DROP TABLE IF EXISTS permissions CASCADE;

-- Drop roles table
DROP TABLE IF EXISTS roles CASCADE;

-- Drop password_reset_tokens table
DROP TABLE IF EXISTS password_reset_tokens CASCADE;

-- Drop email_verification_tokens table
DROP TABLE IF EXISTS email_verification_tokens CASCADE;

-- Drop audit_logs table
DROP TABLE IF EXISTS audit_logs CASCADE;

-- Drop api_keys table
DROP TABLE IF EXISTS api_keys CASCADE;

-- Drop sessions table
DROP TABLE IF EXISTS sessions CASCADE;

-- Drop invitations table
DROP TABLE IF EXISTS invitations CASCADE;

-- Drop environments table
DROP TABLE IF EXISTS environments CASCADE;

-- Drop projects table
DROP TABLE IF EXISTS projects CASCADE;

-- Drop organization_members table
DROP TABLE IF EXISTS organization_members CASCADE;

-- Drop organizations table
DROP TABLE IF EXISTS organizations CASCADE;

-- Drop users table
DROP TABLE IF EXISTS users CASCADE;

-- Drop the updated_at trigger function
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;