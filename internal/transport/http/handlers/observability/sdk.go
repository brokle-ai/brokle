package observability

import (
	"net/http"
	"time"

	"brokle/internal/core/domain/observability"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// ---- query spans ----------------------------------------------------

func (h *SDKHandler) QuerySpans(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context()).String()

	var body SpanQueryRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	limit := body.Limit
	if limit <= 0 {
		limit = 100
	}
	page := body.Page
	if page <= 0 {
		page = 1
	}

	domainReq := &observability.SpanQueryRequest{
		Filter: body.Filter,
		Limit:  limit,
		Page:   page,
	}
	if body.StartTime != "" {
		ts, err := parseRFC3339(body.StartTime, "start_time")
		if err != nil {
			response.WriteError(w, err)
			return
		}
		domainReq.StartTime = ts
	}
	if body.EndTime != "" {
		ts, err := parseRFC3339(body.EndTime, "end_time")
		if err != nil {
			response.WriteError(w, err)
			return
		}
		domainReq.EndTime = ts
	}

	result, err := h.spanQuery.QuerySpans(r.Context(), projectID, domainReq)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, SpanQueryResponse{
		Data:       result.Spans,
		Pagination: response.BuildPagination(page, limit, result.TotalCount),
	})
}

// ---- validate filter -----------------------------------------------

func (h *SDKHandler) ValidateFilter(w http.ResponseWriter, r *http.Request) {
	// RequireSDKAuth middleware guarantees a project is present.
	_ = httpctx.MustGetProjectID(r.Context())

	var body ValidateFilterRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.spanQuery.ValidateFilter(body.Filter); err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, map[string]any{"valid": true, "message": "Filter expression is valid"})
}

// ---- helpers -------------------------------------------------------

func parseRFC3339(v, field string) (*time.Time, error) {
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, appErrors.InvalidParam(field, "must be an RFC3339 timestamp")
	}
	return &ts, nil
}
