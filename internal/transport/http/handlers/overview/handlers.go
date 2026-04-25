// Package overview exposes the project-overview dashboard endpoint
// (stats, charts, onboarding status). Dashboard plane, RequireAuth.
package overview

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"brokle/internal/core/domain/analytics"
	analyticsService "brokle/internal/core/services/analytics"
	"brokle/internal/transport/http/handlers/shared"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type handler struct {
	svc    *analyticsService.OverviewService
	logger *slog.Logger
}

// RegisterRoutes mounts the overview route on r. Expected mount
// context: the authed dashboard chi group (RequireAuth + LimitByUser).
func RegisterRoutes(r chi.Router, svc *analyticsService.OverviewService, logger *slog.Logger) {
	h := &handler{svc: svc, logger: logger}
	r.Get("/api/v1/projects/{projectId}/overview", h.getOverview)
}

func (h *handler) getOverview(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

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
