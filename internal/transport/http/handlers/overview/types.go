package overview

import (
	"brokle/internal/core/domain/analytics"
)

// Huma operation types for the overview package.

type GetOverviewInput struct {
	ProjectID string `path:"projectId" format:"uuid" doc:"Project to overview"`
	TimeRange string `query:"time_range" required:"false" enum:"15m,30m,1h,3h,6h,12h,24h,7d,14d,30d,all" doc:"Relative time-range preset"`
	From      string `query:"from" required:"false" doc:"Custom range start (RFC3339)"`
	To        string `query:"to" required:"false" doc:"Custom range end (RFC3339)"`
}

type GetOverviewOutput struct {
	Body *analytics.OverviewResponse
}
