-- Static queries for organization_members. Composite PK is
-- (user_id, organization_id). Single lifecycle axis: deleted_at
-- (soft-delete on RemoveMember). The legacy `status` flag was retired
-- on 2026-04-30 — see internal/core/domain/auth/auth.go
-- OrganizationMember docstring for rationale.

-- CreateMember inserts a new organization_members row OR restores a
-- previously soft-deleted one. Composite PK is (user_id, organization_id).
-- RemoveMember soft-deletes (sets deleted_at = NOW()), so a bare INSERT
-- on re-invite would collide on the PK with the soft-deleted row and
-- return a unique-violation — the user could never rejoin.
--
-- The UPSERT-WHERE shape (round 17 / project_members CreateProjectMember
-- precedent) absorbs the conflict structurally:
--
--   1. No row → INSERT → rows = 1.
--   2. Soft-deleted row → UPDATE matches WHERE → restore (deleted_at =
--      NULL, fresh role_id / invited_by / joined_at / updated_at) →
--      rows = 1.
--   3. Active row (concurrent add race) → UPDATE WHERE = false → no-op
--      → rows = 0; the repository translates this to AlreadyExists.
--
-- created_at is intentionally NOT in the SET clause — preserves the
-- original creation timestamp through restore (audit-trail consistency
-- with Brokle's soft-delete pattern).
--
-- name: CreateMember :execrows
INSERT INTO organization_members (
    user_id, organization_id, role_id, joined_at, invited_by,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (user_id, organization_id) DO UPDATE
SET role_id    = EXCLUDED.role_id,
    invited_by = EXCLUDED.invited_by,
    joined_at  = EXCLUDED.joined_at,
    updated_at = EXCLUDED.updated_at,
    deleted_at = NULL
WHERE organization_members.deleted_at IS NOT NULL;

-- name: GetMemberByUserAndOrg :one
SELECT user_id, organization_id, role_id, joined_at, invited_by,
       created_at, updated_at, deleted_at
FROM organization_members
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateMember :exec
UPDATE organization_members
SET role_id    = $3,
    invited_by = $4,
    updated_at = NOW()
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL;

-- name: UpdateMemberRole :exec
UPDATE organization_members
SET role_id    = $3,
    updated_at = NOW()
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL;

-- name: SoftDeleteMemberByUserAndOrg :exec
UPDATE organization_members
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL;

-- DeleteProjectMembersByUserAndOrg cascades the org-removal to every
-- project_members row this user holds in the org's projects. Without
-- this cascade, the rows survive the org-membership soft-delete and
-- become "orphaned overrides" — re-inviting the user would silently
-- restore the prior project-level role because IsProjectMember and
-- ListUserEffectivePermissionsInScope read the project_members table
-- directly. Cleaning up at the write boundary is the same shape as
-- Langfuse / GitHub / Notion ("removing from org removes all per-
-- resource grants"); it makes the orphan class structurally
-- impossible. Schema note: project_members has no organization_id
-- column, so the org link is resolved via projects.organization_id.
--
-- name: DeleteProjectMembersByUserAndOrg :exec
DELETE FROM project_members
WHERE user_id = $1
  AND project_id IN (
    SELECT id FROM projects WHERE organization_id = $2
  );

-- name: ListMembersByOrganization :many
SELECT user_id, organization_id, role_id, joined_at, invited_by,
       created_at, updated_at, deleted_at
FROM organization_members
WHERE organization_id = $1 AND deleted_at IS NULL
ORDER BY joined_at ASC, user_id ASC;

-- name: ListMembersByUser :many
SELECT user_id, organization_id, role_id, joined_at, invited_by,
       created_at, updated_at, deleted_at
FROM organization_members
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY joined_at ASC, organization_id ASC;

-- name: ListMembersByOrganizationAndRole :many
SELECT user_id, organization_id, role_id, joined_at, invited_by,
       created_at, updated_at, deleted_at
FROM organization_members
WHERE organization_id = $1 AND role_id = $2 AND deleted_at IS NULL
ORDER BY joined_at ASC;

-- name: IsMember :one
SELECT EXISTS (
    SELECT 1 FROM organization_members
    WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
);

-- name: GetMemberRoleID :one
SELECT role_id FROM organization_members
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: CountMembersByOrganization :one
SELECT COUNT(*)::bigint AS count FROM organization_members
WHERE organization_id = $1 AND deleted_at IS NULL;

-- name: CountMembersByOrganizationAndRole :one
SELECT COUNT(*)::bigint AS count FROM organization_members
WHERE organization_id = $1 AND role_id = $2 AND deleted_at IS NULL;

-- name: ListMembersByRole :many
SELECT user_id, organization_id, role_id, joined_at, invited_by,
       created_at, updated_at, deleted_at
FROM organization_members
WHERE role_id = $1 AND deleted_at IS NULL
ORDER BY organization_id ASC, user_id ASC;

-- name: ListActiveMembersByOrganization :many
-- "Active" here means non-soft-deleted (deleted_at IS NULL). Suspension
-- is no longer modeled — see auth.go OrganizationMember docstring.
SELECT user_id, organization_id, role_id, joined_at, invited_by,
       created_at, updated_at, deleted_at
FROM organization_members
WHERE organization_id = $1
  AND deleted_at IS NULL
ORDER BY joined_at ASC, user_id ASC;

-- name: CountActiveMembersByOrganization :one
SELECT COUNT(*)::bigint AS count FROM organization_members
WHERE organization_id = $1
  AND deleted_at IS NULL;

-- name: CountActiveMembersByRoleName :many
SELECT r.name        AS role_name,
       COUNT(*)::bigint AS count
FROM organization_members om
JOIN roles r ON r.id = om.role_id
WHERE om.organization_id = $1
  AND om.deleted_at IS NULL
GROUP BY r.name;

-- name: ListUserEffectivePermissionsGlobal :many
SELECT DISTINCT (p.resource || ':' || p.action)::text AS name
FROM organization_members om
JOIN roles r             ON r.id  = om.role_id
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p       ON p.id  = rp.permission_id
WHERE om.user_id = $1
  AND om.deleted_at IS NULL;

-- name: ListUserEffectivePermissionsInOrg :many
SELECT DISTINCT (p.resource || ':' || p.action)::text AS name
FROM organization_members om
JOIN roles r             ON r.id  = om.role_id
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p       ON p.id  = rp.permission_id
WHERE om.user_id = $1
  AND om.organization_id = $2
  AND om.deleted_at IS NULL;

-- name: BulkUpdateMemberRoles :exec
UPDATE organization_members AS om
SET role_id    = v.role_id,
    updated_at = NOW()
FROM (
    SELECT UNNEST($1::uuid[]) AS user_id,
           UNNEST($2::uuid[]) AS organization_id,
           UNNEST($3::uuid[]) AS role_id
) AS v
WHERE om.user_id = v.user_id
  AND om.organization_id = v.organization_id
  AND om.deleted_at IS NULL;
