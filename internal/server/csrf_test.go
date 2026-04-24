// Tests for the CSRF (CrossOriginProtection) wiring.
//
// The trusted-origin list is driven by Server.CORSAllowedOrigins so a
// CORS-allowed origin is automatically CSRF-allowed. These tests pin
// that invariant: misconfigured origins panic at boot, untrusted
// origins are rejected, trusted origins pass through, and safe HTTP
// methods bypass the check entirely (so docs/openapi GETs are
// unaffected by misconfiguration).
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/config"
)

// stubDeps returns a Deps with just enough config to build CSRF
// middleware. Origins are caller-supplied so each test can specify
// its own trust set.
func stubDeps(origins ...string) Deps {
	return Deps{
		Config: &config.Config{
			Server: config.ServerConfig{
				CORSAllowedOrigins: origins,
			},
		},
	}
}

// passHandler is a sink handler that records whether it ran.
func passHandler(ran *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		*ran = true
		w.WriteHeader(http.StatusOK)
	}
}

// TestCSRF_TrustsConfiguredOrigin — POST from an origin in
// CORSAllowedOrigins reaches the downstream handler.
func TestCSRF_TrustsConfiguredOrigin(t *testing.T) {
	csrf := crossOriginProtection(stubDeps("http://localhost:3000"))
	var ran bool
	h := csrf.Handler(passHandler(&ran))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, ran, "handler must run for trusted origin")
}

// TestCSRF_RejectsUntrustedOrigin — POST from an unknown origin is
// rejected with 403 before the downstream handler runs.
func TestCSRF_RejectsUntrustedOrigin(t *testing.T) {
	csrf := crossOriginProtection(stubDeps("http://localhost:3000"))
	var ran bool
	h := csrf.Handler(passHandler(&ran))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.False(t, ran, "handler must not run for untrusted origin")
}

// TestCSRF_PanicsOnMalformedOrigin — boot-time panic when a CORS
// origin entry isn't a valid scheme://host[:port]. Matches mustCORS's
// fail-loud-at-boot posture so config typos don't silently 403 the
// dashboard at runtime.
func TestCSRF_PanicsOnMalformedOrigin(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "must panic on malformed origin")
		msg, ok := r.(string)
		require.True(t, ok, "panic value must be a string")
		assert.Contains(t, msg, `server: invalid CSRF trusted origin "not a url"`,
			"panic message must surface the offending origin and prefix")
	}()
	_ = crossOriginProtection(stubDeps("not a url"))
}

// TestCSRF_SafeMethodsBypassCheck — GET/HEAD/OPTIONS pass through
// regardless of trust state. This is why /api/v1/users/me returns
// 401 (auth) not 403 (CSRF) when called from any origin, and why
// /api/v1/openapi.json works for docs viewers from any origin.
func TestCSRF_SafeMethodsBypassCheck(t *testing.T) {
	// No trusted origins configured.
	csrf := crossOriginProtection(stubDeps())

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			var ran bool
			h := csrf.Handler(passHandler(&ran))

			req := httptest.NewRequest(method, "/api/v1/users/me", nil)
			req.Header.Set("Origin", "http://evil.example.com")
			req.Header.Set("Sec-Fetch-Site", "cross-site")

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code,
				"safe method %s must bypass CSRF check", method)
			assert.True(t, ran, "handler must run for safe method %s", method)
		})
	}
}

// TestCSRF_NoOriginHeaderPasses — same-origin browser requests and
// curl-style server-to-server requests have no Origin/Sec-Fetch-Site
// headers; CrossOriginProtection treats them as same-origin and
// passes through. This is the production path for the SDK plane
// and for any same-origin dashboard deployment.
func TestCSRF_NoOriginHeaderPasses(t *testing.T) {
	csrf := crossOriginProtection(stubDeps())
	var ran bool
	h := csrf.Handler(passHandler(&ran))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	// No Origin, no Sec-Fetch-Site.

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, ran, "no-Origin request must pass through (same-origin assumption)")
}

// TestCSRF_MultipleTrustedOrigins — every entry in the slice gets
// added; covers the dev .env case (CORS_ALLOWED_ORIGINS lists three
// localhost ports for the dashboard, docs, and any local-tools UI).
func TestCSRF_MultipleTrustedOrigins(t *testing.T) {
	csrf := crossOriginProtection(stubDeps(
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3006",
	))

	for _, origin := range []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3006",
	} {
		t.Run(origin, func(t *testing.T) {
			var ran bool
			h := csrf.Handler(passHandler(&ran))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
			req.Header.Set("Origin", origin)
			req.Header.Set("Sec-Fetch-Site", "cross-site")

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.True(t, ran)
		})
	}
}
