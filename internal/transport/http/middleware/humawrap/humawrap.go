// Package humawrap adapts chi-style middleware
// (func(http.Handler) http.Handler) for use with Huma v2's per-API
// or per-Group middleware system (func(huma.Context, func(huma.Context))).
//
// Why this exists: humachi.New(router, cfg) captures the chi.Router
// reference at construction time. huma.Register then dispatches via
// the captured router's MethodFunc, so any middleware attached to a
// nested chi subrouter (r.Route / r.Group / r.Use) never runs for
// Huma operations. The fix is to apply per-route auth/rate-limit
// middleware at the Huma layer via huma.Group.UseMiddleware. This
// package is the bridge that lets the existing
// `func(http.Handler) http.Handler` middleware (RequireAuth,
// RequireSDKAuth, LimitBy*) plug in unchanged.
//
// Pattern documented at https://huma.rocks/features/middleware/
// under "Unwrapping Router-Specific Objects".
package humawrap

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
)

// Wrap converts a chi-style middleware into a Huma middleware.
//
// Semantics:
//   - The wrapped chi middleware runs against the unwrapped
//     *http.Request and http.ResponseWriter; headers it writes are
//     visible on the wire.
//   - If the chi middleware writes a response and returns without
//     calling next (e.g. RequireAuth on a missing token writes 401),
//     the Huma handler chain short-circuits — `next` is never invoked.
//   - If the chi middleware calls next.ServeHTTP(w, r.WithContext(...)),
//     any context values it added are propagated to the Huma context
//     via huma.WithContext, so downstream handlers reading via
//     ctx.Value (and helpers like httpctx.MustGetUserID) observe them.
//
// The returned function is safe for concurrent use; closures over mw
// are read-only.
func Wrap(mw func(http.Handler) http.Handler) func(huma.Context, func(huma.Context)) {
	return func(hctx huma.Context, next func(huma.Context)) {
		r, w := humachi.Unwrap(hctx)
		// Rebuild r against the accumulated huma context so chi
		// middlewares see values set by earlier humawrap.Wrap entries
		// in the chain. humachi.Unwrap walks every subContext.Unwrap()
		// wrapper back to the bare *chiContext and returns its original
		// *http.Request — stripping any override installed via
		// huma.WithContext. Without this re-base, chained middlewares
		// (e.g. RequireAuth → LimitByUser) see the pre-chain context
		// and the second middleware's next(huma.WithContext(hctx, r2.Context()))
		// clobbers the first's override.
		r = r.WithContext(hctx.Context())
		mw(http.HandlerFunc(func(_ http.ResponseWriter, r2 *http.Request) {
			next(huma.WithContext(hctx, r2.Context()))
		})).ServeHTTP(w, r)
	}
}

// WrapMany applies Wrap to a slice of chi middlewares, preserving
// order. The first element runs outermost — matching the chi r.Use
// order convention so callers can move existing chains across without
// reordering.
func WrapMany(mws ...func(http.Handler) http.Handler) []func(huma.Context, func(huma.Context)) {
	out := make([]func(huma.Context, func(huma.Context)), len(mws))
	for i, mw := range mws {
		out[i] = Wrap(mw)
	}
	return out
}
