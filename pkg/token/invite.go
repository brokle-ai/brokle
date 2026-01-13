// Package token provides secure token generation and hashing utilities.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	// InviteTokenPrefix is the prefix for invitation tokens
	InviteTokenPrefix = "inv_"
	// InviteTokenBytes is the number of random bytes for invite tokens (256 bits)
	InviteTokenBytes = 32
	// InvitePreviewLength is the length of the token preview (including prefix)
	InvitePreviewLength = 12
)

// InviteToken represents a generated invitation token with its hash and preview
type InviteToken struct {
	// Token is the full plaintext token (send to user, never store)
	Token string
	// Hash is the SHA-256 hash of the token (store in database)
	Hash string
	// Preview is the first characters for display purposes (e.g., "inv_AbCd...")
	Preview string
}

// GenerateInviteToken generates a new secure invitation token
// Returns the plaintext token, its hash, and a preview for display
func GenerateInviteToken() (*InviteToken, error) {
	// Generate random bytes
	bytes := make([]byte, InviteTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to hex and add prefix
	token := InviteTokenPrefix + hex.EncodeToString(bytes)

	// Generate SHA-256 hash
	hash := HashToken(token)

	// Create preview (first N characters)
	preview := token[:InvitePreviewLength] + "..."

	return &InviteToken{
		Token:   token,
		Hash:    hash,
		Preview: preview,
	}, nil
}

// HashToken generates a SHA-256 hash of the given token
func HashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateTokenFormat checks if a token has the expected format
func ValidateTokenFormat(token string) bool {
	// Token should be prefix + 64 hex chars (32 bytes)
	expectedLen := len(InviteTokenPrefix) + (InviteTokenBytes * 2)
	if len(token) != expectedLen {
		return false
	}
	if token[:len(InviteTokenPrefix)] != InviteTokenPrefix {
		return false
	}
	// Check if remainder is valid hex
	_, err := hex.DecodeString(token[len(InviteTokenPrefix):])
	return err == nil
}
