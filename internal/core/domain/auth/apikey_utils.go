package auth

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// API Key format constants (industry-standard pure random)
const (
	APIKeyPrefix       = "bk"
	APIKeySecretLength = 40 // Industry standard (GitHub: 40, Stripe: 32-40)
	APIKeySeparator    = "_"
)

// GenerateAPIKey generates an industry-standard pure random API key.
// Format: bk_{40_char_random_secret}
// Returns: fullKey, error
//
// Example: bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd
func GenerateAPIKey() (string, error) {
	// Generate cryptographically secure random secret
	secret, err := generateSecureSecret(APIKeySecretLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate API key secret: %w", err)
	}

	// Format: bk_{secret}
	fullKey := fmt.Sprintf("%s%s%s", APIKeyPrefix, APIKeySeparator, secret)

	return fullKey, nil
}

// ValidateAPIKeyFormat validates the format of an API key.
// Expected format: bk_{40_chars}
func ValidateAPIKeyFormat(fullKey string) error {
	parts := strings.Split(fullKey, APIKeySeparator)
	if len(parts) != 2 {
		return fmt.Errorf("invalid API key format: expected 2 parts separated by underscore (bk_{secret})")
	}

	if parts[0] != APIKeyPrefix {
		return fmt.Errorf("invalid API key prefix: expected %s", APIKeyPrefix)
	}

	if len(parts[1]) != APIKeySecretLength {
		return fmt.Errorf("invalid secret length: expected %d, got %d", APIKeySecretLength, len(parts[1]))
	}

	return nil
}

// CreateKeyPreview creates a preview version of a full API key for display purposes.
// Input: bk_{40_char_secret}
// Output: bk_xxxx...yyyy (prefix + first 4 chars + ... + last 4 chars)
//
// Example: bk_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCd -> bk_AbCd...AbCd
// This follows GitHub's exact pattern for API key previews
func CreateKeyPreview(fullKey string) string {
	if len(fullKey) <= 11 {
		return fullKey + "..."
	}
	// Show: bk_ + first 4 chars of secret + ... + last 4 chars
	// Example: bk_rvOJ...yym0
	return fullKey[:7] + "..." + fullKey[len(fullKey)-4:]
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
