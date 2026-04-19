// Package apikey is the dashboard-plane API-key management handler
// domain. Exposes /api/v1/projects/{projectId}/api-keys list /
// create / delete operations. Issued keys carry the bk_<40_chars>
// format; the full key appears only in the create response and
// never again.
package apikey

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

type handler struct {
	apiKeySvc authDomain.APIKeyService
	logger    *slog.Logger
}

// RegisterRoutes registers every API-key operation on apiAdmin.
func RegisterRoutes(api huma.API, apiKeySvc authDomain.APIKeyService, logger *slog.Logger) {
	h := &handler{apiKeySvc: apiKeySvc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "list-api-keys",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/api-keys",
		Tags:        []string{"api-keys"},
		Summary:     "List a project's API keys",
		Description: "Returns keys in preview form (bk_xxxx...yyyy) — the full key is never replayed. Offset-paginated; filter by status=active|expired.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID:   "create-api-key",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/api-keys",
		Tags:          []string{"api-keys"},
		Summary:       "Create an API key for a project",
		Description:   "The full key value appears in the response `key` field ONCE and is never stored in plaintext server-side; subsequent list/get calls only return the preview. Expiry is one of 30days / 90days / never.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-api-key",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/api-keys/{keyId}",
		Tags:          []string{"api-keys"},
		Summary:       "Delete an API key",
		Description:   "Permanently revokes the key — SDK clients using it start receiving 401 on their next request.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)
}

// apiKey is the wire shape returned by list/create.
type apiKey struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key,omitempty" doc:"Full API-key value — populated only on the create response"`
	KeyPreview string     `json:"key_preview"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Status     string     `json:"status" enum:"active,expired"`
	LastUsed   *time.Time `json:"last_used,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedBy  uuid.UUID  `json:"created_by"`
}

// keyStatus returns "expired" when the key's expiry has passed,
// otherwise "active". Soft-deleted keys are filtered by the
// repository layer so this function never sees them.
func keyStatus(k authDomain.APIKey) string {
	if k.IsExpired() {
		return "expired"
	}
	return "active"
}

// ----- list-api-keys --------------------------------------------------

type ListAPIKeysInput struct {
	ProjectID string `path:"projectId" format:"uuid" doc:"Project that owns the API keys"`
	Status    string `query:"status" required:"false" enum:"active,expired" doc:"Optional status filter"`
	Page      int    `query:"page" required:"false" minimum:"1" doc:"Page number, 1-indexed"`
	Limit     int    `query:"limit" required:"false" doc:"Items per page (10, 25, 50, 100)"`
	SortBy    string `query:"sort_by" required:"false" enum:"created_at,name,last_used_at" doc:"Sort field"`
	SortDir   string `query:"sort_dir" required:"false" enum:"asc,desc" doc:"Sort direction"`
}

type ListAPIKeysOutput struct {
	Body listAPIKeysResponse
}

type listAPIKeysResponse struct {
	Data []apiKey          `json:"data"`
	Meta listAPIKeysMeta   `json:"meta"`
}

type listAPIKeysMeta struct {
	Pagination *pagination.Params `json:"pagination,omitempty"`
	Total      int64              `json:"total"`
}

func (h *handler) list(ctx context.Context, in *ListAPIKeysInput) (*ListAPIKeysOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	params := parsePagination(in.Page, in.Limit, in.SortBy, in.SortDir)
	filters := &authDomain.APIKeyFilters{ProjectID: &projectID}
	filters.Params = params

	switch in.Status {
	case "active":
		expired := false
		filters.IsExpired = &expired
	case "expired":
		expired := true
		filters.IsExpired = &expired
	}

	keys, err := h.apiKeySvc.GetAPIKeys(ctx, filters)
	if err != nil {
		h.logger.WarnContext(ctx, "apikey: list failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	total, err := h.apiKeySvc.CountAPIKeys(ctx, filters)
	if err != nil {
		h.logger.WarnContext(ctx, "apikey: count failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
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

	return &ListAPIKeysOutput{
		Body: listAPIKeysResponse{
			Data: out,
			Meta: listAPIKeysMeta{
				Pagination: &params,
				Total:      total,
			},
		},
	}, nil
}

// parsePagination normalises the query-param tuple into a
// pagination.Params, applying defaults and validation so the
// handler doesn't carry the logic inline.
func parsePagination(page, limit int, sortBy, sortDir string) pagination.Params {
	p := pagination.Params{
		Page:    1,
		Limit:   50,
		SortBy:  sortBy,
		SortDir: "desc",
	}
	if page >= 1 {
		p.Page = page
	}
	if pagination.IsValidPageSize(limit) {
		p.Limit = limit
	}
	if sortDir == "asc" || sortDir == "desc" {
		p.SortDir = sortDir
	}
	if err := p.Validate(); err != nil {
		if p.GetOffset() > pagination.MaxOffset {
			p.Page = pagination.MaxOffset / p.Limit
		}
		if p.Page < 1 {
			p.Page = 1
		}
	}
	return p
}

// ----- create-api-key -------------------------------------------------

type CreateAPIKeyInput struct {
	ProjectID string `path:"projectId" format:"uuid" doc:"Project the new key belongs to"`
	Body      createAPIKeyBody
}

type createAPIKeyBody struct {
	Name         string `json:"name" minLength:"2" maxLength:"100" doc:"Human-readable name"`
	ExpiryOption string `json:"expiry_option" enum:"30days,90days,never" doc:"Expiry bucket"`
}

type CreateAPIKeyOutput struct {
	Body apiKey
}

func (h *handler) create(ctx context.Context, in *CreateAPIKeyInput) (*CreateAPIKeyOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	var expiresAt *time.Time
	switch in.Body.ExpiryOption {
	case "30days":
		t := time.Now().Add(30 * 24 * time.Hour)
		expiresAt = &t
	case "90days":
		t := time.Now().Add(90 * 24 * time.Hour)
		expiresAt = &t
	case "never":
		expiresAt = nil
	}

	resp, err := h.apiKeySvc.CreateAPIKey(ctx, userID, &authDomain.CreateAPIKeyRequest{
		Name:      in.Body.Name,
		ProjectID: projectID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "apikey: create failed", "user_id", userID, "project_id", projectID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "apikey: created", "user_id", userID, "project_id", projectID, "api_key_id", resp.ID)

	return &CreateAPIKeyOutput{
		Body: apiKey{
			ID:         resp.ID,
			Name:       resp.Name,
			Key:        resp.Key,
			KeyPreview: resp.KeyPreview,
			ProjectID:  resp.ProjectID,
			Status:     "active",
			CreatedAt:  resp.CreatedAt,
			ExpiresAt:  resp.ExpiresAt,
			CreatedBy:  userID,
		},
	}, nil
}

// ----- delete-api-key -------------------------------------------------

type DeleteAPIKeyInput struct {
	ProjectID string `path:"projectId" format:"uuid" doc:"Project that owns the key"`
	KeyID     string `path:"keyId" format:"uuid" doc:"API key to delete"`
}

type DeleteAPIKeyOutput struct{}

func (h *handler) delete(ctx context.Context, in *DeleteAPIKeyInput) (*DeleteAPIKeyOutput, error) {
	projectID, err := uuid.Parse(in.ProjectID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	keyID, err := uuid.Parse(in.KeyID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid API key ID", "keyId must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.apiKeySvc.DeleteAPIKey(ctx, keyID, projectID); err != nil {
		h.logger.WarnContext(ctx, "apikey: delete failed", "user_id", userID, "project_id", projectID, "api_key_id", keyID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "apikey: deleted", "user_id", userID, "project_id", projectID, "api_key_id", keyID)
	return &DeleteAPIKeyOutput{}, nil
}
