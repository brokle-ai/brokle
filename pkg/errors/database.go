package errors

import (
	"errors"
	"strings"
)

// Database-level sentinel errors for common database constraints
var (
	// ErrUniqueConstraintViolation indicates a unique constraint was violated
	ErrUniqueConstraintViolation = errors.New("unique constraint violation")

	// ErrForeignKeyViolation indicates a foreign key constraint was violated
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// IsDatabaseUniqueViolation checks if error is from a unique constraint violation.
// This handles PostgreSQL, SQLite, and MySQL error formats.
func IsDatabaseUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check for wrapped sentinel error first (preferred method)
	if errors.Is(err, ErrUniqueConstraintViolation) {
		return true
	}

	// Fallback to string matching for raw database errors
	// This handles cases where the database driver returns raw errors
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "23505") // PostgreSQL error code for unique violation
}

// IsDatabaseForeignKeyViolation checks if error is from a foreign key constraint violation.
// This handles PostgreSQL, SQLite, and MySQL error formats.
func IsDatabaseForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}

	// Check for wrapped sentinel error first (preferred method)
	if errors.Is(err, ErrForeignKeyViolation) {
		return true
	}

	// Fallback to string matching for raw database errors
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "foreign key") ||
		strings.Contains(errMsg, "23503") // PostgreSQL error code for FK violation
}
