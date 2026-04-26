// Package overview exposes the project-overview dashboard endpoint
// (stats, charts, onboarding status). Dashboard plane, RequireAuth.
package overview

import (
	"log/slog"
	"net/http"

	"brokle/internal/core/domain/analytics"
	analyticsService "brokle/internal/core/services/analytics"
	"brokle/internal/transport/http/handlers/shared"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
)

type Handler struct {
	svc    *analyticsService.OverviewService
	logger *slog.Logger
}

// New constructs a Handler with all required services.
func New(svc *analyticsService.OverviewService, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	q := r.URL.Query()
	fromTime, toTime, err := shared.ParseTimeRange(
		q.Get("from"), q.Get("to"), q.Get("time_range"),
		analytics.TimeRange24Hours,
	)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	overview, err := h.svc.GetOverview(r.Context(), &analytics.OverviewFilter{
		ProjectID: projectID,
		StartTime: fromTime,
		EndTime:   toTime,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "overview: fetch failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, appErrors.NewInternalError("Failed to get project overview", err))
		return
	}

	response.Success(w, overview)
}
