package prompt

import (
	"errors"
	"fmt"
)

// Domain errors for prompt management.
//
// Sentinels are kept only when actively produced by a repository and
// consumed by a service via errors.Is. Same rule for factory helpers
// (NewXxxError) — they live here only when called from outside the
// domain package.
var (
	// Prompt CRUD
	ErrPromptNotFound      = errors.New("prompt not found")
	ErrPromptAlreadyExists = errors.New("prompt already exists")
	ErrInvalidPromptType   = errors.New("invalid prompt type")

	// Versioning
	ErrVersionNotFound = errors.New("version not found")

	// Labels
	ErrLabelNotFound               = errors.New("label not found")
	ErrLabelAlreadyExists          = errors.New("label already exists")
	ErrProtectedLabelAlreadyExists = errors.New("protected label already exists")

	// Templating (used by service-layer factory helpers below)
	ErrInvalidTemplate       = errors.New("invalid template")
	ErrInvalidTemplateFormat = errors.New("invalid template format")
	ErrVariableMissing       = errors.New("required variable missing")

	// Dialect compilation
	ErrUnsupportedDialect = errors.New("unsupported template dialect")
	ErrDialectCompilation = errors.New("template compilation failed")
	ErrTemplateTooLarge   = errors.New("template exceeds maximum size limit")

	// Cache
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheExpired  = errors.New("cache entry expired")
)

// Error codes surfaced to clients via the AppError envelope. Only codes
// with active call sites are kept; anything orphaned was deleted during
// the 2026-04-25 sentinel sweep.
const (
	ErrCodeTemplateTooLarge = "TEMPLATE_TOO_LARGE"
)

// Convenience constructors for contextualized errors.

func NewVariableMissingError(varName string) error {
	return fmt.Errorf("%w: {{%s}}", ErrVariableMissing, varName)
}

func NewInvalidTemplateError(details string) error {
	return fmt.Errorf("%w: %s", ErrInvalidTemplate, details)
}

func NewUnsupportedDialectError(dialect string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedDialect, dialect)
}

func NewDialectCompilationError(dialect, details string) error {
	return fmt.Errorf("%w [%s]: %s", ErrDialectCompilation, dialect, details)
}

func NewTemplateTooLargeError(size, maxSize int) error {
	return fmt.Errorf("%w: size %d exceeds limit %d", ErrTemplateTooLarge, size, maxSize)
}

// IsNotFoundError reports whether err is one of the prompt-domain
// not-found sentinels. Used by callers that need to translate any
// prompt-not-found variant to a 404 AppError without branching on
// each sentinel individually.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrPromptNotFound) ||
		errors.Is(err, ErrVersionNotFound) ||
		errors.Is(err, ErrLabelNotFound) ||
		errors.Is(err, ErrCacheNotFound)
}
