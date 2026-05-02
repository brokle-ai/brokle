---
date: 2026-04-27
status: enacted
tags: [error-taxonomy, repo-boundary, k8s-pattern, structural-fix]
---

# Error-translation re-architecture (Reason at source)

Eliminated the "service forgot to translate" bug class. Symptom: `RequireProjectAccess` middleware swap from `CanUserAccessProject` → `GetProject` in Phase 6 dropped the sentinel→AppError translation, turning every well-formed-but-missing project ID into 500 instead of 404 across `/api/v1/projects/{projectId}/*`.

Surface fix (Phase A): per-method `errors.Is(err, ErrProjectNotFound) → appErrors.NewNotFoundError("project")` translation in `GetProject` + `GetProjectBySlug`. **But that's a workaround for the bug CLASS** — the discipline fails the moment any service writes a one-line `return s.repo.X(...)` passthrough, of which the codebase has dozens.

Web research across **k8s apimachinery `apierrors.StatusError`, Grafana `errutil`, HashiCorp Boundary `internal/errors`, CockroachDB `cockroachdb/errors`** — every major Go project with multi-transport (HTTP + gRPC + CLI) ships errors-with-reason-at-source: a transport-agnostic `Reason` enum carried by the domain error, read mechanically by each transport adapter. Per-method translation has zero major OSS production adopters; it shows up only in "clean architecture" sample apps.

Architectural fix (Phase B): introduced `internal/core/domain/shared/errors` — `*Error{Reason, Resource, Op, Cause}` with constructors (`derrors.NotFound("project")`, `derrors.AlreadyExists("user")`, `derrors.Internal("get x", cause)`) and `Reason.HTTPStatus()` / `Reason.HTTPType()` pure functions. `pkg/response.WriteError buildAPIError` extended with a third branch: `derrors.As(err) != nil` → wire-identical envelope to the equivalent AppError (locked by `TestErrorShape_DomainErrorMatchesAppError` byte-comparison test).

Project domain migrated end-to-end as the reference implementation (repo wraps `pgx.ErrNoRows` → `derrors.NotFound("project")` ONCE; service is now a correct passthrough; per-domain sentinels `ErrProjectNotFound` + `ErrProjectAlreadyExists` deleted).

**Result**: the bug class is structurally impossible for migrated paths — a service writing `return s.repo.GetByID(ctx, id)` IS correct because the repo error self-describes its wire shape. Other domains (user, auth, billing, observability, etc.) migrate piecemeal; legacy "service translates" continues to work until each domain crosses over.

**Generalisable rule (added to Mandatory Development Rules)**: when a discipline-based pattern fails in production, check whether 5+ major Go OSS projects ship the discipline OR a structural alternative — if zero ship the discipline, the discipline is the bug. The k8s/Grafana/Boundary/Cockroach pattern is the industry default; per-method translation never made it past sample-app tutorials.
