// Package response defines the canonical error-response DTO and
// supporting machinery:
//
//   - APIError / ErrorDetail / Pagination — wire types shared by
//     every Brokle HTTP response that has an error body or a list
//     pagination block.
//   - WriteError — stdlib-http helper used by chi middleware that rejects
//     a request before it reaches a Huma operation.
//   - ErrorResponse + InstallHumaErrorFactory — the huma.StatusError
//     implementation and the explicit installer that wires it into
//     Huma's global huma.NewError extension point (see humaerror.go).
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

	"github.com/danielgtaylor/huma/v2"

	appErrors "brokle/pkg/errors"
)

// APIError mirrors the Stripe / OpenAI / Anthropic shape.
//
//   - Type   — closed coarse classification clients switch on for
//     retry / alert behaviour (e.g. "validation_error", "rate_limit").
//   - Code   — open fine-grained domain code (snake_case) for SDK
//     subclassing (e.g. "project_not_found", "quota_exceeded").
//   - Errors — per-field diagnostics lifted from Huma's native
//     ErrorDetailer. Populated on validation failures so clients render
//     "body.items[3].tags: required" next to the offending field;
//     empty on domain-level errors where a single Message suffices.
//   - Details — free-form human-readable elaboration. Kept alongside
//     Errors (not replaced by it) because many AppError call sites
//     populate a single-string detail; machine-readable Errors is
//     additive, not a replacement.
//   - Param  — input field reference for single-field errors that
//     predate the Errors array. New handler code should prefer
//     Errors[].Location.
type APIError struct {
	Type    string        `json:"type" example:"validation_error" description:"Closed coarse error classification"`
	Code    string        `json:"code,omitempty" example:"project_not_found" description:"Open fine-grained domain code (snake_case)"`
	Message string        `json:"message" example:"Invalid request data" description:"Human-readable error message"`
	Details string        `json:"details,omitempty" example:"projectId must be a valid UUID" description:"Additional error context"`
	Param   string        `json:"param,omitempty" example:"projectId" description:"Input field that triggered the error, when applicable"`
	Errors  []ErrorDetail `json:"errors,omitempty" description:"Per-field validation diagnostics (location + message + value)"`
}

// ErrorDetail is re-exported from pkg/errors. Callers that already
// import pkg/response keep compiling; new code is free to import the
// type directly from pkg/errors.
type ErrorDetail = appErrors.ErrorDetail

// Pagination is the offset-paginated list metadata published inline
// on list-response bodies: `{"data": [...], "pagination": {...}}`.
//
// Cursor-based pagination (Stripe-style `has_more` + `url`) is a
// deferred migration — see the Option B plan. For now, offset-based
// is the canonical shape.
type Pagination struct {
	Page       int   `json:"page" example:"1" description:"Current page number (1-indexed)"`
	Limit      int   `json:"limit" example:"50" description:"Items per page"`
	Total      int64 `json:"total" example:"1234" description:"Total number of items"`
	TotalPages int   `json:"total_pages" example:"25" description:"Total number of pages"`
	HasNext    bool  `json:"has_next" example:"true" description:"Whether there are more pages"`
	HasPrev    bool  `json:"has_prev" example:"false" description:"Whether there are previous pages"`
}

// BuildPagination constructs the canonical Pagination metadata from
// the raw pagination inputs (page + limit) and the authoritative
// total count. Used by list handlers so every list endpoint emits
// the same inline `{data, pagination}` shape with identical field
// semantics.
//
// Edge cases:
//
//   - limit == 0: TotalPages is 0 (undefined pagination, e.g. when
//     the caller passed no limit — the endpoint should have defaulted
//     already; this branch just keeps Division by Zero at bay).
//   - total == 0: TotalPages is 0, HasNext/HasPrev both false.
//   - page > totalPages (client asking beyond the end): HasNext is
//     false, HasPrev is true when page > 1 — the caller can walk back.
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
// the HTTP status to write. Shared between WriteError and the Huma
// NewError override in humaerror.go.
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

// fromHumaErrors distils a variadic slice of errors into zero or more
// ErrorDetail entries. Huma validation passes each per-field failure
// as a separate error that implements huma.ErrorDetailer; we unwrap
// those into structured entries. Plain errors are included with only
// Message populated. Nil entries are skipped.
func fromHumaErrors(errs []error) []ErrorDetail {
	if len(errs) == 0 {
		return nil
	}
	out := make([]ErrorDetail, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		if d, ok := err.(huma.ErrorDetailer); ok {
			ed := d.ErrorDetail()
			out = append(out, ErrorDetail{
				Location: ed.Location,
				Message:  ed.Message,
				Value:    ed.Value,
			})
			continue
		}
		out = append(out, ErrorDetail{Message: err.Error()})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
