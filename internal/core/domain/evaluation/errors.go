package evaluation

import "errors"

// Domain errors for evaluation operations
var (
	// ScoreConfig errors
	ErrScoreConfigNotFound   = errors.New("score config not found")
	ErrScoreConfigExists     = errors.New("score config with this name already exists")
	ErrInvalidScoreConfigID  = errors.New("invalid score config ID")
	ErrScoreConfigValidation = errors.New("score config validation failed")

	// Score validation errors (for SDK score ingestion)
	ErrScoreValueOutOfRange = errors.New("score value out of configured range")
	ErrInvalidScoreCategory = errors.New("invalid category for categorical score")
	ErrScoreTypeMismatch    = errors.New("score type does not match config")
)

// Error codes for API responses
const (
	ErrCodeScoreConfigNotFound   = "SCORE_CONFIG_NOT_FOUND"
	ErrCodeScoreConfigExists     = "SCORE_CONFIG_ALREADY_EXISTS"
	ErrCodeInvalidScoreConfig    = "INVALID_SCORE_CONFIG"
	ErrCodeScoreValidationFailed = "SCORE_VALIDATION_FAILED"
)
