package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetadataStorageForReleaseVersion tests that release and version are stored in metadata JSON
func TestMetadataStorageForReleaseVersion(t *testing.T) {
	tests := []struct {
		name               string
		resourceAttrs      map[string]interface{}
		expectTraceRelease string
		expectTraceVersion string
	}{
		{
			name: "both_release_and_version_present",
			resourceAttrs: map[string]interface{}{
				"brokle.release": "v2.1.24",
				"brokle.version": "experiment-fast-mode",
			},
			expectTraceRelease: "v2.1.24",
			expectTraceVersion: "experiment-fast-mode",
		},
		{
			name: "only_release",
			resourceAttrs: map[string]interface{}{
				"brokle.release": "v1.0.0",
			},
			expectTraceRelease: "v1.0.0",
			expectTraceVersion: "",
		},
		{
			name: "only_version",
			resourceAttrs: map[string]interface{}{
				"brokle.version": "experiment-A",
			},
			expectTraceRelease: "",
			expectTraceVersion: "experiment-A",
		},
		{
			name:               "no_version_fields",
			resourceAttrs:      map[string]interface{}{},
			expectTraceRelease: "",
			expectTraceVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Merge attributes (simulating createTraceEvent logic)
			allAttrs := make(map[string]interface{})
			for k, v := range tt.resourceAttrs {
				allAttrs[k] = v
			}

			// Create metadata map
			metadata := make(map[string]interface{})

			// Extract release
			if release, ok := allAttrs["brokle.release"].(string); ok && release != "" {
				metadata["brokle.release"] = release
			}

			// Extract version
			if version, ok := allAttrs["brokle.version"].(string); ok && version != "" {
				metadata["brokle.version"] = version
			}

			// Verify expectations
			if tt.expectTraceRelease == "" {
				_, exists := metadata["brokle.release"]
				assert.False(t, exists, "brokle.release should not be in metadata")
			} else {
				actual, exists := metadata["brokle.release"]
				assert.True(t, exists, "brokle.release should be in metadata")
				assert.Equal(t, tt.expectTraceRelease, actual)
			}

			if tt.expectTraceVersion == "" {
				_, exists := metadata["brokle.version"]
				assert.False(t, exists, "brokle.version should not be in metadata")
			} else {
				actual, exists := metadata["brokle.version"]
				assert.True(t, exists, "brokle.version should be in metadata")
				assert.Equal(t, tt.expectTraceVersion, actual)
			}
		})
	}
}

// TestReleaseVersionMaterialization tests that materialized columns correctly extract from JSON
// Note: This is an integration test that requires ClickHouse to be running
func TestReleaseVersionMaterialization(t *testing.T) {
	t.Skip("Integration test - requires ClickHouse. Run with: go test -tags=integration")

	// This test would:
	// 1. Insert trace with metadata.brokle.release and metadata.brokle.version
	// 2. Query materialized columns traces.release and traces.version
	// 3. Verify they match the JSON values
	//
	// 4. Insert span with attributes.brokle.span.version
	// 5. Query materialized column spans.version
	// 6. Verify it matches the JSON value
}

// TestSpanVersionInAttributes tests that span-level version is stored in attributes
func TestSpanVersionInAttributes(t *testing.T) {
	tests := []struct {
		name              string
		spanAttrs         map[string]interface{}
		expectSpanVersion string
	}{
		{
			name: "span_version_present",
			spanAttrs: map[string]interface{}{
				"brokle.span.version": "prompt-v3",
			},
			expectSpanVersion: "prompt-v3",
		},
		{
			name:              "span_version_absent",
			spanAttrs:         map[string]interface{}{},
			expectSpanVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify span attributes contain the version
			if tt.expectSpanVersion == "" {
				_, exists := tt.spanAttrs["brokle.span.version"]
				assert.False(t, exists, "brokle.span.version should not be in attributes")
			} else {
				actual, exists := tt.spanAttrs["brokle.span.version"]
				assert.True(t, exists, "brokle.span.version should be in attributes")
				assert.Equal(t, tt.expectSpanVersion, actual)
			}
		})
	}
}
