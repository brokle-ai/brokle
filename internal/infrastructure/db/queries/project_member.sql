-- Static queries for project_members. Composite PK is (user_id, project_id).
-- Schema is intentionally simpler than organization_members: no soft delete,
-- no invited_by, no created_at/updated_at — just the four functional columns
-- (user, project, role, status) plus joined_at. Project membership is a
-- lightweight role-elevation override on top of the user's org role
-- (Langfuse MAX semantics — see CheckUserPermissionsInScope).

-- name: CreateProjectMember :exec
INSERT INTO project_members (
    user_id, project_id, role_id, status, joined_at
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetProjectMemberByUserAndProject :one
SELECT * FROM project_members
WHERE user_id = $1 AND project_id = $2
LIMIT 1;

-- name: UpdateProjectMemberRole :exec
UPDATE project_members
SET role_id = $3
WHERE user_id = $1 AND project_id = $2;

-- name: DeleteProjectMember :exec
DELETE FROM project_members
WHERE user_id = $1 AND project_id = $2;

-- name: ListProjectMembersByProject :many
SELECT * FROM project_members
WHERE project_id = $1
ORDER BY joined_at ASC, user_id ASC;

-- name: ListProjectMembersByUser :many
SELECT * FROM project_members
WHERE user_id = $1
ORDER BY joined_at ASC, project_id ASC;

-- name: IsProjectMember :one
SELECT EXISTS (
    SELECT 1 FROM project_members
    WHERE user_id = $1 AND project_id = $2
);

-- name: GetProjectMemberRoleID :one
SELECT role_id FROM project_members
WHERE user_id = $1 AND project_id = $2
LIMIT 1;

-- name: CountProjectMembersByProject :one
SELECT COUNT(*)::bigint AS count FROM project_members
WHERE project_id = $1;

-- ListUserEffectivePermissionsInScope returns the union of permissions the
-- user holds for a given (org, optional project) scope, applying Langfuse
-- MAX semantics: the effective role is the highest of the user's org-level
-- role and (optionally) their project-level override. The query returns
-- the union of permissions across both rows, which is equivalent to MAX
-- because role permissions are nested (Owner ⊇ Admin ⊇ Developer ⊇ Viewer).
--
-- Parameters:
--   $1 user_id
--   $2 organization_id
--   $3 project_id (pass uuid.Nil to skip the project layer)

-- name: ListUserEffectivePermissionsInScope :many
SELECT DISTINCT (p.resource || ':' || p.action)::text AS name
FROM (
    SELECT om.role_id
    FROM organization_members om
    WHERE om.user_id = $1
      AND om.organization_id = $2
      AND om.status = 'active'
      AND om.deleted_at IS NULL

    UNION ALL

    SELECT pm.role_id
    FROM project_members pm
    WHERE pm.user_id = $1
      AND pm.project_id = $3
      AND pm.status = 'active'
      AND $3::uuid <> '00000000-0000-0000-0000-000000000000'::uuid
) AS scoped_roles
JOIN role_permissions rp ON rp.role_id = scoped_roles.role_id
JOIN permissions p       ON p.id       = rp.permission_id;
