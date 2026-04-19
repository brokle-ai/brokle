// Package overview exposes the project-overview dashboard endpoint
// (stats, charts, onboarding status) as a Huma operation on apiAdmin.
package overview

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"brokle/internal/core/domain/analytics"
	"brokle/internal/transport/http/handlers/shared"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	svc    analytics.OverviewService
	logger *slog.Logger
}

// RegisterRoutes registers the overview operation.
func RegisterRoutes(api huma.API, svc analytics.OverviewService, logger *slog.Logger) {
	h := &handler{svc: svc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "get-project-overview",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/overview",
		Tags:        []string{"projects", "overview"},
		Summary:     "Get project overview stats + charts",
		Description: "Time-range scoped. Use `time_range` for presets (15m…30d,all) or `from`/`to` for a custom RFC3339 range. Defaults to 24h when none are supplied.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getOverview)
}

type GetOverviewInput struct {
	ProjectID string `path:"projectId" format:"uuid" doc:"Project to overview"`
	TimeRange string `query:"time_range" required:"false" enum:"15m,30m,1h,3h,6h,12h,24h,7d,14d,30d,all" doc:"Relative time-range preset"`
	From      string `query:"from" required:"false" doc:"Custom range start (RFC3339)"`
	To        string `query:"to" required:"false" doc:"Custom range end (RFC3339)"`
}

type GetOverviewOutput struct {
	Body *analytics.OverviewResponse
}

func (h *handler) getOverview(ctx context.Context, in *GetOverviewInput) (*GetOverviewOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}

	fromTime, toTime, err := shared.ParseTimeRange(in.From, in.To, in.TimeRange, analytics.TimeRange24Hours)
	if err != nil {
		return nil, err
	}

	overview, err := h.svc.GetOverview(ctx, &analytics.OverviewFilter{
		ProjectID: projectID,
		StartTime: fromTime,
		EndTime:   toTime,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "overview: fetch failed", "project_id", projectID, "error", err)
		return nil, appErrors.NewInternalError("Failed to get project overview", err)
	}

	return &GetOverviewOutput{Body: overview}, nil
}
