---
date: 2026-04-30
status: enacted
tags: [error-taxonomy, xor, validation, lint]
---

# XOR validation uses `BadRequest`, not synthetic `param`

Reviewer flagged the score-linkage validation using `InvalidParam("linkage", ...)` — `linkage` is a synthetic field name; no such input exists on the request body. Same shape problem at `playground_service.go:354` (`"credential"` vs the wire field `credential_id`).

Web research surveyed 7 production APIs for XOR / "exactly-one-of" field validation: Stripe (`source` XOR `customer`), OpenAI (`messages` XOR `prompt`), Google `google.rpc.BadRequest` (proto3 `oneof`), Microsoft Graph (RFC 7807), JSON Schema `oneOf` (AJV), HashiCorp Boundary/Vault, gqlgen. **6 of 7 return HTTP 400 with a top-level message naming both branches; ZERO use the `errors[]`-array shape for XOR.** AJV emits per-candidate errors only in verbose mode and explicitly tells clients NOT to render them as field errors. The "synthesise a parent path as a field" pattern (Brokle's `linkage`) has no production precedent.

**Fixes**:
- `score_service.go:46` and `:194` reclassified from `InvalidParam("linkage"/"scores[%d].linkage", ...)` to `BadRequest(...)` (HTTP 400, no `param`, message names both branches).
- `playground_service.go:354` `"credential"` renamed to `"credential_id"` to match the wire field.
- `scripts/lint-conventions.sh` first-arg-discipline check extended to ban `"linkage"` as a curated synthetic field.
- CLAUDE.md error-taxonomy gained the rule: "XOR / cross-field constraints use `BadRequest(message)` naming both branches; never synthesise a `param` for a non-existent field."

**Generalisable rule**: when a Go API has both `BadRequest` and `InvalidParam` constructors, the discriminator is "does the failure point at a single real wire field?" — if yes, `InvalidParam`; if no (request shape, cross-field rule, body XOR), `BadRequest`. The Stripe/OpenAI shape is canonical; a synthetic `param` for a non-field is the same anti-pattern as `InvalidParam("invalid", ...)` (adverb-as-field) the lint already catches.

**Process lesson**: when a constructor's positional arg has a strict semantic ("must be a real wire field"), the lint should encode that semantic as a curated bad-word list — discipline-only enforcement caused 5 successive review rounds of swap-bug regressions; the bad-word list catches them at lint time.
