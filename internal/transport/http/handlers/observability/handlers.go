// Package observability exposes observability operations on three
// auth planes:
//
//   - Dashboard plane (apiAdmin, RequireAuth) — traces browse, spans
//     list/get/delete, scores CRUD, sessions list, filter-preset CRUD.
//     See dashboard.go.
//   - SDK plane (apiPublic, RequireSDKAuth) — span query + filter
//     validation. See sdk.go.
//   - OTLP ingestion (/v1/traces, /v1/logs, /v1/metrics) is HUMA-EXEMPT
//     and mounted as plain chi handlers. See otlp.go.
//
// Public entry points (RegisterRoutes, RegisterSDKRoutes,
// RegisterOTLPChiRoutes) live here; each feature file holds a private
// registerXxxOps helper plus its handler methods.
package observability

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	"brokle/internal/core/domain/observability"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/infrastructure/streams"
)

// ---- handler structs ------------------------------------------------

type dashboardHandler struct {
	traces         *obsServices.TraceService
	scores         *obsServices.ScoreService
	scoreAnalytics *obsServices.ScoreAnalyticsService
	filterPresets  *obsServices.FilterPresetService
	logger         *slog.Logger
}

// sdkHandler exposes SDK-plane (apiPublic, RequireSDKAuth) observability
// operations. The OTLP ingestion endpoints (/v1/traces, /v1/logs, /v1/metrics)
// are NOT registered here — they are HUMA-EXEMPT and mounted as plain
// chi handlers via RegisterOTLPChiRoutes (see otlp.go).
type sdkHandler struct {
	spanQuery *obsServices.SpanQueryService
	logger    *slog.Logger
}

type otlpHandler struct {
	deps OTLPDeps
}

// OTLPDeps bundles all services required by the OTLP ingestion endpoints.
type OTLPDeps struct {
	StreamProducer       *streams.TelemetryStreamProducer
	DeduplicationService observability.TelemetryDeduplicationService
	OTLPConverter        *obsServices.OTLPConverterService
	LogsConverter        *obsServices.OTLPLogsConverterService
	EventsConverter      *obsServices.OTLPEventsConverterService
	MetricsConverter     *obsServices.OTLPMetricsConverterService
	Logger               *slog.Logger
}

// ---- public entry points --------------------------------------------

// RegisterRoutes wires the dashboard-plane observability operations onto
// apiAdmin. All routes require RequireAuth. Routes scoped to a specific
// project carry {projectId}; routes on /traces or /spans expect the
// RequireProjectAccess middleware to pin the project via
// httpctx.WithProjectID (access via query param + middleware).
func RegisterRoutes(
	api huma.API,
	traces *obsServices.TraceService,
	scores *obsServices.ScoreService,
	scoreAnalytics *obsServices.ScoreAnalyticsService,
	filterPresets *obsServices.FilterPresetService,
	logger *slog.Logger,
) {
	h := &dashboardHandler{
		traces:         traces,
		scores:         scores,
		scoreAnalytics: scoreAnalytics,
		filterPresets:  filterPresets,
		logger:         logger,
	}
	registerDashboardOps(api, h)
}

// RegisterSDKRoutes wires the SDK-plane span-query operations onto apiPublic.
// Project ID is derived from the API key via the RequireSDKAuth middleware.
func RegisterSDKRoutes(
	api huma.API,
	spanQuery *obsServices.SpanQueryService,
	logger *slog.Logger,
) {
	h := &sdkHandler{spanQuery: spanQuery, logger: logger}
	registerSDKOps(api, h)
}

// RegisterOTLPChiRoutes mounts the three OTLP HTTP ingestion endpoints as
// plain chi handlers. The caller is responsible for applying
// RequireSDKAuth and rate-limit middleware on the surrounding chi group.
func RegisterOTLPChiRoutes(r chi.Router, deps OTLPDeps) {
	h := &otlpHandler{deps: deps}
	registerOTLPOps(r, h)
}
