# Migration Rules (`migrations/`)

Loaded automatically when Claude reads files under `migrations/`.
Universal repo rules in `/CLAUDE.md`.

## Create migrations only via CLI

```bash
go run cmd/migrate/main.go -db <postgres|clickhouse> -name <name> create
```

Files placed manually under `migrations/` are silently ignored by the migration framework. Lint blocks misnamed files.

## DDL canonical forms (Postgres)

Use the **full canonical names** in all Postgres DDL — not the shortcuts:

| Use this | Not this |
|---|---|
| `TIMESTAMP WITH TIME ZONE` | `TIMESTAMPTZ` |
| `TIMESTAMP WITHOUT TIME ZONE` | `TIMESTAMP` (ambiguous; be explicit) |
| `INTEGER` | `INT` |
| `BOOLEAN` | `BOOL` |
| `NUMERIC(p,s)` | `DECIMAL(p,s)` |

**Why**: sqlc's parser treats short forms as non-`pg_catalog.*` types in `ALTER TABLE ADD COLUMN`. The `pg_catalog.timestamptz` → `time.Time` override silently misses, generated type falls back to `pgtype.Timestamptz`, and the runtime hits scan-time errors. Same trap for the other shortcuts.

`scripts/lint-conventions.sh` blocks `TIMESTAMPTZ` / `DECIMAL` / bare `INT` / `BOOL` in migration files.

## sqlc.yaml — mandatory editing procedure

Override keys for built-in Postgres types use the `pg_catalog.<type>` prefix (e.g. `pg_catalog.numeric`, `pg_catalog.timestamptz`, `pg_catalog.timestamp`). Unprefixed keys never match — sqlc canonicalises to `pg_catalog.*` form before lookup.

Before editing `sqlc.yaml`:

1. Read the existing file. Grep for the type in BOTH prefixed and unprefixed forms.
2. Run `make generate-sqlc` with the current config — do not assume what the output will be.
3. Grep the generated output: `grep -c "pgtype.<Type>" internal/infrastructure/db/gen/*.go`. If count is 0, the existing config works — do not edit.
4. If count is >0, inspect which columns leak and diagnose root cause (DDL shorthand? schema-lie nullable? codec missing?) before editing YAML.
5. When adding an override, use the `pg_catalog.<type>` form.
6. Every Go type mapping in sqlc.yaml MUST have a corresponding pgx codec registered in `internal/infrastructure/db/codecs.go` (via `pgxpool.Config.AfterConnect`). YAML tells codegen what type to emit; pgx codec tells the driver how to decode. Both must be present for runtime correctness — missing codec = compile success + scan-time panic.

## Pre-prod migration policy

There is no production data and backward compatibility is not required.

- **Prefer schema edits over forward+down migration pairs.** If a column is vestigial (no readers, no writers), edit the initial migration file to never have created it instead of carrying a drop migration.
- **Prefer deletion over deprecation.** If a service / endpoint / table has zero callers, delete it (CLAUDE.md "no scaffolded code" rule). Git history preserves the implementation.
- **One-shot historical sweep migrations are still legitimate** when introducing an application-level cascade — the cascade prevents new violations going forward; the sweep remediates state from before the rule existed (GitLab's `Members::CleanupService` precedent).

## Down migrations must be honest

If the up migration drops a column, the down migration cannot symmetrically `ADD COLUMN` it back — the data is gone. `scripts/lint-conventions.sh` "Migration symmetry honesty" check blocks this.

For drops: the down migration is either a no-op (forward-only with a comment explaining why) or restores a CREATE statement with explicit nulls / defaults — never pretends to restore lost data.

## Existing schema-edit precedents

When in doubt about edit-vs-migrate for pre-prod schema changes, check `docs/decisions/2026-04-30-status-column-pre-prod-edit.md` for the reasoning template.

## Reseeding

After editing `seeds/*.yaml`, run `make reseed`. The seeder is NOT auto-invoked by `make dev`. Build-time guards (`internal/seeder/coverage_test.go`, `internal/seeder/floor_scope_test.go`, `internal/seeder/frontend_permissions_test.go`) catch drift on permissions, role floor-scopes, and the frontend-codegenned permission catalog.

For prior context on schema discipline and seeder reconcile-vs-skip, see `docs/decisions/`.
