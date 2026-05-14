---
date: 2026-04-30
status: enacted
tags: [rbac, scope-resolution, override-semantics, floor-scope, multi-round]
note: 16 sub-rounds of P1/P2 reviewer iterations on project RBAC. Most generalisable rules at the bottom of each round.
---

# Project-RBAC resolver — 16 sub-rounds

Two P1 regressions on the project-membership / RBAC surface, then a long series of follow-on rounds that each tightened a different facet. Documents the full arc; the resulting architecture (scope-partitioned hybrid resolver with floor-scope auto-injection + frontend codegenned permission catalog + service-layer pair-validation) is the canonical state.

## Round 1 — `projects:read` guard on single-resource GET broke custom roles

`internal/server/routes.go:341` decorated `GET /api/v1/projects/{projectId}` with `RequirePermission("projects:read")`, and the frontend's URL context resolver (`web/src/lib/context/url-context-manager.ts:148`) calls this endpoint to validate project context for *every* dashboard navigation. On 403 it drops `projectId` from context, so a custom role with `datasets:read` (or any other project-scoped permission) but without `projects:read` couldn't open ANY project page even though downstream feature routes would authorize the user on their own merits.

Web research surveyed 8 production APIs — split ~5/3: GitHub `metadata: read` (auto-included with any other repo permission via the "floor scope" mechanic), GitLab `read_api`, Auth0 `read:clients`, Langfuse `projects:read` enforce explicit floor permission; Linear `GET /workspaces/{id}`, Stripe Connect `GET /accounts/{id}` with `Stripe-Account` header, PostHog `GET /api/projects/{id}` treat membership as sufficient for identity primitives. GitHub's pattern is "permission gate done right" — the floor scope is structurally auto-included with any other repo-scoped permission, preventing the failure mode.

Brokle does NOT implement that auto-include mechanic, and has no custom-role-creation UI today, so building it before the workflow exists is premature complexity.

**Decision**: drop `projects:read` from the single-resource GET only (Linear/Stripe-Connect/PostHog model). Other single-project routes (`PUT`/`DELETE`/`Archive`/`Unarchive`) keep permission guards (state changes are gated). Org-scoped LIST keeps `projects:read` (deciding which projects exist IS gated metadata). Response body (`toProject`) is strictly identity (`id`, `name`, `description`, `organization_id`, `status`, timestamps) — no secrets, settings, plan info, or API keys.

Future upgrade path: when custom-role-creation UI ships, adopt GitHub's floor-scope mechanic — validate at role creation that any role with `<feature>:read` permission auto-includes `projects:read`.

## Round 2 — `IsProjectMember` ignored the active-org filter

Three of the six `project_members` read queries (`ListProjectMembersByProject`, `ListProjectMembersByUser`, `ListUserEffectivePermissionsInScope`) JOINed on `organization_members` with `om.deleted_at IS NULL`, hiding orphan rows; the other three (`IsProjectMember`, `GetProjectMemberRoleID`, `GetProjectMemberByUserAndProject`) read `project_members` directly.

Asymmetry IS the bug class: AddMember's duplicate-check on `IsProjectMember` returns true for orphans the listings hide, blocking re-adds with "user already has a project-level role" while admins can't see the row in the listing to remove it.

Web research evaluated four alternatives: k8s informers (predicate enforced at cache-population, doesn't map to sqlc); SQL views (Sourcegraph rejected for cross-table predicates — codegen friction, EXPLAIN opacity); Rails `default_scope` (no Go equivalent worth building for 6 queries); HashiCorp Boundary repository helpers (production Go idiom — closest match but the JOIN is two lines and inline duplication beats helper abstraction at this scale).

**Decision**: rewrite the three unfiltered queries to use the same active-org JOIN; add a parsed-SQL drift-guard test (`internal/infrastructure/db/project_member_predicate_test.go`) that asserts every `:one`/`:many` query touching `project_members` contains the active-org JOIN OR an explicit `-- noqa: orphan-ok` directive. Two queries get the noqa tag: `CountProjectMembersByProject` (raw analytics count, never used for authz) and `ListUserEffectivePermissionsInScope` (uses an equivalent CTE form the regex can't pattern-match across).

## Generalisable rules (Rounds 1-2)

1. **Single-resource identity GET on a tenancy boundary takes membership, not a separate read scope** — applies to Brokle's resources whose response body is purely identity primitives; the org-scoped LIST and mutations keep their permission guards because they expose *which-resources-can-I-see* and *state changes* respectively. The 2026-04-29 "tenancy ≠ permission" entry was correct for LIST + mutations but mis-fired on identity GET; this round is the reconciliation.
2. **Read-predicate symmetry across all queries on a single table is a structural invariant**, not a per-callsite discipline — when one query gates for authz/identity, every query that reads for authz/identity must gate identically; the asymmetric subset of queries IS the bug class. The drift-guard test is cheaper than the helper abstraction at low query counts (≤10) and matches the sqlc-community recommendation.
3. **noqa-style directives in SQL** (`-- noqa: orphan-ok`) are the right escape hatch for queries that legitimately don't need the predicate (raw counts, equivalent-CTE forms) — they document the exemption inline next to the query, where the next reader will see it.
4. **Same RBAC verb across scopes; scope is the URL path, not the permission string** — project-member DELETE uses `members:remove`, parallel to the org-member DELETE; web research surveyed 6 production peers (Langfuse, GitLab, GitHub, PostHog, Linear, Auth0) and zero prefix membership verbs by scope. Reviewer flagged a P2 mismatch (project-member DELETE was decorated with `members:update` instead of `members:remove`). Cleaned up `members:suspend` from `seeds/permissions.yaml`, `seeds/roles.yaml`, and `internal/core/domain/auth/scope.go` — the last residue of the suspension-deletion sweep.

## Round 5 — Orphan-repair via UPSERT was a phantom-bug fix

The previous round's "orphan-repair via UPSERT" was a phantom-bug fix that introduced a real concurrent-race regression. Reviewer caught it: two admins both pass the filtered-IsProjectMember pre-check; UPSERT silently lets the second writer's role overwrite the first. Both return 200; admin A believes their assignment took effect.

**Verdict**: revert to plain INSERT + classify PK violation as `AlreadyExists` at the repository, translate to `Conflict("project_member", ...)` in the service.

Web research confirms: GitLab `POST /projects/:id/members` returns 409 on duplicate (strict semantics, matches Brokle's POST + AddMember docstring). GitHub `PUT /orgs/{org}/teams/{team}/memberships/{user}` is idempotent, but uses PUT (resource-replacement verb) — NOT the precedent for Brokle's POST. Direction-B (plain INSERT + classify) is the canonical pgx + sqlc idiom across HashiCorp Boundary / Sourcegraph (`appErrors.IsUniqueViolation` predicate, already used at `member_repository.go:49` for org members and `organization_settings_repository.go:39`).

The orphan-repair scenario was unreachable in any production-valid state — `RequireProjectAccess` middleware filters soft-deleted projects + `isOrgMember` check filters non-members BEFORE Create can be called, so a hidden orphan reaching Create requires both a soft-deleted project AND an active org membership simultaneously, which can't happen.

**Generalisable rule (added)**: when proposing a "structural fix" to a write path, audit *every reachable state* the fix's premise depends on — if the premise scenario can't be reached in production, the fix is solving a phantom bug and the structural cost is pure regression surface.

## Round 6 — Endpoint-split was the wrong structural answer

Endpoint-split (`/context` + `/`) is the wrong structural answer for the "frontend resolver vs restrictive role" tension. The previous-previous round dropped `projects:read` from the project detail GET to make the frontend resolver work for custom roles with `datasets:read` but no `projects:read`. The next reviewer correctly flagged that this broke the restrictive-role model.

The right answer is **GitHub's `metadata: read` floor-scope auto-include**, not an endpoint split. Web research surveyed Linear / Vercel / Stripe / Azure REST guidelines: zero peer adoption of the context-vs-detail endpoint split; GitHub's floor-scope is the recognized production pattern.

**Implementation**: restored `projects:read` on `GET /api/v1/projects/{projectId}` + added `internal/seeder/floor_scope_test.go TestRolesCarryProjectsReadFloorScope` build-time invariant (every built-in role with any project-scoped permission MUST also carry `projects:read`). Project-scoped permissions are derived dynamically from `routes.go` via brace-balanced extraction of the `r.Route("/api/v1/projects/{projectId}", func(r chi.Router) { ... })` block — single source of truth, adding a new project-scoped route automatically extends the floor-scope set. The "custom role with `datasets:read` but no `projects:read`" combination is structurally invalid under the floor-scope rule; when a custom-role-creation UI ships, the runtime validator ports the same logic.

**Generalisable rule (added)**: when an endpoint serves two consumers with different authz needs (frontend resolver vs metadata read), the FIRST option to evaluate is "make the role model express the floor scope," not "split the endpoint." Endpoint splits duplicate authz paths and create wide-surface from speculative future need (CLAUDE.md gotchas #14/#37); floor-scope mechanics compress the same idea into one validator + one enforced rule.

## Round 7 — Build-time-only floor-scope enforcement was insufficient

Floor-scope build-time-only enforcement is insufficient when the runtime API is wired — reviewer caught that the previous round's `TestRolesCarryProjectsReadFloorScope` only enforced `projects:read` on built-in roles in `seeds/roles.yaml`, while the runtime custom-role create/update paths (`role_service.go:164` `CreateCustomRole`, `:205` `UpdateCustomRole`) accept arbitrary permission IDs with no validation.

The custom-role API endpoints (`POST/PATCH /api/v1/organizations/{orgId}/roles`) were already wired and reachable, so an org admin could create a custom role with `datasets:read` but no `projects:read` — users assigned that role then 403 on `GET /api/v1/projects/{projectId}` and lose dashboard navigation.

**Fix**: auto-inject the floor permission at the role-service layer (matches GitHub `metadata: read` mechanic, AWS IAM service-linked role baseline, GitLab `read_code` always-on for code-scoped abilities).

**Implementation**:
- (a) added `scope: project | organization` field to every entry in `seeds/permissions.yaml` (data-driven source of truth);
- (b) extended `internal/seeder/types.go PermissionSeed` to read the field;
- (c) replaced the broken hardcoded `projectResources` map in `internal/seeder/seeder.go:452-470` (and the parallel broken prefix-heuristic in `internal/core/domain/auth/scope.go GetScopeLevel`) with the YAML-driven scope;
- (d) added `permissionRepo` to `RoleService` + new `autoInjectFloorScope(ctx, permIDs)` helper that fetches each supplied permission's `ScopeLevel`, and if any is `ScopeLevelProject` and `projects:read` is missing, appends `projects:read`'s ID;
- (e) wired the helper into both `CreateCustomRole` and `UpdateCustomRole` before the bulk permission write;
- (f) added `TestPermissionScopesMatchRoutes` to `internal/seeder/floor_scope_test.go` asserting every permission in the project route block of `routes.go` is tagged `scope: project` in YAML;
- (g) added 5 behavior tests (`role_service_floor_scope_test.go`) covering inject / no-op-already-present / no-op-org-only / empty-input / unknown-permission-rejection.

The auto-injection makes invalid states **unrepresentable** rather than relying on a 422 validator (UI consumer never has to know about the floor).

Web research surveyed GitHub PATs (`metadata: read` auto-inclusion, enforced at PAT creation), AWS IAM (`MalformedPolicyDocument` floor-validation precedent), GitLab custom roles (YAML-driven scope metadata in `Gitlab::CustomRoles::Definition`), Auth0, Hashicorp Boundary, OpenFGA — data-driven scope metadata + auto-injection at write time is the canonical production pattern.

**Generalisable rules**:
1. Build-time invariants on seeded data are necessary but not sufficient when a runtime write-path can mutate the same data shape — the runtime path needs the same invariant as a function of its design, not just a downstream test.
2. Prefer auto-injection over validation when the invariant is a structural floor — auto-injection makes invalid states unrepresentable; validation forces consumers to learn the rule.
3. Scope/category metadata for permissions belongs in YAML and is enforced sync'd to routes.go via build-time test — single source of truth for which permissions decorate which URL trees stays in routes.go; YAML stores per-permission scope; build test guarantees no drift. Replaces three drifting heuristics (seeder hardcoded list, scope.go prefix matcher, scope.go ScopeCategories manual catalog) with one data-driven path.

## Round 8 — Floor-scope auto-injection on shared verb-families is authorization expansion

Reviewer P2: the previous round's auto-injection key (`permissions.scope_level`) tagged `members:read`, `members:update`, `members:remove` as `scope: project` because they appear in the project route block. But `members:*` is the org-rooted membership verb family — used at both `/organizations/{orgId}/members*` AND `/projects/{projectId}/members*`. Tagging the family as project-scoped meant any custom role intended ONLY for org member management (e.g., a hypothetical "Org People Admin" with `[members:read, members:invite, members:remove]`) silently auto-grants `projects:read` — granting project navigation/visibility the role didn't ask for.

The frontend's existing `SCOPE_LEVELS` map already classified all `members:*` as `organization`; backend YAML diverged.

Web research surveyed 5 production peers (GitLab Custom Roles, GitHub, AWS IAM, HashiCorp Boundary, Kubernetes RBAC, OpenFGA) and found **zero peers tag a dual-scope verb with a single scope and apply floor-injection based on that tag** — the previous round's design has no production precedent. GitLab's `available_for: [project, group]` array is the canonical multi-tag answer; Kubernetes' scope-by-binding model is the canonical scope-agnostic answer; GitHub/AWS sidestep via scope-prefixed verb namespaces.

**Decision**: stay with the two-state catalog (`project | organization`) but tag the `members:*` family as `organization` and add a small `sharedVerbResources = {"members"}` exemption to the build-time test. The exemption is the two-state compression of GitLab's `[project, group]` array — same semantics, simpler catalog.

**Generalisable rules**:
1. Classify shared verb-families by SEMANTIC ROOT, not by route-presence accident — `members:*` is org-rooted (an org concept that overrides into project context); even when an individual permission like `members:update` is currently wired only to project routes, family-level classification prevents over-injection.
2. A floor-scope rule that fires on ANY permission with the project tag will over-inject when shared verbs are tagged project-side — verify the rule's behavior across the org-only-intent / project-only-intent / mixed-intent cases before tagging shared verbs.
3. Frontend and backend permission classification must agree — the frontend `SCOPE_LEVELS` map in `web/src/hooks/rbac/use-has-access.ts` is the user-facing source-of-truth for what counts as "org-level vs project-level" in the UI; backend YAML must match or the runtime authz semantics drift from the consumer's expectation.

## Round 9 — Frontend RBAC catalog codegenned from `seeds/permissions.yaml`

Reviewer P2 found 8 frontend `SCOPE_LEVELS` entries misclassified in the wrong direction (`projects:read|write|delete|admin`, `api-keys:read|create|update|delete` tagged `organization` but backend says `project`), causing `useHasAccess` to skip the `projectId` requirement before the access check and render incorrect UI state. 12 more entries misclassified the other way (over-checking but unused), and 21 backend perms were missing from the frontend Scope union entirely.

Web research surveyed 10 production peers (GitHub / GitLab / Linear / Vercel / Stripe / Auth0 / Cal.com / PostHog / Supabase / Outline) — the dominant pattern is **server-computed permissions attached to API responses** (GitLab `userPermissions` GraphQL fragments, Outline `policies.abilities`, PostHog `/users/@me/`, Linear `canEdit`/`canDelete`). Hand-sync with cross-language drift tests is **rare** — most peers compute server-side OR share types via monorepo. Codegen from a YAML/proto source appears in monorepo TS stacks (Cal.com tRPC).

For Brokle's stack (Go backend + separate TS frontend, no monorepo), codegen YAML→TS is the right immediate fix; long-term move to runtime-computed `effectivePermissions` on `/users/me` is the canonical industry pattern (separate ticket).

**Implementation**:
- (a) `cmd/gen-frontend-permissions/main.go` reads `seeds/permissions.yaml` and emits `web/src/generated/permissions.ts` with the `Scope` union + `ScopeLevel` type + `SCOPE_LEVELS` map;
- (b) render logic factored into `internal/seeder/permissions_codegen.go RenderFrontendPermissions` so the drift-guard test can call it in-memory;
- (c) `make gen-frontend-permissions` target wired into `make generate`;
- (d) `internal/seeder/frontend_permissions_test.go TestFrontendPermissionsCatalogInSync` re-runs the render in-memory and asserts byte-equality with the on-disk file — `go test ./...` (already gating CI) catches stale generated files;
- (e) `web/src/hooks/rbac/use-has-access.ts` reduced to importing `Scope`/`ScopeLevel`/`SCOPE_LEVELS` from `@/generated/permissions` (181 lines deleted; hook function unchanged).

Mutation test verified: editing the on-disk TS file without regenerating fails the drift-guard test with a message pointing at `make gen-frontend-permissions`.

**Generalisable rules**:
1. For multi-language permission catalogs, codegen-from-source-of-truth is the structural fix; hand-sync is the workaround. Industry research: zero major peers ship hand-sync + cross-language drift test for RBAC catalogs. Pre-prod is exactly when to invest in the source-of-truth invariant before the catalog grows further.
2. Render logic for codegen lives in an importable package, not the binary's `package main` — the drift-guard test must call the same render function the binary uses; `package main` isn't importable, so factor the rendering into a sibling library package. Brokle's pattern: binary in `cmd/<name>` is I/O wrapper; rendering in `internal/<owner>/<name>_codegen.go`.
3. Generated files commit in-tree with a `// DO NOT EDIT` header pointing at the source-of-truth file and the regen command — first-checkout contributors don't need the codegen toolchain to build; mutation tests in CI catch hand-edits before merge.

## Round 10 — Seeder must reconcile, not skip, when seeded fields gain downstream behavioral significance

Reviewer P2 caught the bug class: `internal/seeder/seeder.go seedPermissions` had skip-on-exists semantics. When the catalog gained `scope_level` (consumed by `RoleService.autoInjectFloorScope` for floor-scope injection AND by `cmd/gen-frontend-permissions` for the frontend `SCOPE_LEVELS` map), upgraded DBs kept their stale `scope_level = 'organization'` (column default) — `make seed` didn't touch them. Result: runtime authz used DB scope while frontend codegen used YAML scope; the two diverged.

Web research surveyed 7 production peers (GitLab fixtures, Discourse SeedFu, Mastodon, Cal.com, Outline, PostHog, Boundary) — two camps emerge: (1) seeder-as-source-of-truth + unconditional upsert by natural key (Mastodon `find_or_initialize_by` + assign + `save!`; Cal.com Prisma `upsert`; Discourse SeedFu `seed`; GitLab fixtures); (2) migration-as-reconciler (PostHog `RunPython`; Boundary `INSERT ... ON CONFLICT DO UPDATE` inside migrations). **100% of surveyed peers update in-place by natural key** for FK-referenced reference data — Mastodon explicitly comments on preserving role IDs across reseeds.

Brokle aligns with camp 1: `make reseed` is wired into the dev workflow + boot-time-reseed pattern is documented; adding a parallel data-migration channel would split the source of truth.

**Decision**: replace skip-on-exists with reconcile-on-exists. Compute desired `scope_level`/`category`/`description` from YAML BEFORE the existence check (same derivation as the create branch); on existing rows, detect drift and call `permissionRepo.Update` (preserves UUID → `role_permissions` FK references stay intact); track `created/updated/unchanged` counters for an aggregate summary log; emit per-row info log only when fields actually changed. Detect-then-update was chosen over unconditional update for log signal-to-noise.

**Generalisable rules**:
1. Skip-on-exists in seeders is correct only when seeded fields have no behavioral significance beyond identity — once a field becomes a runtime decision input (authz, codegen, business logic), the seeder must reconcile to YAML on every reseed or upgraded DBs diverge.
2. Update-in-place by natural key, never delete-recreate for reference rows that are FK-referenced — Mastodon's explicit comment ("preserves role IDs across reseeds") is the canonical pattern.
3. Aggregate summary + per-row log only on real drift is the canonical observability shape.

## Round 11 — Pure OVERRIDE strips org-rooted permissions from the project subtree

Reviewer P1: an org admin (carrying `members:remove`) given a restrictive project viewer override loses `members:remove` on `/api/v1/projects/{projectId}/members*` routes — the override resolver dropped the org role entirely.

Web research surveyed 8 production peers (Langfuse, GitHub, GitLab, AWS IAM, Auth0, HashiCorp Boundary, Kubernetes RBAC, OpenFGA/Cedar/Permify): **7 of 8 use UNION or scope-partitioned hybrid; ZERO use wholesale OVERRIDE across mixed-scope catalogs.** Override appears only WITHIN a single scope partition.

**Critical correction**: Brokle's CLAUDE.md gotcha #39 cited Langfuse as "wholesale OVERRIDE" — verified wrong via `langfuse/web/src/features/rbac/`: Langfuse actually has separate `hasOrganizationAccess` / `hasProjectAccess` entry points (scope-partitioned). This is the second Langfuse mis-citation in Brokle's history (the first was "MAX semantics" in 2026-04-27); the citation drift compounded across rounds because nobody verified against source.

**Decision**: scope-partitioned hybrid resolver. Org-scoped permissions resolve against ORG role only (Langfuse / GitHub / GitLab pattern); project-scoped permissions follow OVERRIDE within their partition.

**Implementation**: rewrote `ListUserEffectivePermissionsInScope` SQL with two scope-filtered CTEs (`org_scope_perms` joins only `active_org` filtered by `WHERE p.scope_level = 'organization'`; `project_scope_perms` does the original OVERRIDE pattern filtered by `WHERE p.scope_level = 'project'`); top-level `UNION` combines them. Active-org-membership guard preserved structurally. Forward-compat: when a future shared-verb family is added (and tagged `scope: organization` per the previous round's exemption pattern), the resolver automatically does the right thing — no per-family code.

**Generalisable rules**:
1. Verify vendor citations against actual source, not just docs/blog posts — citation drift compounded twice across rounds because nobody opened the Langfuse repo. The cost of getting it wrong is years of implementing the wrong semantics.
2. Scope-partitioned resolution is the universal multi-tier RBAC pattern — pure OVERRIDE has zero peer adoption across mixed-scope catalogs because it strips tenancy-rooted concepts (members, billing, etc.) wholesale. The classification follows `permissions.scope_level` from YAML; the resolver consumes it data-drivenly.
3. Once the data model carries scope metadata (post-2026-04-30 round), the resolver MUST consume it — having `permissions.scope_level` populated but ignoring it in the resolver is the structural antecedent of every shared-verb regression.

## Round 12 — Two RBAC consistency fixes (route layer + scope-catalog API used different resolvers)

Reviewer surfaced both in the same round.

**(A)** Project-member-override routes (`/api/v1/projects/{projectId}/members*`) decorated with org-scoped `members:*` verbs only — a custom org role with `members:update` but no `projects:read` could mutate project-level overrides on every project ID it knew, contradicting the floor-scope invariant ("no `projects:read` = no project access"). The shared-verb routes don't transitively get the floor via auto-injection (the previous round's `members:*` reclassification as `scope: organization` was correct for floor-scope auto-injection, but it left these specific routes unguarded). **Fix**: `RequireAllPermissions(authD, "<verb>", "projects:read")` on each of the four project-member-override routes — both the org-scoped verb AND the project-scope floor required.

**(B)** `ScopeService.GetUserScopes` (the resolver behind `POST /api/v1/rbac/users/{id}/scopes/check`, called by the frontend `useHasAccess` hook) ignored project_members overrides — derived ProjectScopes from `orgMemberRepo.GetUserPermissionsInOrganization` filtered by scope_level, never touching `project_members`. The route-layer authz already used `ProjectMemberService.CheckUserPermissionsInScope` → scope-partitioned, override-aware. The two resolvers diverged on every user with a project override (ELEVATE: UI says no but server authorizes; RESTRICT: UI says yes but server rejects). **Fix**: route the project-scope branch of `GetUserScopes` through `projMemberRepo.ListUserEffectivePermissionsInScope` (the same SQL the route layer uses). Single resolver, single source of truth.

Web research surveyed 7 production peers for each issue. **For Issue A**, mature systems express the floor structurally (GitHub's mandatory `metadata:read` at PAT creation; OpenFGA `define edit: [user] and viewer`); middleware-level `RequireAllPermissions` is the lowest-friction tactical answer with the right semantics. **For Issue B**, 6 of 7 peers ship server-attached abilities on entity responses (Outline `policies.abilities`, GitLab `userPermissions`, Linear `canEdit`/`canDelete`); catalog endpoints (Brokle's `POST /scopes/check`) are the divergence-prone outlier (Auth0 explicitly documents the trade-off). The shared-resolver fix is a strict prerequisite to either future direction (catalog-endpoint stays correct, OR migrate to server-attached).

**Generalisable rules**:
1. Shared verbs at route boundaries need explicit floor conjunction — the floor-scope auto-injection at role creation handles single-scope verbs; shared-verb (org-rooted but route-mounted under project URL) need `RequireAllPermissions(verb, floor)` because their org-only scope tag means the floor isn't auto-injected.
2. Two parallel resolvers WILL drift — when the route layer and the scope-catalog API have separate code paths, they diverge silently the moment one path gets a feature update. The fix is single-resolver-source-of-truth — call sites map to the same function.
3. The catalog-endpoint pattern is the wrong long-term shape — server-attached `abilities` on entity responses (Outline / GitLab / Linear / PostHog) eliminates the divergence class structurally. Pre-prod is the right time to plan migration to that shape.

## Round 13 — Stale frontend project endpoints + cross-tenant project-role leak

Reviewer surfaced both in the same review.

**(A)** Frontend `getOrganizationProjects` and `createProject` still called the pre-refactor `/v1/projects?organization_id=...` shape; backend had moved list/create to `/api/v1/organizations/{orgId}/projects` per gotcha #41 (URL-tenancy invariant) AND `Project.List` rejects `?organization_id=` query overrides with 422. The dashboard project-list and project-create flows hit 422/404.

Web research surveyed 6 production peers (GitHub, GitLab, PostHog, Sentry, Grafana, Linear) — hand-written FE fetchers + E2E tests is the dominant non-monorepo pattern; OpenAPI codegen for the in-house SPA is rare (lossy, regen-churn-heavy); Pact has low adoption (GitLab tried + abandoned). **Decision**: hand-sync the two helpers + add a frontend URL-tenancy lint guard (`scripts/lint-conventions.sh` Frontend URL-tenancy invariant section). The lint catches the bug class structurally — any future quoted URL of the form `/v1/<resource>` immediately followed by close-quote/`?` (i.e., flat list/create at a tenant-child resource) fails CI before merge. Same pattern as the existing backend gotcha #41, enforced at the consumer side.

Bonus cleanup: deleted vestigial `API_ENDPOINTS` const in `web/src/lib/constants.ts` (dead code, zero callers, contained the same stale shapes — gotcha #37 "scaffolded-but-unused → delete on sight").

**(B)** `ListUserEffectivePermissionsInScope` `active_project` CTE filtered by `pm.project_id = $3` and gated on `EXISTS(active_org)` for `$2`, but did NOT verify the project actually belonged to the supplied organization. A caller passing `(orgId=A, projectId=B-in-org-C)` could read U's project override from B even though B and A were unrelated tenants. Exploitable via `POST /scopes/check` (catalog endpoint accepts arbitrary inputs from request body); route-layer requests were safe by coincidence (RequireProjectAccess derives orgID from project) but the resolver shouldn't trust caller plumbing.

Web research surveyed 5 production peers (Langfuse, GitHub, GitLab, HashiCorp Boundary, Cedar/OpenFGA/Permify) — universally **never accept redundant tenant identifiers from clients**; canonical pattern is "caller passes leaf resource ID; resolver derives chain." When both are accepted, resolver MUST verify chain via JOIN.

**Decision**: add `JOIN projects p ON p.id = pm.project_id AND p.organization_id = $2 AND p.deleted_at IS NULL` to the `active_project` CTE — Boundary's FK-constrained grant lookup pattern (DB-layer rejection of inconsistent inputs at row read). Cost: one additional PK index hit per resolver call.

Long-term direction (separate ticket): refactor `POST /scopes/check` and `ProjectMemberService.CheckUserPermissionsInScope` signatures to accept `projectId` only — derive `orgId` server-side per Langfuse `hasProjectAccess(session, projectId)` exactly. The SQL JOIN is the prerequisite that makes that refactor safe to do incrementally.

**Generalisable rules**:
1. Frontend helpers go stale silently when backend endpoints move — a backend URL refactor must include a frontend audit, OR the codebase ships a structural lint guard that catches drift at consumer side.
2. Resolvers must self-validate consistency between multiple tenant-identifier inputs — when a resolver takes `(orgID, projectID)` it must verify the relationship via JOIN, not trust the caller.
3. Never accept redundant tenant identifiers from clients in the first place — Langfuse, GitHub, GitLab universally derive parent tenants from leaf resource IDs server-side.

## Round 14 — `GetUserScopes` org-only call must surface project-scoped permissions inherited from the org role

Reviewer P2: the previous round added a `ScopeLevel == organization` filter on the org-scope branch of `ScopeService.GetUserScopes`, which dropped project-scoped permissions (e.g. `projects:read`, `dashboards:*`) from the bare-org catalog response. Built-in `admin`/`developer`/`viewer`/`owner` roles all carry `projects:read`; the regression made `useHasAccess({scope: 'projects:read'})` return false in org-level UI before a specific project was selected, breaking project-list page, "create project" button, etc.

Web research surveyed 7 production peers — **two coherent shapes**: flat catalog with full grant (Langfuse `Membership.scope`, Auth0 `/users/{id}/permissions`, PostHog `effective_permissions`) OR entity-attached abilities (GitHub `permissions` on `/repos/{owner}/{repo}`, GitLab `userPermissions` GraphQL field, Linear `viewer.canX`, Outline `policies.abilities`). **Brokle's `OrganizationScopes` / `ProjectScopes` field-split shape matches ZERO surveyed peers** — it conflates "permission's schema-level `scope_level` (a fact about the permission)" with "context this answer applies to (a fact about the query)" — orthogonal axes that don't need to mirror each other.

**Decision**: (a) tactical fix — populate `ProjectScopes` with the org role's project-scoped baseline when `projectID` is nil (Langfuse partial-context semantics: org-only call returns the inherited baseline); the project branch RESETs `ProjectScopes` before applying the OVERRIDE-aware resolver result so a RESTRICT-style override (org admin → project viewer) correctly drops the admin's project perms. (b) Deferred structural fix: collapse `OrganizationScopes` / `ProjectScopes` field split into single `EffectiveScopes` flat list — Langfuse / Auth0 / PostHog pattern; deferred behind the existing "server-attached abilities on entity responses" migration ticket.

**Generalisable rules**:
1. Catalog APIs under partial tenant context must return inherited baselines, not just the type-of-permission filtered to "matches current context level."
2. Field splits in API response shapes should mirror consumer needs, not schema metadata.
3. Scope-partitioned resolver semantics must apply only when both tenant axes are present — bare-org calls fall back to the org role's full grant; the OVERRIDE-aware resolver runs only when a `projectID` is supplied.

## Round 15 — Cross-tenant fallback leak in resolver + atomic UPSERT-WHERE for orphan-row repair

Reviewer surfaced both in the same round.

**(A)** `ListUserEffectivePermissionsInScope` previously had `JOIN projects p ON p.organization_id = $2` only on the `active_project` CTE — closed the OVERRIDE leak but the `project_scope_perms` org-baseline fallback (`SELECT role_id FROM active_org WHERE NOT EXISTS (active_project)`) STILL fires when active_project is empty for ANY reason (legitimate "no override" OR illegitimate "cross-tenant pair").

Web research surveyed 6 production peers — **6 of 6 structurally prevent mismatch from producing any permissions; ZERO peers fall back to parent permissions on tenant mismatch**. Brokle's fallback was structurally novel and incorrect.

**Fix**: added a `project_in_org` CTE asserting the project belongs to the supplied org and is non-deleted; `project_scope_perms` gates on `EXISTS (project_in_org)` — both the override branch and the org-baseline fallback no-op when the project is foreign. Used `sqlc.arg(project_id)` to disambiguate the param naming (sqlc was naming the param `ID` because `p.id = $3` came first in the WITH clause — `sqlc.arg(project_id)` forces the canonical `ProjectID` field name on the gen struct).

**(B)** AddMember's previous flow — pre-check filtered IsMember + plain INSERT + classify unique-violation as Conflict — failed on hidden orphan rows: pm row physically exists but JOIN filter hides it (e.g., project soft-deleted between layers, race conditions, non-route invocation); pre-check passes, INSERT collides, returns 409 the user can't recover from.

The recheck-and-distinguish pattern (recheck filtered IsMember after INSERT-fail, DELETE+INSERT if hidden) was considered AND REJECTED based on web research — agent flagged a TOCTOU window where two concurrent orphan-repair attempts can DELETE each other's freshly-inserted rows, silently overwriting role assignments. **Zero of 5 surveyed peers do recheck-and-distinguish**: GitHub eager-cascades (no orphans), GitLab revives-in-place (`Member.find_or_initialize_by`), Mastodon/Discourse use background sweeps, Boundary uses `ON CONFLICT DO NOTHING`, Stripe forces caller to two-call `create-or-update`.

The canonical Postgres atomic-upsert-with-state-discriminator pattern (Crunchy Data / Markus Winand) is **`INSERT ... ON CONFLICT DO UPDATE ... WHERE <discriminator>`**.

**Fix**: rewrote `CreateProjectMember` as `:execrows` UPSERT with `WHERE NOT EXISTS (active org_member for project's org AND project not soft-deleted)` — the WHERE clause encodes the orphan-vs-visible-duplicate predicate atomically. Three outcomes: (1) no row → INSERT → RowsAffected=1; (2) hidden orphan → DO UPDATE WHERE matches → role replaced atomically → RowsAffected=1; (3) visible duplicate / concurrent-race loser → DO UPDATE WHERE doesn't match → no-op → RowsAffected=0. Service flow: `n, err := repo.Create(...); if n == 0 → Conflict`. Repository `Create` signature changed from `error` to `(int64, error)`.

**Generalisable rules**:
1. Resolvers must self-validate the entire tenancy chain, not just spot-check one edge — the previous round's JOIN closed the override path but missed the fallback path; multi-CTE resolvers need a single top-level precondition gate that ALL branches respect.
2. Canonical Postgres atomic-upsert-with-discriminator is `ON CONFLICT DO UPDATE WHERE`, not service-layer recheck-and-distinguish — the latter has a TOCTOU window the unique constraint doesn't close because DELETE+INSERT isn't atomic across concurrent repair attempts. Use `:execrows` and check RowsAffected.
3. sqlc parameter names follow column-comparison order in the SQL — when the same `$N` is compared against multiple columns across CTEs, sqlc names the param after the FIRST column it sees; use `sqlc.arg(name)` to force a stable canonical name and avoid `WrongFieldName` regressions when query restructuring changes the order.

## Round 16 — Multi-mode resolver gate must enumerate ALL reachable input modes

Reviewer P1: the previous round's `project_in_org` gate filtered on `p.id = sqlc.arg(project_id) AND project_id <> uuid.Nil` — correct for project-scoped queries (block cross-tenant) but suppressed all project-scoped permissions on org-only resolver calls. The middleware path `RequirePermission → checkPermissions(ctx) → CheckUserPermissionsInScope(userID, orgID, uuid.Nil, perms)` runs on org-scoped routes (`httpctx.ProjectID()` returns `uuid.Nil` outside the project URL tree); resolver returned empty project-scoped perms → routes gated by `projects:read` (project-scoped per YAML) 403'd for owner / admin / developer / viewer / every custom role. Org-scoped project list and create endpoints completely broken.

**Tactical fix**: extend `project_in_org` with `UNION ALL SELECT 1 WHERE sqlc.arg(project_id)::uuid = uuid.Nil` so the gate is satisfied for the org-only case; downstream `project_scope_perms` fallback then correctly surfaces the org role's project-scoped baseline.

Web research surveyed 5 production peers (Langfuse, GitLab `DeclarativePolicy`, HashiCorp Boundary, OpenFGA, Cedar/Permify) — **zero use a single resolver with a nil-sentinel for the resource**; all five split by signature, polymorphic dispatch, explicit scope, or object type. The `UNION ALL synthetic-row sentinel` SQL idiom has no recognized precedent (zero in pgx, sqlc examples, Supabase RLS, PostgREST). Canonical Postgres patterns are `sqlc.narg + WHERE NULL OR ...` or per-mode named queries (Gitea, SigNoz, Grafana).

The tactical fix is acceptable as a bridge; the structural fix (split `ListUserEffectivePermissionsInScope` into `ListUserEffectivePermissionsInOrg` + `ListUserEffectivePermissionsInProject`; service-layer `CheckOrgPermissions` + `CheckProjectPermissions`; middleware dispatches based on `httpctx.ProjectID(ctx)` presence; `uuid.Nil` retires from authz contracts) matches Langfuse exactly + the OpenFGA/Cedar/Boundary consensus and is tracked as the immediate next ticket.

**Generalisable rules**:
1. When adding a precondition gate to a multi-mode resolver, enumerate ALL reachable input modes and assert the gate evaluates correctly in each — middleware's `httpctx.ProjectID() = uuid.Nil` for non-project routes is a load-bearing input mode the previous round's testing missed.
2. `uuid.Nil` as "absent" sentinel is a smell — the zero UUID is a legitimate value, sentinel collisions break silently downstream. Production peers use pointer-as-optional (`*uuid.UUID`), `sqlc.narg` (nullable param), or per-mode named query/service variants.
3. The route-mount-time signal already encodes the resolver mode — threading "is this org-only or project-scoped?" as a sentinel through the middleware → service → SQL chain loses information; per-mode dispatch at the middleware layer (where the URL tree is known) is the canonical answer.

## Round 17 — Write-path symmetry follows read-path symmetry

Once reads filter to hide certain rows (orphans), writes against the same table either need to repair those hidden rows or convert PK collisions into clean errors; otherwise reads-hide + writes-collide produces opaque 500s.

Follow-up P2 reviewer caught this gap on `AddMember` (the previous round changed `IsProjectMember` to filtered but left `CreateProjectMember` as a raw `INSERT`, so hidden-orphan `(user_id, project_id)` pairs surfaced as PK-collision 500s instead of being repaired or rejected).

**Fix**: `CreateProjectMember` is now `INSERT … ON CONFLICT (user_id, project_id) DO UPDATE SET role_id = EXCLUDED.role_id, joined_at = EXCLUDED.joined_at` — single atomic statement, three reachable states (no row → INSERT; visible member → rejected at service pre-check via filtered IsProjectMember; hidden orphan → UPSERT UPDATE replaces stale role). Service code unchanged — pre-check + UPSERT compose correctly.

Industry precedent: GitHub / GitLab / Linear "add member" endpoints are all idempotent on add via the same idiom.

**Generalisable rule (codified)**: when a read predicate explicitly hides rows that physically exist, the write path on the same table must either upsert (repair) or pre-detect and classify (reject) — never silently 500 on PK collisions. The reviewer's "matching cleanup or unique-violation handling" framing is exactly this rule. Detect this class of bug by grepping for tables where some queries JOIN/filter and others don't; check whether INSERT against the table is wrapped in either ON CONFLICT semantics or an `IsAlreadyExists` classifier.

## Round 18 — Scope-catalog endpoints accepting `(orgID, projectID)` must validate the pair upfront

Reviewer P2: `ScopeService.GetUserScopes` populated `OrganizationScopes` from the org role unconditionally when `orgID` was supplied, then the project branch (post-previous-round SQL gate) correctly returned empty for cross-tenant pairs. But OrganizationScopes was never validated against the pair consistency — cross-tenant `(orgId=A, projectId=P_C-in-org-C)` returned `{OrganizationScopes: [members:read, billing:read, ...], ProjectScopes: []}` — partial data for a malformed query.

Web research surveyed 6 production peers (Langfuse, GitLab, GitHub, AWS IAM, Stripe Connect, Auth0) — **6 of 6 return 404 NotFound for cross-tenant identifier mismatches** with uniform reasoning: 403 leaks existence; 422 is for malformed-input-shape not cross-resource consistency; 200-with-empty leaks partial state. Boundary / Sourcegraph / OpenFGA / Cedar all use **service-layer upfront validation as the canonical first layer**; SQL filters are defense-in-depth. Boundary's docs: "scope validation lives where the scope tree is loaded, not where grants are evaluated."

**Tactical fix**: inject `projectRepo orgDomain.ProjectRepository` into `ScopeService`; add upfront pair-validation block at the top of `GetUserScopes` — if `projectID != nil`, load the project and verify `project.OrganizationID == orgID && project.DeletedAt == nil`; on mismatch return `appErrors.NotFound("project")` BEFORE populating any fields. Matches `RequireProjectAccess` middleware semantic + the Langfuse / GitHub / GitLab anti-enumeration consensus.

**Generalisable rules**:
1. Cross-tenant pair validation belongs at the service layer, not the SQL — Boundary's pattern. One service-layer check covers all branches uniformly; SQL filters stay as defense-in-depth.
2. 404 NotFound is the canonical wire shape for cross-tenant mismatches (6/6 surveyed peers); never 403 (leaks existence), never 422 (wrong taxonomy), never 200-with-partial (silent leak).
3. The deeper structural smell is accepting redundant tenant identifiers from clients — zero of 6 surveyed peers do this; the canonical pattern is "the narrowest identifier is canonical; broader scope is derived server-side."

---

## Final state of the resolver (post-round-18)

- **Resolver SQL**: `ListUserEffectivePermissionsInScope` (single function) with two scope-filtered CTEs + active-org-membership precondition + project_in_org self-validation gate.
- **Service**: `ProjectMemberService.CheckUserPermissionsInScope(ctx, userID, orgID, projectID, perms)` — pass `uuid.Nil` as `projectID` to skip the project layer.
- **Middleware**: `RequirePermission` / `RequireAnyPermission` / `RequireAllPermissions` reading `httpctx.MustGetOrganizationID(ctx)` + `httpctx.ProjectID(ctx)`.
- **Floor scope**: auto-injected at custom-role create/update via `RoleService.autoInjectFloorScope` driven by `seeds/permissions.yaml` `scope` field.
- **Frontend catalog**: codegenned at `web/src/generated/permissions.ts` from the same YAML; drift-guard test fails CI on stale generated content.
- **Catalog endpoint**: `POST /api/v1/rbac/users/{id}/scopes/check` validates `(orgID, projectID)` consistency upfront, returns 404 on cross-tenant mismatch.
- **Build-time invariants**: `internal/seeder/coverage_test.go` (every route's permission is seeded), `internal/seeder/floor_scope_test.go` (every built-in role with project-scoped perms carries `projects:read`), `internal/seeder/frontend_permissions_test.go` (frontend codegen in sync), `internal/server/routes_test.go` (chi tree boots without panic).

## Long-term direction (deferred tickets)

1. Per-mode resolver split: `ListUserEffectivePermissionsInOrg` + `ListUserEffectivePermissionsInProject`; middleware dispatches based on `httpctx.ProjectID(ctx)` presence; `uuid.Nil` retires from authz contracts.
2. Scope-catalog endpoint refactor: drop `organizationId` request field, derive from `projectId` server-side (Langfuse `hasProjectAccess(session, projectId)` shape).
3. Server-attached `abilities` on entity responses (Outline / GitLab / Linear / PostHog) — eliminates the catalog-endpoint divergence class structurally.
4. Per-tier verb minting (Auth0 pattern): split `members:*` and `projects:*` into `org_members:*` / `project_members:*` and `org_projects:*` / `project_projects:*`. Retires the shared-verb-family workaround.
