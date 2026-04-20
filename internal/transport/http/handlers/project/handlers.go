// Package project is the dashboard-plane project handler domain.
// Exposes /api/v1/projects CRUD + archive/unarchive on apiAdmin.
//
// Authorisation: RequireAuth guards every operation; membership in
// the project's organisation is verified per-op via the organization
// member service (list/create) or ProjectService.ValidateProjectAccess
// (get/update/delete/archive/unarchive — the service checks membership
// through the project's org).
package project

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"brokle/internal/core/domain/organization"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

type handler struct {
	projectSvc organization.ProjectService
	orgSvc     organization.OrganizationService
	memberSvc  organization.MemberService
	logger     *slog.Logger
}

// RegisterRoutes registers every project operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	projectSvc organization.ProjectService,
	orgSvc organization.OrganizationService,
	memberSvc organization.MemberService,
	logger *slog.Logger,
) {
	h := &handler{
		projectSvc: projectSvc,
		orgSvc:     orgSvc,
		memberSvc:  memberSvc,
		logger:     logger,
	}

	huma.Register(api, huma.Operation{
		OperationID: "list-projects",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects",
		Tags:        []string{"projects"},
		Summary:     "List projects accessible to the authenticated user",
		Description: "Optionally filter by organization, status, or search term. Offset-paginated.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID:   "create-project",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects",
		Tags:          []string{"projects"},
		Summary:       "Create a project in an organization",
		Description:   "Caller must be a member of the target organization.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "get-project",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}",
		Tags:        []string{"projects"},
		Summary:     "Get a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.get)

	huma.Register(api, huma.Operation{
		OperationID: "update-project",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}",
		Tags:        []string{"projects"},
		Summary:     "Update a project (name and description)",
		Description: "Status changes are not accepted here — use the archive / unarchive endpoints.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-project",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}",
		Tags:          []string{"projects"},
		Summary:       "Delete a project",
		Description:   "Soft-deletes the project. Caller must have access to the project's organization.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)

	huma.Register(api, huma.Operation{
		OperationID:   "archive-project",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/archive",
		Tags:          []string{"projects"},
		Summary:       "Archive a project",
		Description:   "Marks the project as archived (read-only, reversible).",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.archive)

	huma.Register(api, huma.Operation{
		OperationID:   "unarchive-project",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/unarchive",
		Tags:          []string{"projects"},
		Summary:       "Unarchive a project",
		Description:   "Restores an archived project to active status.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.unarchive)
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

// ----- list -----------------------------------------------------------

func (h *handler) list(ctx context.Context, in *ListProjectsInput) (*ListProjectsOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	params := parsePagination(in.Page, in.Limit, in.SortBy, in.SortDir)

	var projects []*organization.Project

	if in.OrganizationID != "" {
		orgID, err := uuid.Parse(in.OrganizationID)
		if err != nil {
			return nil, appErrors.NewValidationError("Invalid organization ID", "organization_id must be a valid UUID")
		}

		isMember, err := h.memberSvc.IsMember(ctx, userID, orgID)
		if err != nil {
			h.logger.WarnContext(ctx, "project: membership check failed", "user_id", userID, "org_id", orgID, "error", err)
			return nil, err
		}
		if !isMember {
			return nil, appErrors.NewForbiddenError("You don't have access to this organization")
		}

		projects, err = h.projectSvc.GetProjectsByOrganization(ctx, orgID)
		if err != nil {
			h.logger.WarnContext(ctx, "project: list by org failed", "user_id", userID, "org_id", orgID, "error", err)
			return nil, err
		}
	} else {
		userOrgs, err := h.orgSvc.GetUserOrganizations(ctx, userID)
		if err != nil {
			h.logger.WarnContext(ctx, "project: user-orgs lookup failed", "user_id", userID, "error", err)
			return nil, err
		}
		for _, org := range userOrgs {
			orgProjects, err := h.projectSvc.GetProjectsByOrganization(ctx, org.ID)
			if err != nil {
				h.logger.WarnContext(ctx, "project: skip org due to projects lookup failure", "user_id", userID, "org_id", org.ID, "error", err)
				continue
			}
			projects = append(projects, orgProjects...)
		}
	}

	filtered := make([]*organization.Project, 0, len(projects))
	for _, p := range projects {
		if in.Status != "" && in.Status != "active" {
			// Only active is materially supported until multi-status lands.
			continue
		}
		if in.Search != "" {
			if !strings.Contains(strings.ToLower(p.Name), strings.ToLower(in.Search)) {
				continue
			}
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
	page := []*organization.Project{}
	if offset < total {
		page = filtered[offset:end]
	}

	out := make([]project, len(page))
	for i, p := range page {
		out[i] = toProject(p)
	}

	return &ListProjectsOutput{
		Body: listResponse{
			Data: out,
			Meta: listMeta{
				Pagination: &params,
				Total:      total,
			},
		},
	}, nil
}

// ----- create ----------------------------------------------------------

func (h *handler) create(ctx context.Context, in *CreateProjectInput) (*CreateProjectOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	orgID, err := uuid.Parse(in.Body.OrganizationID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "organization_id must be a valid UUID")
	}

	isMember, err := h.memberSvc.IsMember(ctx, userID, orgID)
	if err != nil {
		h.logger.WarnContext(ctx, "project: membership check failed", "user_id", userID, "org_id", orgID, "error", err)
		return nil, err
	}
	if !isMember {
		return nil, appErrors.NewForbiddenError("You don't have permission to create projects in this organization")
	}

	p, err := h.projectSvc.CreateProject(ctx, orgID, &organization.CreateProjectRequest{
		Name:        in.Body.Name,
		Description: in.Body.Description,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "project: create failed", "user_id", userID, "org_id", orgID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "project: created", "user_id", userID, "org_id", orgID, "project_id", p.ID)
	return &CreateProjectOutput{Body: toProject(p)}, nil
}

// ----- get ------------------------------------------------------------

func (h *handler) get(ctx context.Context, in *GetProjectInput) (*GetProjectOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.projectSvc.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		h.logger.WarnContext(ctx, "project: access denied", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	p, err := h.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &GetProjectOutput{Body: toProject(p)}, nil
}

// ----- update ---------------------------------------------------------

func (h *handler) update(ctx context.Context, in *UpdateProjectInput) (*UpdateProjectOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.projectSvc.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		h.logger.WarnContext(ctx, "project: update access denied", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	if err := h.projectSvc.UpdateProject(ctx, projectID, &organization.UpdateProjectRequest{
		Name:        in.Body.Name,
		Description: in.Body.Description,
	}); err != nil {
		h.logger.WarnContext(ctx, "project: update failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	p, err := h.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	h.logger.InfoContext(ctx, "project: updated", "user_id", userID, "project_id", projectID)
	return &UpdateProjectOutput{Body: toProject(p)}, nil
}

// ----- delete ---------------------------------------------------------

func (h *handler) delete(ctx context.Context, in *DeleteProjectInput) (*DeleteProjectOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.projectSvc.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		h.logger.WarnContext(ctx, "project: delete access denied", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	if err := h.projectSvc.DeleteProject(ctx, projectID); err != nil {
		h.logger.WarnContext(ctx, "project: delete failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "project: deleted", "user_id", userID, "project_id", projectID)
	return &DeleteProjectOutput{}, nil
}

// ----- archive / unarchive -------------------------------------------

func (h *handler) archive(ctx context.Context, in *ArchiveProjectInput) (*ArchiveProjectOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.projectSvc.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	if err := h.projectSvc.ArchiveProject(ctx, projectID); err != nil {
		h.logger.WarnContext(ctx, "project: archive failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "project: archived", "user_id", userID, "project_id", projectID)
	return &ArchiveProjectOutput{}, nil
}

func (h *handler) unarchive(ctx context.Context, in *UnarchiveProjectInput) (*UnarchiveProjectOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.projectSvc.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		return nil, err
	}
	if err := h.projectSvc.UnarchiveProject(ctx, projectID); err != nil {
		h.logger.WarnContext(ctx, "project: unarchive failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "project: unarchived", "user_id", userID, "project_id", projectID)
	return &UnarchiveProjectOutput{}, nil
}

// parsePagination normalises query-param tuple into pagination.Params.
// Mirrors apikey handler's parser; extracted here to avoid a
// cross-domain dependency.
func parsePagination(page, limit int, sortBy, sortDir string) pagination.Params {
	p := pagination.Params{
		Page:    1,
		Limit:   50,
		SortBy:  sortBy,
		SortDir: "desc",
	}
	if page >= 1 {
		p.Page = page
	}
	if pagination.IsValidPageSize(limit) {
		p.Limit = limit
	}
	if sortDir == "asc" || sortDir == "desc" {
		p.SortDir = sortDir
	}
	if err := p.Validate(); err != nil {
		if p.GetOffset() > pagination.MaxOffset {
			p.Page = pagination.MaxOffset / p.Limit
		}
		if p.Page < 1 {
			p.Page = 1
		}
	}
	return p
}
