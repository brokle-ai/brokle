---
date: 2026-04-30
status: enacted
tags: [migrations, rbac, cascade, schema-discipline]
---

# Orphan-cleanup migration

Reviewer flagged the historical-orphan regression. The 2026-04-30 service-layer cascade (`OrganizationMemberService.RemoveMember` + `MemberService.RemoveMember` wrap `DeleteAllInOrgForUser` + `Delete` in `WithinTransaction`) prevents new orphans from forming, but databases that ran the prior code path before the fix landed still hold pre-existing `project_members` orphans.

The list-query JOIN filter (`project_member.sql` 44-49, 64-68) hides them but `IsProjectMember` / `GetProjectMemberRoleID` read raw — so re-inviting a user trips `AddMember`'s `IsProjectMember` precheck with "user already has a project-level role" while admins can't see the row in the listing to delete it.

Web research surveyed GitHub, GitLab, Stripe, Linear, Notion, Discourse, k8s `OwnerReferences`, Postgres FK CASCADE soft-delete constraints: **6 of 7 production schemas use application-level cascade + read-side filter for drift; soft-delete + FK CASCADE is structurally incompatible** (FK CASCADE doesn't fire on UPDATE).

GitLab's `Members::CleanupService` is the canonical precedent — application cascade going forward + periodic / one-shot historical sweep.

Schema-level "denormalize organization_id + trigger cascade" rejected because triggers hide control flow from grep (Rails Paranoia precedent: brittle), and a derived column can drift; GitLab's polymorphic-FK existence proves "structural FK fix" still leaks.

"Dev DBs get reset" rejected because the same dismissal class produced the original bug.

**Fix**: one-shot DELETE migration at `migrations/postgres/20260427165217_cleanup_orphan_project_members.up.sql`. Predicate matches the existing JOIN filter exactly — a row is orphaned if there's no active, non-soft-deleted `organization_members` for `(user_id, project's owning org)`. Idempotent (re-runs on clean DB delete zero); forward-only (down migration is no-op since restoring the rows would resurrect the bug class). The list-query JOIN filter stays as defense-in-depth for any future bug that creates orphans.

**Generalisable rule**: when a cross-aggregate constraint is enforced application-side (cascade in service tx), also add a one-shot historical sweep migration the first time the constraint is introduced — the application-cascade prevents new violations but doesn't remediate state from before the rule existed. The Stripe / GitLab / Linear pattern is "service tx + sweep" together, not either alone.
