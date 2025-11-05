# Repository Guidelines

This document is a concise, practical guide for contributors working in this repository. Favor existing patterns and keep changes narrowly scoped.

## Project Structure & Module Organization
- Backend Go: `cmd/` (entry points like `server`, `migrate`), `internal/` (private app code: `app/`, `config/`, `transport/`, `infrastructure/`, etc.), `pkg/` (reusable helpers).
- Frontend: `web/` (Next.js app: `src/`, `public/`).
- Operations: `configs/`, `deployments/`, `monitoring/`, `migrations/`, `seeders/` + `seeds/`.
- Tooling & Docs: `scripts/`, `docs/` (see `CODING_STANDARDS.md`, `DEVELOPMENT.md`), `tests/` (placeholders for `unit/`, `integration/`, `e2e/`, `load/`).

## Build, Test, and Development Commands
- Setup: `make setup` — install deps, init submodules, start DBs, run migrations, seed dev.
- Dev (full stack): `make dev` — runs server + worker with Air reload (use `make dev-frontend` separately for Next.js).
- Dev (split): `make dev-server`, `make dev-worker`, `make dev-frontend`.
- Build: `make build-oss` or `make build-enterprise`; dev builds via `make build-dev-server` or `make build-dev-worker`.
- Tests: `make test` (all), `make test-coverage`, `make test-unit`, `make test-integration`, `make test-e2e`, `make test-load`.
- Quality: `make lint` (`golangci-lint`, Next.js lint), `make fmt` (Go fmt + goimports), `make fmt-frontend`.
- Docker: `make docker-build`, `make docker-up`, `make docker-down`.

## Coding Style & Naming Conventions
- Go: follow `docs/CODING_STANDARDS.md`. Keep packages lowercase; exported identifiers use PascalCase; tests end with `_test.go`. Format with `make fmt`.
- Frontend: TypeScript + Next.js. Lint via `web/eslint.config.mjs`; format with Prettier (`pnpm run format`).
- Config: prefer `viper`-backed envs; sample values live in `.env.example`.

## Testing Guidelines
- Frameworks: Go standard testing with `testify`. Place unit tests near code (e.g., `internal/<pkg>/*_test.go`).
- Naming: `TestXxx(t *testing.T)`; table-driven tests where possible. Keep integration/E2E under `tests/`.
- Coverage: run `make test-coverage`; include critical paths in PRs.

## Commit & Pull Request Guidelines
- Use Conventional Commits (e.g., `fix(api): standardize pagination response`).
- PRs: clear description, linked issues, reproduction/impact, and screenshots for UI. Include migration notes when touching DBs.
- Checks: ensure `make lint test` pass; update `docs/` when behavior or APIs change.

## Security & Configuration Tips
- Do not commit secrets; use `.env` (local) and keep `.env.example` current.
- Services run via Docker Compose (`postgres`, `clickhouse`, `redis`); use `make setup-databases` locally.

## Agent-Specific Instructions
- Prefer Make targets over ad‑hoc scripts. Do not reformat unrelated files. Place new executables in `cmd/<name>` and new services under `internal/<area>` with tests.
