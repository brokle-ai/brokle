package errors

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strings"
	"testing"
)

// TestEveryReasonHasStatus locks in the invariant that every Reason
// constant has both a non-zero HTTP status mapping AND a non-empty
// HTTPType wire value. The const block IS the source of truth for the
// closed enum; missing a row would silently fall back to 500 / api_error
// in production. Catching that at build time is what keeps the
// closed-enum guarantee real.
func TestEveryReasonHasStatus(t *testing.T) {
	declared := []Reason{
		ReasonInternal,
		ReasonNotFound,
		ReasonAlreadyExists,
		ReasonInvalidInput,
		ReasonBadRequest,
		ReasonUnauthenticated,
		ReasonPermissionDenied,
		ReasonConflict,
		ReasonRateLimit,
		ReasonPaymentRequired,
		ReasonUpstream,
		ReasonUnavailable,
		ReasonNotImplemented,
	}
	for _, r := range declared {
		if r.HTTPStatus() == 0 {
			t.Errorf("Reason %q has no HTTPStatus mapping", r)
		}
		if r.HTTPType() == "" {
			t.Errorf("Reason %q has no HTTPType mapping", r)
		}
	}
}

// TestFromHTTPStatusClassFallback is the regression test for the
// original framework-boundary bug: 4xx statuses (405, 406, 408, 413,
// 415, 416, 417, 451) used to fall through to ReasonInternal. The
// class fallback in reasonFromStatus must keep them in ReasonBadRequest
// (or a more specific 4xx Reason), never let them leak to ReasonInternal.
func TestFromHTTPStatusClassFallback(t *testing.T) {
	cases := map[int]Reason{
		http.StatusMethodNotAllowed:             ReasonBadRequest,
		http.StatusNotAcceptable:                ReasonBadRequest,
		http.StatusRequestTimeout:               ReasonBadRequest,
		http.StatusRequestEntityTooLarge:        ReasonBadRequest,
		http.StatusUnsupportedMediaType:         ReasonBadRequest,
		http.StatusRequestedRangeNotSatisfiable: ReasonBadRequest,
		http.StatusExpectationFailed:            ReasonBadRequest,
		http.StatusUnavailableForLegalReasons:   ReasonBadRequest,
		418: ReasonBadRequest, // I'm a teapot — class fallback
		520: ReasonInternal,   // Cloudflare unknown — 5xx fallback
		599: ReasonInternal,   // far end of 5xx — fallback
		http.StatusGatewayTimeout: ReasonUnavailable,
		http.StatusBadGateway:     ReasonUpstream,
	}
	for status, want := range cases {
		got := reasonFromStatus(status)
		if got != want {
			t.Errorf("reasonFromStatus(%d) = %q, want %q", status, got, want)
		}
		if want != ReasonInternal && got == ReasonInternal {
			t.Errorf("MISCLASSIFICATION REGRESSION: status %d (4xx) leaked to ReasonInternal", status)
		}
	}
}

// TestErrorIs locks in the (Reason, Resource, Code) matching contract
// used by errors.Is — Reason-only matching is the common pattern and
// must not regress to require Resource or Code equality.
func TestErrorIs(t *testing.T) {
	err := NotFound("project", WithCode("project_not_found"))

	if !stderrors.Is(err, &Error{Reason: ReasonNotFound}) {
		t.Error("Reason-only matching against ReasonNotFound should succeed")
	}
	if !stderrors.Is(err, &Error{Reason: ReasonNotFound, Code: "project_not_found"}) {
		t.Error("Reason+Code matching against same code should succeed")
	}
	if stderrors.Is(err, &Error{Reason: ReasonNotFound, Code: "user_not_found"}) {
		t.Error("Reason+Code matching against different code should fail")
	}
	if stderrors.Is(err, &Error{Reason: ReasonInvalidInput}) {
		t.Error("Reason-only matching against different Reason should fail")
	}
	if !stderrors.Is(err, &Error{Reason: ReasonNotFound, Resource: "project"}) {
		t.Error("Reason+Resource matching against same resource should succeed")
	}
	if stderrors.Is(err, &Error{Reason: ReasonNotFound, Resource: "user"}) {
		t.Error("Reason+Resource matching against different resource should fail")
	}
}

// TestErrorIs_NoSentinelCollision is the regression guard for the
// 2026-04-29 review finding: when ReasonInternal was iota=0 the `Is`
// method's "wildcard" branch (`t.Reason == ReasonInternal &&
// t.Resource == "" && t.Code == ""`) made errors.Is treat any *Error
// as Internal. Reserving iota=0 as ReasonUnspecified + deleting the
// wildcard branch makes the collision structurally impossible.
func TestErrorIs_NoSentinelCollision(t *testing.T) {
	internalErr := Internal("oops", stderrors.New("boom"))
	notFoundErr := NotFound("project")

	if !stderrors.Is(internalErr, &Error{Reason: ReasonInternal}) {
		t.Error("internal error should match &Error{Reason: ReasonInternal}")
	}
	if stderrors.Is(notFoundErr, &Error{Reason: ReasonInternal}) {
		t.Error("not-found error must NOT match &Error{Reason: ReasonInternal} (sentinel collision regression)")
	}
	// ReasonUnspecified is iota=0; ensure no constructor produces it.
	// If a constructor ever returns Reason=ReasonUnspecified, the wire
	// envelope falls through to "api_error" / 500 and the iota
	// reservation is broken.
	if internalErr.Reason == ReasonUnspecified {
		t.Error("Internal() must produce Reason=ReasonInternal, not the iota=0 placeholder")
	}
	if notFoundErr.Reason == ReasonUnspecified {
		t.Error("NotFound() must produce Reason=ReasonNotFound, not the iota=0 placeholder")
	}
}

// TestFunctionalOptions sanity-checks the variadic-options pattern —
// each option must mutate only its own field, options must compose,
// and the typed constructors must default Reason consistently.
func TestFunctionalOptions(t *testing.T) {
	cause := stderrors.New("underlying boom")
	err := InvalidParam("projectId", "Invalid project ID",
		WithDetails("projectId must be a valid UUID"),
		WithCode("invalid_project_id"),
		WithCause(cause),
	)

	if err.Reason != ReasonInvalidInput {
		t.Errorf("Reason = %q, want %q", err.Reason, ReasonInvalidInput)
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
		t.Errorf("HTTPStatus() = %d, want %d (ReasonInvalidInput → 422)",
			err.HTTPStatus(), http.StatusUnprocessableEntity)
	}
}

// TestError_MarshalJSON_ProducesCanonicalEnvelope locks the wire
// contract: Stripe/OpenAI-style `{"error": {...}}` with NO top-level
// `success` field (HTTP status is the success signal per RFC 9110).
func TestError_MarshalJSON_ProducesCanonicalEnvelope(t *testing.T) {
	err := Unauthenticated("Invalid email or password")

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	var parsed map[string]any
	if uerr := json.Unmarshal(raw, &parsed); uerr != nil {
		t.Fatalf("unmarshal wire bytes: %v", uerr)
	}

	if _, exists := parsed["success"]; exists {
		t.Errorf("envelope must NOT contain top-level `success` key; got %v", parsed)
	}
	if len(parsed) != 1 {
		t.Errorf("envelope must have exactly one top-level key (\"error\"); got %v", parsed)
	}

	inner, ok := parsed["error"].(map[string]any)
	if !ok {
		t.Fatalf("envelope.error missing or not an object: %v", parsed["error"])
	}
	if inner["type"] != "authentication_error" {
		t.Errorf("error.type = %q, want %q", inner["type"], "authentication_error")
	}
	// `code` is OPTIONAL — Stripe behaviour. The constructor used here
	// did not set Code, so the wire envelope must OMIT the field rather
	// than echo Type.
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

	// No capitalised leaks from a raw-struct marshalling regression.
	for _, leaked := range []string{"Reason", "Resource", "Code", "Message", "Details", "Param", "Op", "Cause"} {
		if _, exists := parsed[leaked]; exists {
			t.Errorf("envelope contains capitalised key %q — raw-struct path regressed", leaked)
		}
		if _, exists := inner[leaked]; exists {
			t.Errorf("envelope.error contains capitalised key %q — raw-struct path regressed", leaked)
		}
	}
}

// TestError_MarshalJSON_OmitsCauseField guarantees wrapped causes
// (database driver errors, upstream provider failures, etc.) never
// reach the HTTP client. The cause remains on the Go struct for
// errors.As / Unwrap / slog emission; only MarshalJSON must scrub it.
func TestError_MarshalJSON_OmitsCauseField(t *testing.T) {
	cause := stderrors.New("internal upstream failure secret-token=abc123")
	err := Internal("Something went wrong", cause)

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	wire := string(raw)
	if strings.Contains(wire, "Cause") {
		t.Errorf("wire bytes contain 'Cause' field — leaks internal cause to clients: %s", wire)
	}
	if strings.Contains(wire, "secret-token") {
		t.Errorf("wire bytes leak the wrapped cause content: %s", wire)
	}
}

// TestError_MarshalJSON_IncludesOptionalFields validates that when
// Details and Param ARE set, they appear on the wire with the correct
// lowercase keys (the omitempty path).
func TestError_MarshalJSON_IncludesOptionalFields(t *testing.T) {
	err := InvalidParam("projectId", "Invalid project ID",
		WithDetails("projectId must be a valid UUID"),
	)

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

// TestError_MarshalJSON_FieldErrors locks the Errors[] wire shape —
// each entry emits as {location, message, value} with omitempty on
// location+value. The top-level envelope carries `errors` only when
// non-empty.
func TestError_MarshalJSON_FieldErrors(t *testing.T) {
	t.Run("populated", func(t *testing.T) {
		err := InvalidFields(
			[]ErrorDetail{
				{Location: "body.name", Message: "required", Value: ""},
				{Location: "body.count", Message: "must be >= 1", Value: 0},
			},
			WithMessage("Validation failed"),
			WithDetails("one or more fields failed validation"),
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
		err := InvalidFields(nil,
			WithMessage("nope"),
			WithDetails("why"),
		)
		raw, _ := json.Marshal(err)
		var parsed map[string]any
		_ = json.Unmarshal(raw, &parsed)
		inner, _ := parsed["error"].(map[string]any)
		if _, exists := inner["errors"]; exists {
			t.Errorf("empty Errors[] must be omitted, got: %v", inner["errors"])
		}
	})
}

// TestError_MarshalJSON_ExplicitCode pins that an explicit Code shows
// up alongside Type on the wire — clients that switch on `error.code`
// therefore see the caller's authored value.
func TestError_MarshalJSON_ExplicitCode(t *testing.T) {
	err := NotFound("project", WithCode("project_not_found"))

	raw, jerr := json.Marshal(err)
	if jerr != nil {
		t.Fatalf("json.Marshal: %v", jerr)
	}

	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	inner, _ := parsed["error"].(map[string]any)

	if inner["type"] != "not_found_error" {
		t.Errorf("error.type = %q, want %q", inner["type"], "not_found_error")
	}
	if inner["code"] != "project_not_found" {
		t.Errorf("error.code = %q, want %q (explicit Code)", inner["code"], "project_not_found")
	}
}
