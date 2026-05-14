---
date: 2026-04-15
status: superseded
tags: [handlers, conventions]
note: Specific helper names reference the gin-era code; the underlying convention (one error pattern + inline parsing) survived the chi migration.
---

# One error pattern, inline parsing across handlers

Handler-layer consistency was enforced by standardizing on ONE error pattern (`response.Error(c, appErrors.NewValidationError(...))`) and inline `uuid.Parse(c.Param(...))` / `c.ShouldBindJSON(&req)` across all 41+ handler files. Shared param extraction helpers were tried and removed — the Go/Gin ecosystem (PhotoPrism, Apache Answer) uses inline parsing, not abstractions. The `evaluation` domain retains a package-local `extractProjectID()` for dual SDK/Dashboard routes.

> **Status note**: the chi-only migration replaced `c.Param` / `c.ShouldBindJSON` / `response.Error(c, ...)` with `request.URLParamUUID(r, ...)` / `request.DecodeJSON(r, ...)` / `response.WriteError(w, err)`, and `appErrors.NewValidationError` was retired in favour of `appErrors.InvalidParam(field, message)`. The underlying principle (inline parsing, one error path) is preserved in `pkg/request` + `pkg/response`.
