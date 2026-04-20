package observability

import (
	"brokle/internal/core/domain/observability"
)

// SDK-plane Huma operation types + span-query request/response DTOs.
// Consumed only by sdk.go.

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
