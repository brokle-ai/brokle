---
date: 2026-04-17
status: enacted
tags: [id-generation, abstractions]
---

# `pkg/uid` minimised — don't wrap `uuid.UUID`

ID type wrappers should only exist when they earn their place. `uuid.UUID` from `google/uuid` already implements `Scan`, `Value`, `MarshalText`, `UnmarshalText` — wrapping it in a custom type (like SigNoz does) forces re-implementing 12+ interface methods for zero benefit. `pkg/uid` has exactly two functions: `uid.New()` (encodes the UUIDv7 decision) and `uid.TimeFromID()` (non-trivial extraction logic). Everything else is `uuid.Parse()`, `uuid.Nil`, `id.String()` directly. Removed `uid.Parse()`, `uid.MustParse()`, `uid.IsZero()`, `uid.Nil` after recognizing they were trivial pass-throughs.
