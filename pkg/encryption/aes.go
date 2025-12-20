// Package encryption provides AES-256-GCM encryption for sensitive data at rest.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrInvalidKey indicates the encryption key is not 32 bytes (256 bits).
	ErrInvalidKey = errors.New("encryption key must be 32 bytes for AES-256")
	// ErrInvalidCiphertext indicates the ciphertext format is invalid.
	ErrInvalidCiphertext = errors.New("invalid ciphertext format")
	// ErrDecryptionFailed indicates decryption failed (wrong key or corrupted data).
	ErrDecryptionFailed = errors.New("decryption failed: authentication tag mismatch")
)

// Service provides AES-256-GCM encryption and decryption.
// AES-256-GCM is an authenticated encryption algorithm that provides
// both confidentiality and integrity protection.
type Service struct {
	key []byte // 32 bytes for AES-256
}

// NewService creates a new encryption service with the provided key.
// Key must be exactly 32 bytes (256 bits) for AES-256.
func NewService(key []byte) (*Service, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, key)
	return &Service{key: keyCopy}, nil
}

// NewServiceFromBase64 creates a service from a base64-encoded key.
// The decoded key must be exactly 32 bytes.
func NewServiceFromBase64(keyBase64 string) (*Service, error) {
	if keyBase64 == "" {
		return nil, errors.New("encryption key cannot be empty")
	}
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, errors.New("invalid base64 encryption key: " + err.Error())
	}
	return NewService(key)
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Returns a base64-encoded string containing: nonce (12 bytes) || ciphertext || auth tag (16 bytes).
// Each encryption uses a unique random nonce, making it safe to encrypt the same plaintext multiple times.
func (s *Service) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext cannot be empty")
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a random 12-byte nonce (96 bits, standard for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal appends the ciphertext and auth tag to the nonce
	// Result format: nonce || ciphertext || tag
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext using AES-256-GCM.
// The ciphertext must be in the format: nonce (12 bytes) || ciphertext || auth tag (16 bytes).
// Returns the original plaintext if decryption succeeds.
func (s *Service) Decrypt(encryptedBase64 string) (string, error) {
	if encryptedBase64 == "" {
		return "", errors.New("ciphertext cannot be empty")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", errors.New("invalid base64 ciphertext: " + err.Error())
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Ciphertext must be at least nonce + tag bytes
	if len(ciphertext) < gcm.NonceSize()+gcm.Overhead() {
		return "", ErrInvalidCiphertext
	}

	// Extract nonce from the beginning
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	// Open decrypts and verifies the auth tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// GenerateKey generates a cryptographically secure 256-bit (32 byte) key.
// Use this to create a new encryption key for production use.
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateKeyBase64 generates a new key and returns it as a base64-encoded string.
// This is suitable for storing in environment variables or configuration files.
func GenerateKeyBase64() (string, error) {
	key, err := GenerateKey()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// MustGenerateKeyBase64 generates a new key and panics if generation fails.
// Use this only in init() or other startup code where failure is unrecoverable.
func MustGenerateKeyBase64() string {
	key, err := GenerateKeyBase64()
	if err != nil {
		panic("failed to generate encryption key: " + err.Error())
	}
	return key
}
