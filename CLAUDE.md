# Brokle — Agent Instructions

Brokle is an LLM observability platform. Monorepo: Go backend (`internal/`, `cmd/`, `pkg/`), Next.js + Vite dashboards (`web/`, `web-vite/`), Python + JavaScript SDKs (`sdk/`).

This file holds **universally applicable** rules — everything Claude needs on every turn. Domain-specific rules live in subdirectory `CLAUDE.md` files (auto-loaded when Claude reads files in that directory). Architectural rationale (the *why* behind structural decisions) lives in `docs/decisions/` and `docs/adr/`.

## Memory map

| Scope | File | Loaded when |
|---|---|---|
| Universal | `CLAUDE.md` (this file) | Every turn |
| Backend | `internal/CLAUDE.md` | Reading `internal/...` |
| Vite frontend | `web-vite/CLAUDE.md` | Reading `web-vite/...` |
| Next.js frontend | `web/CLAUDE.md` | Reading `web/...` |
| SDKs | `sdk/CLAUDE.md` | Reading `sdk/...` |
| Migrations | `migrations/CLAUDE.md` | Reading `migrations/...` |
| Decisions archive | `docs/decisions/INDEX.md` | On demand (grep / read) |
| Formal ADRs | `docs/adr/` | On demand (grep / read) |

## Build & development

```bash
make setup       # install tools, start DBs, run migrations + seeds, generate code
make dev         # backend server + worker with hot reload (air)
make dev-frontend
make test        # all Go tests
make lint        # Go + frontend + convention lint
make generate    # sqlc + frontend permissions codegen
make reseed      # apply seeds/*.yaml (NOT auto-invoked by make dev)
```

Frontend (`cd web` or `cd web-vite`): `pnpm dev`, `pnpm test`, `pnpm test:e2e`, `pnpm type-check`.

## Universal coding conventions

- **Go**: `gofmt`/`goimports`. Lint enforced via `.golangci.yml`. Tests adjacent to code (`*_test.go`).
- **TypeScript**: ESLint + Prettier (`pnpm format`). React component files in kebab-case.
- **Python (SDK)**: Black + isort + flake8 + mypy.
- **Naming**: Go packages lowercase. Migrations timestamp-prefixed. Frontend imports cross-feature ONLY via `@/features/[feature]` barrel.
- **No manual files in `migrations/`** — use `go run cmd/migrate/main.go -db <postgres|clickhouse> -name <name> create`. Lint blocks misnamed files.

## Commit & PR

- Conventional Commits: `feat(scope): ...`, `fix(scope): ...`, `refactor(scope): ...`, `chore: ...`.
- **Subject line only. No body.** The diff says what changed; the subject says what it is. If genuinely load-bearing context won't fit in the subject, put it in the PR description — never in the commit body.
- **Never add `Co-Authored-By:` trailers or any AI/LLM attribution.** Commits read as if a human authored them.
- Keep commits focused. Local tests pass before requesting review.
- PRs follow `.github/pull_request_template.md`.

## Pre-Production Review Charter

External code reviewers (Codex, Claude reviewer agents) default to production-grade hardening — rollback safety, backward compat, defensive layering, migration immutability. In pre-prod Brokle, most of those concerns are speculative cost. Apply this charter when invoking reviewers AND when triaging their output.

**Constraints to state up-front in review prompts:**
- Pre-production: no users, no production data, dev DBs are resettable.
- No backward-compatibility requirement on wire shape, schema, or APIs.
- Prefer **schema edits** over forward+down migration pairs; prefer **deletion** over deprecation; prefer **structural fixes** (lint guards, invariants) over discipline-based recipes.
- Defensive-shape suggestions (StripSlashes, body-side request_id, wrap-everything error types, double-check membership in handler) require 5+ production-peer precedent before adoption.

**Triage by severity, not by count:**
- Act on **P1** correctness/security/data-loss only.
- Defer **P2/P3** unless the fix is cheap AND structural (e.g., adds a lint guard that prevents the bug class).
- If three review rounds attack the same file, **stop fixing the file and ask whether the file should exist** (see `docs/decisions/2026-04-30-status-column-pre-prod-edit.md`).

**Generalisable rule**: an in-house pre-prod consumer base means most defensive-shape arguments are speculative; resist the migration cost until a real workflow demands it. Reviewer feedback in tight loops on the same file is a category-error smell — the file probably shouldn't exist in the shape the review keeps attacking.

## Research-before-rearch (meta-rule)

When a non-trivial design decision lands on the table — new architectural pattern, framework feature usage, wire-contract change, error-shape change — **launch a research agent first** to verify the chosen approach against 5+ production peers in the same shape (LLM observability tier: Anthropic, OpenAI, Langfuse, Helicone; payments/AI: Stripe; orgs/teams: GitHub, Linear, Auth0).

The Brokle-specific evidence pattern: when a reviewer pushes for a defensive-feeling addition (StripSlashes middleware, body-side `request_id`, wrap-everything error types, double-check membership in handler), check whether 5+ peers actually do it before adopting. If zero peers do, the "preserve the previous behaviour" or "add for safety" framing is preserving an anti-pattern.

Vendor-citation discipline: when citing a vendor's model as architectural precedent, verify against their actual docs/source — not blog posts, not summaries. The cost of getting it wrong is years of implementing the wrong semantics. (Two Langfuse mis-citations on RBAC — "MAX semantics" and "wholesale OVERRIDE" — both compounded across rounds before someone re-read the source. See `docs/decisions/2026-04-30-project-rbac-resolver.md` Round 11.)

## Compatibility note

There is no production data; the product has not been released. Backward compatibility on schema, wire shape, or API surface is not required. **Full freedom to remove, modify, or re-architect.**

## Current product focus

- Core observability, evaluation, and analytics flows are P1.
- Website / landing-page features (contact forms, etc.) are P2 — keep working, don't over-engineer.
- SDK improvements (JavaScript, Python) are high priority when explicitly requested.
- Do not spend time hardening deferred features for theoretical completeness.

## When in doubt

- Backend question? → `internal/CLAUDE.md`.
- Frontend question? → `web-vite/CLAUDE.md` (or `web/CLAUDE.md` for the legacy surface).
- SDK question? → `sdk/CLAUDE.md`.
- Migration / sqlc question? → `migrations/CLAUDE.md`.
- "Why does this rule exist?" / "Has this been tried before?" → `grep -r "<keyword>" docs/decisions/` or read `docs/decisions/INDEX.md`.
- Convention enforcement? → `scripts/lint-conventions.sh` is the source of truth for project-wide rules that can't be expressed in golangci-lint.
