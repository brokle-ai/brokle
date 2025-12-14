package prompt

import (
	"errors"
	"fmt"
)

// Domain errors for prompt management
var (
	// Prompt errors
	ErrPromptNotFound      = errors.New("prompt not found")
	ErrPromptAlreadyExists = errors.New("prompt already exists")
	ErrInvalidPromptName   = errors.New("invalid prompt name")
	ErrInvalidPromptType   = errors.New("invalid prompt type")

	// Version errors
	ErrVersionNotFound       = errors.New("version not found")
	ErrVersionImmutable      = errors.New("versions are immutable and cannot be modified")
	ErrInvalidVersionNumber  = errors.New("invalid version number")

	// Label errors
	ErrLabelNotFound         = errors.New("label not found")
	ErrLabelProtected        = errors.New("label is protected")
	ErrLabelAlreadyExists    = errors.New("label already exists")
	ErrInvalidLabelName      = errors.New("invalid label name")
	ErrLatestLabelReserved   = errors.New("'latest' label is auto-managed and cannot be modified")

	// Template errors
	ErrInvalidTemplate       = errors.New("invalid template")
	ErrInvalidTemplateFormat = errors.New("invalid template format")
	ErrVariableMissing       = errors.New("required variable missing")
	ErrInvalidVariableName   = errors.New("invalid variable name")

	// Cache errors
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheExpired  = errors.New("cache entry expired")

	// Execution errors
	ErrExecutionFailed    = errors.New("prompt execution failed")
	ErrProviderNotFound   = errors.New("LLM provider not found")
	ErrInvalidModelConfig = errors.New("invalid model configuration")
)

// Error codes for structured API responses
const (
	ErrCodePromptNotFound      = "PROMPT_NOT_FOUND"
	ErrCodePromptAlreadyExists = "PROMPT_ALREADY_EXISTS"
	ErrCodeVersionNotFound     = "VERSION_NOT_FOUND"
	ErrCodeLabelNotFound       = "LABEL_NOT_FOUND"
	ErrCodeLabelProtected      = "LABEL_PROTECTED"
	ErrCodeInvalidTemplate     = "INVALID_TEMPLATE"
	ErrCodeVariableMissing     = "VARIABLE_MISSING"
	ErrCodeExecutionFailed     = "EXECUTION_FAILED"
)

// Convenience functions for creating contextualized errors

// NewPromptNotFoundError creates a prompt not found error with name context.
func NewPromptNotFoundError(name string) error {
	return fmt.Errorf("%w: %s", ErrPromptNotFound, name)
}

// NewPromptNotFoundByIDError creates a prompt not found error with ID context.
func NewPromptNotFoundByIDError(id string) error {
	return fmt.Errorf("%w: id=%s", ErrPromptNotFound, id)
}

// NewPromptAlreadyExistsError creates a prompt already exists error.
func NewPromptAlreadyExistsError(name, projectID string) error {
	return fmt.Errorf("%w: %s in project %s", ErrPromptAlreadyExists, name, projectID)
}

// NewVersionNotFoundError creates a version not found error.
func NewVersionNotFoundError(promptName string, version int) error {
	return fmt.Errorf("%w: %s version %d", ErrVersionNotFound, promptName, version)
}

// NewLabelNotFoundError creates a label not found error.
func NewLabelNotFoundError(promptName, labelName string) error {
	return fmt.Errorf("%w: %s label '%s'", ErrLabelNotFound, promptName, labelName)
}

// NewLabelProtectedError creates a label protected error.
func NewLabelProtectedError(labelName string) error {
	return fmt.Errorf("%w: '%s' requires admin permission", ErrLabelProtected, labelName)
}

// NewVariableMissingError creates a variable missing error.
func NewVariableMissingError(varName string) error {
	return fmt.Errorf("%w: {{%s}}", ErrVariableMissing, varName)
}

// NewInvalidTemplateError creates an invalid template error with details.
func NewInvalidTemplateError(details string) error {
	return fmt.Errorf("%w: %s", ErrInvalidTemplate, details)
}

// NewExecutionFailedError creates an execution failed error.
func NewExecutionFailedError(details string) error {
	return fmt.Errorf("%w: %s", ErrExecutionFailed, details)
}

// Error classification helpers

// IsNotFoundError checks if the error is any kind of not-found error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrPromptNotFound) ||
		errors.Is(err, ErrVersionNotFound) ||
		errors.Is(err, ErrLabelNotFound) ||
		errors.Is(err, ErrCacheNotFound)
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidPromptName) ||
		errors.Is(err, ErrInvalidPromptType) ||
		errors.Is(err, ErrInvalidLabelName) ||
		errors.Is(err, ErrInvalidTemplate) ||
		errors.Is(err, ErrInvalidTemplateFormat) ||
		errors.Is(err, ErrVariableMissing) ||
		errors.Is(err, ErrInvalidVariableName) ||
		errors.Is(err, ErrInvalidModelConfig)
}

// IsConflictError checks if the error is a conflict error.
func IsConflictError(err error) bool {
	return errors.Is(err, ErrPromptAlreadyExists) ||
		errors.Is(err, ErrLabelAlreadyExists)
}

// IsForbiddenError checks if the error is a forbidden/permission error.
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrLabelProtected) ||
		errors.Is(err, ErrLatestLabelReserved) ||
		errors.Is(err, ErrVersionImmutable)
}
