package project

import (
	"brokle/pkg/pagination"
	"time"

	"github.com/google/uuid"
)

// Huma operation types for the project package.

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

type ListProjectsInput struct {
	OrganizationID string `query:"organization_id" required:"false" format:"uuid" doc:"Optional organization filter"`
	Status         string `query:"status" required:"false" enum:"active,paused,archived" doc:"Filter by project status"`
	Search         string `query:"search" required:"false" doc:"Search by name"`
	Page           int    `query:"page" required:"false" minimum:"1" doc:"Page number, 1-indexed"`
	Limit          int    `query:"limit" required:"false" doc:"Items per page (10, 25, 50, 100)"`
	SortBy         string `query:"sort_by" required:"false" enum:"created_at,name" doc:"Sort field"`
	SortDir        string `query:"sort_dir" required:"false" enum:"asc,desc" doc:"Sort direction"`
}

type ListProjectsOutput struct {
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

type CreateProjectInput struct {
	Body createProjectBody
}

type createProjectBody struct {
	Name           string `json:"name" minLength:"2" maxLength:"100" doc:"Project name"`
	Description    string `json:"description,omitempty" maxLength:"500" doc:"Optional description"`
	OrganizationID string `json:"organization_id" format:"uuid" doc:"Organization the project belongs to"`
}

type CreateProjectOutput struct {
	Body project
}

type GetProjectInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type GetProjectOutput struct {
	Body project
}

type UpdateProjectInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      updateProjectBody
}

type updateProjectBody struct {
	Name        *string `json:"name,omitempty" minLength:"2" maxLength:"100"`
	Description *string `json:"description,omitempty" maxLength:"500"`
}

type UpdateProjectOutput struct {
	Body project
}

type DeleteProjectInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type DeleteProjectOutput struct{}

type ArchiveProjectInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type ArchiveProjectOutput struct{}

type UnarchiveProjectInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type UnarchiveProjectOutput struct{}
