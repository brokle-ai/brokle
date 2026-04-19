// Package response defines the canonical API response envelope used
// by the Huma-based HTTP layer plus a stdlib-http `WriteError` helper
// for chi middleware that rejects a request before it reaches a Huma
// operation. The gin-era helpers (Success, Created, BadRequest, …)
// were removed with the Chi+Huma migration; Huma operations return
// `(*Output, error)` and the envelope is applied centrally via the
// `huma.NewError` override in `internal/server/api_error.go`.
package response

import (
	"encoding/json"
	"net/http"

	appErrors "brokle/pkg/errors"
)

// APIResponse is the on-wire shape of every Brokle HTTP response —
// success responses carry Data + optional Meta; error responses carry
// Error. One envelope keeps the SDK decoder simple.
type APIResponse struct {
	Data    any       `json:"data,omitempty" description:"Response data payload"`
	Error   *APIError `json:"error,omitempty" description:"Error information if request failed"`
	Meta    *Meta     `json:"meta,omitempty" description:"Response metadata"`
	Success bool      `json:"success" example:"true" description:"Indicates if the request was successful"`
}

// APIError mirrors the Stripe / OpenAI / Anthropic shape: Type is the
// closed coarse classification (retry/alert key), Code is the open
// fine-grained identifier (per-error SDK subclass), Param optionally
// points at the offending input field.
type APIError struct {
	Type    string `json:"type" example:"validation_error" description:"Closed coarse error classification"`
	Code    string `json:"code,omitempty" example:"project_not_found" description:"Open fine-grained domain code (snake_case)"`
	Message string `json:"message" example:"Invalid request data" description:"Human-readable error message"`
	Details string `json:"details,omitempty" example:"projectId must be a valid UUID" description:"Additional error context"`
	Param   string `json:"param,omitempty" example:"projectId" description:"Input field that triggered the error, when applicable"`
}

// Pagination is the offset-paginated list metadata published inside
// Meta.Pagination on list-response envelopes.
type Pagination struct {
	Page       int   `json:"page" example:"1" description:"Current page number (1-indexed)"`
	Limit      int   `json:"limit" example:"50" description:"Items per page"`
	Total      int64 `json:"total" example:"1234" description:"Total number of items"`
	TotalPages int   `json:"total_pages" example:"25" description:"Total number of pages"`
	HasNext    bool  `json:"has_next" example:"true" description:"Whether there are more pages"`
	HasPrev    bool  `json:"has_prev" example:"false" description:"Whether there are previous pages"`
}

// Meta is the envelope's response-metadata slot.
type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty" description:"Offset pagination information for list responses"`
	RequestID  string      `json:"request_id,omitempty" example:"req_01h2x3y4z5" description:"Unique request identifier"`
	Timestamp  string      `json:"timestamp,omitempty" example:"2023-12-01T10:30:00Z" description:"Response timestamp in ISO 8601 format"`
	Version    string      `json:"version,omitempty" example:"v1" description:"API version"`
}

// WriteError writes the canonical APIResponse error envelope directly
// to a stdlib http.ResponseWriter. Used by chi middleware that rejects
// a request (auth failure, rate limit, panic) before it reaches a Huma
// operation, where there is no Huma-managed context in scope.
//
// HTTP status is derived from AppError.Type via the canonical mapping;
// non-AppError errors surface as TypeAPIError (HTTP 500). Skips
// encoding the body on 204 (RFC 9110 §15.3.5). JSON-encode errors are
// swallowed — the response is already committed.
func WriteError(w http.ResponseWriter, err error) {
	apiError, statusCode := buildAPIError(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if statusCode == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   apiError,
	})
}

// buildAPIError renders an arbitrary error into the wire APIError plus
// the HTTP status to write. Shared with the Huma `NewError` override in
// `internal/server/api_error.go`.
func buildAPIError(err error) (*APIError, int) {
	if appErr := appErrors.AsAppError(err); appErr != nil {
		return &APIError{
			Type:    string(appErr.Type),
			Code:    appErr.CodeOrType(),
			Message: appErr.Message,
			Details: appErr.Details,
			Param:   appErr.Param,
		}, appErr.HTTPStatus()
	}
	return &APIError{
		Type:    string(appErrors.TypeAPIError),
		Code:    string(appErrors.TypeAPIError),
		Message: "Internal server error",
	}, http.StatusInternalServerError
}
