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
	"time"

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

// project is the wire shape for list / get / create / update.
type project struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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

type ListInput struct {
	OrganizationID string `query:"organization_id" required:"false" format:"uuid" doc:"Optional organization filter"`
	Status         string `query:"status" required:"false" enum:"active,paused,archived" doc:"Filter by project status"`
	Search         string `query:"search" required:"false" doc:"Search by name"`
	Page           int    `query:"page" required:"false" minimum:"1" doc:"Page number, 1-indexed"`
	Limit          int    `query:"limit" required:"false" doc:"Items per page (10, 25, 50, 100)"`
	SortBy         string `query:"sort_by" required:"false" enum:"created_at,name" doc:"Sort field"`
	SortDir        string `query:"sort_dir" required:"false" enum:"asc,desc" doc:"Sort direction"`
}

type ListOutput struct {
	Body listResponse
}

type listResponse struct {
	Data []project `json:"data"`
	Meta listMeta  `json:"meta"`
}

type listMeta struct {
	Pagination *pagination.Params `json:"pagination,omitempty"`
	Total      int                `json:"total"`
}

func (h *handler) list(ctx context.Context, in *ListInput) (*ListOutput, error) {
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

	return &ListOutput{
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

type CreateInput struct {
	Body createBody
}

type createBody struct {
	Name           string `json:"name" minLength:"2" maxLength:"100" doc:"Project name"`
	Description    string `json:"description,omitempty" maxLength:"500" doc:"Optional description"`
	OrganizationID string `json:"organization_id" format:"uuid" doc:"Organization the project belongs to"`
}

type CreateOutput struct {
	Body project
}

func (h *handler) create(ctx context.Context, in *CreateInput) (*CreateOutput, error) {
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
	return &CreateOutput{Body: toProject(p)}, nil
}

// ----- get ------------------------------------------------------------

type GetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type GetOutput struct {
	Body project
}

func (h *handler) get(ctx context.Context, in *GetInput) (*GetOutput, error) {
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
	return &GetOutput{Body: toProject(p)}, nil
}

// ----- update ---------------------------------------------------------

type UpdateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      updateBody
}

type updateBody struct {
	Name        *string `json:"name,omitempty" minLength:"2" maxLength:"100"`
	Description *string `json:"description,omitempty" maxLength:"500"`
}

type UpdateOutput struct {
	Body project
}

func (h *handler) update(ctx context.Context, in *UpdateInput) (*UpdateOutput, error) {
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
	return &UpdateOutput{Body: toProject(p)}, nil
}

// ----- delete ---------------------------------------------------------

type DeleteInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type DeleteOutput struct{}

func (h *handler) delete(ctx context.Context, in *DeleteInput) (*DeleteOutput, error) {
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
	return &DeleteOutput{}, nil
}

// ----- archive / unarchive -------------------------------------------

type ArchiveInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type ArchiveOutput struct{}

func (h *handler) archive(ctx context.Context, in *ArchiveInput) (*ArchiveOutput, error) {
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
	return &ArchiveOutput{}, nil
}

type UnarchiveInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type UnarchiveOutput struct{}

func (h *handler) unarchive(ctx context.Context, in *UnarchiveInput) (*UnarchiveOutput, error) {
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
	return &UnarchiveOutput{}, nil
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
