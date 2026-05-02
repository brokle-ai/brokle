# Brokle Dashboard (web-vite)

Vite 8 + TanStack Router + TanStack Query + openapi-fetch. Replacement
for `web/` under `internal-docs/vite-migration-plan.md`.

## Scripts

- `pnpm dev` — Vite dev server on :3001, proxies `/v1` and `/api/v1` to
  `VITE_PROXY_TARGET` (default `http://localhost:8080`).
- `pnpm build` — type-check + Rolldown production build → `dist/`.
- `pnpm test` / `pnpm test:watch` — Vitest.
- `pnpm test:e2e` — Playwright.
- `pnpm gen:api` — regenerate OpenAPI types from the running backend
  (calls `../scripts/gen-openapi-types.sh`).

## Runtime config

Per-environment settings (`API_URL`, `SENTRY_DSN`, feature flags known
per-env) live in `/config.js`, rendered at container boot from
`public/config.js.template` via `docker-entrypoint.d/10-envsubst-config.sh`.
Read via `getRuntimeConfig()` in `src/lib/config.ts`. Never reach for
`import.meta.env.VITE_*` for per-env values — those are build-time
constants.
