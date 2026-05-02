---
date: 2026-04-28
status: enacted
tags: [error-taxonomy, package-layout, k8s-pattern]
---

# Single `pkg/errors` package unification (morning)

Unified error system into a single `pkg/errors` package — eliminated the `derrors` / `appErrors` two-package architecture introduced on 2026-04-27.

**Drivers**:
1. The two-package shape forced reviewers to remember which alias to use where, with the `d` prefix nominally meaning "domain" but the package was increasingly cross-layer (services constructing it, transport reading it, repos translating into it).
2. Web research across **k8s `apimachinery/pkg/api/errors`, HashiCorp Boundary `internal/errors`, HashiCorp Vault `sdk/helper/errutil`, CockroachDB `cockroachdb/errors`, Grafana `pkg/util/errutil`, SigNoz `pkg/errors`, Upspin top-level `errors/`** — 7 of 7 production codebases place the unified error package at a top-level shared path (NOT under any `domain/` subtree), even when the package carries a domain-style `Reason`/`Kind` enum. Vernon and Evans (DDD) define domain exceptions at the bounded-context level, not in a "shared domain errors" module.

**Approach**: additive migration, not hard cutover. New Reason-based API (`*Error{Reason, Resource, Op, Cause}` + `NotFound` / `AlreadyExists` / `Conflict` / `InvalidInput` / `PermissionDenied` / `Unauthenticated` / `Internal` / `Upstream` / `Unavailable` / `NotImplementedReason` / `RateLimit` constructors) lives in the same `pkg/errors` package as the legacy `*AppError`. Both produce byte-identical wire envelopes (locked by `pkg/response/error_shape_test.go`).

Inventory revealed ~1,330 `appErrors.*` callsites — full hard cutover would have been a multi-day churn fest with no functional benefit.

Migration ran in 5 waves (Wave 1: comment/playground/annotation/dashboard; Wave 2: prompt/credentials/evaluation/observability; Wave 3: billing/user/organization; Wave 4: auth — single agent because of cross-cutting middleware; Wave 5: leaf domains + final cleanup), 3-4 parallel agents per wave by domain to avoid file conflicts.

~330 legacy constructor calls converted to Reason-based; 3 legitimate retentions remain — `pkg/request` validator boundary (per-field `Errors[]` carries data the Reason API doesn't model), one `WithCode("invitation_expired")` programmatic sub-classification, and one `WithParam("orgId")` validator-style test. `forbidigo` rule added to `.golangci.yml` banning future regression to `appErrors.NewXxxError(...)` outside the documented carve-outs.

**Generalisable rules**:
1. **Same-package additive over hard cutover** when migration cost ≫ functional delta — both APIs coexist, the linter prevents new regressions, callsites move opportunistically. The "delete the old API" milestone can come later once the carve-out set is small enough to invert.
2. **Cross-cutting infrastructure (errors, logging, request IDs) belongs at a top-level shared path, not under `domain/`** — verified across 7 production Go codebases; zero put it under `domain/`. The "domain shared" placement is a smell that the package is doing transport/infra work in domain clothing.
3. **Bulk `sed` across ~50 files needs deduplication of imports** — naïve `s|old "x"|new "y"|g` introduces duplicates in files that already import `new`; always grep + dedupe pass after sed.
