---
date: 2026-04-20
status: enacted
tags: [wire-envelope, sdk, frontend]
---

# Wire-contract migration (Stripe / OpenAI shape)

Option B wire-contract migration — dropped the `{success, data, meta}` envelope across backend + frontend + Python SDK + JS SDK in favour of Stripe/OpenAI shape (raw resource on success, `{error:{...}}` on error, HTTP status as the signal, `X-Request-Id` in header). Signal that triggered it: the frontend `extractData` threw on every Huma success response because Huma emits raw bodies while the JS client still asserted `responseData.success !== undefined`. Rather than patch the asymmetry, unified all four surfaces.

Key decisions:
1. No backward-compat shim — there are no production users yet, so a clean cutover beats a dual-shape adapter.
2. Byte-identical invariant across three backend write sites (`AppError.MarshalJSON`, `statusError` via `NewError`, chi-middleware `WriteError`) locked by `pkg/response/error_shape_test.go`.
3. Handler-returned `*AppError` bypasses `huma.NewError` (Huma's `errors.As` path at `huma.go:1100-1105` marshals the error directly), so both `AppError.MarshalJSON` and the envelope factory must produce the same bytes independently.
4. Both SDKs collapsed their per-manager HTTP helpers into one `_http/client.{py,ts}` module with typed-exception error handling (~700 LoC of dead helper code deleted).

The generalisable rule: when a wire contract splits (success raw, error enveloped, or vice versa), unify in one direction — pick the side with more industry precedent (Stripe/OpenAI > internal envelope conventions) and cut over all consumers simultaneously. Half-envelope states are a bug class, not a migration state.
