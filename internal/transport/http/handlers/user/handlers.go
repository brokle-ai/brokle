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

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	organizationService "brokle/internal/core/services/organization"
	userService "brokle/internal/core/services/user"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
	"brokle/pkg/utils"
)

type Handler struct {
	userSvc      *userService.UserService
	profileSvc   *userService.ProfileService
	orgSvc       *organizationService.OrganizationService
	projMemberSvc *authService.ProjectMemberService
	logger       *slog.Logger
}

// New constructs a Handler with all required services.
func New(
	userSvc *userService.UserService,
	profileSvc *userService.ProfileService,
	orgSvc *organizationService.OrganizationService,
	projMemberSvc *authService.ProjectMemberService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		userSvc:       userSvc,
		profileSvc:    profileSvc,
		orgSvc:        orgSvc,
		projMemberSvc: projMemberSvc,
		logger:        logger,
	}
}

// ----- get-user-profile ------------------------------------------------

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	resp, err := h.buildProfile(r.Context())
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// buildProfile is shared between GET and the PATCH echo so the two
// endpoints emit byte-identical response shapes.
func (h *Handler) buildProfile(ctx context.Context) (getProfileResponse, error) {
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

	// Resolve effective scopes for every (org, project) the user
	// belongs to. The frontend caches this via React Query and reads
	// it synchronously for every UI permission gate (Langfuse session-
	// bootstrap pattern). On aggregator failure we degrade gracefully:
	// emit empty scope sets so the UI hides guarded controls — same
	// safe-default the previous /scopes/check error path produced.
	topology := buildScopeTopology(orgsWithProjects)
	scopes, scopeErr := h.projMemberSvc.ListUserEffectiveScopes(ctx, userID, topology)
	if scopeErr != nil {
		h.logger.WarnContext(ctx, "user: effective-scope resolution failed, emitting empty scopes",
			"user_id", userID, "error", scopeErr)
		scopes = map[uuid.UUID]*authService.EffectiveOrgScopes{}
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
		Organizations:         mapOrgsWithProjects(orgsWithProjects, scopes),
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

// buildScopeTopology projects the org-tree shape into the (org → []project)
// map the scope aggregator consumes. Keeps the bootstrap path single-pass
// over the source tree.
func buildScopeTopology(src []*orgDomain.OrganizationWithProjectsAndRole) map[uuid.UUID][]uuid.UUID {
	out := make(map[uuid.UUID][]uuid.UUID, len(src))
	for _, o := range src {
		projectIDs := make([]uuid.UUID, 0, len(o.Projects))
		for _, p := range o.Projects {
			projectIDs = append(projectIDs, p.ID)
		}
		out[o.Organization.ID] = projectIDs
	}
	return out
}

// mapOrgsWithProjects builds the response DTO by joining the org tree
// with the resolved scope tree. Projects with a project_members override
// surface the override RoleName as a display hint; otherwise they inherit
// the org RoleName so the frontend has a meaningful display value without
// a second lookup. Effective `Scopes` are additively resolved server-side
// (UNION of org-projection + override grant — see ADR-0001). Always emits
// non-nil Scopes slices (empty list = "no permissions", distinct from
// absent — required by the frontend hooks' `.includes` path which would
// crash on undefined).
//
// Discovery filter: a project is included only when the user can
// discover it — the org-tier `org_projects:list` authority OR the per-
// project `projects:read` floor scope. Mirrors GitHub/GitLab discovery
// semantics (filter-at-source). The frontend project selector renders
// exactly this list, so projects the caller can't navigate to are not
// enumerated. See EffectiveOrgScopes.CanDiscoverProject.
func mapOrgsWithProjects(
	src []*orgDomain.OrganizationWithProjectsAndRole,
	scopes map[uuid.UUID]*authService.EffectiveOrgScopes,
) []organizationWithProjects {
	out := make([]organizationWithProjects, 0, len(src))
	for _, o := range src {
		orgScopeEntry := scopes[o.Organization.ID]
		projects := make([]projectSummary, 0, len(o.Projects))
		for _, p := range o.Projects {
			if !orgScopeEntry.CanDiscoverProject(p.ID) {
				continue
			}
			projectScopes := []string{}
			projectRole := o.RoleName
			if entry, ok := orgScopeEntry.Projects[p.ID]; ok {
				projectScopes = nullSafeScopes(entry.Scopes)
				if entry.RoleName != "" {
					projectRole = entry.RoleName
				}
			}
			projects = append(projects, projectSummary{
				ID:             p.ID,
				Name:           p.Name,
				CompositeSlug:  utils.GenerateCompositeSlug(p.Name, p.ID),
				Description:    p.Description,
				OrganizationID: p.OrganizationID,
				Status:         p.Status,
				Role:           projectRole,
				Scopes:         projectScopes,
				CreatedAt:      p.CreatedAt,
				UpdatedAt:      p.UpdatedAt,
			})
		}
		orgScopes := []string{}
		if orgScopeEntry != nil {
			orgScopes = nullSafeScopes(orgScopeEntry.Scopes)
		}
		out = append(out, organizationWithProjects{
			ID:            o.Organization.ID,
			Name:          o.Organization.Name,
			CompositeSlug: utils.GenerateCompositeSlug(o.Organization.Name, o.Organization.ID),
			Plan:          o.Organization.Plan,
			Role:          o.RoleName,
			Scopes:        orgScopes,
			CreatedAt:     o.Organization.CreatedAt,
			UpdatedAt:     o.Organization.UpdatedAt,
			Projects:      projects,
		})
	}
	return out
}

// nullSafeScopes guarantees a non-nil JSON array on the wire. The
// resolver may return a nil slice for "no permissions" but the
// frontend's `.includes(scope)` path crashes on JSON null.
func nullSafeScopes(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

// ----- update-user-profile ---------------------------------------------

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) SetDefaultOrganization(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body setDefaultOrgBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	orgID, err := uuid.Parse(body.OrganizationID)
	if err != nil {
		response.WriteError(w, appErrors.InvalidParam("organization_id", "must be a valid UUID"))
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
		response.WriteError(w, appErrors.PermissionDenied("organization", "You are not a member of this organization"))
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
