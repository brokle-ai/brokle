---
date: 2026-04-28
status: enacted
tags: [error-taxonomy, deletion, structural-fix]
---

# Legacy `*AppError` deletion (afternoon)

Eliminated the legacy `*AppError` surface entirely after the morning's unification kept it alive behind forbidigo carve-outs.

**Reviewer's challenge**: "why are we carrying tech legacy debts??" — the carve-outs (pkg/request validator boundary, one `WithCode("invitation_expired")` site, the wire-shape byte-identity test, FromHTTPStatus, recoverer constant) all existed only because the new `*Error` type lacked two fields the legacy `*AppError` had: `Code` (Stripe-style opt-in subcode) and `Errors []ErrorDetail` (per-field validator output).

Per CLAUDE.md ("no fallbacks for scenarios that can't happen", pre-prod with no backward-compat), the right move was to delete legacy entirely instead of justifying its retention.

**Action**: extended `*Error` with `Code`, `Param`, `Details`, `Errors []ErrorDetail` fields + `WithCode`/`WithParam`/`WithDetails`/`WithFieldErrors` options + new `ReasonBadRequest`/`ReasonPaymentRequired` reasons (HTTP 400/402 mapping).

**Deleted**: `*AppError` struct + 14 `New*Error` constructors + `Wrap*` constructors + `ErrorType` enum + `typeToStatus` map + `IsAppError`/`AsAppError` predicates + `New(t, msg, opts...)` factory + `WithCause`/`WithDetails`/`WithParam`/`WithErrors` legacy options + `CodeOrType`/`HTTPStatus(err)`-AppError-branch.

**Rewrote**: `pkg/errors/errors.go` (single 540-LoC file replacing reason.go + errors.go), `pkg/response/buildAPIError` (single-branch *Error), `pkg/request.DecodeJSON` + URL/query helpers (validator boundary now uses `BadRequest` for 400s and `InvalidInput("", WithMessage, WithDetails, WithParam)` for 422s — Resource is empty when the offending input is a field, never `"body"`), `pkg/errors/errors_test.go` (Reason-shape assertions), `pkg/response/error_shape_test.go` (single-path byte test), `pkg/request/request_test.go` (Reason assertions), `internal/transport/http/middleware/recoverer_test.go`, `internal/core/services/auth/api_key_service_test.go` + `playground/service_test.go` (test helpers collapsed to single error-family branch).

Removed forbidigo `appErrors\.New` rule + 4 carve-out exclusions from `.golangci.yml`.

Migration was straightforward: ~10 handler/service callsites with `WithCode`/`WithParam`/`WithDetails` (the morning's "kept" retentions) converted mechanically to the new option names; bulk sed on `WithReasonCause`→`WithCause`, `NotImplementedReason`→`NotImplemented`, `AsReason`→`As`.

**Generalisable rule (added to Mandatory Development Rules in spirit)**: when a "permanent carve-out" requires a forbidigo rule + path exclusions to enforce, the carve-out is a smell, not a stable boundary. The right question is "what data shape gap forces the carve-out?" — extending the canonical type to cover that gap is almost always cheaper than maintaining the parallel API + lint rule + reviewer load forever.

Reviewer pushback on "why are we carrying tech debts?" was the right call; the morning's plan was anchored on cost-of-migration ("~1,330 callsites, multi-day churn") which evaporated once the legacy callsites were already migrated by the morning waves and only ~10 retention sites remained.

**Lesson**: when a migration is 99% complete and the remaining 1% is "data shape gap", finish the data-shape extension instead of declaring the gap permanent.
