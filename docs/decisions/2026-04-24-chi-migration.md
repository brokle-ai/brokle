---
date: 2026-04-24
status: enacted
tags: [chi, http-framework, migration, openapi]
---

# Huma v2 + humachi removal — pure chi v5 + `pkg/request` / `pkg/response`

Dropped Huma v2 + humachi across the entire HTTP surface, converted every handler domain (18 packages, ~280 operations, ~8.3k LoC) to pure chi v5 + `pkg/request` (go-playground/validator) + `pkg/response`.

**Drivers**:
1. **Bus factor** — Huma is single-maintainer and Brokle is pre-production, so the risk is cheap to eliminate now.
2. **Framework friction** — 6 gotchas existed solely because chi + humachi were cooperating on one mux (middleware-ordering landmines, schema-collision panics, `Optional[T]` workaround for Huma's pointer-param panic, error-factory install ordering, humachi binding to initial mux, probe dispatcher built specifically to bypass chi).

Wire contract (Stripe/OpenAI envelope, `{data, pagination}` lists) preserved byte-identical across frontend + Python SDK + JS SDK — zero consumer changes.

**Generalisable rule**: single-maintainer framework dependencies through the request path are a serviceable risk only while pre-production; pay the migration cost before external consumers lock you in.

Removed 2,600 LoC net (Huma + OpenAPI scaffolding deleted > request/validator helpers added). The `pkg/request.DecodeJSON` pattern (MaxBytesReader + DisallowUnknownFields + trailing-data reject + tag-driven validator) is the canonical Go HTTP decoder (Alex Edwards / Brandur references baked in) — every handler now speaks the same body-safety dialect.
