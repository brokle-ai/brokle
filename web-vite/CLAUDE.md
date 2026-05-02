# Frontend Rules — Vite + TanStack (`web-vite/`)

Loaded automatically when Claude reads any file under `web-vite/`.
Universal repo rules in `/CLAUDE.md`. Backend wire contract in `internal/CLAUDE.md`.

## Wire contract — Stripe / OpenAI shape

The backend emits raw resource bodies on success and a single `error` envelope on failure. NO `{success, data, meta}` wrapper exists.

- Success (2xx): body is the resource directly — `{"id":"prj_123","name":"..."}`.
- List endpoints: `{"data":[...], "pagination":{page,limit,total,total_pages,has_next,has_prev}}`.
- Errors (4xx/5xx): `{"error":{"type":"...","message":"...","code"?:"...","details"?:"...","param"?:"...","errors"?:[...]}}`.
- HTTP status is the only success/failure signal.
- `code` is OPTIONAL — present only when the SDK should branch on a sub-classification beyond `type`.
- Request IDs come from the `X-Request-Id` response header. Read via `response.headers.get('x-request-id')` in `src/lib/api/errors.ts`.

NEVER reintroduce `success` / `data` / `meta` envelope assumptions in the API client.

## API client behavior

- `withCredentials: true` — required for httpOnly cookies.
- CSRF tokens extracted from cookies and added to mutation requests (POST/PUT/PATCH/DELETE) only, not GET.
- Auth token refresh is owned by the auth store, not the API client.
- Context headers (`X-Org-ID`, `X-Project-ID`) are **opt-in** — methods require explicit flags (`includeOrgContext`, `includeProjectContext`).

## RBAC catalog is codegenned

`web/src/generated/permissions.ts` is generated from `seeds/permissions.yaml` via `make gen-frontend-permissions`. Contains the `Scope` union, `ScopeLevel` type, and `SCOPE_LEVELS` map.

- **Never hand-edit** `web/src/generated/permissions.ts` — drift-guard test in `internal/seeder/frontend_permissions_test.go` fails CI on stale generated content.
- `useHasAccess` imports `Scope`/`ScopeLevel`/`SCOPE_LEVELS` from `@/generated/permissions` — the hook function holds only the predicate logic.
- When adding a new permission to `seeds/permissions.yaml`, run `make gen-frontend-permissions` (or `make generate` which includes it).

## Tenancy URL convention

Org-scoped list/create endpoints follow REST tenant-parent shape:
- `GET /api/v1/organizations/{orgId}/projects` (NOT `/v1/projects?organization_id=...`).
- `POST /api/v1/organizations/{orgId}/projects`.

The backend rejects `?organization_id=` query overrides and body-side `organization_id` with HTTP 422. The frontend lint guard (`scripts/lint-conventions.sh` Frontend URL-tenancy invariant section) blocks reintroduction of flat `/v1/<resource>?` shapes for tenant-child resources.

## Proxy + auth

`web/src/proxy.ts` (Next.js layout — for the legacy `web/` surface) is **presence-only**: checks cookie presence (`req.cookies.has('access_token') || .has('refresh_token')`) and redirects to `/signin` if absent.

- Do NOT JWT-decode the cookie locally — couples middleware to backend signing scheme, bypasses revocation/blacklist.
- Do NOT check only `access_token` — an expired access + valid refresh must reach the dashboard so the API client's 401 interceptor can refresh.
- Cookie validity is the Go backend's job. The proxy is the optimistic edge guard.

For Vite, the equivalent is the auth-store rehydration + 401 interceptor in `src/lib/api/client.ts`.

## File layout

- `src/features/[feature]/` — features are self-contained: `api/`, `components/`, `hooks/`, `stores/`, `types/`, `utils/`.
- Cross-feature imports: only via `@/features/[feature]` barrel — never into internal subdirectories.
- `src/generated/` — machine-generated; never hand-edit.

## TanStack Query / Router

- Use TanStack Query for server state. Don't mirror server state into Zustand.
- File-based routes live under `src/routes/` (or wherever the project conventions hold).
- Auth-required routes: use the route-level `beforeLoad` guard pattern; never reach into the auth store from a component.

## Tools

- `pnpm dev` — Vite dev server.
- `pnpm test` — Vitest.
- `pnpm test:e2e` — Playwright (if configured).
- `pnpm type-check` — `tsc --noEmit`.

For prior context on the Stripe-shape migration, RBAC catalog codegen, or proxy semantics, see `docs/decisions/`.
