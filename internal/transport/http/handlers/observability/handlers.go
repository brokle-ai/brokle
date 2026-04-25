// Package observability exposes observability operations on three
// surfaces:
//
//   - Dashboard plane (RequireAuth) — traces browse, spans
//     list/get/delete, scores CRUD, sessions list, filter-preset CRUD.
//     See dashboard.go.
//   - SDK plane (RequireSDKAuth) — span query + filter validation.
//     See sdk.go.
//   - OTLP ingestion (/v1/traces, /v1/logs, /v1/metrics) — plain chi
//     handlers that accept raw protobuf bodies. See otlp.go.
package observability

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

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
	DeduplicationService *obsServices.TelemetryDeduplicationService
	OTLPConverter        *obsServices.OTLPConverterService
	LogsConverter        *obsServices.OTLPLogsConverterService
	EventsConverter      *obsServices.OTLPEventsConverterService
	MetricsConverter     *obsServices.OTLPMetricsConverterService
	Logger               *slog.Logger
}

// ---- public entry points --------------------------------------------

// RegisterRoutes mounts the dashboard-plane observability routes on r.
// Expected mount context: the authed dashboard chi group.
func RegisterRoutes(
	r chi.Router,
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
	registerDashboardOps(r, h)
}

// RegisterSDKRoutes mounts the SDK-plane observability routes on r.
// Expected mount context: the SDK-authed chi group.
func RegisterSDKRoutes(
	r chi.Router,
	spanQuery *obsServices.SpanQueryService,
	logger *slog.Logger,
) {
	h := &sdkHandler{spanQuery: spanQuery, logger: logger}
	registerSDKOps(r, h)
}

// RegisterOTLPRoutes mounts the three OTLP HTTP ingestion endpoints
// as plain chi handlers. The caller is responsible for applying
// RequireSDKAuth and rate-limit middleware on the surrounding chi group.
func RegisterOTLPRoutes(r chi.Router, deps OTLPDeps) {
	h := &otlpHandler{deps: deps}
	registerOTLPOps(r, h)
}
