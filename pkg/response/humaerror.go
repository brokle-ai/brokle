// Huma v2 error envelope. Wires the canonical Brokle error DTO into
// every code path Huma emits — handler-returned errors, pipeline
// rejections (request body too large, content-type 415, validation 422,
// …), and the registration-time OpenAPI probe.
//
// Design (anchored against Huma v2.37.3):
//
//   - ErrorResponse is a pure DTO with natural Go field names (Success,
//     Error) and matching JSON tags. Handlers and tests inspect it via
//     ordinary field access — resp.Error.Message, resp.Error.Code — no
//     translation between Go identifier and JSON key. It has no
//     behavioural interfaces attached: one type, one responsibility.
//
//   - statusError is an internal wrapper carrying the HTTP status +
//     the DTO. Implements huma.StatusError (error + GetStatus) and
//     json.Marshaler (delegates to the DTO). Keeps ErrorResponse free
//     of the Error() method / Error field collision that would
//     otherwise force us to rename the field. Callers never see or
//     construct statusError directly.
//
//   - statusError implements huma.SchemaProvider
//     (schema.go:727-732) returning ErrorResponse's schema. This is the
//     load-bearing line — it tells Huma's registry (registry.go:111-115)
//     "this type doesn't get its own ref; use the referenced schema
//     instead". The result:
//       1. OpenAPI spec emits `#/components/schemas/ErrorResponse`
//          with the DTO's fields. SDK codegen (hey-api/openapi-ts,
//          openapi-python-client, fern) produces a clean
//          ErrorResponse type — no `statusError` leakage.
//       2. Huma's SchemaLinkTransformer keys its decoration table by
//          the TYPE registered against an operation's response
//          (ErrorResponse), but checks incoming runtime values by
//          their reflect.TypeOf (statusError). Types differ → the
//          transformer's `t.types[typ]` lookup misses on errors and
//          returns the value unchanged. We get $schema + Link
//          decoration on SUCCESS responses (ordinary DTO types) and
//          NO decoration on ERROR responses, automatically, without
//          writing a custom transformer wrapper.
//
//   - We do NOT implement huma.ContentTypeFilter. Huma's default
//     ErrorModel rewrites application/json → application/problem+json;
//     omitting the method keeps error responses on application/json,
//     matching the success path.
//
//   - Registration-time probe: Huma calls `NewError(0, "")` once per
//     operation (huma.go:1627) to obtain the example used for OpenAPI
//     schema generation. status==0 is normalised to 500 so the example
//     doesn't leak probe-level noise.
//
//   - No init() side effects. InstallHumaErrorFactory is an explicit
//     installer guarded by sync.Once. main() calls it before the first
//     huma.API is built; test harnesses call it via
//     internal/testing/humax. The override is global (huma.NewError is
//     a package-level var; per-API config tracked in huma#755).
package response

import (
	"encoding/json"
	"net/http"
	"reflect"
	"sync"

	"github.com/danielgtaylor/huma/v2"

	appErrors "brokle/pkg/errors"
)

// ErrorResponse is the wire DTO returned on every error path. Pure
// data type — one field (`error`) with the canonical APIError shape.
//
// Matches Stripe / OpenAI / Anthropic: `{"error": {...}}`. There is
// deliberately NO top-level `success` boolean — HTTP status is the
// canonical success/failure signal (RFC 9110 §15), and a redundant
// body field forces every client to double-check what the status
// line already told them.
//
// The Huma error factory wraps ErrorResponse in an internal
// huma.StatusError implementation; callers never need to do that
// themselves.
type ErrorResponse struct {
	Error *APIError `json:"error" description:"Error detail — always populated on error responses"`
}

// errorResponseType caches the reflect.Type used by Schema() below.
// reflect.TypeOf is cheap but called once per operation registration
// per huma.API; caching it trims startup work and makes the intent
// obvious.
var errorResponseType = reflect.TypeFor[ErrorResponse]()

// statusError wraps an ErrorResponse with the HTTP status needed to
// satisfy huma.StatusError. Unexported by design — construction and
// inspection go through the DTO.
type statusError struct {
	status int
	resp   *ErrorResponse
}

// Error satisfies the error interface (and therefore huma.StatusError,
// which embeds error). Reads the wrapped DTO's Error.Message so log /
// trace emission stays consistent with the wire response.
func (e *statusError) Error() string {
	if e.resp != nil && e.resp.Error != nil && e.resp.Error.Message != "" {
		return e.resp.Error.Message
	}
	return http.StatusText(e.status)
}

// GetStatus satisfies huma.StatusError — Huma writes the matching HTTP
// status line.
func (e *statusError) GetStatus() int { return e.status }

// MarshalJSON serializes the wrapped DTO directly so the wire response
// is the flat {success, error} shape. The wrapper fields (status,
// resp) are not part of the wire format.
func (e *statusError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.resp)
}

// Schema satisfies huma.SchemaProvider. Returning ErrorResponse's
// schema here tells Huma's registry (registry.go:111-115) to skip
// creating a component for statusError and reference ErrorResponse
// instead. This is what keeps SDK codegen clean and makes
// SchemaLinkTransformer's type-keyed lookup miss on errors.
func (e *statusError) Schema(r huma.Registry) *huma.Schema {
	return r.Schema(errorResponseType, true, "ErrorResponse")
}

// installHumaErrorOnce guards the global huma.NewError assignment so
// repeat calls from main() + TestMain + test helpers are safe.
var installHumaErrorOnce sync.Once

// InstallHumaErrorFactory replaces huma.NewError with the ErrorResponse
// envelope factory. Idempotent — safe to call from main(), humatest
// helpers, and TestMain hooks. Must be called BEFORE the first huma.API
// is constructed; the documented extension point is at package scope
// (see huma/error.go), not per-API.
func InstallHumaErrorFactory() {
	installHumaErrorOnce.Do(func() {
		huma.NewError = newHumaError
	})
}

// newHumaError is the huma.NewError replacement.
//
// Huma invokes it on three paths:
//
//  1. A handler returns a non-nil error. Huma reads the error's
//     GetStatus() (our *AppError satisfies huma.StatusError) and passes
//     the matched status here; the original error rides along in
//     errs[0]. We extract the AppError and preserve its Type / Code /
//     Message / Details / Param verbatim.
//
//  2. The pipeline rejects a request before any handler runs (request
//     body too large, content type 415, validation 422, method 405, …).
//     The errs slice carries each failure, typically implementing
//     huma.ErrorDetailer. We synthesise an AppError via
//     appErrors.FromHTTPStatus (class-fallback safe — no 4xx ever
//     degrades to a 5xx-flavoured Type) and populate Errors[] with the
//     per-field diagnostics lifted from huma.ErrorDetailer.
//
//  3. Operation registration: Huma calls NewError(0, "") once per
//     operation (huma.go:1627) to produce the OpenAPI example. We
//     normalise status==0 to 500 so the spec example doesn't leak
//     probe-level noise.
func newHumaError(status int, msg string, errs ...error) huma.StatusError {
	if status == 0 {
		status = http.StatusInternalServerError
	}

	if appErr := firstAppError(errs); appErr != nil {
		return wrapAppError(appErr, nil)
	}

	appErr := appErrors.FromHTTPStatus(status, msg, errs...)
	return wrapAppError(appErr, fromHumaErrors(errs))
}

// wrapAppError lifts an *AppError into the DTO and wraps it in the
// internal statusError so Huma sees a huma.StatusError implementation.
func wrapAppError(e *appErrors.AppError, details []ErrorDetail) *statusError {
	return &statusError{
		status: e.HTTPStatus(),
		resp: &ErrorResponse{
			Error: &APIError{
				Type:    string(e.Type),
				Code:    e.CodeOrType(),
				Message: e.Message,
				Details: e.Details,
				Param:   e.Param,
				Errors:  details,
			},
		},
	}
}

// firstAppError returns the first *AppError found in errs, unwrapping
// transparently via errors.As. nil if none.
func firstAppError(errs []error) *appErrors.AppError {
	for _, err := range errs {
		if err == nil {
			continue
		}
		if appErr := appErrors.AsAppError(err); appErr != nil {
			return appErr
		}
	}
	return nil
}
