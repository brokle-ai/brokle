// Cross-path wire-contract test. Two code paths produce an error
// response body:
//
//  1. Handler returns *AppError. AppError.MarshalJSON emits the
//     canonical {error:{...}} envelope directly.
//  2. Middleware or handler calls pkg/response.WriteError(w, err).
//     WriteError builds APIError + ErrorResponse and encodes them.
//
// This test guarantees both paths produce byte-identical envelope
// bytes for the same input AppError, including the per-field
// `errors` array populated by pkg/request.DecodeJSON. If the envelope
// shape ever drifts between paths, SDK consumers will see intermittent
// parse failures that are hell to diagnose — this test is the tripwire.
package response_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
)

// buildAppError produces an AppError rich enough to exercise every
// envelope field (Type, Code, Message, Details, Param) with an
// explicit Code — verifies that explicitly-set codes survive both
// emission paths.
func buildAppError() *appErrors.AppError {
	return &appErrors.AppError{
		Type:    appErrors.TypeNotFound,
		Code:    "project_not_found",
		Message: "Project not found",
		Details: "The project slug you provided does not match any workspace you have access to.",
		Param:   "projectSlug",
	}
}

// TestErrorShape_HandlerAndWriteErrorByteIdentical pins that
// AppError.MarshalJSON (direct json.Marshal) and WriteError emit
// byte-identical envelope bytes for the same input.
func TestErrorShape_HandlerAndWriteErrorByteIdentical(t *testing.T) {
	e := buildAppError()

	handlerBytes, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal(AppError): %v", err)
	}

	rec := httptest.NewRecorder()
	response.WriteError(rec, e)
	// WriteError uses json.NewEncoder which appends a trailing newline;
	// json.Marshal does not. Strip the trailing '\n' for comparison.
	writeErrorBytes := bytes.TrimRight(rec.Body.Bytes(), "\n")

	if !bytes.Equal(handlerBytes, writeErrorBytes) {
		t.Errorf("envelope drift between AppError.MarshalJSON and WriteError:\n"+
			"  handler path: %s\n"+
			"  WriteError:   %s",
			handlerBytes, writeErrorBytes)
	}
}

// TestErrorShape_ResponseEnvelopeMatchesAppError pins that the
// response.ErrorResponse marshalling path stays in lockstep with
// AppError.MarshalJSON when no per-field Errors[] are present.
func TestErrorShape_ResponseEnvelopeMatchesAppError(t *testing.T) {
	e := buildAppError()

	handlerBytes, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal(AppError): %v", err)
	}

	pipelineResp := response.ErrorResponse{
		Error: &response.APIError{
			Type:    string(e.Type),
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
			Param:   e.Param,
		},
	}
	pipelineBytes, err := json.Marshal(pipelineResp)
	if err != nil {
		t.Fatalf("json.Marshal(ErrorResponse): %v", err)
	}

	if !bytes.Equal(handlerBytes, pipelineBytes) {
		t.Errorf("envelope drift between AppError.MarshalJSON and ErrorResponse marshal:\n"+
			"  handler path:  %s\n"+
			"  pipeline path: %s",
			handlerBytes, pipelineBytes)
	}
}

// TestErrorShape_HandlerAndWriteErrorByteIdentical_WithErrors pins
// that AppError.MarshalJSON (handler path) and WriteError emit
// byte-identical envelope bytes when the AppError carries per-field
// Errors[] diagnostics (the go-playground/validator path).
func TestErrorShape_HandlerAndWriteErrorByteIdentical_WithErrors(t *testing.T) {
	e := &appErrors.AppError{
		Type:    appErrors.TypeValidation,
		Code:    "field_required",
		Message: "Validation failed",
		Details: "one or more fields failed validation",
		Param:   "body.name",
		Errors: []appErrors.ErrorDetail{
			{Location: "body.name", Message: "required", Value: ""},
			{Location: "body.api_key", Message: "must be at least 10", Value: "abc"},
		},
	}

	handlerBytes, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal(AppError): %v", err)
	}

	rec := httptest.NewRecorder()
	response.WriteError(rec, e)
	writeErrorBytes := bytes.TrimRight(rec.Body.Bytes(), "\n")

	if !bytes.Equal(handlerBytes, writeErrorBytes) {
		t.Errorf("envelope drift on Errors[] path:\n"+
			"  handler path: %s\n"+
			"  WriteError:   %s",
			handlerBytes, writeErrorBytes)
	}

	// Sanity: both paths include the errors array with both entries.
	if !bytes.Contains(handlerBytes, []byte(`"errors"`)) {
		t.Errorf("handler path missing `errors` key: %s", handlerBytes)
	}
	if !bytes.Contains(handlerBytes, []byte(`"body.name"`)) ||
		!bytes.Contains(handlerBytes, []byte(`"body.api_key"`)) {
		t.Errorf("handler path missing ErrorDetail entries: %s", handlerBytes)
	}
}

// TestErrorShape_NoSuccessField guards against reintroducing the
// `success` boolean on any error path. The Stripe/OpenAI contract is
// strict: the envelope has exactly one top-level key ("error"). HTTP
// status is the success signal.
func TestErrorShape_NoSuccessField(t *testing.T) {
	e := buildAppError()

	bodies := map[string][]byte{}

	raw, _ := json.Marshal(e)
	bodies["AppError.MarshalJSON"] = raw

	rec := httptest.NewRecorder()
	response.WriteError(rec, e)
	bodies["WriteError"] = bytes.TrimRight(rec.Body.Bytes(), "\n")

	raw, _ = json.Marshal(response.ErrorResponse{
		Error: &response.APIError{
			Type:    string(e.Type),
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
			Param:   e.Param,
		},
	})
	bodies["ErrorResponse"] = raw

	for name, body := range bodies {
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("%s produced invalid JSON: %v", name, err)
		}
		if _, exists := parsed["success"]; exists {
			t.Errorf("%s emitted `success` key — Stripe/OpenAI contract violated. Body: %s",
				name, body)
		}
		if len(parsed) != 1 {
			t.Errorf("%s must have exactly one top-level key (\"error\"); got %d keys: %v",
				name, len(parsed), parsed)
		}
		if _, exists := parsed["error"]; !exists {
			t.Errorf("%s missing top-level `error` key. Body: %s", name, body)
		}
	}
}

// TestErrorShape_CodeOmittedWhenUnset pins the post-fix behaviour:
// when a call site does NOT set Code via WithCode, the wire envelope
// MUST omit the field (Stripe behaviour). Regression guard against
// the old `code defaults to type` anti-pattern.
func TestErrorShape_CodeOmittedWhenUnset(t *testing.T) {
	e := appErrors.NewUnauthorizedError("Authentication token required")

	rec := httptest.NewRecorder()
	response.WriteError(rec, e)

	var parsed struct {
		Error map[string]any `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed.Error["type"] != "authentication_error" {
		t.Errorf("type = %v, want authentication_error", parsed.Error["type"])
	}
	if _, hasCode := parsed.Error["code"]; hasCode {
		t.Errorf("code key must be omitted when unset, but got %q",
			parsed.Error["code"])
	}
}

