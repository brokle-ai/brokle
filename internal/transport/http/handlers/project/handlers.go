// Package project is the dashboard-plane project handler domain.
// Exposes /api/v1/projects CRUD + archive/unarchive. Dashboard plane.
//
// Authorisation: RequireAuth guards every operation; membership in
// the project's organisation is verified per-op via the organization
// member service (list/create) or ProjectService.ValidateProjectAccess
// (get/update/delete/archive/unarchive — the service checks membership
// through the project's org).
package project

import (
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"brokle/internal/core/domain/organization"
	organizationService "brokle/internal/core/services/organization"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type handler struct {
	projectSvc *organizationService.ProjectService
	orgSvc     *organizationService.OrganizationService
	memberSvc  *organizationService.MemberService
	logger     *slog.Logger
}

// RegisterRoutes mounts the project routes on r. Expected mount
// context: the authed dashboard chi group (RequireAuth + LimitByUser).
func RegisterRoutes(
	r chi.Router,
	projectSvc *organizationService.ProjectService,
	orgSvc *organizationService.OrganizationService,
	memberSvc *organizationService.MemberService,
	logger *slog.Logger,
) {
	h := &handler{
		projectSvc: projectSvc,
		orgSvc:     orgSvc,
		memberSvc:  memberSvc,
		logger:     logger,
	}

	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{projectId}", h.get)
		r.Put("/{projectId}", h.update)
		r.Delete("/{projectId}", h.delete)
		r.Post("/{projectId}/archive", h.archive)
		r.Post("/{projectId}/unarchive", h.unarchive)
	})
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

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

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

	var projects []*organization.Project
	orgIDStr := q.Get("organization_id")
	status := q.Get("status")
	search := q.Get("search")

	if orgIDStr != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			response.WriteError(w, appErrors.NewValidationError(
				"Invalid organization ID",
				"organization_id must be a valid UUID",
				appErrors.WithParam("organization_id"),
			))
			return
		}

		isMember, err := h.memberSvc.IsMember(r.Context(), userID, orgID)
		if err != nil {
			h.logger.WarnContext(r.Context(), "project: membership check failed",
				"user_id", userID, "org_id", orgID, "error", err)
			response.WriteError(w, err)
			return
		}
		if !isMember {
			response.WriteError(w, appErrors.NewForbiddenError(
				"You don't have access to this organization"))
			return
		}

		projects, err = h.projectSvc.GetProjectsByOrganization(r.Context(), orgID)
		if err != nil {
			h.logger.WarnContext(r.Context(), "project: list by org failed",
				"user_id", userID, "org_id", orgID, "error", err)
			response.WriteError(w, err)
			return
		}
	} else {
		userOrgs, err := h.orgSvc.GetUserOrganizations(r.Context(), userID)
		if err != nil {
			h.logger.WarnContext(r.Context(), "project: user-orgs lookup failed",
				"user_id", userID, "error", err)
			response.WriteError(w, err)
			return
		}
		for _, org := range userOrgs {
			orgProjects, err := h.projectSvc.GetProjectsByOrganization(r.Context(), org.ID)
			if err != nil {
				h.logger.WarnContext(r.Context(),
					"project: skip org due to projects lookup failure",
					"user_id", userID, "org_id", org.ID, "error", err)
				continue
			}
			projects = append(projects, orgProjects...)
		}
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

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body createProjectBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	orgID, err := uuid.Parse(body.OrganizationID)
	if err != nil {
		response.WriteError(w, appErrors.NewValidationError(
			"Invalid organization ID",
			"organization_id must be a valid UUID",
			appErrors.WithParam("organization_id"),
		))
		return
	}

	isMember, err := h.memberSvc.IsMember(r.Context(), userID, orgID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "project: membership check failed",
			"user_id", userID, "org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}
	if !isMember {
		response.WriteError(w, appErrors.NewForbiddenError(
			"You don't have permission to create projects in this organization"))
		return
	}

	p, err := h.projectSvc.CreateProject(r.Context(), orgID, &organization.CreateProjectRequest{
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

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) archive(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) unarchive(w http.ResponseWriter, r *http.Request) {
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
