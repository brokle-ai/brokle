---
date: 2026-04-14
status: superseded
tags: [openapi, dto, swagger]
note: Swagger/swaggo was retired in the chi-only migration; this entry is preserved for historical context only.
---

# Swagger annotations must reference DTOs, not domain entities

Swagger `@Success` annotations must reference the DTO type actually returned by the handler, not the domain entity. Stale annotations cause `make generate` to produce incorrect OpenAPI schemas that mislead SDK clients.

> **Status note (2026-04-30):** the entire swaggo/Huma OpenAPI surface was removed in the chi-only migration. The principle (DTOs are the wire boundary, not domain entities) survives in the current `pkg/response` + handler-DTO conventions; the specific tool-level instruction is no longer applicable.
