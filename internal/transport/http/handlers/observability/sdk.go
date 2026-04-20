package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"brokle/internal/core/domain/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

// registerSDKOps wires the SDK-plane span-query operations onto apiPublic.
// Called from RegisterSDKRoutes in handlers.go.
func registerSDKOps(api huma.API, h *sdkHandler) {
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
