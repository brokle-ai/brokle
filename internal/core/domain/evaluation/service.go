package evaluation

import (
	"context"

	"brokle/pkg/pagination"
	"brokle/pkg/ulid"
)

type ScoreConfigService interface {
	Create(ctx context.Context, projectID ulid.ULID, req *CreateScoreConfigRequest) (*ScoreConfig, error)
	Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *UpdateScoreConfigRequest) (*ScoreConfig, error)
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*ScoreConfig, error)
	GetByName(ctx context.Context, projectID ulid.ULID, name string) (*ScoreConfig, error)
	List(ctx context.Context, projectID ulid.ULID) ([]*ScoreConfig, error)
}

type DatasetService interface {
	Create(ctx context.Context, projectID ulid.ULID, req *CreateDatasetRequest) (*Dataset, error)
	Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *UpdateDatasetRequest) (*Dataset, error)
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Dataset, error)
	List(ctx context.Context, projectID ulid.ULID) ([]*Dataset, error)
	ListWithFilters(ctx context.Context, projectID ulid.ULID, filter *DatasetFilter, params pagination.Params) ([]*DatasetWithItemCount, int64, error)
}

type DatasetItemService interface {
	Create(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetItemRequest) (*DatasetItem, error)
	CreateBatch(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetItemsBatchRequest) (int, error)
	List(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, limit, offset int) ([]*DatasetItem, int64, error)
	Delete(ctx context.Context, id ulid.ULID, datasetID ulid.ULID, projectID ulid.ULID) error

	// Bulk import methods
	ImportFromJSON(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *ImportDatasetItemsFromJSONRequest) (*BulkImportResult, error)
	ImportFromCSV(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *ImportDatasetItemsFromCSVRequest) (*BulkImportResult, error)
	CreateFromTraces(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetItemsFromTracesRequest) (*BulkImportResult, error)
	CreateFromSpans(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetItemsFromSpansRequest) (*BulkImportResult, error)

	// Export method
	ExportItems(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID) ([]*DatasetItem, error)
}

type DatasetVersionService interface {
	// CreateVersion creates a new version snapshot of the current dataset items
	CreateVersion(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetVersionRequest) (*DatasetVersion, error)
	// GetVersion gets a specific version by ID
	GetVersion(ctx context.Context, versionID ulid.ULID, datasetID ulid.ULID, projectID ulid.ULID) (*DatasetVersion, error)
	// ListVersions lists all versions for a dataset
	ListVersions(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID) ([]*DatasetVersion, error)
	// GetLatestVersion gets the most recent version
	GetLatestVersion(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID) (*DatasetVersion, error)
	// GetVersionItems gets items for a specific version with pagination
	GetVersionItems(ctx context.Context, versionID ulid.ULID, datasetID ulid.ULID, projectID ulid.ULID, limit, offset int) ([]*DatasetItem, int64, error)
	// PinVersion pins the dataset to a specific version (nil to unpin)
	PinVersion(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, versionID *ulid.ULID) (*Dataset, error)
	// GetDatasetWithVersionInfo gets a dataset with its version information
	GetDatasetWithVersionInfo(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID) (*DatasetWithVersionResponse, error)
}

type ExperimentService interface {
	Create(ctx context.Context, projectID ulid.ULID, req *CreateExperimentRequest) (*Experiment, error)
	Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *UpdateExperimentRequest) (*Experiment, error)
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Experiment, error)
	List(ctx context.Context, projectID ulid.ULID, filter *ExperimentFilter) ([]*Experiment, error)

	// CompareExperiments compares score metrics across multiple experiments
	CompareExperiments(ctx context.Context, projectID ulid.ULID, experimentIDs []ulid.ULID, baselineID *ulid.ULID) (*CompareExperimentsResponse, error)

	// Rerun creates a new experiment based on an existing one, using the same dataset.
	// The new experiment starts in pending status ready for SDK to run with a new task function.
	Rerun(ctx context.Context, sourceID ulid.ULID, projectID ulid.ULID, req *RerunExperimentRequest) (*Experiment, error)
}

type ExperimentItemService interface {
	CreateBatch(ctx context.Context, experimentID ulid.ULID, projectID ulid.ULID, req *CreateExperimentItemsBatchRequest) (int, error)
	List(ctx context.Context, experimentID ulid.ULID, projectID ulid.ULID, limit, offset int) ([]*ExperimentItem, int64, error)
}

type RuleService interface {
	Create(ctx context.Context, projectID ulid.ULID, userID *ulid.ULID, req *CreateEvaluationRuleRequest) (*EvaluationRule, error)
	Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *UpdateEvaluationRuleRequest) (*EvaluationRule, error)
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*EvaluationRule, error)
	List(ctx context.Context, projectID ulid.ULID, filter *RuleFilter, params pagination.Params) ([]*EvaluationRule, int64, error)
	Activate(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	Deactivate(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetActiveByProjectID(ctx context.Context, projectID ulid.ULID) ([]*EvaluationRule, error)

	// TriggerRule starts a manual evaluation of the rule against matching spans
	TriggerRule(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, opts *TriggerOptions) (*TriggerResponse, error)
}

type RuleExecutionService interface {
	StartExecution(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, triggerType TriggerType) (*RuleExecution, error)
	CompleteExecution(ctx context.Context, executionID ulid.ULID, projectID ulid.ULID, spansMatched, spansScored, errorsCount int) error
	FailExecution(ctx context.Context, executionID ulid.ULID, projectID ulid.ULID, errorMessage string) error
	CancelExecution(ctx context.Context, executionID ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*RuleExecution, error)
	ListByRuleID(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, filter *ExecutionFilter, params pagination.Params) ([]*RuleExecution, int64, error)
	GetLatestByRuleID(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID) (*RuleExecution, error)

	// IncrementCounters atomically increments spans_scored and errors_count for an execution (used by workers)
	IncrementCounters(ctx context.Context, executionID string, projectID ulid.ULID, spansScored, errorsCount int) error

	// StartExecutionWithCount creates an execution with known spans_matched count upfront.
	// Used for automatic evaluations where we know the count before emitting jobs.
	StartExecutionWithCount(ctx context.Context, ruleID ulid.ULID, projectID ulid.ULID, triggerType TriggerType, spansMatched int) (*RuleExecution, error)

	// IncrementAndCheckCompletion atomically increments counters and marks execution as complete
	// if all spans have been processed. Returns true if execution was marked complete.
	IncrementAndCheckCompletion(ctx context.Context, executionID ulid.ULID, projectID ulid.ULID, spansScored, errorsCount int) (bool, error)

	// UpdateSpansMatched updates the spans_matched count for an execution.
	// Used by manual triggers after discovering how many spans will be processed.
	UpdateSpansMatched(ctx context.Context, executionID ulid.ULID, projectID ulid.ULID, spansMatched int) error
}
