// Package response defines the canonical HTTP response helpers and
// error-envelope DTO for Brokle's chi-native handlers.
//
// Public surface:
//
//   - APIError / ErrorDetail / Pagination — wire types shared by
//     every Brokle HTTP response that has an error body or a list
//     pagination block.
//   - ErrorResponse — the top-level `{"error": {...}}` wrapper
//     consumed by WriteError.
//   - WriteError / JSON / Success / Created / NoContent — stdlib
//     http helpers called by every handler.
//   - BuildPagination — constructs the canonical Pagination metadata.
//
// Wire contract (Stripe / OpenAI / Anthropic style):
//
//   - Success: raw resource body, HTTP status 2xx. No envelope.
//   - Error:   `{"error":{"type":"...","code":"...","message":"...",...}}`,
//     HTTP status 4xx/5xx. No `success` boolean — status is the signal.
//   - List:    `{"data":[...],"pagination":{...}}` inline. No outer meta.
//
// Request IDs surface as the `X-Request-Id` response header
// (chi/middleware.RequestID installs it globally).
package response

import (
	"encoding/json"
	"net/http"

	appErrors "brokle/pkg/errors"
)

// APIError mirrors the Stripe / OpenAI / Anthropic shape.
//
//   - Type   — closed coarse classification clients switch on for
//     retry / alert behaviour (e.g. "validation_error", "rate_limit").
//   - Code   — open fine-grained domain code (snake_case) for SDK
//     subclassing (e.g. "project_not_found", "quota_exceeded").
//   - Errors — per-field validation diagnostics populated by
//     pkg/request.DecodeJSON. Empty on domain-level errors where a
//     single Message suffices.
//   - Details — free-form human-readable elaboration. Kept alongside
//     Errors (not replaced by it) because many AppError call sites
//     populate a single-string detail; machine-readable Errors is
//     additive, not a replacement.
//   - Param  — input field reference for single-field errors that
//     predate the Errors array. New handler code should prefer
//     Errors[].Location.
type APIError struct {
	Type    string        `json:"type"`
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message"`
	Details string        `json:"details,omitempty"`
	Param   string        `json:"param,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

// ErrorResponse is the top-level wrapper around APIError. Emitted by
// WriteError and mirrored by AppError.MarshalJSON so both paths
// produce byte-identical output.
type ErrorResponse struct {
	Error *APIError `json:"error"`
}

// ErrorDetail is re-exported from pkg/errors. Callers that already
// import pkg/response keep compiling; new code is free to import the
// type directly from pkg/errors.
type ErrorDetail = appErrors.ErrorDetail

// Pagination is the offset-paginated list metadata published inline
// on list-response bodies: `{"data": [...], "pagination": {...}}`.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// BuildPagination constructs the canonical Pagination metadata from
// the raw pagination inputs (page + limit) and the authoritative
// total count.
//
// Edge cases:
//
//   - limit == 0: TotalPages is 0 (undefined pagination — the endpoint
//     should have defaulted already; this branch just keeps division
//     by zero at bay).
//   - total == 0: TotalPages is 0, HasNext/HasPrev both false.
//   - page > totalPages: HasNext is false, HasPrev is true when page > 1.
func BuildPagination(page, limit int, total int64) *Pagination {
	totalPages := 0
	if limit > 0 && total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// WriteError writes the canonical Brokle error envelope directly to a
// stdlib http.ResponseWriter. Used by every chi handler + every chi
// middleware that rejects a request (auth failure, rate limit, panic).
//
// Output shape matches AppError.MarshalJSON exactly — bytes pinned by
// error_shape_test.go:
//
//	{"error":{"type":"...","code":"...","message":"...",...}}
//
// HTTP status derives from AppError.Type via the canonical mapping;
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
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: apiError})
}

// JSON writes status + payload as JSON. The generic 2xx helper.
//
// Prefer the semantic helpers (Success, Created, NoContent) where one
// matches the intended status; JSON is the escape hatch for the less
// common 2xx codes (e.g. 202 Accepted, 207 Multi-Status). Writes an
// empty body on 204 per RFC 9110 §15.3.5. JSON-encode errors are
// swallowed because the response line is already committed.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent || payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

// Success writes 200 OK with a JSON-encoded body. Canonical helper
// for GET / PUT / PATCH happy paths.
func Success(w http.ResponseWriter, payload any) {
	JSON(w, http.StatusOK, payload)
}

// Created writes 201 Created with a JSON-encoded body. Canonical
// helper for POST operations that create a resource.
func Created(w http.ResponseWriter, payload any) {
	JSON(w, http.StatusCreated, payload)
}

// NoContent writes 204 No Content with no body. Canonical helper
// for DELETE operations and for PUT/PATCH where the caller
// explicitly opts out of an echo body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// buildAPIError renders an arbitrary error into the wire APIError plus
// the HTTP status to write.
func buildAPIError(err error) (*APIError, int) {
	if appErr := appErrors.AsAppError(err); appErr != nil {
		return &APIError{
			Type:    string(appErr.Type),
			Code:    appErr.CodeOrType(),
			Message: appErr.Message,
			Details: appErr.Details,
			Param:   appErr.Param,
			Errors:  appErr.Errors,
		}, appErr.HTTPStatus()
	}
	return &APIError{
		Type:    string(appErrors.TypeAPIError),
		Code:    string(appErrors.TypeAPIError),
		Message: "Internal server error",
	}, http.StatusInternalServerError
}
