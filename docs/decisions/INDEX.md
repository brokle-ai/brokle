# Decisions Archive

Date-prefixed lessons learned, lifted from the pre-2026-04-30 root `CLAUDE.md`. Each entry preserves the original rationale verbatim — the *why* behind a structural fix or convention. Loaded on demand (grep / `Read` by Claude when relevant), not on every turn.

For formal architectural decisions with full context+alternatives sections, see `docs/adr/` (numbered).

## How to use this archive

- **Searching by topic**: `grep -l '<keyword>' docs/decisions/`. Entries are tagged with topic prefixes in their filenames.
- **Searching by date**: filenames are `YYYY-MM-DD-<slug>.md`.
- **Adding new entries**: when shipping a non-obvious structural fix, write a new file here with the same shape — frontmatter + verbatim narrative. The `Lessons Learned` section was retired from CLAUDE.md to keep that file scannable; new lessons land here directly.

## By topic

### Error taxonomy & wire envelope
- [2026-04-19 — AppError redesign (closed Type + open Code)](2026-04-19-apperror-redesign.md)
- [2026-04-20 — Wire-contract migration (Stripe/OpenAI shape)](2026-04-20-wire-contract.md)
- [2026-04-26 — `code` field demoted to opt-in; tenancy-from-URL hardening](2026-04-26-code-field-and-tenancy.md)
- [2026-04-27 — Error-translation re-architecture (Reason at source)](2026-04-27-error-translation.md)
- [2026-04-28 (morning) — Single `pkg/errors` package unification](2026-04-28-error-unification-morning.md)
- [2026-04-28 (afternoon) — Legacy `*AppError` deletion](2026-04-28-error-unification-afternoon.md)
- [2026-04-28 (evening) — `InvalidParam` constructor sweep](2026-04-28-invalidparam-sweep.md)
- [2026-04-28 (late evening) — Message-as-positional sweep](2026-04-28-message-positional-sweep.md)
- [2026-04-28 (final) — `InvalidInput` deletion + `InvalidFields` constructor](2026-04-28-invalidinput-deletion.md)
- [2026-04-29 — RBAC bypass + sentinel collision + state-as-validation fixes](2026-04-29-four-fixes.md)
- [2026-04-30 (xor) — XOR validation uses `BadRequest`, not synthetic `param`](2026-04-30-xor-bad-request.md)
- [2026-04-30 (final) — Domain `errors.go` files deletion](2026-04-30-domain-errors-deletion.md)

### RBAC & membership
- [2026-04-27 — RBAC + membership edge-case fixes](2026-04-27-rbac-membership.md)
- [2026-04-30 — Project-RBAC resolver (16 sub-rounds, scope-partitioned hybrid)](2026-04-30-project-rbac-resolver.md)
- [2026-04-30 — Suspension-deletion sweep](2026-04-30-suspension-deletion.md)
- [2026-04-30 — Orphan-cleanup migration](2026-04-30-orphan-cleanup.md)
- [2026-04-30 — Stale `project_members` post-org-removal + sentence-style first-arg ban](2026-04-30-stale-project-members.md)

### HTTP framework / chi v5
- [2026-04-24 — Huma v2 + humachi removal, pure chi v5 + `pkg/request`/`pkg/response`](2026-04-24-chi-migration.md)
- [2026-04-24 — chi `r.Route` Mount collision diagnosis](2026-04-24-chi-route-mount-collision.md)

### Service layer & abstractions
- [2026-04-24 — Service-interface sweep (return concrete, not interface)](2026-04-24-service-interfaces.md)

### Repositories, schemas, nullable fields
- [2026-04-18 — Nullable-domain alignment sweep](2026-04-18-nullable-domain-alignment.md)
- [2026-04-18 — Nested-actor LEFT JOIN hydration](2026-04-18-nested-actor-hydration.md)
- [2026-04-18 — 91% omitempty empirical finding](2026-04-18-omitempty-convention.md)
- [2026-04-18 — Scaffolded `BillingService` deletion](2026-04-18-billingservice-deletion.md)
- [2026-04-30 — Status-column pre-prod schema edit (delete the migration)](2026-04-30-status-column-pre-prod-edit.md)

### Handler layer conventions
- [2026-04-15 — One error pattern, inline parsing across handlers](2026-04-15-handler-conventions.md)
- [2026-04-17 — `Must*` auth helpers (no defensive 401 checks)](2026-04-17-must-helpers.md)
- [2026-04-17 — `pkg/uid` minimised (don't wrap `uuid.UUID`)](2026-04-17-uid-minimised.md)
- [2026-04-17 — Bulk `sed` ordering pitfalls](2026-04-17-sed-ordering.md)
- [2026-04-17 — Scaffolded admin/token routes deletion](2026-04-17-admin-tokens-deletion.md)

### Async, email, workers
- [2026-04-14 — Detached email-send goroutine pattern](2026-04-14-detached-email-send.md)
- [2026-04-14 — Domain-entity field-type changes require handler audit](2026-04-14-domain-entity-field-changes.md)
- [2026-04-14 — Swagger annotations must reference DTOs, not domain entities](2026-04-14-swagger-dto-annotations.md)

### SDKs
- [2026-04-20 — SDK error-hierarchy audit (one shared family + extends)](2026-04-20-sdk-error-hierarchy.md)

### Rate limiting / probes
- [2026-04-20 — Rate-limit layering audit (per-IP for unauth, per-principal for authed)](2026-04-20-rate-limit-layering.md)
- [2026-04-20 — Probe-bypass-middleware via outer `http.ServeMux` dispatcher](2026-04-20-probe-bypass.md)

## Chronological

| Date | Slug | Topic |
|---|---|---|
| 2026-04-14 | `2026-04-14-detached-email-send.md` | async/email |
| 2026-04-14 | `2026-04-14-domain-entity-field-changes.md` | domain/handlers |
| 2026-04-14 | `2026-04-14-swagger-dto-annotations.md` | OpenAPI/DTOs |
| 2026-04-15 | `2026-04-15-handler-conventions.md` | handlers |
| 2026-04-17 | `2026-04-17-must-helpers.md` | auth/middleware |
| 2026-04-17 | `2026-04-17-sed-ordering.md` | tooling/process |
| 2026-04-17 | `2026-04-17-uid-minimised.md` | ID generation |
| 2026-04-17 | `2026-04-17-admin-tokens-deletion.md` | scaffolded code |
| 2026-04-18 | `2026-04-18-nullable-domain-alignment.md` | repos/schema |
| 2026-04-18 | `2026-04-18-billingservice-deletion.md` | scaffolded code |
| 2026-04-18 | `2026-04-18-nested-actor-hydration.md` | repos/perf |
| 2026-04-18 | `2026-04-18-omitempty-convention.md` | JSON convention |
| 2026-04-19 | `2026-04-19-apperror-redesign.md` | error taxonomy |
| 2026-04-20 | `2026-04-20-wire-contract.md` | wire envelope |
| 2026-04-20 | `2026-04-20-sdk-error-hierarchy.md` | SDKs |
| 2026-04-20 | `2026-04-20-rate-limit-layering.md` | middleware |
| 2026-04-20 | `2026-04-20-probe-bypass.md` | ops/probes |
| 2026-04-24 | `2026-04-24-service-interfaces.md` | abstractions |
| 2026-04-24 | `2026-04-24-chi-route-mount-collision.md` | routing |
| 2026-04-24 | `2026-04-24-chi-migration.md` | HTTP framework |
| 2026-04-26 | `2026-04-26-code-field-and-tenancy.md` | error wire / tenancy |
| 2026-04-27 | `2026-04-27-error-translation.md` | error taxonomy |
| 2026-04-27 | `2026-04-27-rbac-membership.md` | RBAC |
| 2026-04-28 | `2026-04-28-error-unification-morning.md` | error taxonomy |
| 2026-04-28 | `2026-04-28-error-unification-afternoon.md` | error taxonomy |
| 2026-04-28 | `2026-04-28-invalidparam-sweep.md` | error taxonomy |
| 2026-04-28 | `2026-04-28-message-positional-sweep.md` | error taxonomy |
| 2026-04-28 | `2026-04-28-invalidinput-deletion.md` | error taxonomy |
| 2026-04-29 | `2026-04-29-four-fixes.md` | RBAC + errors |
| 2026-04-30 | `2026-04-30-status-column-pre-prod-edit.md` | schema discipline |
| 2026-04-30 | `2026-04-30-suspension-deletion.md` | scaffolded code |
| 2026-04-30 | `2026-04-30-orphan-cleanup.md` | RBAC + migrations |
| 2026-04-30 | `2026-04-30-xor-bad-request.md` | error taxonomy |
| 2026-04-30 | `2026-04-30-domain-errors-deletion.md` | error taxonomy |
| 2026-04-30 | `2026-04-30-project-rbac-resolver.md` | RBAC (16 sub-rounds) |
| 2026-04-30 | `2026-04-30-stale-project-members.md` | RBAC + naming |
