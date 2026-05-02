// Package project is the dashboard-plane project handler domain.
//
// Routes (mounted in internal/server/routes.go):
//
//   - GET    /api/v1/organizations/{orgId}/projects   — list   (org-scoped)
//   - POST   /api/v1/organizations/{orgId}/projects   — create (org-scoped)
//   - GET    /api/v1/projects/{projectId}             — get
//   - PUT    /api/v1/projects/{projectId}             — update
//   - DELETE /api/v1/projects/{projectId}             — delete
//   - POST   /api/v1/projects/{projectId}/archive
//   - POST   /api/v1/projects/{projectId}/unarchive
//
// Authorisation:
//   - List/Create are mounted under RequireOrganizationAccess, which
//     pins orgID into ctx after verifying membership. Handlers read
//     orgID via httpctx.MustGetOrganizationID — never from the body or
//     query string. URL is the single source of truth for tenancy
//     (Stripe / GitHub / PostHog convention).
//   - Get/Update/Delete/Archive/Unarchive are mounted under
//     RequireProjectAccess, which validates membership in the
//     project's org and pins both projectID and orgID into ctx.
//
// No handler-level membership re-checks: the middleware enforcement
// is the single layer; the project service uses
// ValidateProjectAccess for the project-scoped ops as a service-side
// invariant guard, not as defence-in-depth duplication.
package project

import (
	"log/slog"
	"net/http"
	"sort"
	"strings"

	organizationService "brokle/internal/core/services/organization"

	"brokle/internal/core/domain/organization"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	projectSvc *organizationService.ProjectService
	logger     *slog.Logger
}

// New constructs the project handler.
func New(projectSvc *organizationService.ProjectService, logger *slog.Logger) *Handler {
	return &Handler{projectSvc: projectSvc, logger: logger}
}

func toProject(p *organization.Project) project {
	return project{
		ID:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		OrganizationID: p.OrganizationID,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// ----- list ------------------------------------------------------------

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orgID := httpctx.MustGetOrganizationID(r.Context())

	// Reject query overrides of URL scope. If a future caller mistakenly
	// sends `?organization_id=` thinking it controls scope, fail loudly
	// with 422 — silent ignore was exactly the bug class that produced
	// the cross-org regression after the URL refactor.
	if r.URL.Query().Has("organization_id") {
		response.WriteError(w, appErrors.InvalidParam("organization_id", "is set by the URL path, not the query string"))
		return
	}

	q := r.URL.Query()
	page, limit, err := request.QueryPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	params := pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}
	if params.SortDir != "asc" && params.SortDir != "desc" {
		params.SortDir = "desc"
	}
	status := q.Get("status")
	search := q.Get("search")

	projects, err := h.projectSvc.GetProjectsByOrganization(r.Context(), orgID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "project: list by org failed",
			"org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}

	filtered := make([]*organization.Project, 0, len(projects))
	for _, p := range projects {
		if status != "" && status != "active" {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(search)) {
			continue
		}
		filtered = append(filtered, p)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if params.SortDir == "asc" {
			return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)
	offset := params.GetOffset()
	end := offset + params.Limit
	if end > total {
		end = total
	}
	pageSlice := []*organization.Project{}
	if offset < total {
		pageSlice = filtered[offset:end]
	}

	out := make([]project, len(pageSlice))
	for i, p := range pageSlice {
		out[i] = toProject(p)
	}

	response.Success(w, projectListBody{
		Data:       out,
		Pagination: response.BuildPagination(params.Page, params.Limit, int64(total)),
	})
}

// ----- create ----------------------------------------------------------

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	orgID := httpctx.MustGetOrganizationID(r.Context())

	// pkg/request.DecodeJSON enforces DisallowUnknownFields. A client
	// sending `organization_id` in the body gets a 422 here — the
	// cross-tenant write attack class is structurally impossible
	// because createProjectBody has no tenancy field at all.
	var body createProjectBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	p, err := h.projectSvc.CreateProject(r.Context(), orgID, userID, &organization.CreateProjectRequest{
		Name:        body.Name,
		Description: body.Description,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "project: create failed",
			"user_id", userID, "org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "project: created",
		"user_id", userID, "org_id", orgID, "project_id", p.ID)
	response.Created(w, toProject(p))
}

// ----- get -------------------------------------------------------------

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.projectSvc.ValidateProjectAccess(r.Context(), userID, projectID); err != nil {
		h.logger.WarnContext(r.Context(), "project: access denied",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	p, err := h.projectSvc.GetProject(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toProject(p))
}

// ----- update ----------------------------------------------------------

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.projectSvc.ValidateProjectAccess(r.Context(), userID, projectID); err != nil {
		h.logger.WarnContext(r.Context(), "project: update access denied",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	var body updateProjectBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.projectSvc.UpdateProject(r.Context(), projectID, &organization.UpdateProjectRequest{
		Name:        body.Name,
		Description: body.Description,
	}); err != nil {
		h.logger.WarnContext(r.Context(), "project: update failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	p, err := h.projectSvc.GetProject(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "project: updated",
		"user_id", userID, "project_id", projectID)
	response.Success(w, toProject(p))
}

// ----- delete ----------------------------------------------------------

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.projectSvc.ValidateProjectAccess(r.Context(), userID, projectID); err != nil {
		h.logger.WarnContext(r.Context(), "project: delete access denied",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	if err := h.projectSvc.DeleteProject(r.Context(), projectID); err != nil {
		h.logger.WarnContext(r.Context(), "project: delete failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "project: deleted",
		"user_id", userID, "project_id", projectID)
	response.NoContent(w)
}

// ----- archive / unarchive --------------------------------------------

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.projectSvc.ValidateProjectAccess(r.Context(), userID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.projectSvc.ArchiveProject(r.Context(), projectID); err != nil {
		h.logger.WarnContext(r.Context(), "project: archive failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "project: archived",
		"user_id", userID, "project_id", projectID)
	response.NoContent(w)
}

func (h *Handler) Unarchive(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.projectSvc.ValidateProjectAccess(r.Context(), userID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.projectSvc.UnarchiveProject(r.Context(), projectID); err != nil {
		h.logger.WarnContext(r.Context(), "project: unarchive failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "project: unarchived",
		"user_id", userID, "project_id", projectID)
	response.NoContent(w)
}
