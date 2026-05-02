---
date: 2026-04-24
status: enacted
tags: [services, abstractions, go-idiom]
---

# Service-interface sweep — return concrete, not interface

Codebase-wide service-interface sweep, grounded in the Go idiom "accept interfaces, return concrete" (Dave Cheney, Rob Pike, stdlib). Audit surfaced 51 domain-level service interfaces across 13 domain packages; of those, 43 were returned from their constructor and 0 had any polymorphism story — zero mocks, zero alternate implementations, zero external plugin boundaries. All orphan.

Earlier in the same session I'd moved in the *opposite* direction twice (created a new domain interface for `ScoreAnalyticsService` because a boundary-crossing service "should" have one; created `pkg/db/pgerrors` to classify pg errors at the repo boundary which required importing HTTP-coupled types into the domain layer). Both reverted once the ecosystem data showed Brokle is a web-app-shape codebase where Gitea/Prometheus/stdlib patterns fit — concrete returns, consumer-side interfaces only when polymorphism actually exists.

The rule now codified as a Mandatory Development Rule: constructors return `*XxxService`; domain-level service interfaces require ≥2 impls **today**, a real mock, or an external plugin boundary.

**Generalisable process lesson**: when introducing an abstraction, first check the ecosystem for precedent in codebases with the same shape, not just the theoretically cleanest design. A single-impl no-mock interface is indirection pretending to be architecture; in a DDD codebase without polymorphism it's the same cargo-cult a misread of clean-architecture tutorials causes elsewhere.

Repository interfaces are the legitimate exception — they satisfy criterion (a) via the service-consumption pattern (service accepts `xxxRepository`, prod uses sqlc+pgx, tests can stub).

Also kept from the same session arc:
- Goroutine-ctx detachment fix in `prompt/execution_service.go:146` (streaming LLM calls now use `context.WithTimeout(context.Background(), 5*time.Minute)`).
- Handler cleanup (listAPIKeysResponse → types.go, projectId UUID-parse normalisation).
- Gitea-pattern service-boundary error translation (repos wrap UNIQUE violations to `domain.ErrAlreadyExists` via `appErrors.IsUniqueViolation(err)`, services translate to `NewConflictError`).
