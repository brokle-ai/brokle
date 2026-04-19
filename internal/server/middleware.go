package server

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"brokle/internal/transport/http/middleware"
)

// installGlobalMiddleware applies the process-wide middleware stack
// on mux. Called from server.New EXACTLY ONCE, before any route is
// registered on mux — and critically, before humachi.New constructs
// the Huma APIs (humachi's constructor registers /openapi, /docs,
// and /schemas routes immediately, sealing the mux for further
// chi.Mux.Use calls at go-chi/chi/v5/mux.go:100-104).
//
// The order mirrors the request flow documented on addRoutes:
//
//   - RequestID — generates or propagates X-Request-ID first so every
//     downstream log line / metric carries it
//   - RealIP — normalises r.RemoteAddr from X-Forwarded-For before
//     RequestMetadata captures it
//   - RequestMetadata — stuffs client IP + User-Agent into
//     r.Context() so httpctx accessors work in handlers
//   - RequestLogger — emits one slog line per request; installed
//     OUTSIDE Recoverer so a panic in a downstream handler still
//     produces a log line (Recoverer writes a 500 after catching,
//     RequestLogger then sees the 500 and logs it on its way out)
//   - Recoverer — slog-shaped panic recovery with request_id
//     correlation; re-raises http.ErrAbortHandler so SSE/WebSocket
//     abort semantics survive (see middleware/recoverer.go)
//   - Metrics — Prometheus counters + histograms keyed by the chi
//     route pattern; runs innermost so label cardinality reflects
//     the matched route, not an intermediate wrapper
//
// Ops-plane paths (/livez, /readyz, /healthz, /metrics) are served
// by the outer dispatcher in newProbeDispatcher and never reach this
// middleware stack — that is the whole point of the dispatcher.
func installGlobalMiddleware(mux *chi.Mux, deps Deps) {
	mux.Use(chimw.RequestID)
	mux.Use(chimw.RealIP)
	mux.Use(middleware.RequestMetadata(nil))
	mux.Use(middleware.RequestLogger(deps.Logger))
	mux.Use(middleware.Recoverer(deps.Logger))
	mux.Use(middleware.Metrics())
}
