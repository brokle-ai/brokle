---
date: 2026-04-28
status: enacted
tags: [error-taxonomy, sweep, k8s-pattern]
---

# `InvalidParam` constructor sweep (evening)

Added `InvalidParam(field, message string, opts...)` constructor and swept ~315 callsites that had been using `InvalidInput("", WithMessage(...), WithDetails(...), WithParam(...))` — the same field-vs-entity conflation the morning's wave had retired in different clothing.

**Reviewer challenged a single line**: *"why this: appErrors.InvalidInput(\"\", ????"* — the empty-string sentinel + 3-line option-bag setup for one bad field was the smell.

**Inventory**: 379 `InvalidInput(...)` callsites total, 315 (83%) used the empty-resource form.

Web research surveyed 7 production sources (Stripe, OpenAI, Google `google.rpc.BadRequest`, k8s `apimachinery/util/validation/field`, Stripe-Go SDK, HashiCorp Boundary, Grafana errutil): the **k8s `field.Required/Invalid/Duplicate` precedent** is the strongest match for hand-written Go callsites — distinct constructors per cause type collapse 3-4 lines into 1 while keeping the wire envelope identical. Stripe/OpenAI use one shape because they're JSON marshalling (no Go ergonomics cost).

**Action**: added `InvalidParam(field, message, opts...)` (4 LoC, sugar over `InvalidInput("", WithParam(field), WithMessage(message), opts...)`); swept callsites in 4 parallel agents by zone (handlers A/B/C, services D); ~298 sites collapsed to one-liners. Promoted `pkg/request.paramValidationError` private helper to inline `InvalidParam(...)` calls at 8 sites. Added `scripts/lint-conventions.sh` rule banning empty-resource `InvalidInput("",...)` outside the `WithFieldErrors` carve-out (validator multi-field output).

Two state-machine sites (`project_service.ArchiveProject` / `UnarchiveProject`) re-classified from `InvalidInput("",...)` to `Conflict("project",...)` — they were always state-change conflicts, not validation. Three auth sites reclassified to entity-level (`InvalidInput("oauth_session"|"login_session"|"signup", ...)`) — they're rejecting a multi-field flow state, not a single field. One presence check (`auth/sdk.go` API key missing) reclassified to `Unauthenticated()` — it was always an auth failure (401), not a 422.

Wire shape preserved across all 311 InvalidParam conversions (Reason=ReasonInvalidInput → 422 + `error.param` populated). Final state: 373 `InvalidParam` calls, 7 `InvalidInput` calls (4 entity-level + 2 carve-outs + 1 retired).

**Generalisable rules**:
1. The smell of a positional arg that 83% of callsites pass empty for is "the API has two distinct cases pretending to be one" — split it, don't document the empty-string convention.
2. Field-name-as-resource (`InvalidInput("status", ...)`) is the same conflation in disguise — the lint must catch both forms or the codebase regresses to it within weeks.
3. k8s `apimachinery/util/validation/field` is the canonical Go precedent for field-level validation constructors when you have dozens of hand-written sites — Stripe-Go-style "one constructor + optional fields" works for SDK consumers because they marshal JSON, not for hand-written backend code where the per-call ergonomics tax compounds.

Reviewer's instinct ("why this?") was load-bearing: each time the same *kind* of "this is verbose for no reason" question surfaces, the answer is almost always that the API design papered over a real semantic split.
