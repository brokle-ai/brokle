---
date: 2026-04-30
status: enacted
tags: [scaffolded-code, deletion, schema-discipline, rbac]
---

# Suspension-deletion sweep

Reviewer flagged a P1 data-loss bug + P2 admin-visibility regression: my orphan-cleanup migration predicate `om.status = 'active' AND om.deleted_at IS NULL` would delete `project_members` rows for currently-suspended users (irreversible), and the same predicate in `ListProjectMembersByProject` / `ListProjectMembersByUser` JOINs hid suspended users' overrides from admin listings.

**Investigation**: `MemberStatusSuspended`, `IsSuspended()`, `Suspend()`, `Activate()`, `SuspendMember`, `ActivateMember`, `setMemberStatus` were defined in domain/repo/service but had **zero external callers** — no HTTP route, no CLI command, no business flow used them.

Web research surveyed GitHub, GitLab, Slack, Linear, Notion, Auth0, WorkOS, Clerk + Postgres modeling consensus (Dimitri Fontaine, Karwin): **no major production peer puts suspension on the membership row** — they all do it at the user level (`users.suspended_at` / `users.state` / `users.is_restricted`). Brokle's per-org suspension had no production precedent.

**Decision** (per CLAUDE.md gotcha #37 "scaffolded-but-unused → delete on sight"): delete the entire suspension feature rather than fix predicates around dead code.

**Deleted**: `MemberStatusActive` / `MemberStatusInvited` / `MemberStatusSuspended` constants + `IsActive()` / `IsInvited()` / `IsSuspended()` / `Activate()` / `Suspend()` methods on `OrganizationMember` and `ProjectMember`; `ActivateMember` / `SuspendMember` from auth-domain repo interface and impl; `setMemberStatus` helper; `UpdateMemberStatus` SQL query; `Status` field from `OrganizationMember` / `ProjectMember` / `Member` (org-domain) Go types; `Status` field from `memberResponse` / `projectMemberResponse` wire DTOs; `?status=` query filter from list-members handler; `om.status = 'active'` / `pm.status = 'active'` from every SQL predicate (12 sites in `member.sql` + `project_member.sql`); `status` column INSERT/UPDATE/SELECT from queries.

New migration `20260427171335_drop_member_status_column.{up,down}.sql` drops the column from both tables.

Orphan-cleanup migration (`20260427165217_cleanup_orphan_project_members.up.sql`) predicate fixed to single-axis on `deleted_at` only — preserves the data-loss-prevention semantics correctly even though suspension is no longer modeled (defensive: the column gets dropped in the next migration anyway).

## Generalisable rules

1. **Scaffolded code with zero callers gets deleted, not patched** — adding predicate fixes around unused features is the same anti-pattern as the discipline-based bug class CLAUDE.md repeatedly warns against.
2. **Soft-delete (`deleted_at`) and status-flag suspension are independent axes** — never collapse them into a single predicate. Permission-eligible = both filters; row-visible-to-admin = `deleted_at IS NULL` only; orphan = `deleted_at IS NOT NULL` only.
3. **If suspension is genuinely needed, put it on the user**, not the membership — GitHub / GitLab / Slack / Auth0 / WorkOS / Clerk all do this. Per-membership suspension has no production precedent and is a schema artifact, not a product requirement.