package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	authDomain "brokle/internal/core/domain/auth"
	appErrors "brokle/pkg/errors"
)

// SDK-plane operations for the auth domain. Mounted on apiPublic
// (/v1/openapi.json) rather than apiAdmin — these are the
// endpoints the Brokle JavaScript / Python SDKs call, not the
// dashboard frontend.

// RegisterSDKRoutes registers every SDK-plane auth operation on
// apiPublic. Mount against apiPublic only — these endpoints use
// X-API-Key authentication (or Authorization: Bearer <key>) and
// have no dashboard-cookie semantics.
func RegisterSDKRoutes(api huma.API, apiKeySvc authDomain.APIKeyService, logger *slog.Logger) {
	h := &sdkHandler{apiKeySvc: apiKeySvc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "validate-api-key",
		Method:      http.MethodPost,
		Path:        "/v1/auth/validate-key",
		Tags:        []string{"sdk-auth"},
		Summary:     "Introspect an API key and return its auth context",
		Description: "Public SDK endpoint for key validation. Accepts the key via `X-API-Key` (canonical) or `Authorization: Bearer <key>`. Rate-limited by IP and by hashed key-prefix at the route-group level to prevent brute force.",
	}, h.validateAPIKey)
}

// ----- validate-api-key ---------------------------------------------

func (h *sdkHandler) validateAPIKey(ctx context.Context, in *ValidateAPIKeyInput) (*ValidateAPIKeyOutput, error) {
	apiKey := in.XAPIKey
	if apiKey == "" {
		if v, ok := strings.CutPrefix(in.Authorization, "Bearer "); ok {
			apiKey = v
		}
	}
	if apiKey == "" {
		return nil, appErrors.NewValidationError("API key required", "provide X-API-Key header or Authorization: Bearer <key>")
	}

	result, err := h.apiKeySvc.ValidateAPIKey(ctx, apiKey)
	if err != nil {
		h.logger.WarnContext(ctx, "validate-api-key failed", "error", err)
		return nil, err
	}

	return &ValidateAPIKeyOutput{
		Body: validateAPIKeyResponse{
			AuthContext:    result.AuthContext,
			ProjectID:      result.ProjectID.String(),
			OrganizationID: result.OrganizationID.String(),
		},
	}, nil
}
