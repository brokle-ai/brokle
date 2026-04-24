// Package request bundles HTTP request-binding helpers for chi-native
// handlers: JSON body decoding + struct-tag validation, URL parameter
// parsing, and query parameter parsing. Every helper returns
// *appErrors.AppError (via the error interface) so handlers hand the
// return straight to pkg/response.WriteError — no translation layer,
// one error taxonomy end-to-end.
//
// Design anchors (cited rather than reinvented):
//
//   - Alex Edwards, "How to Parse a JSON Request Body in Go"
//     (https://www.alexedwards.net/blog/how-to-properly-parse-a-json-
//     request-body) — body size cap, DisallowUnknownFields, classified
//     decoder errors, and trailing-data rejection together form the
//     canonical production decoder shape in the Go ecosystem.
//
//   - Brandur Leach, "Web APIs: Enriched DX By Disallowing Unknown
//     Fields" (https://brandur.org/disallow-unknown-fields) — silently
//     ignoring unknown fields hides client typos until production;
//     rejecting them with a 422 surfaces integration bugs at
//     development time. Stripe's decoder behaves this way; so does
//     ours.
//
//   - go-playground/validator/v10 — the most widely used Go struct
//     validator (dep of Gin/Echo/Fiber). One process-wide instance
//     configured at init; safe for concurrent Struct() calls per
//     validator v10 docs. RegisterTagNameFunc is wired so error paths
//     report the JSON tag name ("body.name", not "Body.Name").
//
// Handlers compose these primitives:
//
//	func (h *handler) create(w http.ResponseWriter, r *http.Request) {
//	    orgID, err := request.URLParamUUID(r, "orgId")
//	    if err != nil { response.WriteError(w, err); return }
//
//	    var body createFooBody
//	    if err := request.DecodeJSON(r, &body); err != nil {
//	        response.WriteError(w, err); return
//	    }
//	    ...
//	}
//
// Raw-protobuf endpoints (OTLP ingestion) bypass this package entirely
// and keep their own decoders + size caps; they share only the error
// taxonomy with DecodeJSON, not the body machinery.
package request

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	appErrors "brokle/pkg/errors"
)

// maxBodyBytes caps the decoded JSON body. 1 MiB is well above any
// dashboard or SDK payload we expect and far under the point where
// accepting the body would be a DoS risk. If a future endpoint
// legitimately needs more, expose a DecodeJSONWithLimit variant —
// do not raise the global cap.
const maxBodyBytes = 1 << 20

// validate is the process-wide validator. Configured once at package
// init; safe for concurrent Struct() calls (validator v10 doc:
// "Validate is designed to be thread-safe and used as a singleton").
var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	// Use the JSON tag as the field name in validation errors — so
	// clients see "body.api_key" (matching the request body key they
	// sent) rather than "Body.APIKey" (the Go field name).
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// DecodeJSON reads, size-limits, decodes, and validates r.Body into
// dst. On success dst is populated and nil is returned. On failure
// the returned error is a typed *appErrors.AppError ready for
// response.WriteError — no translation needed at the call site.
//
// Safety layers, in order:
//
//  1. http.MaxBytesReader caps the request body at maxBodyBytes
//     (1 MiB) before any decode work happens. A rogue client
//     streaming 10 GB hits the cap mid-read and receives a 413.
//
//  2. Decoder.DisallowUnknownFields rejects JSON keys not present on
//     dst. A client typing "user_nam":"x" gets a 422 with the
//     typoed field named, instead of having the value silently
//     ignored.
//
//  3. After the first value is decoded, dec.More() is checked: a
//     second JSON document on the same body is rejected. Prevents
//     HTTP request smuggling shenanigans where the second value is
//     payload intended for a different endpoint.
//
//  4. validator.Struct enforces the `validate:"..."` struct tags.
//     Error → per-field ErrorDetail entries on the returned
//     AppError.Errors, with Location using the JSON tag path.
//
// Error classification:
//
//	io.EOF / empty body               → 400 invalid_request_error
//	io.ErrUnexpectedEOF               → 400 invalid_request_error
//	*json.SyntaxError                 → 400 invalid_request_error
//	*json.UnmarshalTypeError          → 422 validation_error (Param=field)
//	"json: unknown field X"           → 422 validation_error (Param=X)
//	*http.MaxBytesError               → 400 invalid_request_error (413 coverage)
//	trailing-data                     → 400 invalid_request_error
//	validator.ValidationErrors        → 422 validation_error (Errors[])
//	anything else                     → 400 invalid_request_error (wrapped)
func DecodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return classifyDecodeError(err)
	}
	if dec.More() {
		return appErrors.NewBadRequestError(
			"Invalid request body",
			"request body must contain only a single JSON document",
		)
	}

	if err := validate.Struct(dst); err != nil {
		return classifyValidationError(err)
	}
	return nil
}

func classifyDecodeError(err error) error {
	var (
		syntaxErr    *json.SyntaxError
		unmarshalErr *json.UnmarshalTypeError
		maxBytesErr  *http.MaxBytesError
	)
	switch {
	case stderrors.Is(err, io.EOF):
		return appErrors.NewBadRequestError(
			"Request body is empty",
			"expected a JSON document",
		)
	case stderrors.Is(err, io.ErrUnexpectedEOF):
		return appErrors.NewBadRequestError(
			"Malformed JSON",
			"request body ended unexpectedly",
		)
	case stderrors.As(err, &syntaxErr):
		return appErrors.NewBadRequestError(
			"Malformed JSON",
			fmt.Sprintf("syntax error at byte offset %d", syntaxErr.Offset),
		)
	case stderrors.As(err, &unmarshalErr):
		field := unmarshalErr.Field
		if field == "" {
			field = unmarshalErr.Type.String()
		}
		return appErrors.NewValidationError(
			fmt.Sprintf("Invalid type for field %q", field),
			fmt.Sprintf("expected %s", unmarshalErr.Type.String()),
			appErrors.WithParam(field),
		)
	case stderrors.As(err, &maxBytesErr):
		return appErrors.NewBadRequestError(
			"Request body too large",
			fmt.Sprintf("body must not exceed %d bytes", maxBytesErr.Limit),
		)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		// stdlib does not expose a named error type for unknown
		// fields. Parse the canonical message: `json: unknown field "X"`.
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		field = strings.Trim(field, `"`)
		return appErrors.NewValidationError(
			fmt.Sprintf("Unknown field %q", field),
			"remove the field or correct its name",
			appErrors.WithParam(field),
		)
	default:
		return appErrors.NewBadRequestError(
			"Invalid request body",
			err.Error(),
		)
	}
}

func classifyValidationError(err error) error {
	var ve validator.ValidationErrors
	if !stderrors.As(err, &ve) {
		// validator.Struct itself returned a non-ValidationErrors —
		// typically an InvalidValidationError from passing a nil
		// pointer. Surface it as a generic validation failure rather
		// than a 500 because the caller's input did reach validation.
		return appErrors.NewValidationError("Validation failed", err.Error())
	}
	details := make([]appErrors.ErrorDetail, 0, len(ve))
	for _, fe := range ve {
		details = append(details, appErrors.ErrorDetail{
			Location: fieldLocation(fe),
			Message:  humanize(fe),
			Value:    fe.Value(),
		})
	}
	param := ""
	if len(details) > 0 {
		param = details[0].Location
	}
	return appErrors.NewValidationError(
		"Validation failed",
		"one or more fields failed validation",
		appErrors.WithParam(param),
		appErrors.WithErrors(details),
	)
}

// fieldLocation renders a validator.FieldError as a document-rooted
// dotted path ("body.items[3].tags"). The raw Namespace() from
// validator is "StructName.field.nested", where StructName is the
// Go type name of the root value passed to Struct(). We strip the
// Go type segment (not part of the wire document) and prepend
// "body." so the client sees a path it can map back to the JSON
// body it sent.
func fieldLocation(fe validator.FieldError) string {
	if _, rest, ok := strings.Cut(fe.Namespace(), "."); ok {
		return "body." + rest
	}
	// Unusual: struct-level validation with no field segment. Fall
	// back to the field name so we never emit a bare "body." entry.
	return "body." + fe.Field()
}

// humanize turns a validator.FieldError into a human-readable
// diagnostic string. Coverage targets the tags we actually use across
// the handler layer; unknown tags fall back to a readable default
// rather than panicking.
func humanize(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "min":
		return "must be at least " + fe.Param()
	case "max":
		return "must be at most " + fe.Param()
	case "len":
		return "must have length " + fe.Param()
	case "oneof":
		return "must be one of: " + fe.Param()
	case "email":
		return "must be a valid email"
	case "url":
		return "must be a valid URL"
	case "uri":
		return "must be a valid URI"
	case "uuid", "uuid4":
		return "must be a valid UUID"
	case "eqfield":
		return "must equal " + fe.Param()
	case "nefield":
		return "must not equal " + fe.Param()
	case "gte":
		return "must be >= " + fe.Param()
	case "lte":
		return "must be <= " + fe.Param()
	case "gt":
		return "must be > " + fe.Param()
	case "lt":
		return "must be < " + fe.Param()
	case "alphanum":
		return "must contain only letters and digits"
	case "hexadecimal":
		return "must be a hexadecimal string"
	case "numeric":
		return "must be numeric"
	default:
		if fe.Param() == "" {
			return "failed validation: " + fe.Tag()
		}
		return "failed validation: " + fe.Tag() + "=" + fe.Param()
	}
}

// -----------------------------------------------------------------------------
// URL-path parameters
// -----------------------------------------------------------------------------

// URLParamUUID parses a chi URL parameter as a uuid.UUID. Returns a
// typed validation error with Param populated so clients can map the
// failure back to the URL segment.
func URLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	raw := chi.URLParam(r, key)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError(
			fmt.Sprintf("Invalid %s", key),
			fmt.Sprintf("%s must be a valid UUID", key),
			appErrors.WithParam(key),
		)
	}
	return id, nil
}

// URLParamInt parses a chi URL parameter as an int.
func URLParamInt(r *http.Request, key string) (int, error) {
	raw := chi.URLParam(r, key)
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, appErrors.NewValidationError(
			fmt.Sprintf("Invalid %s", key),
			fmt.Sprintf("%s must be an integer", key),
			appErrors.WithParam(key),
		)
	}
	return v, nil
}

// URLParamString returns the raw chi URL parameter. Empty string if
// absent (per chi semantics). Present for symmetry with the other
// typed helpers so handlers don't need to import chi for this one
// operation.
func URLParamString(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// -----------------------------------------------------------------------------
// Query parameters
// -----------------------------------------------------------------------------

// QueryString returns r.URL.Query().Get(key). Empty when absent.
func QueryString(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// QueryOptionalBool returns nil when the query param is absent, *bool
// when present and parseable, or a typed validation error on malformed
// input. Accepts anything strconv.ParseBool understands: 1/0, t/f,
// true/false, TRUE/FALSE.
func QueryOptionalBool(r *http.Request, key string) (*bool, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, appErrors.NewValidationError(
			fmt.Sprintf("Invalid %s", key),
			fmt.Sprintf("%s must be a boolean (true/false)", key),
			appErrors.WithParam(key),
		)
	}
	return &v, nil
}

// QueryOptionalInt — nil when absent, *int when present and parseable.
func QueryOptionalInt(r *http.Request, key string) (*int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, appErrors.NewValidationError(
			fmt.Sprintf("Invalid %s", key),
			fmt.Sprintf("%s must be an integer", key),
			appErrors.WithParam(key),
		)
	}
	return &v, nil
}

// QueryOptionalUUID — nil when absent, *uuid.UUID when present and valid.
func QueryOptionalUUID(r *http.Request, key string) (*uuid.UUID, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	v, err := uuid.Parse(raw)
	if err != nil {
		return nil, appErrors.NewValidationError(
			fmt.Sprintf("Invalid %s", key),
			fmt.Sprintf("%s must be a valid UUID", key),
			appErrors.WithParam(key),
		)
	}
	return &v, nil
}

// QueryInt returns the parsed int, or the default when absent.
func QueryInt(r *http.Request, key string, def int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, appErrors.NewValidationError(
			fmt.Sprintf("Invalid %s", key),
			fmt.Sprintf("%s must be an integer", key),
			appErrors.WithParam(key),
		)
	}
	return v, nil
}

// QueryPagination reads ?page=&limit= with the Brokle-standard
// defaults (page=1, limit=50) and bounds (page≥1, 1≤limit≤1000).
// Returns a validation error on out-of-range or unparseable input.
//
// Kept in pkg/request (not a handler-level helper) because every list
// endpoint needs identical pagination semantics and identical error
// wording — the handler layer should never reinvent this.
func QueryPagination(r *http.Request) (page, limit int, err error) {
	page, err = QueryInt(r, "page", 1)
	if err != nil {
		return 0, 0, err
	}
	limit, err = QueryInt(r, "limit", 50)
	if err != nil {
		return 0, 0, err
	}
	if page < 1 {
		return 0, 0, appErrors.NewValidationError(
			"Invalid page",
			"page must be >= 1",
			appErrors.WithParam("page"),
		)
	}
	if limit < 1 || limit > 1000 {
		return 0, 0, appErrors.NewValidationError(
			"Invalid limit",
			"limit must be between 1 and 1000",
			appErrors.WithParam("limit"),
		)
	}
	return page, limit, nil
}
