---
date: 2026-04-14
status: enacted
tags: [domain, handlers, dto]
---

# Domain-entity field-type changes require handler audit

Changing a domain entity field type (e.g., `string` → `json.RawMessage`) requires auditing ALL handlers that serialize that entity. DTO conversion helpers (`toScoreResponse()`) must be used at every endpoint — missing one creates a serialization regression. Swagger annotations must also be updated to reference the DTO type, not the domain entity.
