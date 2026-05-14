---
date: 2026-04-20
status: enacted
tags: [ops, probes, middleware, observability]
---

# Probe-bypass-middleware via outer `http.ServeMux` dispatcher

Probe-bypass-middleware is the Go ops-plane convention, verified at Loki (`pkg/util/server/server.go` `DoNotLogURL` regex), Consul (`agent/http.go` `wrap()` allowlist), Vault (`vault/http/handler.go`), and the Kubernetes apiserver (`pkg/endpoints/filters/longrunning.go`). Brokle's `newProbeDispatcher` uses an outer `http.ServeMux` (Go 1.22+ pattern routing) wrapping the chi mux — the same shape.

This replaces an earlier attempt that registered `/livez`, `/readyz`, `/healthz`, `/metrics` ON the chi mux "before global middleware" to keep probes silent — that design is incompatible with chi's "middleware before routes" invariant and panics at boot. The wrapper pattern keeps chi's middleware contract intact AND gives ops-plane paths a clean bypass.

Separate-admin-port (OpenTelemetry Collector, SigNoz) is the minority variant reserved for deployments with distinct probe/API trust boundaries — not applicable to a single-port k8s Service like Brokle's today, but reversible to later in ~20 LoC.
