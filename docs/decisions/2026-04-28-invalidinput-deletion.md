---
date: 2026-04-28
status: enacted
tags: [error-taxonomy, k8s-pattern, deletion]
---

# `InvalidInput` deletion + `InvalidFields` constructor (final)

Deleted `InvalidInput(resource, opts...)` constructor entirely; added `InvalidFields(details []ErrorDetail, opts...)` for the validator multi-field case.

The earlier "InvalidInput keeps the entity-level case" reasoning was wrong: an audit of the surviving 4 entity-level sites found that ALL of them were mis-categorized — `signup` XOR is a request-shape problem (`BadRequest`), incomplete OAuth/login sessions are state-machine problems (`Conflict`), and `user/handlers.go body.organization_id` is a single-field UUID parse (`InvalidParam`). With those reclassified, the only legitimate `InvalidInput("",...)` use was the validator multi-field boundary.

Web research surveyed stdlib `net/http`, k8s `apierrors`, cockroachdb/errors, grpc/status, go-hclog, go-chi/chi: **k8s `apierrors.NewInvalid(kind, name, errs)` is the canonical multi-field validator adapter** — separate named constructor for the rare boundary case, not a struct literal.

Reviewer challenged "is `InvalidFields` a workaround for one caller?" — the right answer (per k8s, cockroachdb, grpc/status precedent) is no: constructor count isn't the cost; *inconsistent construction shape* is. 313 sites using `InvalidParam(...)` + 1 site using `&Error{Reason: ..., Errors: ..., Op: ...}` literal would become the grep-anomaly future contributors copy. The k8s pattern is "named constructor per shape, struct literals only for in-package internal mutation."

Final API: `InvalidParam` (single, 313 sites) + `InvalidFields` (multi, 1+ future sites). `InvalidInput` deleted; `scripts/lint-conventions.sh` blocks reintroduction.

**Generalisable rule (added to development discipline)**: when designing an error API, **the wire shape distinction (Param-set vs Errors[]-set) drives the constructor count, not callsite count**. A 1-caller constructor with a clear semantic charter is cheaper than a 1-callsite struct-literal anomaly.

The reviewer's "is this workaround?" question forced two flips before I committed to the research-backed answer; the lesson is to lean on production precedent (k8s/cockroach/grpc) over my own equivocation when the data and pattern align.
