package utils

import (
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // MD5 supported for legacy compatibility only
	cryptoRand "crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA1 supported for legacy compatibility only
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"time"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

// HashAlgorithm represents different hashing algorithms
type HashAlgorithm string

const (
	AlgorithmMD5    HashAlgorithm = "md5"
	AlgorithmSHA1   HashAlgorithm = "sha1"
	AlgorithmSHA256 HashAlgorithm = "sha256"
	AlgorithmSHA512 HashAlgorithm = "sha512"
)

// HashPassword hashes a password using bcrypt with default cost
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword validates a password against its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := cryptoRand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateHexToken generates a cryptographically secure random token in hex format
func GenerateHexToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := cryptoRand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate hex token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateAPIKey generates a prefixed API key with ULID
func GenerateAPIKey(prefix string) (string, error) {
	if prefix == "" {
		prefix = "bk"
	}

	id := ulid.MustNew(ulid.Timestamp(time.Now()), cryptoRand.Reader)

	// Create API key format: prefix_env_ulid (e.g., bk_live_01ARZ3NDEKTSV4RRFFQ69G5FAV)
	return fmt.Sprintf("%s_live_%s", prefix, id.String()), nil
}

// Hash computes a hash using the specified algorithm
func Hash(data []byte, algorithm HashAlgorithm) (string, error) {
	var hasher hash.Hash

	switch algorithm {
	case AlgorithmMD5:
		hasher = md5.New()
	case AlgorithmSHA1:
		hasher = sha1.New()
	case AlgorithmSHA256:
		hasher = sha256.New()
	case AlgorithmSHA512:
		hasher = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashString computes a hash of a string using the specified algorithm
func HashString(data string, algorithm HashAlgorithm) (string, error) {
	return Hash([]byte(data), algorithm)
}

// HMAC computes an HMAC using the specified algorithm
func HMAC(data []byte, key []byte, algorithm HashAlgorithm) (string, error) {
	var hasher hash.Hash

	switch algorithm {
	case AlgorithmSHA1:
		hasher = hmac.New(sha1.New, key)
	case AlgorithmSHA256:
		hasher = hmac.New(sha256.New, key)
	case AlgorithmSHA512:
		hasher = hmac.New(sha512.New, key)
	default:
		return "", fmt.Errorf("unsupported HMAC algorithm: %s", algorithm)
	}

	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HMACString computes an HMAC of a string using the specified algorithm
func HMACString(data string, key string, algorithm HashAlgorithm) (string, error) {
	return HMAC([]byte(data), []byte(key), algorithm)
}

// GenerateHMAC generates an HMAC-SHA256 signature for webhook validation
func GenerateHMAC(data []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateHMAC validates an HMAC-SHA256 signature
func ValidateHMAC(data []byte, secret string, expectedMAC string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	expectedBytes, err := hex.DecodeString(expectedMAC)
	if err != nil {
		return false
	}
	return hmac.Equal(mac.Sum(nil), expectedBytes)
}

// EncodeBase64 encodes data to base64
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes base64 data
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// EncodeBase64URL encodes data to base64 URL encoding
func EncodeBase64URL(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeBase64URL decodes base64 URL encoded data
func DecodeBase64URL(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}

// GenerateSalt generates a cryptographic salt
func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := cryptoRand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// SecureRandomBytes generates cryptographically secure random bytes
func SecureRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(cryptoRand.Reader, bytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure random bytes: %w", err)
	}
	return bytes, nil
}

// CompareHashes performs a constant-time comparison of two hashes
func CompareHashes(hash1, hash2 string) bool {
	return hmac.Equal([]byte(hash1), []byte(hash2))
}

// GenerateNonce generates a cryptographic nonce
func GenerateNonce() (string, error) {
	return GenerateSecureToken(16)
}

// HashAPIKey creates a consistent hash of an API key for storage
func HashAPIKey(apiKey string) (string, error) {
	return HashString(apiKey, AlgorithmSHA256)
}

// ULID Generation Functions

// GenerateULID generates a new ULID
func GenerateULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), cryptoRand.Reader).String()
}

// GenerateULIDWithTime generates a ULID with specific timestamp
func GenerateULIDWithTime(t time.Time) string {
	return ulid.MustNew(ulid.Timestamp(t), cryptoRand.Reader).String()
}

// ParseULID parses a ULID string and returns the ULID
func ParseULID(s string) (ulid.ULID, error) {
	return ulid.Parse(s)
}

// ULIDTime extracts the timestamp from a ULID
func ULIDTime(id string) (time.Time, error) {
	parsed, err := ulid.Parse(id)
	if err != nil {
		return time.Time{}, err
	}
	return ulid.Time(parsed.Time()), nil
}

// GenerateTestAPIKey generates a test API key with ULID
func GenerateTestAPIKey(prefix string) (string, error) {
	if prefix == "" {
		prefix = "bk"
	}

	id := ulid.MustNew(ulid.Timestamp(time.Now()), cryptoRand.Reader)

	return fmt.Sprintf("%s_test_%s", prefix, id.String()), nil
}
