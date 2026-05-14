---
date: 2026-04-30
status: enacted
tags: [migrations, schema-discipline, pre-prod]
---

# Status-column pre-prod schema edit (delete the migration)

Three review rounds chased rollback-compatibility hazards (positional `SELECT *` scans + appended `ADD COLUMN` on rollback + dishonest down-migration symmetry) for the `drop_member_status_column` migration before the user pointed out the obvious answer: **pre-prod with no production data, no rollback constraint, and zero code paths reading the column means there's no migration to write — just edit the initial schema to never have the column.**

Removed `status VARCHAR(20) DEFAULT 'active'` from `organization_members` and `project_members` in `20250908140000_normalized_rbac_schema.up.sql` (and the trailing `idx_org_members_status` index); deleted `20260427171335_drop_member_status_column.{up,down}.sql` entirely; trimmed the corresponding `UPDATE … SET status = 'active'` + `ALTER … ALTER COLUMN status SET NOT NULL` lines from `20260417091106_enforce_not_null_on_defaulted_columns.{up,down}.sql`.

The user then applied the same reasoning to a sibling migration: `20260427165217_cleanup_orphan_project_members.{up,down}.sql` — a one-shot DELETE for "orphan project_members rows left behind by pre-2026-04-30 code paths" — also deleted, because in pre-prod those rows only exist in developer dev DBs (resettable) and the service-layer cascade prevents new ones structurally.

**Bonus structural improvements that stay because they're load-bearing for forward correctness**: explicit-column lists replacing `SELECT *` in `member.sql` (6 sites), `project_member.sql` (1 site), `role.sql` (7), `user.sql` (4 + a `u.*` JOIN site), `usage_budget.sql` (4), `filter_preset.sql` (1), `prompt.sql` (2 protected-label sites); two `scripts/lint-conventions.sh` guards that catch the bug class for *future* migrations: "Migration symmetry honesty" (down ADDs a column the up DROPs → fail) and the scoped `SELECT *` ban (member + project_member files; broaden the allowlist as the codebase-wide sweep progresses).

## Generalisable rules

1. **CLAUDE.md "Compatibility Notes" load-bearing for migration design** — "no production data, backward compatibility not required" means the right answer to "this migration's down direction is unsafe" is almost always "delete the migration; edit the initial schema instead." Carrying a forward + down pair to drop a vestigial column reads "we're shipping defensive infrastructure for a constraint we don't have."

2. **Reviewer feedback in tight loops is a smell** — three rounds of P1s on the same migration, each time fixing a distinct surface concern (column-position scan / dishonest down / no-op-still-doesn't-restore-INSERT-targets), pointed at a fundamental category error: the migration shouldn't exist. The next time three review rounds attack the same file, **stop fixing the file and ask whether the file should exist**. The over-engineering tax across three rounds was ~4 hours of work for a 6-line CREATE TABLE diff.

3. **Pre-production is the privilege to delete schema, not just patch it** — every major Go project's migration discipline (Stripe, GitLab, Shopify, Rails) treats migrations as immutable once shipped; pre-prod Brokle has no such constraint, and editing initial migrations is the cheapest possible refactor. The structural improvements that stayed (explicit columns, lint guards) are the ones that have value independent of this migration — they prevent the bug class for *future* schema changes that *will* be immutable.
