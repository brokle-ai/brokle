# Frontend Rules — Next.js dashboard (`web/`)

Loaded automatically when Claude reads any file under `web/`.
Universal repo rules in `/CLAUDE.md`. Backend wire contract in `internal/CLAUDE.md`.
Frontend wire / API client / RBAC rules in `web-vite/CLAUDE.md` apply to this surface too — only Next.js-specific items live here.

> **Migration status.** This surface (Next.js) is being migrated to `web-vite/` (Vite + TanStack). New features land in `web-vite/`. Edits to `web/` should be limited to fixes for the dashboard's still-active flows.

## Next.js specifics

- `output: 'standalone'` — frontend builds as a standalone Node.js app, not static files. Docker uses `node .next/standalone`.
- `next.config.ts` has `ignoreBuildErrors: true` — TypeScript build errors won't fail the build. Run `pnpm type-check` locally to catch them.
- One `proxy.ts`, in `src/`. Next.js 16 accepts `proxy.ts` at the project root OR inside `src/`; if both exist, the root file silently shadows the `src/` one (no warning, no build error). Brokle uses the `src/` layout, so `web/src/proxy.ts` is canonical. NEVER create `web/proxy.ts` at the root. Same rule for `middleware.ts`, `instrumentation.ts`.

## Inherited rules

The following rules apply to BOTH `web/` and `web-vite/` and live in `web-vite/CLAUDE.md`:
- Wire contract (Stripe / OpenAI raw shape, no envelope)
- API client cookie + CSRF behavior
- RBAC catalog from `web/src/generated/permissions.ts` (codegenned, never hand-edit)
- Tenancy URL convention (org/project IDs in path, never body or query)
- Feature module layout
- Proxy cookie-presence check (no JWT decode)

For prior context, see `docs/decisions/`.
