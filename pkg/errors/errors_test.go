package errors

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strings"
	"testing"
)

// TestEveryErrorTypeHasStatus locks in the invariant that every value
// listed in the ErrorType const block has a typeToStatus row. The
// const block IS the source of truth for the closed enum; missing a
// row would silently fall back to 500 in production. Catching that at
// build time is what keeps the closed-enum guarantee real.
func TestEveryErrorTypeHasStatus(t *testing.T) {
	declared := []ErrorType{
		TypeInvalidRequest,
		TypeValidation,
		TypeAuthentication,
		TypePermission,
		TypeNotFound,
		TypeConflict,
		TypePaymentRequired,
		TypeRateLimit,
		TypeUpstreamProvider,
		TypeServiceUnavailable,
		TypeNotImplemented,
		TypeAPIError,
	}
	for _, tt := range declared {
		if _, ok := typeToStatus[tt]; !ok {
			t.Errorf("ErrorType %q has no typeToStatus row — add it to the map in errors.go", tt)
		}
	}
	if len(declared) != len(typeToStatus) {
		t.Errorf("declared types (%d) != typeToStatus entries (%d) — exhaustive list out of sync",
			len(declared), len(typeToStatus))
	}
}

// TestFromHTTPStatusClassFallback is the regression test for the
// original bug: framework-emitted 4xx statuses (405, 406, 408, 413,
// 415, 416, 417, 451) used to fall through to INTERNAL_ERROR. The
// class fallback in typeFromStatus must keep them in TypeInvalidRequest
// (or a more specific 4xx Type), never let them leak to TypeAPIError.
func TestFromHTTPStatusClassFallback(t *testing.T) {
	// Statuses commonly emitted at the framework boundary (chi/net-http
	// 405, request-decode 413/415/422, content negotiation 406), plus a
	// fronting CDN's 520 (Cloudflare).
	cases := map[int]ErrorType{
		http.StatusMethodNotAllowed:           TypeInvalidRequest,
		http.StatusNotAcceptable:              TypeInvalidRequest,
		http.StatusRequestTimeout:             TypeInvalidRequest,
		http.StatusRequestEntityTooLarge:      TypeInvalidRequest,
		http.StatusUnsupportedMediaType:       TypeInvalidRequest,
		http.StatusRequestedRangeNotSatisfiable: TypeInvalidRequest,
		http.StatusExpectationFailed:          TypeInvalidRequest,
		http.StatusUnavailableForLegalReasons: TypeInvalidRequest,
		418:                                   TypeInvalidRequest, // I'm a teapot — class fallback
		520:                                   TypeAPIError,       // Cloudflare unknown — 5xx fallback
		599:                                   TypeAPIError,       // far end of 5xx — fallback
		http.StatusGatewayTimeout:             TypeServiceUnavailable,
		http.StatusBadGateway:                 TypeUpstreamProvider,
	}
	for status, want := range cases {
		got := typeFromStatus(status)
		if got != want {
			t.Errorf("typeFromStatus(%d) = %q, want %q", status, got, want)
		}
		if want != TypeAPIError && got == TypeAPIError {
			t.Errorf("MISCLASSIFICATION REGRESSION: status %d (4xx) leaked to TypeAPIError", status)
		}
	}
}

// TestAppErrorIs locks in the (Type, Code) matching contract used by
// errors.Is — type-only matching is a common pattern and must not
// regress to require code equality.
func TestAppErrorIs(t *testing.T) {
	err := NewNotFoundError("project", WithCode("project_not_found"))

	if !stderrors.Is(err, &AppError{Type: TypeNotFound}) {
		t.Error("type-only matching against TypeNotFound should succeed")
	}
	if !stderrors.Is(err, &AppError{Type: TypeNotFound, Code: "project_not_found"}) {
		t.Error("type+code matching against same code should succeed")
	}
	if stderrors.Is(err, &AppError{Type: TypeNotFound, Code: "user_not_found"}) {
		t.Error("type+code matching against different code should fail")
	}
	if stderrors.Is(err, &AppError{Type: TypeValidation}) {
		t.Error("type-only matching against different type should fail")
	}
}

// TestFunctionalOptions sanity-checks the variadic-options pattern —
// each option must mutate only its own field, options must compose,
// and the typed constructors must default Type / Code consistently.
func TestFunctionalOptions(t *testing.T) {
	cause := stderrors.New("underlying boom")
	err := NewValidationError("Invalid project ID", "projectId must be a valid UUID",
		WithCode("invalid_project_id"),
		WithParam("projectId"),
		WithCause(cause),
	)

	if err.Type != TypeValidation {
		t.Errorf("Type = %q, want %q", err.Type, TypeValidation)
	}
	if err.Code != "invalid_project_id" {
		t.Errorf("Code = %q, want %q", err.Code, "invalid_project_id")
	}
	if err.Param != "projectId" {
		t.Errorf("Param = %q, want %q", err.Param, "projectId")
	}
	if err.Details != "projectId must be a valid UUID" {
		t.Errorf("Details = %q, want the supplied details string", err.Details)
	}
	if !stderrors.Is(err, cause) {
		t.Error("WithCause should make errors.Is(err, cause) succeed via Unwrap")
	}
	if err.HTTPStatus() != http.StatusUnprocessableEntity {
		t.Errorf("HTTPStatus() = %d, want %d (TypeValidation → 422)",
			err.HTTPStatus(), http.StatusUnprocessableEntity)
	}
}

// TestAppError_MarshalJSON_ProducesCanonicalEnvelope locks the wire
// contract: Stripe/OpenAI-style `{"error": {...}}` with NO top-level
// `success` field (HTTP status is the success signal per RFC 9110).
// Regression guard for (a) the prior bug where AppError marshalled as
// raw Go struct with capitalised keys, and (b) any future drift back
// toward the envelope-style `{success: false, error: ...}` shape.
func TestAppError_MarshalJSON_ProducesCanonicalEnvelope(t *testing.T) {
	err := NewUnauthorizedError("Invalid email or password")

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	var parsed map[string]any
	if uerr := json.Unmarshal(raw, &parsed); uerr != nil {
		t.Fatalf("unmarshal wire bytes: %v", uerr)
	}

	// Exactly one top-level key: "error". No `success`, no envelope.
	if _, exists := parsed["success"]; exists {
		t.Errorf("envelope must NOT contain top-level `success` key "+
			"(Stripe/OpenAI-style shape); got %v", parsed)
	}
	if len(parsed) != 1 {
		t.Errorf("envelope must have exactly one top-level key (\"error\"); got %v", parsed)
	}

	inner, ok := parsed["error"].(map[string]any)
	if !ok {
		t.Fatalf("envelope.error missing or not an object: %v", parsed["error"])
	}
	if inner["type"] != string(TypeAuthentication) {
		t.Errorf("error.type = %q, want %q", inner["type"], TypeAuthentication)
	}
	// `code` is OPTIONAL — Stripe behaviour. The constructor used here
	// (NewUnauthorizedError) does not set Code, so the wire envelope
	// must OMIT the field (omitempty) rather than echo Type. This pins
	// the post-fix behaviour against regressing back to the old
	// CodeOrType fallback.
	if _, exists := inner["code"]; exists {
		t.Errorf("error.code must be omitted when unset; got %v", inner["code"])
	}
	if inner["message"] != "Invalid email or password" {
		t.Errorf("error.message = %q, want %q", inner["message"], "Invalid email or password")
	}
	if _, exists := inner["details"]; exists {
		t.Errorf("error.details should be omitted when empty, got %v", inner["details"])
	}
	if _, exists := inner["param"]; exists {
		t.Errorf("error.param should be omitted when empty, got %v", inner["param"])
	}

	// No capitalised leaks from the old raw-struct marshalling path.
	for _, leaked := range []string{"Type", "Code", "Message", "Details", "Param", "Err"} {
		if _, exists := parsed[leaked]; exists {
			t.Errorf("envelope contains capitalised key %q — raw-struct path regressed", leaked)
		}
		if _, exists := inner[leaked]; exists {
			t.Errorf("envelope.error contains capitalised key %q — raw-struct path regressed", leaked)
		}
	}
}

// TestAppError_MarshalJSON_OmitsErrField guarantees wrapped causes
// (database driver errors, upstream provider failures, etc.) never
// reach the HTTP client. The cause remains on the Go struct for
// errors.As / Unwrap / slog emission; only MarshalJSON must scrub it.
// Before this fix, `"Err":null` was leaking on every response and a
// real wrapped cause would have leaked as `"Err":"...cause..."`.
func TestAppError_MarshalJSON_OmitsErrField(t *testing.T) {
	cause := stderrors.New("internal upstream failure secret-token=abc123")
	err := NewInternalError("Something went wrong", cause)

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	wire := string(raw)
	if strings.Contains(wire, "Err") {
		t.Errorf("wire bytes contain 'Err' field — leaks internal cause to clients: %s", wire)
	}
	if strings.Contains(wire, "secret-token") {
		t.Errorf("wire bytes leak the wrapped cause content: %s", wire)
	}
}

// TestAppError_MarshalJSON_IncludesOptionalFields validates that
// when Details and Param ARE set, they appear on the wire with the
// correct lowercase keys (the omitempty path).
func TestAppError_MarshalJSON_IncludesOptionalFields(t *testing.T) {
	err := NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	err.Param = "projectId"

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	inner, _ := parsed["error"].(map[string]any)

	if inner["details"] != "projectId must be a valid UUID" {
		t.Errorf("error.details = %v, want %q", inner["details"], "projectId must be a valid UUID")
	}
	if inner["param"] != "projectId" {
		t.Errorf("error.param = %v, want %q", inner["param"], "projectId")
	}
}

// TestAppError_MarshalJSON_Errors locks the Errors[] wire shape —
// each entry emits as {location, message, value} with omitempty on
// location+value. The top-level envelope carries `errors` only when
// non-empty (omitempty), so nil + empty slices disappear from the
// wire rather than appearing as `"errors":[]` (which would drift
// from the Stripe/OpenAI envelope shape).
func TestAppError_MarshalJSON_Errors(t *testing.T) {
	t.Run("populated", func(t *testing.T) {
		err := NewValidationError("Validation failed", "one or more fields failed validation",
			WithParam("body.name"),
			WithErrors([]ErrorDetail{
				{Location: "body.name", Message: "required", Value: ""},
				{Location: "body.count", Message: "must be >= 1", Value: 0},
			}),
		)
		raw, jerr := json.Marshal(err)
		if jerr != nil {
			t.Fatalf("json.Marshal: %v", jerr)
		}

		var parsed map[string]any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		inner, _ := parsed["error"].(map[string]any)
		entries, ok := inner["errors"].([]any)
		if !ok {
			t.Fatalf("error.errors missing or not an array: %v", inner["errors"])
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d: %v", len(entries), entries)
		}
		first, _ := entries[0].(map[string]any)
		if first["location"] != "body.name" || first["message"] != "required" {
			t.Errorf("first entry = %v", first)
		}
	})

	t.Run("empty slice omitted", func(t *testing.T) {
		err := NewValidationError("nope", "why", WithErrors(nil))
		raw, _ := json.Marshal(err)
		var parsed map[string]any
		_ = json.Unmarshal(raw, &parsed)
		inner, _ := parsed["error"].(map[string]any)
		if _, exists := inner["errors"]; exists {
			t.Errorf("empty Errors[] must be omitted, got: %v", inner["errors"])
		}
	})
}

// TestAppError_MarshalJSON_CodeOverridesType locks the CodeOrType
// fallback: when the caller sets an explicit Code (e.g. a fine-grained
// domain code like "project_not_found"), it wins over the coarse Type
// string on the wire. Clients that switch on `error.code` therefore see
// the domain-authored value instead of the type name.
func TestAppError_MarshalJSON_CodeOverridesType(t *testing.T) {
	err := &AppError{
		Type:    TypeNotFound,
		Code:    "project_not_found",
		Message: "Project not found",
	}

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	inner, _ := parsed["error"].(map[string]any)

	if inner["type"] != string(TypeNotFound) {
		t.Errorf("error.type = %q, want %q", inner["type"], TypeNotFound)
	}
	if inner["code"] != "project_not_found" {
		t.Errorf("error.code = %q, want %q (explicit Code)", inner["code"], "project_not_found")
	}
}
