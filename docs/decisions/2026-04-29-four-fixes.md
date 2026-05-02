---
date: 2026-04-29
status: enacted
tags: [rbac, errors, sentinel-collision, security]
---

# RBAC bypass + sentinel collision + state-as-validation fixes

Code-review fix for 4 issues — 2× P1 RBAC bypass, 1× P2 sentinel collision, 1× P2 state-error-as-validation.

## (1) RBAC bypass on project read endpoints

`GET /api/v1/organizations/{orgId}/projects` and `GET /api/v1/projects/{projectId}` — both routes had `RequirePermission("projects:read")` dropped during a previous "tenancy primitive" refactor. Web research surveyed AWS IAM, GitHub OAuth scopes, GCP IAM, k8s RBAC, Stripe Connect, Auth0, Langfuse, PostHog, Linear: **every single one requires BOTH tenancy AND explicit permission** ("tenancy implies read" exists only in single-tier systems like Slack pre-Enterprise Grid). Fix: re-wrap both reads with `RequirePermission(authD, "projects:read")`. The built-in roles (owner/admin/developer/viewer) all carry the perm so the dashboard loader works; custom roles that omit it correctly 403.

## (2) `errors.Is` sentinel collision

`pkg/errors/errors.go` — `ReasonInternal Reason = iota` was 0 (zero value), and the `Is` method had a "wildcard" branch `if t.Reason == ReasonInternal && t.Resource == "" && t.Code == "" { return true }` intended to mean "&Error{} matches any *Error". Result: `errors.Is(err, &Error{Reason: ReasonInternal})` falsely matched every *Error because the target's zero values triggered the wildcard.

Web research surveyed k8s `apimachinery/pkg/api/errors`, CockroachDB `cockroachdb/errors`, HashiCorp Boundary `internal/errors`, gRPC `google.golang.org/grpc/status`, pgx `*pgconn.PgError`, stdlib `syscall.Errno`/`os.PathError`: **zero of them implement a wildcard branch inside `Is`** — they all expose package-level predicates (`IsNotFound`) + `errors.As` for type-only checks.

Two-part structural fix: (a) reserve `iota=0` as `ReasonUnspecified` per proto3 / k8s `StatusReasonUnknown` / gRPC `codes.Unknown` / Boundary `Unknown Code = 0` (real codes start at 100) convention — even if a future contributor re-adds the wildcard, the collision is structurally impossible because `ReasonInternal != 0`; (b) delete the wildcard branch entirely from `Is`, simplifying to clean exact-field-match semantics. Added `TestErrorIs_NoSentinelCollision` regression test pinning both invariants.

## (3) State-error-as-validation

`internal/core/services/organization/project_service.go:78` and 3 sites in `internal/core/services/annotation/item_service.go` — all used `InvalidParam("cannot", "...")` (HTTP 422 with `error.param=cannot` pointing at a non-existent form field) for state-machine rejections that should be `Conflict("project"|"annotation_queue"|"annotation_item", "...")` (HTTP 409). Sibling `ArchiveProject`/`UnarchiveProject` already-state checks already used the correct shape. Added `scripts/lint-conventions.sh` guard banning `InvalidParam("cannot"|"already"|"archived"|"locked"|"deleted"|"inactive"|"expired"|"disabled", ...)` — state-words as field names are the smell.

## Generalisable rules

1. **iota=0 is ALWAYS `XxxUnspecified`** for any new domain enum (proto3 convention; defensive against sentinel collisions).
2. **`Is` wildcards are non-idiomatic in Go** — package-level predicates + `errors.As` are the canonical pattern across k8s/gRPC/Boundary/pgx.
3. **Tenancy ≠ permission** — even when the built-in roles all carry the permission, the route MUST explicitly enforce it because custom roles can omit any permission.
4. **State words as field names are the smell** — when reaching for `InvalidParam` and the "field" name is a verb or state, you want `Conflict` (state-machine, HTTP 409) or `BadRequest` (request shape, HTTP 400) instead.
