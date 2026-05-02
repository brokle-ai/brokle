-- Static queries for project_members. Composite PK is (user_id, project_id).
-- Schema is intentionally simpler than organization_members: no soft delete,
-- no invited_by, no created_at/updated_at — just the three functional columns
-- (user, project, role) plus joined_at. Project membership is an ADDITIVE
-- per-project role grant: the role's project-tier scopes UNION with the
-- user's org-role projection — see CheckUserPermissionsInScope. Both
-- branches require non-soft-deleted org membership.
--
-- The legacy `status` column (always 'active') was retired on 2026-04-30 —
-- see internal/core/domain/auth/auth.go ProjectMember docstring.

-- CreateProjectMember is `INSERT ... ON CONFLICT DO UPDATE WHERE` — the
-- canonical Postgres atomic-upsert-with-discriminator pattern. Three
-- reachable outcomes, all serialized by the unique index:
--
--   1. No row exists                → INSERT branch  → RowsAffected = 1.
--   2. Hidden orphan row exists     → DO UPDATE WHERE matches → role
--      replaced atomically          → RowsAffected = 1. The orphan
--      condition: NOT EXISTS active organization_members for the
--      project's owning org (or project soft-deleted) — the same
--      filter `IsProjectMember` uses to hide orphans on read.
--   3. Visible duplicate / concurrent-race loser → DO UPDATE WHERE
--      doesn't match → no-op       → RowsAffected = 0. Service
--      translates 0 rows to 409 Conflict.
--
-- A plain INSERT + service-layer recheck-and-distinguish was considered
-- and rejected: it has a TOCTOU window where two concurrent orphan-
-- repair attempts can DELETE each other's freshly-inserted rows.
-- Single-statement ON CONFLICT WHERE has no such window — Postgres
-- serializes the UPSERT decision at the unique index. Pattern: Crunchy
-- Data / Markus Winand canonical Postgres "atomic upsert with state
-- discriminator." See CLAUDE.md 2026-04-30 (project-rbac).

-- name: CreateProjectMember :execrows
INSERT INTO project_members (
    user_id, project_id, role_id, joined_at
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (user_id, project_id) DO UPDATE
SET role_id   = EXCLUDED.role_id,
    joined_at = EXCLUDED.joined_at
WHERE NOT EXISTS (
    SELECT 1
    FROM organization_members om
    JOIN projects p ON p.organization_id = om.organization_id
    WHERE om.user_id          = project_members.user_id
      AND p.id                = project_members.project_id
      AND om.deleted_at IS NULL
      AND p.deleted_at IS NULL
);

-- GetProjectMemberByUserAndProject returns the project_members row for a
-- (user, project) pair only if the user holds an active (non-soft-deleted)
-- organization_members row in the project's owning org. Same predicate as
-- ListProjectMembersByProject — see project_member.sql header for the
-- "predicate symmetry" rationale.

-- name: GetProjectMemberByUserAndProject :one
SELECT pm.user_id, pm.project_id, pm.role_id, pm.joined_at
FROM project_members pm
INNER JOIN projects p
        ON p.id = pm.project_id
       AND p.deleted_at IS NULL
INNER JOIN organization_members om
        ON om.user_id = pm.user_id
       AND om.organization_id = p.organization_id
       AND om.deleted_at IS NULL
WHERE pm.user_id = $1 AND pm.project_id = $2
LIMIT 1;

-- name: UpdateProjectMemberRole :exec
UPDATE project_members
SET role_id = $3
WHERE user_id = $1 AND project_id = $2;

-- name: DeleteProjectMember :exec
DELETE FROM project_members
WHERE user_id = $1 AND project_id = $2;

-- ListProjectMembersByProject returns project_members rows whose user
-- still has a non-soft-deleted organization_members row in the project's
-- owning org. Orphan rows (left behind by pre-2026-04-30 code paths)
-- are hidden — they grant no permissions via the resolver, and showing
-- them invites confused remove-clicks on rows that already grant
-- nothing.

-- name: ListProjectMembersByProject :many
SELECT pm.user_id, pm.project_id, pm.role_id, pm.joined_at
FROM project_members pm
INNER JOIN projects p
        ON p.id = pm.project_id
       AND p.deleted_at IS NULL
INNER JOIN organization_members om
        ON om.user_id = pm.user_id
       AND om.organization_id = p.organization_id
       AND om.deleted_at IS NULL
WHERE pm.project_id = $1
ORDER BY pm.joined_at ASC, pm.user_id ASC;

-- ListProjectMembersByUser returns the user's project_members rows for
-- projects whose owning org still has a non-soft-deleted
-- organization_members row for the user. Same orphan-filter as
-- ListProjectMembersByProject so the user-self view doesn't leak ghost
-- overrides on orgs they've been removed from.

-- name: ListProjectMembersByUser :many
SELECT pm.user_id, pm.project_id, pm.role_id, pm.joined_at
FROM project_members pm
INNER JOIN projects p
        ON p.id = pm.project_id
       AND p.deleted_at IS NULL
INNER JOIN organization_members om
        ON om.user_id = pm.user_id
       AND om.organization_id = p.organization_id
       AND om.deleted_at IS NULL
WHERE pm.user_id = $1
ORDER BY pm.joined_at ASC, pm.project_id ASC;

-- IsProjectMember and GetProjectMemberRoleID gate on the same active-org
-- predicate as the list queries: a project_members row is "real" only if
-- the user holds a non-soft-deleted organization_members row in the
-- project's owning org. Without this filter, orphan rows (left behind by
-- pre-cascade code paths or any future bug that breaks the application-
-- level cascade) would leak into AddMember's duplicate-check, blocking
-- legitimate re-adds. Keeping all 6 read predicates symmetric is the
-- structural fix; see project_member_predicate_test.go for the drift guard.

-- name: IsProjectMember :one
SELECT EXISTS (
    SELECT 1 FROM project_members pm
    INNER JOIN projects p
            ON p.id = pm.project_id
           AND p.deleted_at IS NULL
    INNER JOIN organization_members om
            ON om.user_id = pm.user_id
           AND om.organization_id = p.organization_id
           AND om.deleted_at IS NULL
    WHERE pm.user_id = $1 AND pm.project_id = $2
);

-- name: GetProjectMemberRoleID :one
SELECT pm.role_id
FROM project_members pm
INNER JOIN projects p
        ON p.id = pm.project_id
       AND p.deleted_at IS NULL
INNER JOIN organization_members om
        ON om.user_id = pm.user_id
       AND om.organization_id = p.organization_id
       AND om.deleted_at IS NULL
WHERE pm.user_id = $1 AND pm.project_id = $2
LIMIT 1;

-- noqa: orphan-ok — raw count for analytics / capacity, never used for
-- authz. Including orphan rows here is harmless and matches the pre-fix
-- behavior; gating would require a 2-table JOIN that adds cost without
-- changing any decision the count drives.

-- name: CountProjectMembersByProject :one
SELECT COUNT(*)::bigint AS count FROM project_members
WHERE project_id = $1;

-- ListUserEffectivePermissionsInScope returns the permissions the user
-- holds for a given (org, optional project) scope, applying SCOPE-
-- PARTITIONED ADDITIVE semantics. Three structural rules:
--
--   1. Non-soft-deleted-org-membership precondition. BOTH the org branch
--      and the project branch require the user to currently hold a
--      non-soft-deleted `organization_members` row in the org. A stale
--      `project_members` row (left behind after the user was removed
--      from the org) grants ZERO permissions on its own. This closes
--      the privilege-retention bug where org-removal didn't cascade to
--      project_members.
--
--   2. Org-scoped permissions resolve against the ORG ROLE only. A
--      project membership row has NO effect on org-scoped permission
--      resolution. This is the structural fix for the regression where
--      a restrictive project override (e.g., project viewer) stripped
--      org-rooted verbs like `members:remove` from an org admin —
--      breaking project-member-management routes that legitimately use
--      org-scoped verbs (gotcha #41 / Langfuse-precedent shared verbs).
--
--   3. Project-scoped permissions are ADDITIVE — the org role's
--      project-tier projection UNIONs with any project_members override
--      grant. Per-resource grants can ONLY ADD permissions, never
--      reduce them. This is the AWS IAM / Cedar / OpenFGA / GitHub /
--      GitLab / Linear / Stripe-Connect convention (13 of 13 modern
--      policy engines and SaaS authz models surveyed).
--
-- Migration history: round 11 shipped OVERRIDE (project_members.role
-- REPLACES org projection on the project), mirroring Langfuse's
-- `resolveProjectRole`. Round 22 hit the regression class endemic to
-- OVERRIDE — a per-resource grant downgraded users with richer org
-- roles (e.g., owner has `projects:delete`, the seeded admin override
-- does not). Round 24 flipped the resolver to additive, structurally
-- eliminating the regression class. See docs/adr/0001-rbac-additive-
-- semantics.md for the full rationale + industry survey.
--
-- Restriction support (e.g., "demote org admin to project viewer for a
-- sensitive PII project") is no longer expressible via project_members
-- alone — additive resolvers can only ADD. A future deny mechanism
-- (`project_denies` table, AWS IAM `Effect: Deny` shape) will land if
-- and when a real restriction use case appears. ADR-0001 §D1 has the
-- trigger condition + approach.
--
-- Parameters:
--   $1 user_id
--   $2 organization_id
--   $3 project_id (pass uuid.Nil to skip the project layer)

-- noqa: orphan-ok — uses an equivalent CTE form (active_org / active_project
-- with `om.deleted_at IS NULL`) instead of a direct JOIN. The structural
-- invariant is preserved; the linter just can't pattern-match across the
-- CTE boundary.

-- name: ListUserEffectivePermissionsInScope :many
WITH active_org AS (
    SELECT om.role_id
    FROM organization_members om
    WHERE om.user_id = $1
      AND om.organization_id = $2
      AND om.deleted_at IS NULL
    LIMIT 1
),
-- project_in_org self-validates the (orgID, projectID) pair: the project
-- must exist, belong to the supplied organization, and not be soft-
-- deleted. The `project_scope_perms` CTE gates on `EXISTS
-- (project_in_org)` — without this, a cross-tenant pair (orgId=A,
-- projectId=B-in-org-C) would still leak A's project-scoped baseline as
-- if it applied to B. Pattern: HashiCorp Boundary's FK-constrained
-- grant lookup (DB-layer rejection of inconsistent inputs at row read);
-- zero of 6 surveyed peers fall back to parent permissions on tenant
-- mismatch. See CLAUDE.md 2026-04-30 (project-rbac).
--
-- project_in_org gates `project_scope_perms` against three input modes:
--
--   1. Org-only (project_id = uuid.Nil): synthetic 1-row branch fires.
--      `active_project` stays empty (no specific project to look up);
--      the `project_scope_perms` org-baseline (`SELECT role_id FROM
--      active_org`) surfaces the user's org-role project-scoped
--      baseline. This is the "what project-scoped permissions does the
--      user hold in this org generally?" semantic — middleware on
--      org-scoped routes (e.g., /api/v1/organizations/{orgId}/projects
--      gated by `projects:read`) passes uuid.Nil for projectID and
--      depends on this baseline.
--   2. Same-tenant (project_id set, project belongs to org): first
--      branch fires. `active_project` resolves any override grant;
--      `active_org` always contributes the org-projection. UNION = full
--      effective set (additive).
--   3. Cross-tenant (project_id set, project NOT in org): both branches
--      empty → project_scope_perms blocked. Cross-tenant leak prevented.
--
-- The two branches of project_in_org are mutually exclusive on
-- `sqlc.arg(project_id) = Nil`, so project_in_org has 0 or 1 rows.
-- UNION ALL avoids dedupe overhead.
project_in_org AS (
    SELECT 1
    FROM projects p
    WHERE p.id = sqlc.arg(project_id)::uuid
      AND p.organization_id = $2
      AND p.deleted_at IS NULL
      AND sqlc.arg(project_id)::uuid <> '00000000-0000-0000-0000-000000000000'::uuid
    UNION ALL
    SELECT 1
    WHERE sqlc.arg(project_id)::uuid = '00000000-0000-0000-0000-000000000000'::uuid
),
-- active_project gates on the same project_in_org precondition as a
-- defense-in-depth (the project_scope_perms top-level gate is the
-- load-bearing check; this keeps active_project's invariants explicit).
active_project AS (
    SELECT pm.role_id
    FROM project_members pm
    WHERE pm.user_id = $1
      AND pm.project_id = sqlc.arg(project_id)::uuid
      AND sqlc.arg(project_id)::uuid <> '00000000-0000-0000-0000-000000000000'::uuid
      AND EXISTS (SELECT 1 FROM active_org)
      AND EXISTS (SELECT 1 FROM project_in_org)
    LIMIT 1
),
-- Org-scoped permissions: resolved against ORG role ONLY. Project
-- membership has no effect on org-scoped resolution (Langfuse / GitHub /
-- GitLab pattern). A project role granting an org-scoped permission is
-- silently inert for this branch — see CLAUDE.md 2026-04-30 (project-rbac).
org_scope_perms AS (
    SELECT DISTINCT (p.resource || ':' || p.action)::text AS name
    FROM active_org
    JOIN role_permissions rp ON rp.role_id = active_org.role_id
    JOIN permissions p       ON p.id       = rp.permission_id
    WHERE p.scope_level = 'organization'
),
-- Project-scoped permissions: ADDITIVE — org role's project-tier
-- projection UNIONs with any project_members override grant. ELEVATE
-- (org viewer → project admin) is naturally expressible: the override
-- adds scopes the org role didn't grant. RESTRICT (org admin → project
-- viewer for sensitive carve-outs) is NOT expressible via this resolver
-- alone; if needed, layer an explicit deny pass per ADR-0001 §D1.
--
-- The top-level `EXISTS (project_in_org)` gate is the load-bearing
-- cross-tenant guard: when the project doesn't belong to the supplied
-- org, both the override branch and the org-baseline branch yield
-- empty (no rows), preventing the resolver from leaking the org's
-- project-scoped baseline as permissions on a foreign project.
--
-- Both subqueries always evaluate; the UNION ALL inside is the additive
-- operator. DISTINCT at the outer level deduplicates permissions that
-- both the org role and the override grant.
project_scope_perms AS (
    SELECT DISTINCT (p.resource || ':' || p.action)::text AS name
    FROM (
        SELECT role_id FROM active_project
        UNION ALL
        SELECT role_id FROM active_org
    ) src
    JOIN role_permissions rp ON rp.role_id = src.role_id
    JOIN permissions p       ON p.id       = rp.permission_id
    WHERE p.scope_level = 'project'
      AND EXISTS (SELECT 1 FROM project_in_org)
)
SELECT name FROM org_scope_perms
UNION
SELECT name FROM project_scope_perms;
