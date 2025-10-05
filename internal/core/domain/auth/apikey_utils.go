package auth

import (
	"crypto/rand"
	"fmt"
	"strings"

	"brokle/pkg/ulid"
)

// API Key format constants
const (
	APIKeyPrefix        = "bk"
	APIKeyScope         = "proj"
	APIKeySecretLength  = 32
	APIKeySeparator     = "_"
)

// GenerateProjectScopedAPIKey generates a new project-scoped API key.
// Returns: fullKey, keyID, secret, error
func GenerateProjectScopedAPIKey(projectID ulid.ULID) (string, string, string, error) {
	// Create key ID: bk_proj_{project_id}
	keyID := fmt.Sprintf("%s%s%s%s%s",
		APIKeyPrefix, APIKeySeparator,
		APIKeyScope, APIKeySeparator,
		projectID.String())

	// Generate secure secret
	secret, err := generateSecureSecret(APIKeySecretLength)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate secret: %w", err)
	}

	// Combine: bk_proj_{project_id}_{secret}
	fullKey := fmt.Sprintf("%s%s%s", keyID, APIKeySeparator, secret)

	return fullKey, keyID, secret, nil
}

// ParseAPIKey parses a full API key into its components.
func ParseAPIKey(fullKey string) (*ParsedAPIKey, error) {
	// Split by underscore: [bk, proj, {project_id}, {secret}]
	parts := strings.Split(fullKey, APIKeySeparator)
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid API key format: expected 4 parts, got %d", len(parts))
	}

	prefix := parts[0]       // "bk"
	scope := parts[1]        // "proj"
	projectIDStr := parts[2] // project ULID
	secret := parts[3]       // secret

	// Validate prefix
	if prefix != APIKeyPrefix {
		return nil, fmt.Errorf("invalid API key prefix: expected %s, got %s", APIKeyPrefix, prefix)
	}

	// Validate scope
	if scope != APIKeyScope {
		return nil, fmt.Errorf("invalid API key scope: expected %s, got %s", APIKeyScope, scope)
	}

	// Parse project ID
	projectID, err := ulid.Parse(projectIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID in API key: %w", err)
	}

	// Validate secret length
	if len(secret) != APIKeySecretLength {
		return nil, fmt.Errorf("invalid secret length: expected %d, got %d", APIKeySecretLength, len(secret))
	}

	// Construct keyID
	keyID := fmt.Sprintf("%s%s%s%s%s", prefix, APIKeySeparator, scope, APIKeySeparator, projectIDStr)

	return &ParsedAPIKey{
		Prefix:    prefix,
		Scope:     scope,
		ProjectID: projectID,
		Secret:    secret,
		KeyID:     keyID,
	}, nil
}

// ValidateAPIKeyFormat validates the format of an API key without parsing project ID.
func ValidateAPIKeyFormat(fullKey string) error {
	parts := strings.Split(fullKey, APIKeySeparator)
	if len(parts) != 4 {
		return fmt.Errorf("invalid API key format: expected 4 parts separated by underscore")
	}

	if parts[0] != APIKeyPrefix {
		return fmt.Errorf("invalid API key prefix")
	}

	if parts[1] != APIKeyScope {
		return fmt.Errorf("invalid API key scope")
	}

	if len(parts[2]) != 26 { // ULID length
		return fmt.Errorf("invalid project ID length")
	}

	if len(parts[3]) != APIKeySecretLength {
		return fmt.Errorf("invalid secret length")
	}

	return nil
}

// ExtractProjectIDFromFullKey extracts the project ID from a full API key.
// Format: bk_proj_{project_id}_{secret}
func ExtractProjectIDFromFullKey(fullKey string) (ulid.ULID, error) {
	parts := strings.Split(fullKey, APIKeySeparator)
	if len(parts) != 4 {
		return ulid.ULID{}, fmt.Errorf("invalid API key format: expected 4 parts")
	}

	if parts[0] != APIKeyPrefix || parts[1] != APIKeyScope {
		return ulid.ULID{}, fmt.Errorf("invalid API key prefix/scope")
	}

	return ulid.Parse(parts[2])
}

// CreateKeyPreview creates a preview version of a full API key for display purposes.
// Input: bk_proj_{project_id}_{secret} (full key)
// Output: bk_proj_...WXYZ (first 8 chars + ... + last 4 chars)
func CreateKeyPreview(fullKey string) string {
	if len(fullKey) <= 12 {
		return fullKey + "..."
	}
	// Show first 8 characters (bk_proj_) and last 4 characters of full key
	return fullKey[:8] + "..." + fullKey[len(fullKey)-4:]
}

// generateSecureSecret generates a cryptographically secure random string.
func generateSecureSecret(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

// IsProjectScopedKey checks if a key follows the project-scoped format.
func IsProjectScopedKey(fullKey string) bool {
	parts := strings.Split(fullKey, APIKeySeparator)
	return len(parts) == 4 &&
		   parts[0] == APIKeyPrefix &&
		   parts[1] == APIKeyScope
}

// GetKeyScope returns the scope of an API key (currently always "proj").
func GetKeyScope(fullKey string) (string, error) {
	parts := strings.Split(fullKey, APIKeySeparator)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid API key format")
	}
	return parts[1], nil
}