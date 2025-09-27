package auth

import (
	"testing"

	"brokle/pkg/ulid"
)

func TestGenerateProjectScopedAPIKey(t *testing.T) {
	// Create a test project ID
	projectID := ulid.New()

	// Generate API key
	fullKey, keyID, secret, err := GenerateProjectScopedAPIKey(projectID)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// Validate key structure
	expectedKeyID := "bk_proj_" + projectID.String()
	if keyID != expectedKeyID {
		t.Errorf("Expected keyID %s, got %s", expectedKeyID, keyID)
	}

	// Validate secret length
	if len(secret) != APIKeySecretLength {
		t.Errorf("Expected secret length %d, got %d", APIKeySecretLength, len(secret))
	}

	// Validate full key format
	expectedFullKey := keyID + "_" + secret
	if fullKey != expectedFullKey {
		t.Errorf("Expected full key %s, got %s", expectedFullKey, fullKey)
	}

	t.Logf("Generated API key: %s", fullKey)
	t.Logf("Key ID: %s", keyID)
	t.Logf("Secret: %s", secret)
}

func TestParseAPIKey(t *testing.T) {
	// Create a test project ID
	projectID := ulid.New()

	// Generate API key
	fullKey, expectedKeyID, expectedSecret, err := GenerateProjectScopedAPIKey(projectID)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// Parse the key
	parsed, err := ParseAPIKey(fullKey)
	if err != nil {
		t.Fatalf("Failed to parse API key: %v", err)
	}

	// Validate parsed components
	if parsed.Prefix != APIKeyPrefix {
		t.Errorf("Expected prefix %s, got %s", APIKeyPrefix, parsed.Prefix)
	}

	if parsed.Scope != APIKeyScope {
		t.Errorf("Expected scope %s, got %s", APIKeyScope, parsed.Scope)
	}

	if parsed.ProjectID != projectID {
		t.Errorf("Expected project ID %s, got %s", projectID, parsed.ProjectID)
	}

	if parsed.Secret != expectedSecret {
		t.Errorf("Expected secret %s, got %s", expectedSecret, parsed.Secret)
	}

	if parsed.KeyID != expectedKeyID {
		t.Errorf("Expected keyID %s, got %s", expectedKeyID, parsed.KeyID)
	}

	t.Logf("Successfully parsed API key")
	t.Logf("Project ID: %s", parsed.ProjectID)
	t.Logf("Key ID: %s", parsed.KeyID)
}

func TestParseAPIKeyInvalidFormat(t *testing.T) {
	testCases := []struct {
		name   string
		key    string
		errMsg string
	}{
		{
			name:   "Too few parts",
			key:    "bk_proj_invalid",
			errMsg: "expected 4 parts",
		},
		{
			name:   "Invalid prefix",
			key:    "invalid_proj_01234567890123456789012345_secret123456789012345678901234567890",
			errMsg: "invalid API key prefix",
		},
		{
			name:   "Invalid scope",
			key:    "bk_invalid_01234567890123456789012345_secret123456789012345678901234567890",
			errMsg: "invalid API key scope",
		},
		{
			name:   "Invalid project ID",
			key:    "bk_proj_invalid_secret123456789012345678901234567890",
			errMsg: "invalid project ID",
		},
		{
			name:   "Invalid secret length",
			key:    "bk_proj_01234567890123456789012345_short",
			errMsg: "invalid secret length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseAPIKey(tc.key)
			if err == nil {
				t.Errorf("Expected error for invalid key: %s", tc.key)
			}
			if err != nil && !contains(err.Error(), tc.errMsg) {
				t.Errorf("Expected error message to contain '%s', got: %s", tc.errMsg, err.Error())
			}
		})
	}
}

func TestValidateAPIKeyFormat(t *testing.T) {
	// Generate a valid key
	projectID := ulid.New()
	validKey, _, _, err := GenerateProjectScopedAPIKey(projectID)
	if err != nil {
		t.Fatalf("Failed to generate valid key: %v", err)
	}

	// Test valid key
	if err := ValidateAPIKeyFormat(validKey); err != nil {
		t.Errorf("Valid key failed validation: %v", err)
	}

	// Test invalid keys
	invalidKeys := []string{
		"invalid",
		"bk_proj_invalid",
		"invalid_proj_01234567890123456789012345_secret123456789012345678901234567890",
		"bk_invalid_01234567890123456789012345_secret123456789012345678901234567890",
		"bk_proj_short_secret123456789012345678901234567890",
		"bk_proj_01234567890123456789012345_short",
	}

	for _, key := range invalidKeys {
		if err := ValidateAPIKeyFormat(key); err == nil {
			t.Errorf("Invalid key passed validation: %s", key)
		}
	}
}

func TestExtractProjectID(t *testing.T) {
	// Create a test project ID
	projectID := ulid.New()

	// Generate API key
	fullKey, _, _, err := GenerateProjectScopedAPIKey(projectID)
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	// Extract project ID
	extractedID, err := ExtractProjectID(fullKey)
	if err != nil {
		t.Fatalf("Failed to extract project ID: %v", err)
	}

	if extractedID != projectID {
		t.Errorf("Expected project ID %s, got %s", projectID, extractedID)
	}

	t.Logf("Successfully extracted project ID: %s", extractedID)
}

func TestCreateKeyPreview(t *testing.T) {
	testCases := []struct {
		name     string
		keyID    string
		expected string
	}{
		{
			name:     "Normal key ID",
			keyID:    "bk_proj_01234567890123456789012345",
			expected: "bk_proj_0123456789012345678901...2345",
		},
		{
			name:     "Short key ID",
			keyID:    "short",
			expected: "short...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			preview := CreateKeyPreview(tc.keyID)
			if preview != tc.expected {
				t.Errorf("Expected preview %s, got %s", tc.expected, preview)
			}
		})
	}
}

func TestIsProjectScopedKey(t *testing.T) {
	// Generate a valid project-scoped key
	projectID := ulid.New()
	validKey, _, _, err := GenerateProjectScopedAPIKey(projectID)
	if err != nil {
		t.Fatalf("Failed to generate valid key: %v", err)
	}

	// Test valid project-scoped key
	if !IsProjectScopedKey(validKey) {
		t.Error("Valid project-scoped key not recognized")
	}

	// Test invalid keys
	invalidKeys := []string{
		"bk_live_1234567890abcdef",
		"invalid_format",
		"bk_proj_short",
		"other_proj_01234567890123456789012345_secret123456789012345678901234567890",
	}

	for _, key := range invalidKeys {
		if IsProjectScopedKey(key) {
			t.Errorf("Invalid key incorrectly identified as project-scoped: %s", key)
		}
	}
}

func TestGetKeyScope(t *testing.T) {
	// Generate a valid project-scoped key
	projectID := ulid.New()
	validKey, _, _, err := GenerateProjectScopedAPIKey(projectID)
	if err != nil {
		t.Fatalf("Failed to generate valid key: %v", err)
	}

	// Test getting scope
	scope, err := GetKeyScope(validKey)
	if err != nil {
		t.Fatalf("Failed to get key scope: %v", err)
	}

	if scope != APIKeyScope {
		t.Errorf("Expected scope %s, got %s", APIKeyScope, scope)
	}

	// Test invalid key
	_, err = GetKeyScope("invalid")
	if err == nil {
		t.Error("Expected error for invalid key format")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 (len(s) > len(substr) && containsSubstring(s, substr)))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}