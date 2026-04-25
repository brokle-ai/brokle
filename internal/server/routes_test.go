package server

import (
	"testing"

	"github.com/go-chi/chi/v5"

	"brokle/internal/config"
	observabilityService "brokle/internal/core/services/observability"
)

// TestAddRoutes_NoMountCollision is a boot-time regression guard.
//
// On 2026-04-24, the server panicked at boot with
//
//	panic: chi: attempting to Mount() a handler on an existing path, '/api/v1/auth'
//
// Root cause: the auth handler's RegisterPublicRoutes and
// RegisterProtectedRoutes each called r.Route("/api/v1/auth", ...) on
// sibling chi.Group sub-routers that share their parent mux's routing
// tree. chi's Mount check panics when the same exact pattern is
// Mounted twice. A second latent collision existed at
// /api/v1/projects/{projectId} between evaluation and observability.
//
// Neither showed up in CI because no test previously exercised the
// full route tree. This file is the chi-era replacement for that gap.
//
// This test boots the whole tree against typed-nil service deps and
// asserts addRoutes does not panic. Registration code only stashes
// services on handler structs, it never calls them — so the nils are
// safe at registration time. If a future handler dereferences a
// service field while registering, this test catches that bug too.
//
// Any new duplicate r.Route prefix or any registration-time NPE will
// panic here before it ever ships.
func TestAddRoutes_NoMountCollision(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("addRoutes panicked during route registration: %v", r)
		}
	}()

	mux := chi.NewRouter()
	deps := Deps{
		Config: &config.Config{
			Server: config.ServerConfig{
				CORSAllowedOrigins: []string{"http://localhost:3000"},
				CORSAllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				CORSAllowedHeaders: []string{"Content-Type", "Authorization"},
			},
		},
		// Observability registry must be non-nil because the OTLP
		// handler registration reads (but does not call) several
		// service fields from it. An empty ServiceRegistry leaves those
		// fields as nil interfaces — safe at registration time because
		// the handler only stashes them.
		Observability: &observabilityService.ServiceRegistry{},
		// All other service and infrastructure fields left at their
		// zero values. addRoutes + installGlobalMiddleware never
		// dereference services at registration time — they only
		// capture pointers into handler structs and construct
		// middleware closures.
	}

	installGlobalMiddleware(mux, deps)
	addRoutes(mux, deps)
}
