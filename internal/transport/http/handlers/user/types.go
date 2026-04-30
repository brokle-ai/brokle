package user

import (
	"time"

	"github.com/google/uuid"
)

// ----- get-user-profile response shape ---------------------------------

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
	// Scopes is the user's resolved org-tier permission set for this
	// organization (Langfuse session-bootstrap pattern). The frontend's
	// useHasOrganizationAccess hook reads this directly — no separate
	// /scopes/check round trip. Always emitted (never omitempty); empty
	// list is a meaningful "no org-tier permissions" signal.
	Scopes        []string         `json:"scopes"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	// Projects is the user's DISCOVERABLE subset of the organization's
	// projects: a project is included iff the user holds either
	// `org_projects:list` on the org (full-list authority) OR
	// `projects:read` on the project (per-project floor scope —
	// auto-injected by RoleService.autoInjectFloorScope when any
	// project-tier permission is granted). Mirrors GitHub/GitLab
	// filter-at-discovery semantics: projects the caller can't
	// navigate to are NOT enumerated, so the frontend project
	// selector renders exactly the navigable set.
	Projects []projectSummary `json:"projects"`
}

type projectSummary struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CompositeSlug  string    `json:"composite_slug"`
	Description    *string   `json:"description,omitempty"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Status         string    `json:"status"`
	// Role is the most-specific project role for display:
	// project_members.role when the user has a per-resource grant,
	// otherwise the inherited org RoleName. Always populated. Display
	// hint only — effective scopes are additively resolved server-side
	// and live in `scopes`.
	Role           string    `json:"role"`
	// Scopes is the user's additively-resolved project-tier permission
	// set for this project (UNION of org-projection + any project_members
	// grant). Empty list = no project-tier permissions; UI hides
	// project-scoped controls. Always emitted.
	Scopes         []string  `json:"scopes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ----- update-user-profile body ----------------------------------------

type updateUserProfileBody struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty"  validate:"omitempty,min=1,max=100"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty"   validate:"omitempty,min=2,max=2"`
}

// ----- set-default-organization body -----------------------------------

type setDefaultOrgBody struct {
	OrganizationID string `json:"organization_id" validate:"required"`
}
