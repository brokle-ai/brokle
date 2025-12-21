package evaluation

import (
	"context"

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
	Delete(ctx context.Context, id ulid.ULID, datasetID ulid.ULID) error
	CountByDataset(ctx context.Context, datasetID ulid.ULID) (int64, error)
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
