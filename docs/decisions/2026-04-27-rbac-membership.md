---
date: 2026-04-27
status: enacted
tags: [rbac, membership, scope-validation]
---

# RBAC + membership edge-case fixes

Three behavioural regressions and one mis-cited model.

## (1) `ProjectMemberService.AddMember` accepted any role

No scope_type / scope_id check, so an admin could attach a system role OR an org-B custom role to an org-A project (privilege escalation). Fixed by adding `validateProjectAssignableRole` — rejects scope=system, rejects org-custom roles whose scope_id != project's org, rejects project-scoped roles whose scope_id != project. Pattern verified at GitHub teams (org-bound), Snowflake (`GRANT DATABASE ROLE` rejects cross-DB), Microsoft Entra (`directoryScopeId`), Auth0 (tenant-bound API token).

## (2) Stale `project_members` after `RemoveMember`

`OrganizationMemberService.RemoveMember` only soft-deleted the org_members row; the user's project_members survived and the permission resolver UNION'd them in (privilege retention). Fixed in SQL — `ListUserEffectivePermissionsInScope` now requires an active org_members row in BOTH branches; a stale project_members row alone grants zero. Picked query-time gate over eager cascade because it preserves audit trail and re-invitation UX (Snowflake / Langfuse implicit model; GitHub eagerly cascades inherited team membership but keeps direct grants).

## (3) Stale `role_permissions` after YAML expansion (65→86 perms)

Blocked existing dev users from any route guarded by a new perm. The seeder's `seedRoles → UpdateRolePermissions` already does atomic delete+reinsert (`role_repository.go:207-224`), but the `make dev` flow needs to ensure boot-time reseed runs (Kubernetes pattern: *"At each start-up, the API server updates default cluster roles with any missing permissions"*). Documented + smoke test added that asserts every `RequirePermission(authD, "...")` permission in `routes.go` is present in `seeds/permissions.yaml` — build-time drift guard.

## (4) Mis-cited "Langfuse MAX semantics"

The code, gotcha #39, and 2026-04-26 entry all claimed Brokle implemented "Langfuse MAX/UNION semantics", but the actual Langfuse RBAC docs (langfuse.com/docs/administration/rbac) describe **OVERRIDE semantics** (project role REPLACES org role when present, not UNION). Brokle was implementing UNION (project role can only elevate). For an LLM observability platform where some projects may carry data even org admins shouldn't see (the load-bearing restrict case), override is the right model — strictly more expressive. SQL rewritten to override in the same change.

**Generalisable rule**: when citing a vendor's model as architectural precedent, verify against their actual docs/source — the vendor name is a load-bearing reference future contributors trust. The cost of getting it wrong is they implement the wrong semantics for years before someone re-reads the source.
