// Package apikey is the dashboard-plane API-key management handler
// domain. Exposes /api/v1/projects/{projectId}/api-keys list /
// create / delete operations. Issued keys carry the bk_<40_chars>
// format; the full key appears only in the create response and
// never again.
package apikey

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type handler struct {
	apiKeySvc authDomain.APIKeyService
	logger    *slog.Logger
}

// RegisterRoutes mounts API-key routes on r. Expected mount context:
// the authed dashboard chi group (RequireAuth + LimitByUser).
func RegisterRoutes(r chi.Router, apiKeySvc authDomain.APIKeyService, logger *slog.Logger) {
	h := &handler{apiKeySvc: apiKeySvc, logger: logger}

	r.Route("/api/v1/projects/{projectId}/api-keys", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Delete("/{keyId}", h.delete)
	})
}

// apiKey is the wire shape returned by list/create.
type apiKey struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty"`
	KeyPreview string     `json:"key_preview"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Status     string     `json:"status"`
	LastUsed   *time.Time `json:"last_used,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedBy  uuid.UUID  `json:"created_by"`
}

type listAPIKeysResponse struct {
	Data       []apiKey             `json:"data"`
	Pagination *response.Pagination `json:"pagination"`
}

func keyStatus(k authDomain.APIKey) string {
	if k.IsExpired() {
		return "expired"
	}
	return "active"
}

// list returns a paginated view of the project's API keys. Status
// filter (active/expired) is optional.
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	page, limit, err := request.QueryPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	params := pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  r.URL.Query().Get("sort_by"),
		SortDir: r.URL.Query().Get("sort_dir"),
	}
	if params.SortDir != "asc" && params.SortDir != "desc" {
		params.SortDir = "desc"
	}

	filters := &authDomain.APIKeyFilters{ProjectID: &projectID}
	filters.Params = params

	switch r.URL.Query().Get("status") {
	case "active":
		expired := false
		filters.IsExpired = &expired
	case "expired":
		expired := true
		filters.IsExpired = &expired
	}

	keys, err := h.apiKeySvc.GetAPIKeys(r.Context(), filters)
	if err != nil {
		h.logger.WarnContext(r.Context(), "apikey: list failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	total, err := h.apiKeySvc.CountAPIKeys(r.Context(), filters)
	if err != nil {
		h.logger.WarnContext(r.Context(), "apikey: count failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	out := make([]apiKey, len(keys))
	for i, k := range keys {
		out[i] = apiKey{
			ID:         k.ID,
			Name:       k.Name,
			KeyPreview: k.KeyPreview,
			ProjectID:  k.ProjectID,
			Status:     keyStatus(*k),
			LastUsed:   k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
			ExpiresAt:  k.ExpiresAt,
			CreatedBy:  k.UserID,
		}
	}

	response.Success(w, listAPIKeysResponse{
		Data:       out,
		Pagination: response.BuildPagination(params.Page, params.Limit, total),
	})
}

// create mints a new API key. The response includes the full key
// value ONCE — it is never replayed on subsequent reads.
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body createAPIKeyBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	var expiresAt *time.Time
	switch body.ExpiryOption {
	case "30days":
		t := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = &t
	case "90days":
		t := time.Now().Add(90 * 24 * time.Hour)
		expiresAt = &t
	case "never":
		expiresAt = nil
	}

	resp, err := h.apiKeySvc.CreateAPIKey(r.Context(), userID, &authDomain.CreateAPIKeyRequest{
		Name:      body.Name,
		ProjectID: projectID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "apikey: create failed",
			"user_id", userID, "project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "apikey: created",
		"user_id", userID, "project_id", projectID, "api_key_id", resp.ID)

	response.Created(w, apiKey{
		ID:         resp.ID,
		Name:       resp.Name,
		Key:        resp.Key,
		KeyPreview: resp.KeyPreview,
		ProjectID:  resp.ProjectID,
		Status:     "active",
		CreatedAt:  resp.CreatedAt,
		ExpiresAt:  resp.ExpiresAt,
		CreatedBy:  userID,
	})
}

// delete revokes an API key. SDK clients using it start receiving
// 401 on their next request.
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	keyID, err := request.URLParamUUID(r, "keyId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.apiKeySvc.DeleteAPIKey(r.Context(), keyID, projectID); err != nil {
		h.logger.WarnContext(r.Context(), "apikey: delete failed",
			"user_id", userID, "project_id", projectID,
			"api_key_id", keyID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "apikey: deleted",
		"user_id", userID, "project_id", projectID, "api_key_id", keyID)
	response.NoContent(w)
}
