// Package humax is the humatest harness Brokle tests use for
// Huma-operation unit tests. It hides two pieces of boilerplate
// every handler-level test would otherwise duplicate:
//
//  1. Installing the `huma.NewError` override so validation and
//     pipeline errors emit the canonical APIResponse envelope rather
//     than Huma's default RFC 9457 problem+json shape. Handler tests
//     don't import internal/server (the production installer), so we
//     re-install here.
//  2. Producing a `humatest.TestAPI` with a sensible default config —
//     matching the OpenAPI metadata served by internal/server is not
//     the test's concern.
//
// Usage (inside a `<domain>_test.go`):
//
//	func newTestAPI(t *testing.T, /* fake services */) humatest.TestAPI {
//	    t.Helper()
//	    api := humax.NewAPI(t)
//	    handler.RegisterRoutes(api, /* fakes */, slog.Default())
//	    return api
//	}
//
// NewAPI is deliberately thin — it returns only the `humatest.TestAPI`
// (not a tuple) so each domain's test file owns its own service-fake
// wiring instead of stuffing everything through a shared builder. See
// `internal/transport/http/handlers/credentials/handlers_test.go` for
// the reference shape.
package humax

import (
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	"brokle/pkg/response"
)

// NewAPI returns a humatest.TestAPI with the APIResponse error
// envelope override installed. Idempotent — calling it from many test
// packages in the same process is safe (response.InstallHumaErrorFactory
// guards with sync.Once).
func NewAPI(t *testing.T) humatest.TestAPI {
	t.Helper()
	response.InstallHumaErrorFactory()
	_, api := humatest.New(t, huma.DefaultConfig("brokle-test", "0.0.0"))
	return api
}
