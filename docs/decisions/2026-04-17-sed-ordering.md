---
date: 2026-04-17
status: enacted
tags: [tooling, process]
---

# Bulk `sed` ordering pitfalls

Bulk `sed` replacements on SQL migrations require replacing longer patterns first (`VARCHAR(26)` before `CHAR(26)`), otherwise `CHAR(26)` inside `VARCHAR(26)` matches first and produces invalid types like `VARUUID`. Same applies to frontend: `sed 's/ulid/uuid/g'` on TS files corrupted `import('ulid')` → `import('uuid')` (package didn't exist) and `ulid()` → `uuid()` (wrong API). Always verify dynamic imports and runtime calls after bulk renames.
