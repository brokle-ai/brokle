-- Rollback: invitation_audit_events
-- WARNING: This will delete all invitation audit logs

-- Drop indexes first
DROP INDEX IF EXISTS idx_audit_events_invitation;
DROP INDEX IF EXISTS idx_audit_events_actor;
DROP INDEX IF EXISTS idx_audit_events_created;
DROP INDEX IF EXISTS idx_audit_events_type;

-- Drop the table
DROP TABLE IF EXISTS invitation_audit_events;
