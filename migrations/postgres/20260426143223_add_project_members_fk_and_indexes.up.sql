-- Migration: add_project_members_fk_and_indexes
-- Created: 2026-04-26T14:32:23+05:30
--
-- project_members was created in 20250908140000_normalized_rbac_schema.up.sql
-- with the FK to projects deferred ("project_id FK will be added when projects
-- table is created"). Projects exists since the initial schema, so we can now
-- enforce referential integrity. Phase 6 (RBAC) wires this table into the
-- effective-permission resolver, so the FK becomes load-bearing.

ALTER TABLE project_members
    ADD CONSTRAINT project_members_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
