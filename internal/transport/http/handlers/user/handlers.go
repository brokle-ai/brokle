// Package user is the dashboard-plane user-profile handler domain.
// Exposes /api/v1/users/me operations (get enriched profile, patch
// profile fields, set default organization).
//
// Distinct from the auth domain's /api/v1/auth/profile + /me — auth
// returns a session-aware shape (expiry metadata, change-password
// hooks), user returns an org-hierarchy-aware shape (the dashboard's
// sidebar navigation reads this on every page load).
package user

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/utils"
)

type handler struct {
	userSvc    user.UserService
	profileSvc user.ProfileService
	orgSvc     orgDomain.OrganizationService
	logger     *slog.Logger
}

// RegisterRoutes registers every user operation on apiAdmin.
func RegisterRoutes(
	api huma.API,
	userSvc user.UserService,
	profileSvc user.ProfileService,
	orgSvc orgDomain.OrganizationService,
	logger *slog.Logger,
) {
	h := &handler{userSvc: userSvc, profileSvc: profileSvc, orgSvc: orgSvc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "get-user-profile",
		Method:      http.MethodGet,
		Path:        "/api/v1/users/me",
		Tags:        []string{"user"},
		Summary:     "Get the authenticated user's profile + organization hierarchy",
		Description: "Returns user core fields, extended profile, profile-completeness score, and organizations + nested projects. Dashboard reads this on every page load for the sidebar + org switcher.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getProfile)

	huma.Register(api, huma.Operation{
		OperationID: "update-user-profile",
		Method:      http.MethodPatch,
		Path:        "/api/v1/users/me",
		Tags:        []string{"user"},
		Summary:     "Update the authenticated user's name and preferences",
		Description: "Partial update. Name fields write to the user record; timezone/language write to the profile record; profile-layer failures are soft-logged and do not abort the user-layer update.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateProfile)

	huma.Register(api, huma.Operation{
		OperationID: "set-default-organization",
		Method:      http.MethodPut,
		Path:        "/api/v1/users/me/default-organization",
		Tags:        []string{"user"},
		Summary:     "Set the authenticated user's default organization",
		Description: "Verifies membership before persisting; 403 when the user is not a member (doesn't leak whether the org exists).",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.setDefaultOrganization)
}

// ----- get-user-profile ---------------------------------------------

type GetProfileOutput struct {
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

func (h *handler) getProfile(ctx context.Context, _ *struct{}) (*GetProfileOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	u, err := h.userSvc.GetUser(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "user: get-profile user fetch failed", "user_id", userID, "error", err)
		return nil, err
	}

	profile, profileErr := h.profileSvc.GetProfile(ctx, userID)
	if profileErr != nil {
		h.logger.DebugContext(ctx, "user: profile not yet created", "user_id", userID)
		profile = nil
	}

	completeness, completenessErr := h.profileSvc.GetProfileCompleteness(ctx, userID)
	if completenessErr != nil {
		h.logger.WarnContext(ctx, "user: completeness lookup failed", "user_id", userID, "error", completenessErr)
	}

	orgsWithProjects, orgErr := h.orgSvc.GetUserOrganizationsWithProjects(ctx, userID)
	if orgErr != nil {
		h.logger.WarnContext(ctx, "user: org hierarchy lookup failed, returning empty", "user_id", userID, "error", orgErr)
		orgsWithProjects = []*orgDomain.OrganizationWithProjectsAndRole{}
	}

	resp := getProfileResponse{
		ID:                    u.ID,
		Email:                 u.Email,
		Name:                  u.GetFullName(),
		FirstName:             u.FirstName,
		LastName:              u.LastName,
		IsEmailVerified:       u.IsEmailVerified,
		IsActive:              u.IsActive,
		CreatedAt:             u.CreatedAt,
		LastLoginAt:           u.LastLoginAt,
		DefaultOrganizationID: u.DefaultOrganizationID,
		Organizations:         mapOrgsWithProjects(orgsWithProjects),
	}

	if profile != nil {
		resp.Profile = &profileData{
			Bio:         profile.Bio,
			Location:    profile.Location,
			Website:     profile.Website,
			TwitterURL:  profile.TwitterURL,
			LinkedInURL: profile.LinkedInURL,
			GitHubURL:   profile.GitHubURL,
			Timezone:    profile.Timezone,
			Language:    profile.Language,
			Theme:       profile.Theme,
		}
	}
	if completeness != nil {
		resp.Completeness = completeness.OverallScore
	}

	return &GetProfileOutput{Body: resp}, nil
}

func mapOrgsWithProjects(src []*orgDomain.OrganizationWithProjectsAndRole) []organizationWithProjects {
	out := make([]organizationWithProjects, 0, len(src))
	for _, o := range src {
		projects := make([]projectSummary, 0, len(o.Projects))
		for _, p := range o.Projects {
			projects = append(projects, projectSummary{
				ID:             p.ID,
				Name:           p.Name,
				CompositeSlug:  utils.GenerateCompositeSlug(p.Name, p.ID),
				Description:    p.Description,
				OrganizationID: p.OrganizationID,
				Status:         p.Status,
				CreatedAt:      p.CreatedAt,
				UpdatedAt:      p.UpdatedAt,
			})
		}
		out = append(out, organizationWithProjects{
			ID:            o.Organization.ID,
			Name:          o.Organization.Name,
			CompositeSlug: utils.GenerateCompositeSlug(o.Organization.Name, o.Organization.ID),
			Plan:          o.Organization.Plan,
			Role:          o.RoleName,
			CreatedAt:     o.Organization.CreatedAt,
			UpdatedAt:     o.Organization.UpdatedAt,
			Projects:      projects,
		})
	}
	return out
}

// ----- update-user-profile ------------------------------------------

type UpdateProfileInput struct {
	Body updateProfileBody
}

type updateProfileBody struct {
	FirstName *string `json:"first_name,omitempty" minLength:"1" maxLength:"100"`
	LastName  *string `json:"last_name,omitempty" minLength:"1" maxLength:"100"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty" minLength:"2" maxLength:"2"`
}

type UpdateProfileOutput struct {
	Body getProfileResponse
}

func (h *handler) updateProfile(ctx context.Context, in *UpdateProfileInput) (*UpdateProfileOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	if in.Body.FirstName != nil || in.Body.LastName != nil {
		if _, err := h.userSvc.UpdateUser(ctx, userID, &user.UpdateUserRequest{
			FirstName: in.Body.FirstName,
			LastName:  in.Body.LastName,
		}); err != nil {
			h.logger.WarnContext(ctx, "user: update-profile user update failed", "user_id", userID, "error", err)
			return nil, err
		}
	}

	if in.Body.Timezone != nil || in.Body.Language != nil {
		if _, err := h.profileSvc.UpdateProfile(ctx, userID, &user.UpdateProfileRequest{
			Timezone: in.Body.Timezone,
			Language: in.Body.Language,
		}); err != nil {
			h.logger.DebugContext(ctx, "user: update-profile profile update skipped (profile may not exist yet)", "user_id", userID, "error", err)
		}
	}

	out, err := h.getProfile(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &UpdateProfileOutput{Body: out.Body}, nil
}

// ----- set-default-organization -------------------------------------

type SetDefaultOrgInput struct {
	Body setDefaultOrgBody
}

type setDefaultOrgBody struct {
	OrganizationID string `json:"organization_id" format:"uuid" doc:"Organization to set as default"`
}

type SetDefaultOrgOutput struct {
	Body messageResponse
}

type messageResponse struct {
	Message string `json:"message"`
}

func (h *handler) setDefaultOrganization(ctx context.Context, in *SetDefaultOrgInput) (*SetDefaultOrgOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	orgID, err := uuid.Parse(in.Body.OrganizationID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid organization ID", "organization_id must be a valid UUID")
	}

	isMember, err := h.userSvc.ValidateUserOrgMembership(ctx, userID, orgID)
	if err != nil {
		h.logger.WarnContext(ctx, "user: org membership check failed", "user_id", userID, "org_id", orgID, "error", err)
		return nil, err
	}
	if !isMember {
		h.logger.WarnContext(ctx, "user: set-default denied (not a member)", "user_id", userID, "org_id", orgID)
		return nil, appErrors.NewForbiddenError("You are not a member of this organization")
	}

	if err := h.userSvc.SetDefaultOrganization(ctx, userID, orgID); err != nil {
		h.logger.WarnContext(ctx, "user: set-default persist failed", "user_id", userID, "org_id", orgID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "user: default organization updated", "user_id", userID, "org_id", orgID)
	return &SetDefaultOrgOutput{Body: messageResponse{Message: "Default organization updated successfully"}}, nil
}
