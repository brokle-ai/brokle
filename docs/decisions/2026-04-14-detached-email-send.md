---
date: 2026-04-14
status: enacted
tags: [async, email, context]
---

# Detached email-send goroutine pattern

Synchronous email sends using the HTTP request context are silently dropped when the client disconnects. Always detach email sends into a goroutine with `context.WithTimeout(context.Background(), 30*time.Second)` — matching the invitation service pattern at `internal/core/services/organization/invitation_service.go:147`.
