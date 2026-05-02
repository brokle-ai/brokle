---
date: 2026-04-30
status: enacted
tags: [error-taxonomy, deletion, k8s-pattern]
---

# Domain `errors.go` files deletion

Deleted all 11 `internal/core/domain/*/errors.go` files. The audit revealed 3 doc-only stubs (comment/dashboard/playground — pure historical markers, violated CLAUDE.md "no removed-comments-for-removed-code"), 28 orphan sentinels, and ~57 alive sentinels across 8 domains — but the alive ones were ALL classifications (NotFound / AlreadyExists / Conflict / InvalidInput / Unauthenticated) that `pkg/errors` already expressed via Reason+predicates. None were genuine flow-control signals.

Web research surveyed Gitea, GitLab Runner, CockroachDB, k8s `apimachinery`, HashiCorp Boundary/Vault: **k8s ships zero domain sentinels** (single `apierrors.IsNotFound` across 40+ resource types); Cockroach uses `cockroachdb/errors` library + pgcode reasons; Boundary single typed struct + `Code` enum. Only Gitea uses typed structs per domain (and it's the outlier).

The "domain sentinel for ergonomic Is checks" pattern is the discipline-based duplication the 2026-04-27 entry warned about — Brokle had retired it in spirit but kept the residue.

**Action**: 4 parallel agents migrated ~250 callsites across producer (repository) and consumer (service/handler/test) layers — every `return fmt.Errorf("...: %w", domain.ErrXxxNotFound)` became `return appErrors.NotFound("xxx", appErrors.WithOp("repo.xxx.method"))`; every `errors.Is(err, domain.ErrXxxNotFound)` became `appErrors.IsNotFound(err)`.

Non-NotFound/AlreadyExists sentinels mapped per case: `ErrInvalidCredentials`/`ErrTokenExpired`/`ErrTokenInvalid` → `Unauthenticated(...)`; `ErrInvalidPromptType`/`ErrInvalidTemplateFormat` → `InvalidParam(...)`; `ErrInvitationResendCooldown` → `RateLimit(...)`; `ErrInvitationResendLimit` → `Conflict(...)`; `ErrCacheExpired` → `Conflict(...)`.

The `observability/errors.go` file (which held a validation-diagnostic struct, not sentinels) renamed to `validation.go` to match content.

`scripts/lint-conventions.sh` orphan-sentinel check replaced with a structural ban: any `internal/core/domain/*/errors.go` file fails lint.

Wire shape preserved (NotFound→404, AlreadyExists→409, etc.).

**Generalisable rule (codified)**: when a refactor migrates the producer side (repos return typed errors), the consumer side and the legacy sentinel files must be cleaned up in the same commit — leaving the residue is how discipline-based patterns become re-fossilised. The k8s/Cockroach/Boundary "central classification + predicates" model is the proven Go shape for transport-agnostic error types; per-domain `errors.go` files are pre-unification cargo. Reviewer's question ("why still keeping these?") forced the audit; the audit forced the cleanup. The discipline-based pattern survives until structural enforcement catches it.
