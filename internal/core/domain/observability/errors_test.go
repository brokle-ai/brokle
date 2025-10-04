package observability

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// HIGH-VALUE TESTS: Error Wrapping and Chaining
// ============================================================================

// TestErrorWrapping tests that errors can be unwrapped correctly
func TestErrorWrapping(t *testing.T) {
	originalErr := fmt.Errorf("database connection failed")
	wrappedErr := NewObservabilityErrorWithCause(
		ErrCodeBatchOperationFailed,
		"failed to process batch",
		originalErr,
	)

	// Test that we can unwrap to get the original error
	assert.True(t, errors.Is(wrappedErr, originalErr))

	// Test that error message includes both wrapped and original
	assert.Contains(t, wrappedErr.Error(), "failed to process batch")
	assert.Contains(t, wrappedErr.Error(), "database connection failed")
}

// TestErrorChaining tests multi-level error wrapping
func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	dbErr := fmt.Errorf("connection timeout")
	repoErr := NewObservabilityErrorWithCause(ErrCodeBatchOperationFailed, "repository error", dbErr)
	serviceErr := NewObservabilityErrorWithCause(ErrCodeBatchOperationFailed, "service error", repoErr)

	// Verify we can unwrap through the chain
	assert.True(t, errors.Is(serviceErr, repoErr))
	assert.True(t, errors.Is(serviceErr, dbErr))
}

// TestObservabilityError_WithDetail tests detail accumulation
func TestObservabilityError_WithDetail(t *testing.T) {
	err := NewObservabilityError("TEST_CODE", "test message")

	// Add first detail
	err.WithDetail("key1", "value1")
	assert.Equal(t, "value1", err.Details["key1"])

	// Add second detail
	err.WithDetail("key2", 123)
	assert.Equal(t, "value1", err.Details["key1"])
	assert.Equal(t, 123, err.Details["key2"])

	// Add detail with nil Details (should initialize)
	err2 := &ObservabilityError{
		Code:    "TEST",
		Message: "test",
		Details: nil,
	}
	err2.WithDetail("key", "value")
	assert.NotNil(t, err2.Details)
	assert.Equal(t, "value", err2.Details["key"])
}
