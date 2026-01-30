package evaluation

import (
	"context"

	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

type ScoreConfigRepository interface {
	Create(ctx context.Context, config *ScoreConfig) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*ScoreConfig, error)
	GetByName(ctx context.Context, projectID ulid.ULID, name string) (*ScoreConfig, error)
	List(ctx context.Context, projectID ulid.ULID, offset, limit int) ([]*ScoreConfig, int64, error)
	Update(ctx context.Context, config *ScoreConfig, projectID ulid.ULID) error
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error)
}

type DatasetRepository interface {
	Create(ctx context.Context, dataset *Dataset) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Dataset, error)
	GetByName(ctx context.Context, projectID ulid.ULID, name string) (*Dataset, error)
List(ctx context.Context, projectID ulid.ULID, filter *DatasetFilter, offset, limit int) ([]*Dataset, int64, error)
	ListWithFilters(ctx context.Context, projectID ulid.ULID, filter *DatasetFilter, params pagination.Params) ([]*DatasetWithItemCount, int64, error)
	Update(ctx context.Context, dataset *Dataset, projectID ulid.ULID) error
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error)
}

type DatasetItemRepository interface {
	Create(ctx context.Context, item *DatasetItem) error
	CreateBatch(ctx context.Context, items []*DatasetItem) error
	GetByID(ctx context.Context, id ulid.ULID, datasetID ulid.ULID) (*DatasetItem, error)
	GetByIDForProject(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*DatasetItem, error)
	List(ctx context.Context, datasetID ulid.ULID, limit, offset int) ([]*DatasetItem, int64, error)
	ListAll(ctx context.Context, datasetID ulid.ULID) ([]*DatasetItem, error)
	Delete(ctx context.Context, id ulid.ULID, datasetID ulid.ULID) error
	CountByDataset(ctx context.Context, datasetID ulid.ULID) (int64, error)
	FindByContentHash(ctx context.Context, datasetID ulid.ULID, contentHash string) (*DatasetItem, error)
	FindByContentHashes(ctx context.Context, datasetID ulid.ULID, contentHashes []string) (map[string]bool, error)
}

type DatasetVersionRepository interface {
	// Create creates a new dataset version
	Create(ctx context.Context, version *DatasetVersion) error
	// GetByID gets a version by its ID
	GetByID(ctx context.Context, id ulid.ULID, datasetID ulid.ULID) (*DatasetVersion, error)
	// GetByVersionNumber gets a version by dataset ID and version number
	GetByVersionNumber(ctx context.Context, datasetID ulid.ULID, versionNum int) (*DatasetVersion, error)
	// GetLatest gets the latest version for a dataset
	GetLatest(ctx context.Context, datasetID ulid.ULID) (*DatasetVersion, error)
	// List lists all versions for a dataset
	List(ctx context.Context, datasetID ulid.ULID) ([]*DatasetVersion, error)
	// GetNextVersionNumber returns the next version number for a dataset
	GetNextVersionNumber(ctx context.Context, datasetID ulid.ULID) (int, error)

	// Item-Version associations
	// AddItems associates items with a version (batch)
	AddItems(ctx context.Context, versionID ulid.ULID, itemIDs []ulid.ULID) error
	// GetItemIDs gets all item IDs for a version
	GetItemIDs(ctx context.Context, versionID ulid.ULID) ([]ulid.ULID, error)
	// GetItems gets all items for a version with pagination
	GetItems(ctx context.Context, versionID ulid.ULID, limit, offset int) ([]*DatasetItem, int64, error)
}

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *Experiment) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Experiment, error)
	List(ctx context.Context, projectID ulid.ULID, filter *ExperimentFilter, offset, limit int) ([]*Experiment, int64, error)
	Update(ctx context.Context, experiment *Experiment, projectID ulid.ULID) error
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error

	// Progress tracking methods
	// SetTotalItems sets the total number of items for an experiment
	SetTotalItems(ctx context.Context, id, projectID ulid.ULID, total int) error
	// IncrementCounters atomically increments completed and/or failed counters
	IncrementCounters(ctx context.Context, id, projectID ulid.ULID, completed, failed int) error
	// IncrementCountersAndUpdateStatus atomically increments counters and updates status if complete.
	// Returns true if the experiment was marked as complete (completed, failed, or partial).
	IncrementCountersAndUpdateStatus(ctx context.Context, id, projectID ulid.ULID, completed, failed int) (bool, error)
	// GetProgress gets minimal experiment data for progress polling
	GetProgress(ctx context.Context, id, projectID ulid.ULID) (*Experiment, error)
}

type ExperimentItemRepository interface {
	Create(ctx context.Context, item *ExperimentItem) error
	CreateBatch(ctx context.Context, items []*ExperimentItem) error
	List(ctx context.Context, experimentID ulid.ULID, limit, offset int) ([]*ExperimentItem, int64, error)
	CountByExperiment(ctx context.Context, experimentID ulid.ULID) (int64, error)
}

// ExperimentConfigRepository handles persistence for experiment configurations created via the wizard.
type ExperimentConfigRepository interface {
	// Create creates a new experiment config
	Create(ctx context.Context, config *ExperimentConfig) error
	// GetByID gets an experiment config by its ID
	GetByID(ctx context.Context, id ulid.ULID) (*ExperimentConfig, error)
	// GetByExperimentID gets the config for a specific experiment
	GetByExperimentID(ctx context.Context, experimentID ulid.ULID) (*ExperimentConfig, error)
	// Update updates an existing experiment config
	Update(ctx context.Context, config *ExperimentConfig) error
	// Delete deletes an experiment config
	Delete(ctx context.Context, id ulid.ULID) error
}

type RuleRepository interface {
	Create(ctx context.Context, rule *EvaluationRule) error
	Update(ctx context.Context, rule *EvaluationRule) error
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*EvaluationRule, error)
	GetByProjectID(ctx context.Context, projectID ulid.ULID, filter *RuleFilter, params pagination.Params) ([]*EvaluationRule, int64, error)
	GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*EvaluationRule, error)
	ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error)
}

type RuleExecutionRepository interface {
	Create(ctx context.Context, execution *RuleExecution) error
	Update(ctx context.Context, execution *RuleExecution) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*RuleExecution, error)
	GetByRuleID(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, filter *ExecutionFilter, params pagination.Params) ([]*RuleExecution, int64, error)
	GetLatestByRuleID(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID) (*RuleExecution, error)

	// IncrementCounters atomically increments spans_scored and errors_count counters
	IncrementCounters(ctx context.Context, id ulid.ULID, projectID ulid.ULID, spansScored, errorsCount int) error

	// IncrementCountersAndComplete atomically increments counters and marks execution as completed
	// if spans_scored + errors_count >= spans_matched. Returns true if execution was marked complete.
	IncrementCountersAndComplete(ctx context.Context, id ulid.ULID, projectID ulid.ULID, spansScored, errorsCount int) (bool, error)

	// UpdateSpansMatched updates only the spans_matched field for an execution.
	// Used by manual triggers after discovering how many spans will be processed.
	UpdateSpansMatched(ctx context.Context, id ulid.ULID, projectID ulid.ULID, spansMatched int) error
}
