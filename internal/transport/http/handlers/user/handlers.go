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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
	"brokle/pkg/utils"
)

type handler struct {
	userSvc    user.UserService
	profileSvc user.ProfileService
	orgSvc     orgDomain.OrganizationService
	logger     *slog.Logger
}

// RegisterRoutes mounts the user-profile routes on r. Expected mount
// context: the authed dashboard chi group (RequireAuth + LimitByUser).
func RegisterRoutes(
	r chi.Router,
	userSvc user.UserService,
	profileSvc user.ProfileService,
	orgSvc orgDomain.OrganizationService,
	logger *slog.Logger,
) {
	h := &handler{userSvc: userSvc, profileSvc: profileSvc, orgSvc: orgSvc, logger: logger}

	r.Route("/api/v1/users/me", func(r chi.Router) {
		r.Get("/", h.getProfile)
		r.Patch("/", h.updateProfile)
		r.Put("/default-organization", h.setDefaultOrganization)
	})
}

// ----- get-user-profile ------------------------------------------------

func (h *handler) getProfile(w http.ResponseWriter, r *http.Request) {
	resp, err := h.buildProfile(r.Context())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// buildProfile is shared between GET and the PATCH echo so the two
// endpoints emit byte-identical response shapes.
func (h *handler) buildProfile(ctx context.Context) (getProfileResponse, error) {
	userID := httpctx.MustGetUserID(ctx)

	u, err := h.userSvc.GetUser(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "user: get-profile user fetch failed",
			"user_id", userID, "error", err)
		return getProfileResponse{}, err
	}

	profile, profileErr := h.profileSvc.GetProfile(ctx, userID)
	if profileErr != nil {
		h.logger.DebugContext(ctx, "user: profile not yet created", "user_id", userID)
		profile = nil
	}

	completeness, completenessErr := h.profileSvc.GetProfileCompleteness(ctx, userID)
	if completenessErr != nil {
		h.logger.WarnContext(ctx, "user: completeness lookup failed",
			"user_id", userID, "error", completenessErr)
	}

	orgsWithProjects, orgErr := h.orgSvc.GetUserOrganizationsWithProjects(ctx, userID)
	if orgErr != nil {
		h.logger.WarnContext(ctx, "user: org hierarchy lookup failed, returning empty",
			"user_id", userID, "error", orgErr)
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

	return resp, nil
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

// ----- update-user-profile ---------------------------------------------

func (h *handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body updateUserProfileBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if body.FirstName != nil || body.LastName != nil {
		if _, err := h.userSvc.UpdateUser(r.Context(), userID, &user.UpdateUserRequest{
			FirstName: body.FirstName,
			LastName:  body.LastName,
		}); err != nil {
			h.logger.WarnContext(r.Context(), "user: update-profile user update failed",
				"user_id", userID, "error", err)
			response.WriteError(w, err)
			return
		}
	}

	if body.Timezone != nil || body.Language != nil {
		if _, err := h.profileSvc.UpdateProfile(r.Context(), userID, &user.UpdateUserProfileRequest{
			Timezone: body.Timezone,
			Language: body.Language,
		}); err != nil {
			h.logger.DebugContext(r.Context(),
				"user: update-profile profile update skipped (profile may not exist yet)",
				"user_id", userID, "error", err)
		}
	}

	resp, err := h.buildProfile(r.Context())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// ----- set-default-organization ----------------------------------------

func (h *handler) setDefaultOrganization(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body setDefaultOrgBody
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

	isMember, err := h.userSvc.ValidateUserOrgMembership(r.Context(), userID, orgID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "user: org membership check failed",
			"user_id", userID, "org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}
	if !isMember {
		h.logger.WarnContext(r.Context(), "user: set-default denied (not a member)",
			"user_id", userID, "org_id", orgID)
		response.WriteError(w, appErrors.NewForbiddenError("You are not a member of this organization"))
		return
	}

	if err := h.userSvc.SetDefaultOrganization(r.Context(), userID, orgID); err != nil {
		h.logger.WarnContext(r.Context(), "user: set-default persist failed",
			"user_id", userID, "org_id", orgID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "user: default organization updated",
		"user_id", userID, "org_id", orgID)

	resp, err := h.buildProfile(r.Context())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}
