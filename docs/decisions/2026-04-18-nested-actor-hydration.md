---
date: 2026-04-18
status: enacted
tags: [repos, performance, n-plus-one, json-shape]
---

# Nested-actor LEFT JOIN hydration

Nested-actor hydration via LEFT JOIN — for list endpoints that need inviter/author/reviewer display, populate a `*ActorRef` field on the domain type via a hydrated SQL query rather than looping `GetUser()` per row. Eliminates both N+1 query storms and an entire class of nil-deref panics on nullable FKs with `ON DELETE SET NULL`. Canonical shape: `Invitation { Inviter *InviterRef; Role *RoleRef }` with `InviterRef.FullName()` convenience method. Reject the flat `inviter_email` / `inviter_name` alternative — breaks GitHub/GitLab/Stripe JSON conventions.
