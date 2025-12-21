package evaluation

import "errors"

var (
	ErrScoreConfigNotFound   = errors.New("score config not found")
	ErrScoreConfigExists     = errors.New("score config with this name already exists")
	ErrInvalidScoreConfigID  = errors.New("invalid score config ID")
	ErrScoreConfigValidation = errors.New("score config validation failed")

	ErrScoreValueOutOfRange = errors.New("score value out of configured range")
	ErrInvalidScoreCategory = errors.New("invalid category for categorical score")
	ErrScoreTypeMismatch    = errors.New("score type does not match config")

	ErrDatasetNotFound = errors.New("dataset not found")
	ErrDatasetExists   = errors.New("dataset with this name already exists")

	ErrDatasetItemNotFound = errors.New("dataset item not found")

	ErrExperimentNotFound     = errors.New("experiment not found")
	ErrExperimentItemNotFound = errors.New("experiment item not found")
)
