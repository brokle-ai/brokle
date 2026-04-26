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

// createProjectBody — POST /api/v1/organizations/{orgId}/projects.
//
// Tenancy lives in the URL path, not the body — Stripe / GitHub /
// PostHog convention. NEVER add `organization_id` here. The orgID is
// pinned into ctx by RequireOrganizationAccess upstream and read by
// the handler via httpctx.MustGetOrganizationID. pkg/request.DecodeJSON
// enforces DisallowUnknownFields, so any client sending
// `organization_id` in the body gets a 422.
type createProjectBody struct {
	Name        string `json:"name"                  validate:"required,min=2,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// updateProjectBody — PUT /api/v1/projects/{projectId}.
type updateProjectBody struct {
	Name        *string `json:"name,omitempty"        validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}
