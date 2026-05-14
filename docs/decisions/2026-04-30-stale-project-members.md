---
date: 2026-04-30
status: enacted
tags: [rbac, cascade, error-naming, lint]
---

# Stale `project_members` post-org-removal + sentence-style first-arg ban

Two more code-review fixes that emerged from the same series of error-API + RBAC reviews.

## (1) Stale `project_members` after org removal (P2)

When a user was removed from an organization (`OrganizationMemberRepository.Delete` / `MemberRepository.DeleteByUserAndOrg`), only the `organization_members` row was soft-deleted. The user's `project_members` rows survived. A previous mitigation added an `INNER JOIN organization_members` on the LIST queries (`ListProjectMembersByProject`, `ListProjectMembersByUser`) and the permission resolver (`ListUserEffectivePermissionsInScope`), so orphans were *hidden* — but `IsProjectMember` and `GetProjectMemberRoleID` still read `project_members` directly without the join. Result: re-inviting the user silently restored their prior project-level role because those direct-read predicates returned the stale row. The query-time mitigation was treating the symptom (hide the row) instead of the cause (delete the row at the write boundary).

**Fix**: added `DeleteProjectMembersByUserAndOrg` SQL in `internal/infrastructure/db/queries/member.sql` (resolves the project→organization link via `WHERE project_id IN (SELECT id FROM projects WHERE organization_id = $2)` since `project_members` has no `organization_id` column), and called it from BOTH org-removal paths (`internal/infrastructure/repository/auth/organization_member_repository.go Delete` AND `internal/infrastructure/repository/organization/member_repository.go DeleteByUserAndOrg`) BEFORE the soft-delete. Cascade is atomic via the request-scoped `TxManager`.

With orphans now structurally impossible, the join in the LIST/effective-permissions queries becomes pure defence-in-depth (kept). Mirrors Langfuse / GitHub / Notion semantics: "removing from org removes all per-resource grants."

**Generalisable rule**: when a write produces a state that downstream reads must filter against, the fix is always to make the write produce the right state, not to add filters at every read site — read filters drift, writes are atomic.

## (2) Sentence-style first-arg to `NotFound` / `AlreadyExists` (P3)

The new `NotFound(resource string, opts ...Option)` derives the public message via `<resource> not found`. Calling `NotFound("user not found")` produced "user not found not found" on the wire. Audit found 13+ sites across `auth_service.go`, `role_service.go`, `permission_service.go`, `session_service.go`, `oauth_session.go` using this shape.

**Fix**: bulk perl-converted simple cases (`NotFound("user not found")` → `NotFound("user")`); attached the original prose via `WithMessage(...)` when the phrasing was load-bearing (OAuth session "expired or invalid", "no executions found for this evaluator"). Added `scripts/lint-conventions.sh` guard banning `appErrors.NotFound("X not found"|"X already exists"|"X does not exist"|"X missing")` so the sentence-style misuse can't return.

The lint compounds: with the existing `InvalidParam` state-word guard, the `InvalidInput` deletion guard, the empty-resource guard, and now this sentence-style guard, the four most common error-API mis-shapes are all structurally enforced.

**Generalisable rule (codified)**: every public constructor with a positional "name/identifier" arg gets a lint check that the arg is a single entity name (single word, snake_case allowed, no whitespace, no "not found" / "already exists" / "expired" suffix). Naming-discipline lints scale; reviewer-discipline doesn't.
