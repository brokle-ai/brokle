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

	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/infrastructure/streams"
)

// ---- handler structs ------------------------------------------------

type DashboardHandler struct {
	traces         *obsServices.TraceService
	scores         *obsServices.ScoreService
	scoreAnalytics *obsServices.ScoreAnalyticsService
	filterPresets  *obsServices.FilterPresetService
	logger         *slog.Logger
}

// NewDashboard constructs a DashboardHandler.
func NewDashboard(
	traces *obsServices.TraceService,
	scores *obsServices.ScoreService,
	scoreAnalytics *obsServices.ScoreAnalyticsService,
	filterPresets *obsServices.FilterPresetService,
	logger *slog.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		traces:         traces,
		scores:         scores,
		scoreAnalytics: scoreAnalytics,
		filterPresets:  filterPresets,
		logger:         logger,
	}
}

type SDKHandler struct {
	spanQuery *obsServices.SpanQueryService
	logger    *slog.Logger
}

// NewSDK constructs an SDKHandler.
func NewSDK(spanQuery *obsServices.SpanQueryService, logger *slog.Logger) *SDKHandler {
	return &SDKHandler{spanQuery: spanQuery, logger: logger}
}

type OTLPHandler struct {
	deps OTLPDeps
}

// NewOTLP constructs an OTLPHandler from a deps bundle.
func NewOTLP(deps OTLPDeps) *OTLPHandler {
	return &OTLPHandler{deps: deps}
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
