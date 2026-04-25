package evaluation

import "errors"

var (
	ErrScoreConfigNotFound = errors.New("score config not found")
	ErrScoreConfigExists   = errors.New("score config with this name already exists")

	ErrDatasetNotFound = errors.New("dataset not found")
	ErrDatasetExists   = errors.New("dataset with this name already exists")

	ErrDatasetVersionNotFound = errors.New("dataset version not found")
	ErrDatasetVersionExists   = errors.New("dataset version already exists")

	ErrDatasetItemNotFound = errors.New("dataset item not found")

	ErrExperimentNotFound       = errors.New("experiment not found")
	ErrExperimentConfigNotFound = errors.New("experiment config not found")

	ErrEvaluatorNotFound = errors.New("evaluator not found")
	ErrEvaluatorExists   = errors.New("evaluator with this name already exists")

	ErrExecutionNotFound = errors.New("evaluator execution not found")
)
