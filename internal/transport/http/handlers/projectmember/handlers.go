// Package projectmember exposes project-level role overrides.
//
// A project membership is OPTIONAL — by default users access projects
// through their organization role (resolved via Langfuse MAX semantics
// in CheckUserPermissionsInScope). A project_members row is created
// only when an org admin wants to elevate a specific user's role inside
// one project. Removing the row reverts the user to their org-level role.
//
// Mount context: nested under /api/v1/projects/{projectId} with
// RequireProjectAccess upstream — handlers read projectID via
// httpctx.MustGetProjectID.
package projectmember

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	authService "brokle/internal/core/services/auth"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// Handler bundles the project-member service + logger.
type Handler struct {
	svc    *authService.ProjectMemberService
	logger *slog.Logger
}

// New constructs the handler.
func New(svc *authService.ProjectMemberService, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// ----- DTOs --------------------------------------------------------------

type addMemberBody struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type updateMemberRoleBody struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type projectMemberResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	ProjectID uuid.UUID `json:"project_id"`
	RoleID    uuid.UUID `json:"role_id"`
	Status    string    `json:"status"`
	JoinedAt  string    `json:"joined_at"`
}

// ----- Handler methods --------------------------------------------------

// List returns every project_members row for the resolved project — the
// users with explicit role overrides. Org-level-only members are NOT in
// this list (use the org members endpoint for those).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	members, err := h.svc.ListProjectMembers(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	out := make([]projectMemberResponse, 0, len(members))
	for _, m := range members {
		out = append(out, projectMemberResponse{
			UserID:    m.UserID,
			ProjectID: m.ProjectID,
			RoleID:    m.RoleID,
			Status:    m.Status,
			JoinedAt:  m.JoinedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	response.Success(w, out)
}

// Add assigns a project-level role to an existing org member.
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	orgID := httpctx.MustGetOrganizationID(r.Context())

	var body addMemberBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	member, err := h.svc.AddMember(r.Context(), body.UserID, projectID, orgID, body.RoleID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, projectMemberResponse{
		UserID:    member.UserID,
		ProjectID: member.ProjectID,
		RoleID:    member.RoleID,
		Status:    member.Status,
		JoinedAt:  member.JoinedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateRole changes the project-level role of an existing project member.
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body updateMemberRoleBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.svc.UpdateMemberRole(r.Context(), userID, projectID, body.RoleID); err != nil {
		response.WriteError(w, err)
		return
	}

	member, err := h.svc.GetMember(r.Context(), userID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, projectMemberResponse{
		UserID:    member.UserID,
		ProjectID: member.ProjectID,
		RoleID:    member.RoleID,
		Status:    member.Status,
		JoinedAt:  member.JoinedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// Remove deletes the project-level role override; the user reverts to
// their org-level access. Returns 204 on success (DELETE convention).
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.svc.RemoveMember(r.Context(), userID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}
