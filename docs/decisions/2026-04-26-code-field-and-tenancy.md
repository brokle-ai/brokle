---
date: 2026-04-26
status: enacted
tags: [error-envelope, tenancy, security, BOLA]
---

# `code` field demoted to opt-in; tenancy-from-URL hardening

Three-part error-envelope + tenancy hardening pass after browser smoke surfaced regressions.

## (1) `code` field demoted to opt-in

Every backend error was emitting `code` equal to `type` (e.g., `code:"authentication_error"` mirroring `type:"authentication_error"`) because the `CodeOrType()` helper at the wire-marshal site fell back `Code` to `string(Type)` when unset — and zero call sites in `internal/` ever set `Code` (verified `grep -rn "WithCode(Code"` → 0 hits across 1542 `appErrors.*` references). The defensive fallback meant every response paid two-fields-worth of bytes for one field of information.

Fix: dropped the fallback at both wire sites (`pkg/errors/errors.go MarshalJSON`, `pkg/response/response.go buildAPIError`); deleted the `CodeOrType()` method to prevent the same anti-pattern from being reintroduced; updated the `internalErrorBody` constant in the recoverer; added `TestErrorShape_CodeOmittedWhenUnset` regression guard. Matches Stripe behaviour: `code` is present only when programmatically actionable (`card_declined`, `quota_exceeded`); never echoes `type`.

## (2) `request_id` body-vs-header decision

Initially added `request_id` to the error body alongside `X-Request-Id` header (Anthropic / OpenAI / AWS S3 pattern), which required threading `*http.Request` into `WriteError(w, r, err)` across 808 call sites. Reverted after research showed 9 of 14 surveyed APIs are header-only (Stripe / OpenAI / GitHub / Twilio / Cloudflare / Helicone — Brokle's actual peer set). The header is already populated by `chimw.RequestID`; the frontend already reads it via `response.headers.get('x-request-id')`. Body duplication only justified by SSE-streaming or external-SDK-paste-into-support-ticket workflows — neither exists in pre-prod Brokle.

**Generalisable rule**: an in-house pre-prod consumer base means the "support ticket paste" use case for body-side request IDs is speculative; resist the migration cost (808 sites of compile-time-enforced churn) until a real workflow needs it.

## (3) Tenancy-from-URL-only enforcement (gotcha #41)

The project handler still read `organization_id` from the request body (Create) and the query string with cross-org fan-out fallback (List), even though both routes are mounted under `RequireOrganizationAccess` which already pins the URL's orgID into ctx.

After the Phase-2 URL refactor, the frontend stopped sending the query param, so List silently fanned out across every org the user belonged to ("open org A, see projects from A+B+C"). Create accepted body-side `organization_id` so a user could POST to `/api/v1/organizations/{orgA}/projects` and write into orgB.

Both fixed structurally:
- `createProjectBody` no longer has any tenancy field (DisallowUnknownFields → 422 on rogue body keys).
- List rejects `?organization_id=` query overrides with 422.
- Both handlers read orgID from `httpctx.MustGetOrganizationID(ctx)` only.

The `orgSvc` and `memberSvc` fields on `project.Handler` became orphan and were deleted along with the constructor args.

**Cross-cutting principle reinforced**: single source of truth per concern — tenancy in path, authorization in middleware (+ future repo-side `WHERE organization_id = ?`), cross-tenant views in their own dedicated endpoints. Pattern verified at Stripe Connect, GitHub `/orgs/{org}/repos`, PostHog, Auth0, Supabase, Jira; defence-in-depth at the SAME layer (handler re-checking what middleware enforced) is cargo-cult per Loki / Vault / Consul precedent and CLAUDE.md's `Must*` invariant rule.

**Process lesson**: when handlers and routes evolve at different times, the regression class is "handler reads pre-refactor state from places the new mount no longer carries" — search for `c.Body.<TenantID>`, `URL.Query().Get("organization_id")`, etc. as part of any URL restructure review.
