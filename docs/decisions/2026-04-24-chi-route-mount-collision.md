---
date: 2026-04-24
status: enacted
tags: [chi, routing, http-framework]
---

# chi `r.Route` Mount collision diagnosis

chi Mount collisions at boot after the Huma migration — `auth/routes.go` had two `r.Route("/api/v1/auth", ...)` wrappers (RegisterPublicRoutes + RegisterProtectedRoutes) on sibling `chi.Group` sub-routers that share the parent mux's routing tree, and observability + evaluation dashboard handlers each wrapped their endpoints in a duplicate `r.Route("/api/v1/projects/{projectId}", ...)`. Panic originates at `go-chi/chi/v5/mux.go` Mount's `findPattern` guard.

The minimal fix is also the right fix: **delete the `r.Route` wrappers.** Only `r.Route` / `r.Mount` call chi's internal Mount; direct `r.Post` / `r.Get` registrations at absolute paths write concrete leaves to the trie and never collide, even across sibling `chi.Group`s with different middleware stacks.

Auth's two Register functions now register concrete `/api/v1/auth/*` paths directly — the bootstrap's two-chi.Group posture architecture (dashPublic LimitByIP, dashAuth RequireAuth+LimitByUser) is preserved unchanged.

Observability's project-scoped routes are split into sibling Mounts at distinct deeper prefixes (`/scores`, `/sessions`, `/filter-presets`) — chi's radix trie routes static-deeper before wildcard-shallower, so they coexist with evaluation's `/api/v1/projects/{projectId}/*` catchall.

**False-start worth recording**: my first attempt consolidated auth's two Register functions into one with inner `r.Group` blocks and introduced a `MiddlewareDeps` contract so the merged function could declare its own middleware — reverted after recognising it was overbuilt. The chi `_examples/rest` pattern (one `r.Route` with inner `r.Group`) is for **single-function** route-tree construction and doesn't apply to split-registration cases; checking the simpler primitive (direct `r.Post` on a shared tree) FIRST would have produced a two-line diff instead of a ~20-file refactor plan.

**Why CI didn't catch the original bug**: the Huma-era `schema_collision_test.go` was the only test that booted the full route tree, and was deleted as Huma-specific without a chi replacement. Restored as `internal/server/routes_test.go` — boots `addRoutes(chi.NewRouter(), stubDeps)` against typed-nil service deps with a `recover()` assertion, and catches registration-time NPEs too. Also added `scripts/lint-conventions.sh` rule flagging any `r.Route("/api/v1/...")` prefix that appears in more than one handler file (catches future collisions at lint, before boot).

**Generalisable rule** (both technical and process): when decomposing routes by middleware posture leads to two `r.Route` calls at the same prefix, drop the `r.Route` — concrete-path registration handles sibling-group coexistence natively; don't reach for consolidation or new abstractions before checking whether the simpler chi primitive already covers the case.
