# AGENTS.md

This repository's authoritative agent instructions live in `CLAUDE.md` plus subdirectory `CLAUDE.md` files. AGENTS.md exists so non-Claude agents (OpenAI Codex, Cursor, Aider, etc.) discover the same rules.

For project structure, build commands, conventions, gotchas, and the pre-production review charter:

- `CLAUDE.md` — universal rules, build & dev commands, commit/PR style, review charter
- `internal/CLAUDE.md` — Go backend (chi v5, error taxonomy, RBAC, Repo→Service→Handler, sqlc + pgx)
- `web-vite/CLAUDE.md` — Vite + TanStack frontend (wire contract, RBAC catalog, tenancy URL)
- `web/CLAUDE.md` — Next.js dashboard (legacy surface, being migrated to web-vite/)
- `sdk/CLAUDE.md` — Python + JavaScript SDK conventions
- `migrations/CLAUDE.md` — DDL canonical forms, sqlc.yaml editing procedure
- `docs/decisions/` — date-prefixed lessons (architectural rationale; loaded on demand)
- `docs/adr/` — formal ADRs (numbered, full context+alternatives sections)

Subdirectory `CLAUDE.md` files load automatically when Claude reads a file in that directory. For non-Claude agents without that mechanism, read the relevant subdirectory file directly when working in that area.
