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
	List(ctx context.Context, projectID ulid.ULID) ([]*ScoreConfig, error)
	Update(ctx context.Context, config *ScoreConfig, projectID ulid.ULID) error
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error)
}

type DatasetRepository interface {
	Create(ctx context.Context, dataset *Dataset) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Dataset, error)
	GetByName(ctx context.Context, projectID ulid.ULID, name string) (*Dataset, error)
	List(ctx context.Context, projectID ulid.ULID) ([]*Dataset, error)
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

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *Experiment) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Experiment, error)
	List(ctx context.Context, projectID ulid.ULID, filter *ExperimentFilter) ([]*Experiment, error)
	Update(ctx context.Context, experiment *Experiment, projectID ulid.ULID) error
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
}

type ExperimentItemRepository interface {
	Create(ctx context.Context, item *ExperimentItem) error
	CreateBatch(ctx context.Context, items []*ExperimentItem) error
	List(ctx context.Context, experimentID ulid.ULID, limit, offset int) ([]*ExperimentItem, int64, error)
	CountByExperiment(ctx context.Context, experimentID ulid.ULID) (int64, error)
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
