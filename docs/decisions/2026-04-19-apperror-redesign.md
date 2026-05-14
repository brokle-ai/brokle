---
date: 2026-04-19
status: superseded
tags: [error-taxonomy, wire-envelope]
superseded-by: 2026-04-28-error-unification-morning, 2026-04-28-invalidinput-deletion
---

# AppError redesign — closed `Type` enum + open `Code` field

AppError redesign — closed `ErrorType` enum + open `Code` field (OpenAI/Stripe shape) replaces the old single `AppErrorType` that did double duty as both type and code. The original gin error path had a status→code map with a `default → INTERNAL_ERROR` branch, which silently misclassified every framework-emitted 4xx (405/406/408/413/415/416/417/451) as a server fault. The fix is structural: HTTP status is now a pure function of `Type` via `ErrorType.HTTPStatus()`, framework errors arrive via `appErrors.FromHTTPStatus(status, msg, errs...)` which uses an explicit map + RFC 9110 class fallback (4xx→`TypeInvalidRequest`, 5xx→`TypeAPIError`). The misclassification bug class is now structurally impossible. Sentinel test `TestFromHTTPStatusClassFallback` locks the invariant.

Stub-only handler packages get deleted, not converted. Signal: `grep -c "TODO\|\"message\":" handler.go` high + `grep -c "response.Success(c, gin.H" handler.go` high = the package is scaffolding. `handlers/logs/` (172 LoC, 100% TODO) and `handlers/analytics/` (247 LoC, 100% TODO) were deleted rather than converted to Huma. Converting adds zero value — the operations didn't do anything and the migration would have just carried the TODO forward in a new shape.

> **Status note (2026-04-28)**: this design evolved further — the closed `ErrorType` was retired in favour of a closed `Reason` enum + named constructors (`NotFound`, `AlreadyExists`, etc.); see the 2026-04-28 entries for the unification arc.
