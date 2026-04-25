// Invariant tests for the outer probe dispatcher.
//
// The dispatcher's job is to route /livez, /readyz, /healthz, and
// /metrics to raw handlers so they bypass chi and its global
// middleware stack (RequestLogger + Metrics + …). Getting this wrong
// once caused a chi "all middlewares must be defined before routes on
// a mux" panic at boot. These tests guard the invariants at a wire
// level: they exercise the dispatcher + chi + middleware composition
// exactly as production does, just without the service-layer
// dependencies.
package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/config"
)

// newProbeDispatcherForTest builds a minimal dispatcher + chi stack
// that mirrors production wiring but skips the service-layer Deps
// (DBs, auth services, …) the full Server needs. Good enough to
// exercise the middleware-bypass and ready-flip invariants.
//
// Returns the handler, the log buffer the logger writes to, and the
// ready state so tests can flip it.
func newProbeDispatcherForTest(t *testing.T) (http.Handler, *bytes.Buffer, *readyState) {
	t.Helper()
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mux := chi.NewRouter()
	installGlobalMiddleware(mux, Deps{
		Logger: logger,
		Config: stubProbeConfig(),
	})
	// A trivial chi route we use to prove middleware fires on
	// non-probe paths.
	mux.Get("/chi/echo", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	ready := newReadyState()
	ready.MarkReady()
	// healthDeps with nil pointers skips every dependency check, so
	// /readyz returns 200 purely on the readyState flag.
	hd := healthDeps{Logger: logger}

	return newProbeDispatcher(mux, ready, hd), &buf, ready
}

// Probes must NOT emit request-log lines. Kubernetes hits these
// ~6/min × pod count; logging every one floods structured logs.
func TestProbes_BypassRequestLogger(t *testing.T) {
	handler, buf, _ := newProbeDispatcherForTest(t)

	for _, path := range []string{"/livez", "/readyz", "/healthz"} {
		for i := 0; i < 5; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			handler.ServeHTTP(rec, req)
			require.Equalf(t, http.StatusOK, rec.Code, "%s call %d", path, i+1)
		}
	}

	assert.Empty(t, buf.String(),
		"probe responses must not produce request-log lines — middleware leaked through the dispatcher")
}

// /metrics must bypass the Metrics middleware — scrape traffic must
// not appear in http_requests_total, or the counter observes itself
// and inflates monotonically on every Prometheus scrape.
func TestMetrics_ScrapeIsNotCountedAsRequest(t *testing.T) {
	handler, _, _ := newProbeDispatcherForTest(t)

	// Drive a real chi route so we know the counter has at least
	// one series and the middleware stack works.
	{
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/chi/echo", nil)
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	// Scrape /metrics (hit the dispatcher directly, same as a real
	// Prometheus scrape would).
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()

	// The counter for the chi route must be present — proves the
	// middleware ran for non-probe traffic.
	assert.Contains(t, body, `path="/chi/echo"`,
		"http_requests_total must include the chi route label — middleware was not wired")

	// The counter must NOT mention the four ops-plane paths — proves
	// the dispatcher bypassed middleware for them.
	for _, bypass := range []string{"/livez", "/readyz", "/healthz", "/metrics"} {
		needle := `path="` + bypass + `"`
		assert.NotContainsf(t, body, needle,
			"http_requests_total must not include ops-plane path %q — middleware leaked", bypass)
	}
}

// Non-probe traffic MUST be logged + metered. If this test fails,
// the global middleware stack is not wired onto the chi mux (a
// regression of the original chi panic would look like this).
func TestNonProbe_RunsGlobalMiddleware(t *testing.T) {
	handler, buf, _ := newProbeDispatcherForTest(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/chi/echo", nil)
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	log := buf.String()
	assert.NotEmpty(t, log, "non-probe requests must produce a request-log line")
	assert.Contains(t, log, "http_request", "log record must be the RequestLogger's http_request event")
	assert.True(t,
		strings.Contains(log, "/chi/echo") || strings.Contains(log, "chi/echo"),
		"log line must mention the request path: %q", log)
}

// stubProbeConfig returns the minimal *config.Config installGlobalMiddleware
// needs without panicking. The probe tests only exercise paths that
// fall outside /api/v1 and /v1, so the path-prefix middleware
// (CORS, CSRF, IP-limit) constructed below is short-circuited per
// request — the values just need to satisfy jub0bs/cors's validator
// and rate_limit's nil checks at construction time.
func stubProbeConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			CORSAllowedOrigins: []string{"https://example.com"},
			CORSAllowedMethods: []string{http.MethodGet, http.MethodPost},
			CORSAllowedHeaders: []string{"Content-Type"},
		},
	}
}

// Readiness flips to 503 after MarkNotReady — the foundation of the
// two-phase graceful shutdown. kubelet must see 503 before the
// listener closes so the LB deregisters the pod cleanly.
func TestReadyz_FlipsOnMarkNotReady(t *testing.T) {
	handler, _, ready := newProbeDispatcherForTest(t)

	{
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		require.Equal(t, http.StatusOK, rec.Code, "ready state → /readyz must be 200")
	}

	ready.MarkNotReady()

	{
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		require.Equal(t, http.StatusServiceUnavailable, rec.Code,
			"after MarkNotReady → /readyz must be 503")
	}
}
