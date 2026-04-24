package auth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
)

// SDK-plane operations for the auth domain. Mounted on the SDK-public
// chi group rather than the authed surface — validate-key is how
// SDKs bootstrap their credentials before they can authenticate to
// anything else.

// RegisterSDKRoutes mounts the SDK-plane auth routes on r. Expected
// mount context: the SDK-public chi group (LimitByIP +
// LimitByKeyPrefix). X-API-Key / Authorization: Bearer accepted.
func RegisterSDKRoutes(r chi.Router, apiKeySvc authDomain.APIKeyService, logger *slog.Logger) {
	h := &sdkHandler{apiKeySvc: apiKeySvc, logger: logger}
	r.Post("/v1/auth/validate-key", h.validateAPIKey)
}

// ----- validate-api-key ----------------------------------------------

func (h *sdkHandler) validateAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		if v, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
			apiKey = v
		}
	}
	if apiKey == "" {
		response.WriteError(w, appErrors.NewValidationError(
			"API key required",
			"provide X-API-Key header or Authorization: Bearer <key>",
		))
		return
	}

	result, err := h.apiKeySvc.ValidateAPIKey(r.Context(), apiKey)
	if err != nil {
		h.logger.WarnContext(r.Context(), "validate-api-key failed", "error", err)
		response.WriteError(w, err)
		return
	}

	response.Success(w, validateAPIKeyResponse{
		AuthContext:    result.AuthContext,
		ProjectID:      result.ProjectID.String(),
		OrganizationID: result.OrganizationID.String(),
	})
}
