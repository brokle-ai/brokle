# Backend Rules (`internal/`)

Loaded automatically when Claude reads any file under `internal/`.
Universal repo rules live in `/CLAUDE.md`. Architectural rationale lives in `docs/decisions/` (date-prefixed) and `docs/adr/` (formal, numbered).

## Architecture flow

Repository → Service → Handler. Each layer owns a distinct responsibility:

- **Repository**: data access. Wraps `pgx.ErrNoRows` / unique-violation classifier into `pkg/errors.*Error` at this single boundary. Always tag with `WithOp("repo.<domain>.<method>")`. Never bubble raw `pgx`/`sqlc` errors up to the service.
- **Service**: business rules. Constructs `*pkg/errors.Error` directly for service-only conditions (validation, business-rule rejection). For repository errors, **passes through** — `return s.repo.GetByID(ctx, id)` is correct because the repo error self-describes its wire shape.
- **Handler**: HTTP boundary only. Decodes input via `pkg/request`, invokes service, responds via `pkg/response`. No business logic, no per-method error translation.

## Error taxonomy

Single package, single type, single alias:
- Import as `appErrors "brokle/pkg/errors"` — everywhere (domain, repo, service, handler, middleware, worker).
- One type: `*pkg/errors.Error` carrying `Reason`, `Resource`, optional `Message`/`Details`/`Code`/`Param`/`Errors []ErrorDetail`, `Op`, `Cause`.
- `Reason` enum is closed: `ReasonUnspecified` (iota=0, never set on real errors — proto3/k8s convention), `ReasonInternal`, `ReasonNotFound`, `ReasonAlreadyExists`, `ReasonInvalidInput`, `ReasonBadRequest`, `ReasonUnauthenticated`, `ReasonPermissionDenied`, `ReasonConflict`, `ReasonRateLimit`, `ReasonPaymentRequired`, `ReasonUpstream`, `ReasonUnavailable`, `ReasonNotImplemented`.

Constructor cheat-sheet:

| Shape | Constructor | Use when |
|---|---|---|
| Default-message-OK | `NotFound(resource, opts...)` | Default `"X not found"` works at most call sites. |
| Default-message-OK | `AlreadyExists(resource, opts...)` | Default `"X already exists"`. |
| Default-message-OK | `Unavailable(resource, opts...)` | Default `"X unavailable"`. |
| Default-message-OK | `PaymentRequired(opts...)` | Default `"payment required"`. |
| Message-required | `Conflict(resource, message, opts...)` | State-machine rejection (HTTP 409). |
| Message-required | `PermissionDenied(resource, message, opts...)` | RBAC failure (HTTP 403). |
| Message-required | `NotImplemented(resource, message, opts...)` | Stub endpoint. |
| Message-required | `InvalidParam(field, message, opts...)` | Single named wire field bad. Sets `error.param`. |
| Message-required | `Unauthenticated(message, opts...)` | Auth failure (HTTP 401). |
| Message-required | `BadRequest(message, opts...)` | Request shape, XOR / cross-field rule, not a single field. |
| Message-required | `RateLimit(message, opts...)` | Throttled. |
| Message-required | `Upstream(resource, cause, opts...)` | Third-party call failed. |
| Message-required | `Internal(message, cause, opts...)` | Bug or unknown failure. |
| Multi-field | `InvalidFields(details []ErrorDetail, opts...)` | Validator returned N field errors in one decode. |

**XOR / cross-field rules** (e.g. `score_id` XOR `experiment_id`, `org_name` XOR `invite_token`) use `BadRequest(message)` naming both branches. Never synthesise a `param` for a non-existent field — `scripts/lint-conventions.sh` blocks the lint-curated bad words (`linkage`, `cannot`, `already`, etc.).

**`InvalidParam` first-arg discipline**: must be a real wire field that exists on the request body. Synthetic field names are forbidden by lint.

**Predicates**: `appErrors.As(err) *Error`, `appErrors.IsReason(err, ReasonX)`, `appErrors.IsNotFound(err)`, `IsAlreadyExists(err)`, `IsConflict(err)`, `IsUniqueViolation(err)` (pgx classifier).

**Domain `errors.go` files are banned** — `scripts/lint-conventions.sh` enforces structurally. All classification flows through `pkg/errors` constructors.

## HTTP framework — chi v5

Stack: `chi v5` + `pkg/request` (JSON decode + `go-playground/validator/v10`) + `pkg/response` (envelope helpers). Two surfaces:
- `/v1/*` — SDK plane (X-API-Key auth)
- `/api/v1/*` — dashboard plane (cookie + JWT auth)

Handler signature: `func(w http.ResponseWriter, r *http.Request)`.

Common ops:
- Path params: `request.URLParamUUID(r, "id")`, `request.URLParamInt(r, "n")`, `chi.URLParam(r, "slug")`.
- Query params: `request.QueryInt(r, "limit", 50)`, `request.QueryOptionalBool(r, "has_error")`, `request.QueryPagination(r)`.
- Body: `var body fooBody; if err := request.DecodeJSON(r, &body); err != nil { response.WriteError(w, err); return }`. Enforces `MaxBytesReader` (1 MiB), `DisallowUnknownFields`, trailing-data reject, validator tags.
- Response: `response.Success(w, body)` (200), `response.Created(w, body)` (201), `response.NoContent(w)` (204), `response.WriteError(w, err)`.

**DELETE handlers** return `response.NoContent(w)` — never `Success(w, gin.H{"message": ...})`.
**UPDATE handlers** return the updated entity via `response.Success(w, entity)` — never message-only (auth flows like logout/password-reset are the exception).

### chi gotchas

- `Mux.Use` must precede every route registration. `installGlobalMiddleware` in `internal/server/middleware.go` runs BEFORE `addRoutes`. Never call `r.Use(...)` on the top-level mux inside `addRoutes` — sub-routers via `r.Route` / `r.Group` are fine.
- Probe + `/metrics` paths bypass chi entirely via the outer `http.ServeMux` dispatcher in `internal/server/health.go`. Don't add API routes to it.
- One `r.Route` per URL prefix per chi tree — sibling `chi.Group` sub-routers share their parent's routing tree. Two `r.Route("/api/v1/auth", ...)` collide. Use direct `r.Get/Post(...)` registrations across siblings; use `r.Route` only when one function owns the whole subtree.
- Routes register **without** trailing slashes: `r.Get("/api/v1/users/me", h)` — no slash. Two carve-outs: `r.Get("/", h)` index handlers inside `r.Route("/foo/{id}", ...)` are slash-safe automatically; never add `chimw.StripSlashes` or `RedirectSlashes`.
- Routes are centralised in `internal/server/routes.go` `addRoutes` (Mat Ryer pattern). Handler packages export only `Handler` struct + `New(...)` + methods. `RegisterRoutes` helpers under `internal/transport/http/handlers/` are banned by lint.

## Auth context

`Must*` helpers panic on misconfigured routes (caught by `middleware.Recoverer` → 500). Use them on routes protected by `RequireAuth` / `RequireSDKAuth`:

- Dashboard: `httpctx.MustGetUserID(ctx)`, `MustGetAuthContext(ctx)`, `MustGetTokenClaims(ctx)`.
- Project-scoped: `httpctx.MustGetProjectID(ctx)`, `MustGetOrganizationID(ctx)`.
- SDK: `httpctx.MustGetSDKAuthContext(ctx)`.

Tuple-return forms (`httpctx.UserID(ctx) (uuid.UUID, bool)`) are reserved for `OptionalAuth` routes and audit-field population. Never combine `Must*` with an existence check.

**SDK vs dashboard context keys are distinct.** Calling `MustGetUserID(ctx)` in an SDK handler panics; same for `MustGetSDKAuthContext` in a dashboard handler. Match the getter to the route group.

`UserID`/`ProjectID`/`OrganizationID` store `uuid.UUID` **by value**. Only `APIKeyID` stores `*uuid.UUID` (legitimately nullable for session auth).

## RBAC — scope-partitioned hybrid resolver

Two-tier model: `organization_members` (base) + `project_members` (optional per-project override). Single resolver: `ProjectMemberService.CheckUserPermissionsInScope(ctx, userID, orgID, projectID, perms)`. Pass `uuid.Nil` as `projectID` to skip the project layer.

Three structural rules in `ListUserEffectivePermissionsInScope`:
1. **Active-org-membership precondition** — both branches require an active `organization_members` row. A stale `project_members` row alone grants zero.
2. **Org-scoped permissions** (`p.scope_level = 'organization'`, e.g. `members:*`, `billing:*`, `organizations:*`) resolve against the org role only — project membership is irrelevant.
3. **Project-scoped permissions** (`p.scope_level = 'project'`, e.g. `traces:*`, `datasets:*`) follow OVERRIDE within the partition: a `project_members` role REPLACES the org role's projection (supports both ELEVATE and RESTRICT).

Enforcement is at the route layer via `r.With(middleware.RequirePermission(authD, "<resource>:<action>"))` — handlers do NOT call `IsMember` / `CheckUserPermission` themselves. `RequireAllPermissions(verb, "projects:read")` is required for shared-verb routes mounted under project URLs (e.g. `/api/v1/projects/{projectId}/members*`).

**Floor-scope auto-injection**: any custom role created with a project-scoped permission auto-gets `projects:read` (GitHub `metadata:read` mechanic). `internal/seeder/floor_scope_test.go` is the build-time invariant.

**Permission catalog**: `seeds/permissions.yaml` is the source of truth (with `scope: project|organization` field). Codegen via `make gen-frontend-permissions` produces `web/src/generated/permissions.ts`. Drift-guard test fails CI on stale generated file.

**Tenancy is in the URL — never in body, never in query.** For routes under `/api/v1/organizations/{orgId}/...` or `/projects/{projectId}/...`, read tenancy via `httpctx.MustGetOrganizationID(ctx)` / `MustGetProjectID(ctx)`. Reject `?organization_id=` / `?project_id=` query overrides with 422. Request DTOs MUST NOT contain a tenancy field — `DisallowUnknownFields` will 422 on body keys. See OWASP API1:2023 BOLA.

## Repositories

PostgreSQL repos take `*db.TxManager`:
- `r.tm.Queries(ctx)` for sqlc-generated queries.
- `r.tm.DB(ctx)` for squirrel-built dynamic queries.
- Transactions propagate via ctx (set by `TxManager.WithinTransaction(ctx, fn)`).

ClickHouse repos use `clickhouse.Conn` directly — no transaction scoping.

**Nullable field pattern** — schema nullability drives Go type:
- **Category 1 (required)**: `NOT NULL` schema (no `DEFAULT`) + value type + no `omitempty` on JSON + constructor validates.
- **Category 2 (optional)**: nullable schema + pointer type + `omitempty` on JSON + nil-check at read sites.

Never use `NOT NULL DEFAULT ''` to paper over Go type decisions. Never write `derefX` / `emptyToNilX` bridging helpers — fix the domain type instead. Forbidigo blocks them at lint time.

**Cross-domain primitives** live in `internal/core/domain/shared/`: `shared.NilIfEmpty(s string) *string`, `shared.Ptr[T any](v T) *T`. A helper earns a seat after ≥2 domains need it.

**Service constructors return concrete** — `func NewXxxService(...) *XxxService { ... }`. Domain-level service interfaces require ≥2 implementations OR a real mock OR an external plugin boundary today. Repository interfaces are the deliberate exception (services consume them as interfaces; sqlc+pgx prod, in-memory test).

**Nested-actor hydration**: list endpoints needing inviter/author/reviewer display populate a `*ActorRef` field via LEFT JOIN in the repo (e.g. `Invitation.Inviter *InviterRef`, `Invitation.Role *RoleRef`). Eliminates N+1 + nil-deref panics. Never flatten `inviter_email` / `inviter_name` at the top level.

## ID generation

`uid.New()` (UUIDv7, monotonic). Do not introduce new ULID usage; `pkg/ulid` has been removed.

## JSON convention for response DTOs

- Pointer types for optional fields use `omitempty` (matches Kubernetes; 91% of pointer-typed domain fields already do).
- Required fields use value types with no `omitempty`.
- Collection fields in **response** DTOs do NOT use `omitempty` — `[]` means "queried, zero results" and is distinct from absent.
- Collections in **request** DTOs may use `omitempty` (absent = don't update).

`json.RawMessage` fields in domain entities (`observability.Score.Metadata`, `prompt.ModelConfig.Tools`, etc.) require DTO conversion in handlers — never serialize the raw entity.

## Wire contract (Stripe / OpenAI shape)

Success bodies: raw resource directly. List endpoints inline pagination: `{"data":[...], "pagination":{...}}`. NO `{success, data, meta}` envelope.

Error bodies: `{"error":{"type":"...","message":"...","code"?:"...","details"?:"...","param"?:"...","errors"?:[...]}}`. HTTP status is the only success/failure signal. `code` is opt-in via `WithCode(...)` only when SDK consumers should branch on a sub-classification beyond `type`.

Request IDs surface via the `X-Request-Id` response header (`chimw.RequestID`). Not in the body.

Pre-authoritative: `pkg/response/error_shape_test.go` locks byte-identical envelope across `*Error.MarshalJSON`, `ErrorResponse` direct marshal, and `WriteError` paths.

## slog message case

- **Service layer** (`internal/core/services/...`): lowercase-first. Mirrors the `*Error` it returns on the same path.
- **Transport / worker / bootstrap** (`internal/transport/`, `internal/workers/`, `internal/app/`, `internal/server/`, `internal/migration/`, `internal/seeder/`): Capital-first. Operational events, not error strings.
- Acronyms stay uppercase: `"API key invalid"`, `"OAuth session expired"`.

Drift check: `grep -rn 'logger\.\(Error\|Warn\|Info\|Debug\)("[A-Z]' internal/core/services/` should return zero hits.

## Background work

All `emailSender.Send()` calls triggered by HTTP requests use a detached context:
```go
go func() {
  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()
  // ...
}()
```
Never use the request context for async work that must complete after the handler returns.

Worker DLQ semantics (`telemetry_stream_consumer.go`): ACK on success OR `ErrMovedToDLQ` (data preserved). Leave pending on other errors.

## Seeders

`make reseed` after editing `seeds/*.yaml` — the seeder is NOT auto-invoked by `make dev`. Build-time guard `internal/seeder/coverage_test.go` asserts every permission referenced by a `RequirePermission(...)` route exists in `seeds/permissions.yaml`. Floor-scope guard `internal/seeder/floor_scope_test.go` asserts every built-in role with project-scoped perms also carries `projects:read`.

Seeder is **reconciling**, not skip-on-exists, for permissions: when a seeded field gains downstream behavioral significance (e.g., `scope_level` driving floor-scope auto-injection), the seeder must update existing rows in place by natural key. Update-in-place via `permissionRepo.Update` — never delete-recreate (FK references would cascade-invalidate).

## Other gotchas

- **Dual ports**: HTTP API on `8080`, gRPC OTLP on `4317`. Independent servers.
- **Enterprise build tag**: `-tags="enterprise"` gates SSO/RBAC/compliance in `internal/ee/`. OSS builds have stubs.
- **Config validation is mode-specific**: `APP_MODE=worker` skips server-mode validations (e.g. `JWT_SECRET`). Promoting a worker to server requires all server-mode validations to pass.
- **No platform-admin concept**: all roles in `seeds/roles.yaml` are `scope_type: "organization"`. `annotation.RoleAdmin` is an unrelated per-queue reviewer role.
- **No OpenAPI spec is emitted by the server.** SDKs are hand-written; frontend uses raw `fetch`. If a future need arises, hand-write `openapi.yaml` and serve via Scarf/Redoc — do not reintroduce runtime spec generation.
- **Stub-only handler packages get deleted, not preserved.** Signal: no `RegisterRoutes`/wiring caller, or 90% TODO bodies. Precedents: `handlers/logs/`, `handlers/analytics/`, `handlers/enterprise/`, `handlers/websocket/`, `handlers/admin/` all deleted. Git history preserves them.

## Convention lint

`scripts/lint-conventions.sh` (run via `make lint-conventions`, wired into `make lint`) enforces:
- DDL canonical forms (no `TIMESTAMPTZ` / `DECIMAL` / bare `INT` / `BOOL`)
- Migration filename convention
- No `^func Register` under `internal/transport/http/handlers/`
- No `internal/core/domain/*/errors.go` files
- `InvalidParam` first-arg bad-word list (state-words, synthetic field names)
- Frontend URL-tenancy invariant (no `/v1/<resource>?` flat tenant-child shapes)
- Score-linkage / similar XOR rules use `BadRequest`, not `InvalidParam("synthetic", ...)`

New project-wide rules that can't be expressed in golangci-lint go here.

---

For deep architectural context behind these rules, see `docs/decisions/` (date-prefixed lessons) and `docs/adr/` (formal ADRs).
