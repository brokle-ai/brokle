package humawrap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"brokle/internal/transport/http/middleware/humawrap"
)

// ctxKey is a typed key used to verify chi-middleware context
// mutations propagate to downstream Huma handlers.
type ctxKey string

const userKey ctxKey = "user"

// helloOutput is the Huma response body for the test operation.
type helloOutput struct {
	Body struct {
		Greeting string `json:"greeting"`
	}
}

// newTestAPI returns a chi router + huma.API pair built the same way
// production wiring does so humawrap.Wrap can call humachi.Unwrap on
// the live adapter.
func newTestAPI(t *testing.T) (*chi.Mux, huma.API) {
	t.Helper()
	mux := chi.NewRouter()
	api := humachi.New(mux, huma.DefaultConfig("humawrap-test", "0.0.0"))
	return mux, api
}

// registerHello registers a single GET /hello operation against the
// supplied API (or huma.Group). The handler reads ctxKey "user" if
// present and includes it in the greeting so tests can assert that
// chi-middleware-set context values propagated through.
func registerHello(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "hello",
		Method:      http.MethodGet,
		Path:        "/hello",
	}, func(ctx context.Context, _ *struct{}) (*helloOutput, error) {
		out := &helloOutput{}
		out.Body.Greeting = "anonymous"
		if v, ok := ctx.Value(userKey).(string); ok {
			out.Body.Greeting = v
		}
		return out, nil
	})
}

// TestWrap_Passthrough — a chi middleware that calls next must allow
// the Huma handler to run.
func TestWrap_Passthrough(t *testing.T) {
	mux, api := newTestAPI(t)

	called := false
	chiMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.Wrap(chiMW))
	registerHello(grp)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, called, "chi middleware must run")
	assert.Contains(t, rec.Body.String(), "anonymous")
}

// TestWrap_Abort — a chi middleware that writes a response and
// returns without calling next must short-circuit the Huma handler.
func TestWrap_Abort(t *testing.T) {
	mux, api := newTestAPI(t)

	handlerRan := false
	chiMW := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.Wrap(chiMW))
	huma.Register(grp, huma.Operation{
		OperationID: "guarded",
		Method:      http.MethodGet,
		Path:        "/guarded",
	}, func(_ context.Context, _ *struct{}) (*helloOutput, error) {
		handlerRan = true
		return &helloOutput{}, nil
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/guarded", nil))

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerRan, "handler must not run when middleware aborts")
}

// TestWrap_ContextPropagation — context values written by a chi
// middleware via r.WithContext must reach the Huma handler.
func TestWrap_ContextPropagation(t *testing.T) {
	mux, api := newTestAPI(t)

	chiMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userKey, "alice")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.Wrap(chiMW))
	registerHello(grp)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "alice",
		"handler must observe context value set by chi middleware")
}

// TestWrap_HeaderPropagation — response headers written by a chi
// middleware must appear in the final response.
func TestWrap_HeaderPropagation(t *testing.T) {
	mux, api := newTestAPI(t)

	chiMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom", "yes")
			next.ServeHTTP(w, r)
		})
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.Wrap(chiMW))
	registerHello(grp)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "yes", rec.Header().Get("X-Custom"))
}

// TestWrap_ContextPropagation_Chained — a value set in request
// context by an outer chi middleware MUST be observable inside an
// inner chi middleware and in the final Huma handler. Reproduces the
// production panic "user ID not found — protected route is missing
// RequireAuth middleware" that occurred when RequireAuth +
// LimitByUser were chained via humawrap.WrapMany: humachi.Unwrap
// walks past subContext overrides back to the bare *chiContext, so
// without re-basing r against hctx.Context() every subsequent
// middleware saw the pre-chain request and clobbered the override.
func TestWrap_ContextPropagation_Chained(t *testing.T) {
	mux, api := newTestAPI(t)

	outer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userKey, "alice")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	innerSaw := ""
	inner := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v, ok := r.Context().Value(userKey).(string); ok {
				innerSaw = v
			}
			next.ServeHTTP(w, r)
		})
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.WrapMany(outer, inner)...)
	registerHello(grp)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "alice", innerSaw,
		"inner chi middleware must see context values set by outer chi middleware")
	assert.Contains(t, rec.Body.String(), "alice",
		"handler must still observe the outer-set context after the inner middleware runs")
}

// TestWrap_InnerMiddlewareSeesOuterContext — the inner chi middleware
// can act on the outer-set value (here: write a response header) even
// if it never calls next. Proves chi-middleware code paths — not just
// the Huma handler — receive the accumulated context.
func TestWrap_InnerMiddlewareSeesOuterContext(t *testing.T) {
	mux, api := newTestAPI(t)

	outer := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userKey, "bob")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	inner := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v, ok := r.Context().Value(userKey).(string); ok {
				w.Header().Set("X-Seen-User", v)
			}
			next.ServeHTTP(w, r)
		})
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.WrapMany(outer, inner)...)
	registerHello(grp)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "bob", rec.Header().Get("X-Seen-User"),
		"inner middleware must observe outer middleware's context at the chi layer")
}

// TestWrapMany_OuterToInnerOrder — WrapMany must preserve order: the
// first element runs outermost, mirroring the chi r.Use convention.
func TestWrapMany_OuterToInnerOrder(t *testing.T) {
	mux, api := newTestAPI(t)

	var trace []string
	mk := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, "before:"+name)
				next.ServeHTTP(w, r)
				trace = append(trace, "after:"+name)
			})
		}
	}

	grp := huma.NewGroup(api)
	grp.UseMiddleware(humawrap.WrapMany(mk("a"), mk("b"), mk("c"))...)
	registerHello(grp)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t,
		[]string{"before:a", "before:b", "before:c", "after:c", "after:b", "after:a"},
		trace,
		"first element must be outermost (matches chi r.Use ordering)")
}
