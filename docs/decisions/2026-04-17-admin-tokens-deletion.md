---
date: 2026-04-17
status: enacted
tags: [scaffolded-code, deletion]
---

# Scaffolded admin/token routes deletion

Scaffolded-but-unreachable code must be deleted, not kept "for later." The `admin/token_admin.go` package with four `/admin/tokens/*` routes was guarded by `admin:manage` — a permission never seeded. No user could ever call it; always 403. Kept for months until an audit found it. CLAUDE.md rule: "don't design for hypothetical future requirements." Git history preserves the implementation if a real need arrives.
