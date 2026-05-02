package middleware

import (
	"bytes"
	"encoding/json"
	"testing"

	appErrors "brokle/pkg/errors"
)

// TestInternalErrorBody_ByteIdenticalToError pins the recoverer's
// hand-written 500 envelope to the wire shape Error.MarshalJSON
// emits for ReasonInternal errors. Drift is a real bug — clients
// switch on the envelope's type/code fields, so the recoverer body
// MUST be indistinguishable from the structured error path.
//
// This test sits alongside pkg/response/error_shape_test.go (which
// pins the Error → WriteError byte-identity). With this added, all
// paths that emit a Brokle error envelope to the wire are locked.
func TestInternalErrorBody_ByteIdenticalToError(t *testing.T) {
	t.Parallel()

	// The reference envelope: *Error with ReasonInternal and the
	// same message the recoverer writes. No Param, no Errors[],
	// no Details, no cause — the recoverer body is the bare-minimum
	// 500 shape (the panic itself is logged separately, never lifted
	// into the body).
	want := &appErrors.Error{
		Reason:  appErrors.ReasonInternal,
		Message: "Internal server error",
	}

	wantBytes, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("json.Marshal(*Error): %v", err)
	}

	if !bytes.Equal(wantBytes, []byte(internalErrorBody)) {
		t.Errorf("recoverer envelope drift:\n  Error.MarshalJSON: %s\n  internalErrorBody: %s",
			wantBytes, internalErrorBody)
	}
}
