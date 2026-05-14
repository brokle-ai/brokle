---
date: 2026-04-18
status: enacted
tags: [domain, schema, nullable, lint]
---

# Nullable-domain alignment sweep

Nullable-domain alignment sweep across 8 domains (auth, user, credentials, website, organization, billing, observability, evaluation) eliminated 63 forbidigo `^deref[A-Z]` / `^emptyToNil[A-Z]` findings. The rule: schema nullability drives Go type (Category 1 value + `NOT NULL` vs Category 2 `*T` + nullable). Bridging helpers (`derefString`, `emptyToNilString`) are a schema-lie smell — when you reach for one, fix the domain type instead. Cross-domain primitives like `NilIfEmpty(s string) *string` live in `internal/core/domain/shared/`, not duplicated per-package.
