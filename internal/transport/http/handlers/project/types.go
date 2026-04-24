package project

import (
	"time"

	"github.com/google/uuid"

	"brokle/pkg/response"
)

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

// projectListBody — inline {data, pagination} list-response shape.
type projectListBody struct {
	Data       []project            `json:"data"`
	Pagination *response.Pagination `json:"pagination"`
}

// createProjectBody — POST /api/v1/projects.
type createProjectBody struct {
	Name           string `json:"name"                  validate:"required,min=2,max=100"`
	Description    string `json:"description,omitempty" validate:"omitempty,max=500"`
	OrganizationID string `json:"organization_id"       validate:"required"`
}

// updateProjectBody — PUT /api/v1/projects/{projectId}.
type updateProjectBody struct {
	Name        *string `json:"name,omitempty"        validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}
