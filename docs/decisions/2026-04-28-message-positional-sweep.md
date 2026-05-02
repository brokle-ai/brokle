---
date: 2026-04-28
status: enacted
tags: [error-taxonomy, api-design, sweep]
---

# Message-as-positional sweep (late evening)

Promoted `message` to a positional arg in 6 more constructors (`Conflict`, `PermissionDenied`, `NotImplemented`, `Unauthenticated`, `BadRequest`, `RateLimit`) after the InvalidParam wave revealed the same pattern.

**Empirical signal that motivated it**: a 1-line audit (`grep -A3 'appErrors\.<X>(' \| grep -c WithMessage`) showed every one of these constructors had **100% of callsites setting WithMessage** — i.e., the default message ("X conflict", "permission denied for X", "authentication required", "rate limit exceeded", "X not implemented") was effectively unused in practice.

Two constructors (`NotFound` 34%, `AlreadyExists` 15%) genuinely benefit from default messages and stayed `(resource, opts...)`.

**The new uniform rule across all 13 constructors**: positional args = always-required, opts = optional, decided by empirical "is this always set?" rather than aesthetic symmetry. Matches `Internal(message, cause, opts...)` and `Upstream(resource, cause, opts...)` precedent that already existed for the same reason.

**Sweep**: 145 callsites mechanically converted via two perl passes (`Conflict`/`PermissionDenied`/`NotImplemented` had `(resource, ...)` so the regex was `("res", WithMessage("M"))` → `("res", "M")`; `Unauthenticated`/`BadRequest`/`RateLimit` had no resource so `(WithMessage("M"))` → `("M")`). Wire shape preserved exactly.

**Generalisable rule**: when designing a multi-positional constructor, **don't lead with aesthetic symmetry — lead with empirical args usage**. A 100%-set option is a signal the API design hasn't matched the actual call distribution. Single-positional `(name, opts...)` is right when defaults are usable; two-positional `(name, message, opts...)` is right when defaults are vestigial. The cost of the wrong choice is hundreds of lines of `WithMessage(...)` boilerplate at call sites that the constructor could have absorbed with one line of factory code.
