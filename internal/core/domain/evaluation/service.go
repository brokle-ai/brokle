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
}

type DatasetItemService interface {
	Create(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetItemRequest) (*DatasetItem, error)
	CreateBatch(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *CreateDatasetItemsBatchRequest) (int, error)
	List(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, limit, offset int) ([]*DatasetItem, int64, error)
	Delete(ctx context.Context, id ulid.ULID, datasetID ulid.ULID, projectID ulid.ULID) error
}

type ExperimentService interface {
	Create(ctx context.Context, projectID ulid.ULID, req *CreateExperimentRequest) (*Experiment, error)
	Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *UpdateExperimentRequest) (*Experiment, error)
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*Experiment, error)
	List(ctx context.Context, projectID ulid.ULID, filter *ExperimentFilter) ([]*Experiment, error)

	// CompareExperiments compares score metrics across multiple experiments
	CompareExperiments(ctx context.Context, projectID ulid.ULID, experimentIDs []ulid.ULID, baselineID *ulid.ULID) (*CompareExperimentsResponse, error)
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
}
