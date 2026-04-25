package observability

import (
	"brokle/internal/core/domain/observability"
)

// SDK-plane request/response DTOs.

type SpanQueryRequest struct {
	Filter    string `json:"filter"               validate:"required,min=1,max=2000"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Limit     int    `json:"limit,omitempty"      validate:"omitempty,min=0,max=10000"`
	Page      int    `json:"page,omitempty"       validate:"omitempty,min=0"`
}

type SpanQueryResponse struct {
	Spans      []*observability.Span `json:"spans"`
	TotalCount int64                 `json:"total_count"`
	HasMore    bool                  `json:"has_more"`
}

type ValidateFilterRequest struct {
	Filter string `json:"filter" validate:"required,min=1,max=2000"`
}
