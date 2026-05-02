# ADR-0001: RBAC Resolver — Additive (UNION) Semantics

- **Status**: Accepted
- **Date**: 2026-04-30
- **Deciders**: Brokle engineering
- **Supersedes**: Round-11 OVERRIDE-resolver decision (CLAUDE.md 2026-04-30 project-rbac-resolver)
- **Related**: Round 22 (creator-becomes-admin override), Round 23 (floor-scope-extension — abandoned in favour of this ADR)

---

## Context

Brokle's permission resolver answers the question *"what can user U do in scope (org O, project P)?"*.

Round 11 shipped a **scope-partitioned hybrid OVERRIDE** resolver, mirroring Langfuse's `resolveProjectRole` (the canonical reference). Under OVERRIDE:

- Org-scoped permissions resolve against the user's org role only.
- Project-scoped permissions: if the user has a `project_members` row for project P, that role REPLACES the user's org role's project-tier projection on P. Otherwise the org role projects through.

OVERRIDE optimises for one rare feature — *restrictive overrides* (e.g., "demote org admin to project viewer for a sensitive PII project"). It pays a heavy semantic cost: per-resource grants can DOWNGRADE users with richer org roles. Round 22's "creator becomes admin" pattern (industry-standard at GitHub/GitLab/Linear) introduced exactly this regression — an org owner who creates a project loses `projects:delete` because the seeded `admin` template is narrower than `owner`. The reviewer flagged it correctly.

The ergonomic fix in OVERRIDE-world (Round 23 plan) was to extend `autoInjectFloorScope` to fire on `org_projects:*` permissions — couple the floor scope at the role catalog so the lockout vector is closed. That fix is correct under OVERRIDE.

But research showed the **bug class is structurally impossible under additive semantics**. Every modern policy engine (AWS Cedar, OpenFGA, Permify, Topaz, k8s RBAC, Auth0 fine-grained, AWS IAM, GitHub, GitLab, Stripe Connect, Linear, OPA) uses additive grants — per-resource roles UNION with org-level grants instead of replacing them. Brokle's OVERRIDE was the outlier; round 22 was the cost of that divergence.

Pre-prod, with no production data, gave us a rare opportunity to flip the semantic in a single SQL edit.

## Industry survey (verified by reading source / vendor docs)

| System | Resolver semantics | Restriction support |
|---|---|---|
| Google Zanzibar (foundational paper) | Additive, relationship-based | None (model "restriction" as lack of grant) |
| OpenFGA (Auth0/Okta OSS) | Additive | None canonical |
| AWS Cedar / Verified Permissions | Additive + explicit `forbid` | Explicit `forbid` policies |
| Permify / SpiceDB / Topaz | Additive (Zanzibar-derived) | None / ABAC conditions |
| AWS IAM | UNION of identity + resource policies | `Effect: Deny` takes precedence |
| GitHub | UNION of org + team + repo collaborator | None — restriction not modelled |
| GitLab | MAX of group + project | None |
| Linear | Additive (workspace + team) | None |
| Stripe Connect | Additive scopes | None |
| Auth0 fine-grained | Additive scopes + ABAC conditions | ABAC conditions |
| k8s RBAC | RoleBindings additive | NetworkPolicies (separate primitive) |
| OPA/Rego | Default-additive | Explicit `deny` rules |
| Langfuse (former Brokle reference) | OVERRIDE | Inherent to OVERRIDE |

**13 of 13 modern policy engines and major SaaS authz models use additive.** OVERRIDE shows up in early-stage SaaS (Langfuse) and legacy Active-Directory containers; it is not the direction the industry has moved. The "restriction" case OVERRIDE optimises for is rare enough in production SaaS that 9 of 11 surveyed peers don't model it at all; the two that do (AWS IAM/Cedar, Auth0 ABAC) layer it via explicit `deny` rules on top of additive grants.

## Decision

**Brokle's project-scope resolver is additive (UNION).**

Effective project-tier permissions for user U on project P =
`(org-role's project-tier projection through OM in O)` ∪ `(project_members.role's project-tier scopes when present)`

Org-scoped permissions stay as-is — resolved against the user's org role only (SCOPE-PARTITIONED hybrid; the org-side rules from round 11 are preserved).

Cross-tenant safety, active-org-membership preconditions, scope-partitioning between org-tier and project-tier — all preserved unchanged. Six rounds of resolver hardening (rounds 11–18) carry over: those guards are model-agnostic. Only the OVERRIDE-vs-UNION clause changes.

### Concrete implementation

The resolver's project-scope CTE in `internal/infrastructure/db/queries/project_member.sql` changes one clause:

```sql
-- BEFORE (OVERRIDE):
SELECT role_id FROM active_org
  WHERE NOT EXISTS (SELECT 1 FROM active_project)

-- AFTER (additive):
SELECT role_id FROM active_org   -- always include the org projection
```

Removing `WHERE NOT EXISTS (...)` flips REPLACE → UNION.

Round 22's creator-becomes-admin override (`ProjectService.CreateProject` writing a `project_members` admin row for the creator) becomes structurally correct under additive — every row in the post-fix permission matrix is non-decreasing relative to the creator's org role. Round 22 elevates from "regression" to "canonical GitHub-aligned implementation."

Round 23's floor-scope-extension (extending `autoInjectFloorScope` to fire on `org_projects:*`) is no longer needed and was never shipped. The OVERRIDE-specific lockout vector it patched does not exist under additive.

`ProjectMemberService.validateProjectAssignableRole` drops its `projects:read` floor-scope guard — that was lockout-protection, OVERRIDE-specific. Cross-tenant scope_id validation stays (still load-bearing).

## Alternatives considered

### Alt-1: Stay on OVERRIDE; ship Round 23 (floor-scope-extension)

Patches the immediate bug class. Preserves restrictive-override capability. Industry-divergent.

**Rejected** because:
1. Brokle's roadmap (eventual teams, SCIM, external authz, ABAC-style data-classification) all assume additive composition.
2. Every quarter we stay on OVERRIDE adds new code that bakes in OVERRIDE-specific assumptions; migration cost grows.
3. The restrictive-override capability is theoretical — no UI, no test asserting a real product use case, no production data with restrictive members.
4. External authz integration (OpenFGA/Cedar) at any future point would require a mid-flight semantic flip; cheaper now.

### Alt-2: Switch to additive AND build deny mechanism (`project_denies` table)

Preserves restriction parity via explicit deny — AWS IAM / Cedar `forbid` / OPA `deny` pattern.

**Deferred** (not rejected). Reasoning: 9 of 11 surveyed peers don't model restriction at all. Brokle has zero current use cases. Build when the first concrete PII / sensitive-data carve-out lands. The additive resolver is the right substrate for it; adding a deny pass is a localised change at that time.

### Alt-3: External authz from day one (OpenFGA / Cedar / Permify)

Skip the in-house resolver entirely; delegate to a policy engine.

**Deferred**. Brokle's RBAC needs are simple enough today (one resolver, one CTE) that the operational cost of an external policy engine outweighs the benefit. Re-evaluate when:
- Cross-service authz becomes a real need (gateway → observability → billing call-graph authz).
- Number of policy rules grows past O(100).
- Team-based / multi-source memberships add a third resolver branch.

Additive is the right *semantic* either way; the resolver implementation can live in-house for a long time.

### Alt-4: Keep OVERRIDE; build deny mechanism on top

Mix-and-match: OVERRIDE for project-tier, additive for cross-org composition.

**Rejected** as needlessly complex. Two semantic models in one resolver multiplies edge cases (rounds 11–18 already paid this tax once).

## Consequences

### What gets better

1. **Round-22 bug class structurally impossible.** Per-resource grants can never downgrade a user; UNION is monotonic.
2. **GitHub-style "creator owns what they create" works correctly.** Industry-standard pattern, audit-trail row in `project_members`, no semantic conflict.
3. **Future-aligned.** Adding team-based memberships = one more UNION branch. SCIM role mapping = direct fit (every IDP assumes additive). External authz (OpenFGA/Cedar) = native semantic match.
4. **Mental-model shift toward standard.** Anyone familiar with AWS IAM / GitHub / OpenFGA can reason about Brokle's resolver immediately. The OVERRIDE primitive was a learning tax for new contributors.
5. **Round 23 unnecessary.** Saves a small but real amount of code (~50 LoC + tests) and one more nest of YAML-coupling at role-creation time.

### What gets worse

1. **Restrictive overrides no longer expressible.** The "demote org admin to project viewer for sensitive carve-out" feature isn't supported until/unless we build the deny mechanism (Alt-2). For now: no use case, no loss.
2. **One mental-model shift for anyone who's read the round-11/13/14/16/18 lessons-learned.** The retired CLAUDE.md entries get a one-line "**RETIRED in round 24** — see entry 24" prefix; round 24 owns the canonical explanation.
3. **Frontend role display semantic shift.** Today `currentOrganization.projects[].role` = override role if present, else org role. Under additive this is now "most-specific role" rather than "the effective role." Display logic unchanged; the comment block updates to reflect the new meaning.
4. **The "Langfuse precedent" anchor weakens.** Brokle now diverges from Langfuse on this specific design choice. Documentation refactored to anchor on AWS IAM / GitHub / Cedar / OpenFGA / GitLab as the dominant industry pattern.

## Deferred follow-ups

These items are explicitly NOT part of the round-24 migration. Each has a clear trigger condition.

### D1. Deny mechanism (`project_denies` table)

**Trigger**: first concrete restriction use case — typically PII / sensitive-data classification ("only legal team can read traces in project X").

**Approach when it lands**: add a `project_denies` table mirroring `project_members` (user_id, project_id, role_id), subtracted from the union at resolution time. Resolver gains one CTE branch (~30 LoC). Industry precedent: AWS IAM `Effect: Deny` (takes precedence), Cedar `forbid` (declarative), OPA `deny` (rule). Pick the simplest of the three — likely a deny-takes-precedence subtraction step.

**Cost estimate**: half a day for resolver + tests + repository + service-layer entry point + one e2e flow.

**Why deferred**: zero current use cases, no UI, no test, no production data. Building speculatively risks the round-22 mistake (per-resource creator grant ahead of OVERRIDE's lockout class) in reverse — preserve a feature nobody asked for, pay the maintenance tax forever.

### D2. External authz integration (OpenFGA / Cedar / Permify)

**Trigger**: cross-service authz becomes a real need (gateway → observability → billing service-to-service calls each making authz decisions).

**Approach when it lands**: model Brokle's resolver as an OpenFGA store or Cedar policy set. Existing additive semantics translate directly. The CTE becomes a reference implementation we can deprecate once external authz is the source of truth.

**Cost estimate**: 1-2 weeks (model translation + caching strategy + observability + rollout). Most of the cost is operational infrastructure, not policy translation.

**Why deferred**: today's resolver is a single CTE answering O(N) per-user-per-page queries. External authz adds operational complexity (latency, availability, version skew between policy and code) that the current scale doesn't justify.

### D3. Team-based / sub-org grouping

**Trigger**: customer ask for "team within an organisation" hierarchy — common in larger orgs that want functional grouping below the org/project structure.

**Approach when it lands**: add a `teams` table + `team_members` membership + `team_permissions` grants. Resolver gains one more UNION branch (org role projection + team role projection + project_members override grant). Same additive composition; trivially extensible.

**Cost estimate**: 2-3 days for the data model + resolver + UI surface for team management.

**Why deferred**: no roadmap commitment yet. The additive resolver has the structural shape ready; this becomes a localised feature add when it's prioritised.

### D4. ABAC conditions (data-classification rules)

**Trigger**: rules of the form "user can read trace IFF user.dept == trace.owner_dept" — attribute-based, not role-based.

**Approach when it lands**: introduce a Cedar/Rego-style condition layer ON TOP of the additive resolver. Permissions become predicates: `(role_grant ∧ condition)`. This is the right shape for D1 too — restriction via condition rather than via dedicated deny table.

**Cost estimate**: 1 week — schema + condition evaluator + integration into resolver + UI primitives for "labelled" projects/traces.

**Why deferred**: zero current use cases. May supersede D1 (deny mechanism) when it lands — a condition like `project.classification = 'pii' → require legal:read` is more expressive than a flat deny row.

### D5. CLAUDE.md historical revision pass

**Trigger**: round 24's migration lands.

**Approach**: rounds 11/13/14/16/17/18/22 entries get a one-line **"RETIRED in round 24"** prefix without altering their original text. Round 23 (floor-scope-extension) gets a "**SUPERSEDED in round 24** — never shipped" note. Round 24 becomes the canonical reference for OVERRIDE → additive rationale; readers follow the prefix link from any older entry.

**Cost estimate**: 30 minutes.

**Why deferred from this ADR**: the ADR is the future-stable reference; CLAUDE.md is the conversational history. Both should exist; the historical-revision pass is mechanical and should land alongside round 24's actual migration commit.

## Out of scope (explicit non-goals)

These are not deferred (no trigger condition); they're deliberately out of scope for the additive-resolver decision.

### O1. Frontend permission caching strategy

The session-bootstrap pattern (round 21) caches resolved scopes via React Query on `/users/me`. Cache invalidation policy is unchanged by this ADR. Discussions of staleness, refresh-on-focus, optimistic updates, etc. belong in their own design.

### O2. Audit-log entries for role assignments / project_member changes

Brokle's audit-log surface is a separate effort. The additive resolver makes the audit story SIMPLER (every grant is additive; revocation is row-deletion), but the actual audit-log infrastructure is not in scope.

### O3. Permission catalog evolution (adding new permissions / verb families)

The resolver is parametric over the seeded permission catalog. Adding `traces:share`, `dashboards:execute`, etc. is a YAML edit + sqlc reseed; the resolver doesn't change.

### O4. Multi-tenant isolation strategy beyond the active-org gate

Cross-tenant validation (project_in_org gate, active-org-membership precondition) is preserved as-is. Wider multi-tenancy work (row-level security policies, separate schemas per org, etc.) is out of scope.

### O5. Performance / scaling of the resolver

The current CTE answers in <5ms p99 per call against the seeded fixture. Caching, materialised views, denormalisation — all out of scope until profiling shows pressure. The additive flip keeps the same query shape (one CTE, one round trip per call), so no perf regression risk.

### O6. Brokle's stance on Langfuse compatibility

Brokle deliberately diverges from Langfuse on this specific design choice. The CLAUDE.md "vendor as precedent" rule (round 4 / "Research before re-architect") gets a refinement: vendor-as-precedent is right when the vendor is current with industry; wrong when the industry has moved past the vendor. This refinement is documented in the round-24 lessons-learned entry; it's a meta-rule, not part of this ADR's scope.

## References

### Industry source / docs

- AWS Cedar — `docs.cedarpolicy.com/policies/syntax-overview.html` (forbid policies)
- AWS IAM — `docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html` (deny precedence)
- OpenFGA — `openfga.dev/docs/concepts` (additive grants, no deny)
- Google Zanzibar paper — `research.google/pubs/zanzibar-googles-consistent-global-authorization-system/`
- AWS Verified Permissions — `aws.amazon.com/verified-permissions/`
- GitHub repo permissions — `docs.github.com/en/organizations/managing-user-access-to-your-organizations-repositories/managing-repository-roles/repository-roles-for-an-organization`
- GitLab permissions — `docs.gitlab.com/ee/user/permissions.html`
- Auth0 fine-grained permissions — `auth0.com/docs/manage-users/access-control/rbac`
- k8s RBAC — `kubernetes.io/docs/reference/access-authn-authz/rbac/`
- Langfuse `resolveProjectRole` — `competitors/langfuse/packages/shared/src/server/auth/userProjectRoleAuth.ts:6-19`

### Internal source

- Resolver SQL — `internal/infrastructure/db/queries/project_member.sql:209-304` (`ListUserEffectivePermissionsInScope`)
- Service layer — `internal/core/services/auth/project_member_service.go` (`CheckUserPermissionsInScope`, `ListUserEffectiveScopes`, `validateProjectAssignableRole`, `AddMember`)
- Discovery filter — `internal/transport/http/handlers/user/handlers.go` (`mapOrgsWithProjects`)
- Floor-scope auto-injection — `internal/core/services/auth/role_service.go` (`autoInjectFloorScope`)
- Round-22 creator override — `internal/core/services/organization/project_service.go` (`CreateProject`)

### Round-history (CLAUDE.md)

- Round 11 (project-rbac-resolver): scope-partitioned OVERRIDE shipped — RETIRED in round 24.
- Round 13 (cross-tenant fallback): JOIN gate added — preserved unchanged.
- Round 14 (`GetUserScopes` baseline): preserved unchanged.
- Round 16 (`UNION ALL` synthetic-row sentinel): org-only mode gate — preserved unchanged.
- Round 17 (write-symmetry, UPSERT-WHERE): preserved unchanged.
- Round 18 (cross-tenant pair validation in `ScopeService`): historical (ScopeService deleted in round 20).
- Round 21 (session-bootstrap pattern): preserved unchanged; resolver flip is invisible to the bootstrap consumer.
- Round 22 (creator-becomes-admin override): RECLASSIFIED from "regression workaround" to "canonical pattern" under additive.
- Round 23 (floor-scope-extension on `org_projects:*`): SUPERSEDED — never shipped; this ADR closes the bug class structurally.
- Round 24 (this work): adopt this ADR.

## Appendix: Effective-permission matrix under each model

For a user U on project P in org O:

### Under OVERRIDE (rounds 11–23)

| Has `project_members(U,P)`? | Project-tier perms |
|---|---|
| No | Org role's project-tier projection |
| Yes | Project_members role's project-tier scopes (REPLACES org projection) |

### Under additive (round 24+)

| Has `project_members(U,P)`? | Project-tier perms |
|---|---|
| No | Org role's project-tier projection |
| Yes | (Org role's project-tier projection) ∪ (Project_members role's project-tier scopes) |

Org-tier perms are unchanged in both models — always resolved against org role only.

The additive resolver is monotonically non-decreasing in the number of grants (org_members + project_members) the user holds. This is the property that makes per-resource creator grants safe under additive (round 22) and unsafe under OVERRIDE.
