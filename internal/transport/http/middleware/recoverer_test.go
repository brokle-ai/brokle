package middleware

import (
	"bytes"
	"encoding/json"
	"testing"

	appErrors "brokle/pkg/errors"
)

// TestInternalErrorBody_ByteIdenticalToAppError pins the recoverer's
// hand-written 500 envelope to the wire shape AppError.MarshalJSON
// emits for type=api_error errors. Drift is a real bug — clients
// switch on the envelope's type/code fields, so the recoverer body
// MUST be indistinguishable from the structured error path.
//
// This test sits alongside pkg/response/error_shape_test.go (which
// pins the AppError → WriteError → statusError byte-identity). With
// this added, all three paths that emit a Brokle error envelope to
// the wire are locked.
func TestInternalErrorBody_ByteIdenticalToAppError(t *testing.T) {
	t.Parallel()

	// The reference envelope: AppError of type api_error with the
	// same message the recoverer writes. No Param, no Errors[],
	// no Details, no cause — the recoverer body is the bare-minimum
	// 500 shape (the panic itself is logged separately, never lifted
	// into the body).
	want := appErrors.New(appErrors.TypeAPIError, "Internal server error")

	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("json.Marshal(AppError): %v", err)
	}

	if !bytes.Equal(wantBytes, []byte(internalErrorBody)) {
		t.Errorf("recoverer envelope drift:\n  AppError.MarshalJSON: %s\n  internalErrorBody:    %s",
			wantBytes, internalErrorBody)
	}
}
