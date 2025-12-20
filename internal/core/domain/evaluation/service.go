package evaluation

import (
	"context"

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
