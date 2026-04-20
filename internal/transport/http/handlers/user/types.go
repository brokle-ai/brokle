package user

import (
	"time"

	"github.com/google/uuid"

	"brokle/internal/transport/http/handlers/shared"
)

// Huma operation types for the user package. Input/Output wrappers
// are the operation signatures; body/response shapes carry the JSON
// contracts.

// ----- get-user-profile ---------------------------------------------

type GetUserProfileOutput struct {
	Body getProfileResponse
}

type getProfileResponse struct {
	ID                    uuid.UUID                  `json:"id"`
	Email                 string                     `json:"email"`
	Name                  string                     `json:"name"`
	FirstName             string                     `json:"first_name"`
	LastName              string                     `json:"last_name"`
	AvatarURL             string                     `json:"avatar_url"`
	IsEmailVerified       bool                       `json:"is_email_verified"`
	IsActive              bool                       `json:"is_active"`
	CreatedAt             time.Time                  `json:"created_at"`
	LastLoginAt           *time.Time                 `json:"last_login_at,omitempty"`
	DefaultOrganizationID *uuid.UUID                 `json:"default_organization_id,omitempty"`
	Completeness          int                        `json:"completeness"`
	Profile               *profileData               `json:"profile,omitempty"`
	Organizations         []organizationWithProjects `json:"organizations"`
}

type profileData struct {
	Bio         *string `json:"bio,omitempty"`
	Location    *string `json:"location,omitempty"`
	Website     *string `json:"website,omitempty"`
	TwitterURL  *string `json:"twitter_url,omitempty"`
	LinkedInURL *string `json:"linkedin_url,omitempty"`
	GitHubURL   *string `json:"github_url,omitempty"`
	Timezone    string  `json:"timezone"`
	Language    string  `json:"language"`
	Theme       string  `json:"theme"`
}

type organizationWithProjects struct {
	ID            uuid.UUID        `json:"id"`
	Name          string           `json:"name"`
	CompositeSlug string           `json:"composite_slug"`
	Plan          string           `json:"plan"`
	Role          string           `json:"role"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Projects      []projectSummary `json:"projects"`
}

type projectSummary struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CompositeSlug  string    `json:"composite_slug"`
	Description    *string   `json:"description,omitempty"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ----- update-user-profile ------------------------------------------

type UpdateUserProfileInput struct {
	Body updateUserProfileBody
}

type updateUserProfileBody struct {
	FirstName *string `json:"first_name,omitempty" minLength:"1" maxLength:"100"`
	LastName  *string `json:"last_name,omitempty" minLength:"1" maxLength:"100"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty" minLength:"2" maxLength:"2"`
}

type UpdateUserProfileOutput struct {
	Body getProfileResponse
}

// ----- set-default-organization -------------------------------------

type SetDefaultOrgInput struct {
	Body setDefaultOrgBody
}

type setDefaultOrgBody struct {
	OrganizationID string `json:"organization_id" format:"uuid" doc:"Organization to set as default"`
}

type SetDefaultOrgOutput struct {
	Body shared.MessageResponse
}
