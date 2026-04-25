# Architecture Reference — LLM Observability Platform

**Audience:** Coding agents and engineers building on this platform.
**Status:** Authoritative. When in doubt, follow this document. If something here seems wrong, flag it — do not silently deviate.

---

## 1. Product context

We are building an LLM observability platform. Customers send us traces of their LLM calls (prompts, completions, tool uses, metadata) and we give them dashboards, search, evals, and alerting on top of that data.

Two API consumers:

1. **Public SDKs** (JavaScript, Python, and future languages) used by third-party developers. Authenticated via API keys.
2. **First-party dashboard** (Next.js web app) used by our customers to view their data. Authenticated via session cookie + short-lived JWT.

Compliance targets: **US and EU data residency**, SOC 2 Type II, GDPR. HIPAA-ready architecture (may not be certified day one).

---

## 2. Top-level topology

Three planes. Keep them mentally and physically separate.

```
                    ┌─────────────────────────────────────────┐
                    │         CONTROL PLANE (global)          │
                    │  yourcompany.com                        │
                    │  app.yourcompany.com      ← region picker│
                    │  api.yourcompany.com/v1/  ← account/billing only
                    │                                          │
                    │  Stores: org metadata, billing, auth,    │
                    │          region pinning                  │
                    │  Never stores: prompts, completions,     │
                    │                traces, any customer data │
                    └─────────────────────────────────────────┘
                               │                    │
                ┌──────────────┘                    └──────────────┐
                ▼                                                  ▼
    ┌──────────────────────┐                        ┌──────────────────────┐
    │   US DATA PLANE      │                        │   EU DATA PLANE      │
    │                      │                        │                      │
    │ us.api.yourcompany   │                        │ eu.api.yourcompany   │
    │   /v1/...            │                        │   /v1/...            │
    │   /internal/...      │                        │   /internal/...      │
    │                      │                        │                      │
    │ us.dashboard.your…   │                        │ eu.dashboard.your…   │
    │                      │                        │                      │
    │ Independent:         │                        │ Independent:         │
    │  - DB                │                        │  - DB                │
    │  - object storage    │                        │  - object storage    │
    │  - queues            │                        │  - queues            │
    │  - k8s cluster       │                        │  - k8s cluster       │
    │  - secrets           │                        │  - secrets           │
    └──────────────────────┘                        └──────────────────────┘
```

**Rule:** Customer observability data (prompts, completions, traces, evals, logs of their LLM calls) lives **only** in the regional plane the org is pinned to. It never transits the control plane. It never crosses regions.

---

## 3. URL structure

### Public API (SDKs)

```
https://us.api.yourcompany.com/v1/...
https://eu.api.yourcompany.com/v1/...
```

- Region in the subdomain, version in the path.
- SDKs accept a `region` parameter (or encode it in the API key) and construct the base URL from it.
- `v1` in US and `v1` in EU are the **same contract**. Never ship a version to one region ahead of the other.

### Dashboard API (internal)

```
https://us.api.yourcompany.com/internal/...
https://eu.api.yourcompany.com/internal/...
```

- Same host as the public API, different path prefix.
- **Not versioned.** Deploys lockstep with the dashboard.
- Called only by `us.dashboard.yourcompany.com` / `eu.dashboard.yourcompany.com`.

### Control plane

```
https://yourcompany.com                    ← marketing
https://app.yourcompany.com                ← region picker, post-login redirect
https://api.yourcompany.com/v1/...         ← account, billing, org metadata
```

- Handles signup, billing, org creation, region selection, user auth.
- Stores metadata only. **No observability payloads, ever.**

### Hard rules

- **The dashboard must never call `/v1/`.** If the dashboard needs something, add it to `/internal/`. This keeps the public surface small and prevents leakage of admin-shaped endpoints into the SDK contract.
- **Do not use path-based regions** (e.g. `api.yourcompany.com/eu/v1/`). The regional boundary must be at DNS and infrastructure, not in application routing code.
- **Do not let the control plane call into a regional plane to fetch customer data.** If the control plane needs to display something region-specific, redirect the user to the regional dashboard.

---

## 4. Versioning

### Public API: `/v1/` in the path, from day one.

- Path-based (`/v1/`, `/v2/`). No header-based versioning.
- A breaking change requires a new version. Non-breaking changes (adding optional fields, adding endpoints, adding enum values with a documented "unknown" fallback) go into the current version.

**Breaking changes include:**
- Removing or renaming a field, endpoint, or parameter
- Changing a field's type or semantics
- Changing error response shape or error codes
- Tightening validation (rejecting input that was previously accepted)
- Changing default values
- Changing pagination, sorting, or filtering semantics

**Non-breaking:**
- Adding optional request fields
- Adding response fields (clients must ignore unknown fields — document this)
- Adding new endpoints
- Loosening validation

Old versions stay supported for a documented deprecation window (suggested: 12 months minimum after the next version ships, announced via changelog and SDK warnings).

### Dashboard API: unversioned.

Deploys with the dashboard. If the contract changes, both sides change in the same PR.

### SDKs

SDKs have their own semver independent of API version. `sdk@2.3.0` can target API `v1`. Bumping API to `v2` means a new major SDK release that targets `v2`.

---

## 5. Separation of surfaces

Whatever the existing code organization, the following boundaries must hold:

- **Public routes (`/v1/*`), internal routes (`/internal/*`), and control-plane routes are three distinct surfaces.** Each has its own auth middleware, its own rate-limit policy, and its own review standards. Do not merge handlers across surfaces.
- **Business logic is shared; HTTP handling is not.** Route handlers stay thin — auth, validation, response shaping. The domain logic underneath should be callable from any surface and must not know which surface invoked it.
- **Storage access goes through a single layer.** Domain code must not reach into database clients or object storage SDKs directly. This is what makes per-region deployment, CMEK, and retention enforcement tractable later.
- **OpenAPI is the source of truth for `/v1/*`.** If a PR changes a public route's shape without updating OpenAPI, it is incomplete. SDKs are generated, not hand-written.
- **Shared types, not OpenAPI, for `/internal/*`.** TypeScript types (or equivalent) shared between the dashboard and internal routes are sufficient and lower-overhead.

---

## 6. Authentication

### Public API (`/v1/*`)

- **API keys.** Format: `sk_live_<region>_<random>` or `sk_test_<region>_<random>`.
  - Region encoded in the prefix so a misconfigured base URL fails loudly.
  - Stored hashed at rest (argon2id or bcrypt). Never log the raw key. Log a fingerprint (first 8 chars + last 4) for debugging.
- Middleware validates the key, loads the org, and attaches `{ orgId, region, keyId, scopes }` to the request context.
- Rate limit: per-key, sliding window. Tiered by plan.

### Dashboard API (`/internal/*`)

- **Session cookie** (httpOnly, Secure, SameSite=Lax) established at login.
- Dashboard exchanges the session for a short-lived JWT (5–15 min) for API calls.
- Middleware validates the JWT, loads user + org memberships, attaches `{ userId, orgId, region, roles }`.
- Rate limit: per-user.

### Control plane

- Session cookie for web, scoped API keys for M2M (e.g., billing webhooks).
- Has its own user/org tables. Regional planes receive user/org identity via signed tokens; they do not query the control plane synchronously on the hot path.

### Cross-cutting

- Every authenticated request has a resolved `orgId` and `region`. Handlers must verify that the region they are running in matches the org's pinned region. If it doesn't, return 404 (not 403 — don't leak existence).
- A region-guard middleware enforces this at the edge of every regional service.

---

## 7. Region pinning and data residency

- Every org has a `region` field set at creation time. **Immutable** for now — migration tooling is a separate project.
- SDKs must be configured with the correct region. Mismatches fail fast at auth.
- Dashboard: after login in the control plane, user is redirected to `us.dashboard.*` or `eu.dashboard.*` based on the org they selected. There is no single dashboard that shows both regions' data side by side.
- Customer data (anything tagged as observability payload — see section 9) **must not** leave its region. This is enforced by:
  - Network policies between clusters (no direct connectivity).
  - Per-region object storage buckets in the region's geography.
  - Per-region databases.
  - CI checks that flag new outbound calls from regional services to non-regional hostnames.

### What the control plane may hold globally

- Org name, org ID, region pin, plan, billing status.
- User identity (email, name, hashed password / SSO identifiers).
- User → org memberships and roles.
- Audit log of control-plane actions (logins, billing changes).

### What the control plane must never hold

- Prompts, completions, tool calls, embeddings, traces, spans.
- Eval inputs or outputs.
- Any customer-uploaded content.
- Anything the customer would consider "their data in our product."

---

## 8. Compliance scaffolding

Decisions to honor in code from day one, even if the feature ships later:

- **PII redaction at ingest.** SDK supports a client-side redaction hook. Server supports a configurable redaction pipeline (regex + optional model-based) that runs before persistence. Redacted fields are replaced with typed placeholders (e.g., `<EMAIL>`), not deleted, so shape is preserved.
- **Customer-managed encryption keys (CMEK).** Storage layer supports per-tenant KMS keys. Day one we use a platform key; the interface is built so CMEK is a config change, not a rewrite.
- **Per-org data retention.** All observability records have a TTL derived from org config. A scheduled job enforces deletion. Retention is org-configurable within plan limits.
- **Right to erasure (GDPR Article 17).** Every observability record is keyed by `orgId` and has a `subjectId` field (optional, set by customer) to support subject-level deletion. Deletion is a first-class operation, not an afterthought.
- **Audit log.** Every admin action (key creation, member invite, retention change, data export, deletion) is logged to an append-only store with actor, timestamp, action, target, and IP.
- **Subprocessor list.** Maintained in a tracked document. Adding a new vendor requires updating it and notifying customers per DPA terms.

---

## 9. Data classification

Tag every field and table with one of:

- **Observability payload** — prompts, completions, traces, spans, tool I/O, evals. Regional only. Encrypted at rest. Subject to retention and erasure.
- **Customer metadata** — org settings, dashboard configs, saved queries. Regional. Not subject to retention limits but subject to erasure on account deletion.
- **Identity** — users, orgs, memberships. Control plane. Subject to erasure on account deletion.
- **Billing** — invoices, payment methods, usage counters. Control plane. Retained per tax/accounting requirements (typically 7 years), overrides erasure for the retained fields.
- **Operational** — logs, metrics, traces *of our own system*. Must be scrubbed of customer payloads before leaving the region. If an error log would contain a prompt, redact it.

When adding a new table or field, the PR description must state which classification it falls under. If unclear, default to the most restrictive and ask in review.

---

## 10. SDK generation

- Source of truth: OpenAPI 3.1 spec, checked in and versioned.
- JS SDK: generated via `openapi-typescript` + a thin hand-written client wrapper for auth, retries, and region handling.
- Python SDK: generated via `openapi-python-client` + equivalent wrapper.
- Do not hand-edit generated files. Extend via the wrapper layer.
- SDK release gates: spec lints cleanly, generated code compiles, contract tests pass against a staging deploy, changelog entry exists.

### SDK UX requirements

- Region is required at client construction. No silent default.
- Retries with exponential backoff on 5xx and 429, respecting `Retry-After`.
- Idempotency keys supported on all write endpoints.
- Timeouts configurable, with sensible defaults (30s total, 10s connect).
- User-Agent identifies SDK name, version, and language runtime.
- Errors surface status code, error code, and request ID.

---

## 11. Operational expectations

- Every request gets a `X-Request-ID` (generate if absent, propagate if present). Logged on both sides. Returned in error responses.
- Structured logging. No `console.log` / `print` in production code paths.
- Metrics tagged with `region`, `surface` (`public` | `internal` | `control`), and `route`. Never tag with `orgId` directly (cardinality); use hashed bucketing if needed.
- Runbooks are region-aware. An incident in US does not imply an incident in EU.
- Deploys go to staging → one region → the other region. Never both regions simultaneously for anything riskier than a config change.
- Database migrations run per-region. Migration code must be backwards-compatible with the previous app version (expand/contract pattern) because regions deploy at different times.

---

## 12. Decisions that are NOT open

These have been made. Do not relitigate in a PR description. If you genuinely think one is wrong, open a design doc.

1. Region in subdomain, not path.
2. `/v1/` path versioning on public API.
3. Dashboard API unversioned.
4. Dashboard never calls `/v1/`.
5. Control plane holds no observability data.
6. SDKs generated from OpenAPI.
7. Region is immutable on an org (for now).
8. API keys include region in the prefix.

---

## 13. Open questions (flag, don't guess)

If your task touches any of these, stop and ask:

- Billing model for data volume vs. seats — affects rate limiting and plan enforcement.
- Whether we expose a webhook system in v1 or defer to v2.
- SSO provider support matrix for enterprise.
- Self-hosted / on-prem offering — architectural implications are large.
- HIPAA BAA scope — which features are in-scope vs. out.

---

**Last updated:** 2026-04-24
**Owner:** Platform team
