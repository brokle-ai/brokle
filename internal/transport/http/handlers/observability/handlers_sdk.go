package observability

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"brokle/internal/core/domain/observability"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

// sdkHandler exposes SDK-plane (apiPublic, RequireSDKAuth) observability
// operations. The OTLP ingestion endpoints (/v1/traces, /v1/logs, /v1/metrics)
// are NOT registered here — they are HUMA-EXEMPT and mounted as plain
// chi handlers via RegisterOTLPChiRoutes (see otlp.go).
type sdkHandler struct {
	spanQuery *obsServices.SpanQueryService
	logger    *slog.Logger
}

// RegisterSDKRoutes wires the SDK-plane span-query operations onto apiPublic.
// Project ID is derived from the API key via the RequireSDKAuth middleware.
func RegisterSDKRoutes(
	api huma.API,
	spanQuery *obsServices.SpanQueryService,
	logger *slog.Logger,
) {
	h := &sdkHandler{spanQuery: spanQuery, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "sdk-query-spans",
		Method:      http.MethodPost,
		Path:        "/v1/spans/query",
		Tags:        []string{"SDK - span-query"},
		Summary:     "Query spans using filter expressions",
		Description: "Query production telemetry data using human-readable filter syntax. " +
			"Supports operators: =, !=, >, <, >=, <=, CONTAINS, IN, EXISTS. " +
			"Logical operators AND, OR with parentheses grouping.",
		Security: []map[string][]string{{"apiKey": {}}},
	}, h.querySpans)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-validate-span-filter",
		Method:      http.MethodPost,
		Path:        "/v1/spans/query/validate",
		Tags:        []string{"SDK - span-query"},
		Summary:     "Validate a filter expression",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.validateFilter)
}

// ---- query spans ---------------------------------------------------

type SpanQueryRequest struct {
	Filter    string `json:"filter" minLength:"1" maxLength:"2000" doc:"Filter expression, e.g. service.name=chatbot AND gen_ai.system=openai"`
	StartTime string `json:"start_time,omitempty" doc:"RFC3339 timestamp"`
	EndTime   string `json:"end_time,omitempty" doc:"RFC3339 timestamp"`
	Limit     int    `json:"limit,omitempty" minimum:"0" maximum:"10000"`
	Page      int    `json:"page,omitempty" minimum:"0"`
}

type QuerySpansInput struct {
	Body SpanQueryRequest
}

type QuerySpansOutput struct {
	Body SpanQueryResponse
}

type SpanQueryResponse struct {
	Spans      []*observability.Span `json:"spans"`
	TotalCount int64                 `json:"total_count"`
	HasMore    bool                  `json:"has_more"`
}

func (h *sdkHandler) querySpans(ctx context.Context, in *QuerySpansInput) (*QuerySpansOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx).String()

	limit := in.Body.Limit
	if limit <= 0 {
		limit = 100
	}
	page := in.Body.Page
	if page <= 0 {
		page = 1
	}

	domainReq := &observability.SpanQueryRequest{
		Filter: in.Body.Filter,
		Limit:  limit,
		Page:   page,
	}
	if in.Body.StartTime != "" {
		ts, err := parseRFC3339(in.Body.StartTime, "start_time")
		if err != nil {
			return nil, err
		}
		domainReq.StartTime = ts
	}
	if in.Body.EndTime != "" {
		ts, err := parseRFC3339(in.Body.EndTime, "end_time")
		if err != nil {
			return nil, err
		}
		domainReq.EndTime = ts
	}

	result, err := h.spanQuery.QuerySpans(ctx, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &QuerySpansOutput{Body: SpanQueryResponse{
		Spans:      result.Spans,
		TotalCount: result.TotalCount,
		HasMore:    result.HasMore,
	}}, nil
}

// ---- validate filter -----------------------------------------------

type ValidateFilterRequest struct {
	Filter string `json:"filter" minLength:"1" maxLength:"2000"`
}

type ValidateFilterInput struct {
	Body ValidateFilterRequest
}

type ValidateFilterOutput struct {
	Body struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message"`
	}
}

func (h *sdkHandler) validateFilter(ctx context.Context, in *ValidateFilterInput) (*ValidateFilterOutput, error) {
	// RequireSDKAuth middleware guarantees a project is present; use the
	// Must* accessor so a misconfiguration panics loudly.
	_ = httpctx.MustGetProjectID(ctx)

	if err := h.spanQuery.ValidateFilter(in.Body.Filter); err != nil {
		return nil, err
	}
	out := &ValidateFilterOutput{}
	out.Body.Valid = true
	out.Body.Message = "Filter expression is valid"
	return out, nil
}

// ---- helpers -------------------------------------------------------

func parseRFC3339(v, field string) (*time.Time, error) {
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid "+field, field+" must be an RFC3339 timestamp")
	}
	return &ts, nil
}
