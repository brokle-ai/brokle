---
date: 2026-04-17
status: enacted
tags: [auth, middleware, handlers]
---

# `Must*` auth helpers (no defensive 401 checks)

Defensive handler-level auth checks (`if !exists { 401 }` after `middleware.GetXxx(c)`) mask programming errors and add boilerplate. The idiomatic Go fix is the `Must*` convention from `regexp.MustCompile` / `template.Must`: handlers trust the middleware invariant and call `middleware.MustGetUserID(c)` etc. directly; a misconfigured route (missing `RequireAuth`) now panics → `middleware.Recovery` → HTTP 500 with stack trace, surfacing the bug instead of silently returning a misleading 401. Tuple-return `Get*` forms are reserved for `OptionalAuth` routes. Before adding a new "defensive" 401 in a handler, ask whether the middleware invariant is guaranteed — if yes, use `Must*`.
