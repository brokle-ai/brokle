package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jub0bs/cors"

	"brokle/internal/transport/http/middleware"
)

// installGlobalMiddleware applies the process-wide middleware stack
// on mux. Called from server.New EXACTLY ONCE, before addRoutes
// registers any route. chi panics once the mux has any route
// (go-chi/chi/v5/mux.go:100-104), so mux-level middleware MUST
// precede all route registration.
//
// Two layers of middleware are installed here:
//
//  1. Truly global middleware that wraps every chi route — RequestID,
//     RealIP, RequestMetadata, RequestLogger, Recoverer, Metrics. The
//     order mirrors the request flow:
//
//     - RequestID — generates or propagates X-Request-ID first so
//     every downstream log line / metric carries it.
//     - RealIP — normalises r.RemoteAddr from X-Forwarded-For before
//     RequestMetadata captures it.
//     - RequestMetadata — stuffs client IP + User-Agent into
//     r.Context() so httpctx accessors work in handlers.
//     - RequestLogger — emits one slog line per request; installed
//     OUTSIDE Recoverer so a panic in a downstream handler still
//     produces a log line (Recoverer writes a 500 after catching,
//     RequestLogger then sees the 500 and logs it on its way out).
//     - Recoverer — slog-shaped panic recovery with request_id
//     correlation; re-raises http.ErrAbortHandler so SSE/WebSocket
//     abort semantics survive (see middleware/recoverer.go).
//     - Metrics — Prometheus counters + histograms keyed by the chi
//     route pattern; runs innermost so label cardinality reflects
//     the matched route, not an intermediate wrapper.
//
//  2. Path-prefix-scoped cross-cutting middleware for the two API
//     surfaces — CORS + CSRF on /api/v1/*. These MUST live at the
//     mux level rather than on chi subrouters because humachi binds
//     its adapter to the captured router reference at construction
//     time; routes registered via huma.Register land on the mux, not
//     on any subrouter built later via r.Route / r.Group.
//
//     The /api/v1 stack is, outer-to-inner:
//     - corsAdmin — CORS preflight handling + allow headers on
//     actual responses.
//     - http.CrossOriginProtection — Go 1.25 CSRF check (no-op for
//     GET/HEAD/OPTIONS; enforces Sec-Fetch-Site / Origin on
//     state-changing methods).
//
//     The /v1 SDK plane gets no browser-origin protections — SDK
//     callers are server-side.
//
// Rate limiting is intentionally NOT mounted here. Industry practice
// (GitHub, Stripe, OpenAI, Anthropic, Cloudflare WAF) scopes limits
// to the authenticated principal post-auth and to IP pre-auth —
// never layered on the same request. A mux-level LimitByIP would
// run against every request including authenticated ones, collapsing
// all shared-egress callers (corporate NAT, CGNAT, mobile carriers,
// Next.js SSR pods, serverless functions) into one bucket. That's
// the SSR-loopback 429 bug class. IP limits therefore live on the
// PUBLIC Huma groups only (dashPublic, sdkPublic) in addRoutes;
// authed groups (dashAuth, sdkAuth) run with LimitByUser /
// LimitByAPIKey alone.
//
// Ops-plane paths (/livez, /readyz, /healthz, /metrics) are served
// by the outer dispatcher in newProbeDispatcher and never reach this
// middleware stack — that is the whole point of the dispatcher.
func installGlobalMiddleware(mux *chi.Mux, deps Deps) {
	mux.Use(chimw.RequestID)
	mux.Use(chimw.RealIP)
	mux.Use(echoRequestID)
	mux.Use(middleware.RequestMetadata(nil))
	mux.Use(middleware.RequestLogger(deps.Logger))
	mux.Use(middleware.Recoverer(deps.Logger))
	mux.Use(middleware.Metrics())

	corsAdmin := mustCORS(deps, "dashboard")
	csrf := crossOriginProtection(deps)

	// Dashboard plane (/api/v1/*) — CORS + CSRF only. Rate limiting
	// is attached per-Huma-group in addRoutes.
	// Dashboard plane (/api/v1/*) — CORS + CSRF only. Rate limiting
	// is attached per-chi-group in addRoutes.
	mux.Use(pathPrefix("/api/v1", corsAdmin.Wrap))
	mux.Use(pathPrefix("/api/v1", csrf.Handler))
}

// echoRequestID copies the request ID out of the chi context (set by
// chimw.RequestID upstream) into the `X-Request-Id` response header.
// Chi's RequestID only writes to context by default; this middleware
// is how the ID surfaces to clients (frontend, SDKs, log correlators).
//
// Matches Stripe/OpenAI/GitHub convention of returning request IDs in
// headers rather than the response body — keeps the body clean of
// metadata when success paths are raw resources.
//
// Must run AFTER chimw.RequestID (which sets the context value) and
// BEFORE the handler writes any bytes (headers must be set first).
func echoRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := chimw.GetReqID(r.Context()); id != "" {
			w.Header().Set("X-Request-Id", id)
		}
		next.ServeHTTP(w, r)
	})
}

// pathPrefix returns a middleware that runs mw only when r.URL.Path
// starts with prefix; otherwise it forwards to next unmodified.
//
// Used to scope cross-cutting middleware (CORS, CSRF, IP rate-limit)
// to a specific API surface without touching the other surface or
// the ops-plane paths handled by the outer probe dispatcher.
//
// The exact prefix match is via strings.HasPrefix, which is the same
// semantics chi uses for its mounted subrouters — `/api/v1` matches
// `/api/v1`, `/api/v1/users`, `/api/v1/openapi.json`, but not
// `/api/v1foo` (because no real route would have that shape and the
// extra cost of a stricter check isn't justified).
func pathPrefix(prefix string, mw func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		wrapped := mw(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, prefix) {
				wrapped.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// mustCORS constructs a jub0bs/cors middleware from config or panics
// at boot if the config is invalid. The panic is intentional:
// jub0bs/cors's whole value proposition is config validation, and a
// CORS misconfig at runtime is exactly the silent-vulnerability
// failure mode we adopted the library to prevent.
//
// "plane" labels which surface (sdk / dashboard) the CORS instance
// belongs to so the panic message points at the right config knob.
func mustCORS(d Deps, plane string) *cors.Middleware {
	mw, err := cors.NewMiddleware(cors.Config{
		Origins:         d.Config.Server.CORSAllowedOrigins,
		Methods:         d.Config.Server.CORSAllowedMethods,
		RequestHeaders:  append(d.Config.Server.CORSAllowedHeaders, "X-CSRF-Token"),
		Credentialed:    true,
		MaxAgeInSeconds: 300, // 5 minutes — matches existing config
	})
	if err != nil {
		panic("server: invalid CORS config for " + plane + " plane: " + err.Error())
	}
	return mw
}

// crossOriginProtection returns a configured Go 1.25
// http.CrossOriginProtection enforcing the Sec-Fetch-Site +
// Origin header check on every non-idempotent request to /api/v1/*.
//
// Trusted-origin list mirrors Server.CORSAllowedOrigins. CORS-trusted
// implies CSRF-trusted: if the browser is allowed to send credentialed
// cross-origin requests under CORS, it must also be allowed past the
// CSRF check, otherwise legitimate dashboard mutations land as 403s
// that look like permission failures (the original symptom this
// helper was reshaped to fix). The two lists are one trust decision
// expressed in two RFCs; keeping them in lockstep is the security
// invariant.
//
// Same-origin deployments need no entries — CrossOriginProtection
// passes same-origin requests through regardless. Cross-origin
// deployments (subdomain split prod, dev with frontend on :3000 and
// API on :8080 when not using the Next.js rewrites proxy, mobile
// webviews) require the origin to appear here AND in the CORS
// allowlist; both are driven by the same env var (CORS_ALLOWED_ORIGINS).
//
// AddTrustedOrigin returns an error on malformed origins — trailing
// slashes, query strings, paths, anything that isn't bare
// scheme://host[:port]. We panic at boot because a typo in env config
// should fail loudly during deploy, not silently 403 the dashboard
// at runtime. Same posture as mustCORS for the same reason.
//
// GET/HEAD/OPTIONS bypass the check regardless of trusted-origin
// state, so docs and openapi paths pass through without configuration.
// Bypass patterns (future webhook receivers) go through
// AddInsecureBypassPattern when the corresponding handler registers.
func crossOriginProtection(d Deps) *http.CrossOriginProtection {
	p := http.NewCrossOriginProtection()
	for _, origin := range d.Config.Server.CORSAllowedOrigins {
		if err := p.AddTrustedOrigin(origin); err != nil {
			panic("server: invalid CSRF trusted origin " +
				strconv.Quote(origin) + ": " + err.Error())
		}
	}
	return p
}
