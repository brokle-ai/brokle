package evaluation

import (
	"context"

	"brokle/pkg/ulid"
)

// ScoreConfigRepository defines the interface for score config data access.
// Implemented by PostgreSQL repository.
type ScoreConfigRepository interface {
	// Create creates a new score config
	Create(ctx context.Context, config *ScoreConfig) error

	// GetByID retrieves a score config by ID within a project
	GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*ScoreConfig, error)

	// GetByName retrieves a score config by name within a project
	// Returns nil, nil if not found (for uniqueness checks)
	GetByName(ctx context.Context, projectID ulid.ULID, name string) (*ScoreConfig, error)

	// List retrieves all score configs for a project
	List(ctx context.Context, projectID ulid.ULID) ([]*ScoreConfig, error)

	// Update updates an existing score config
	Update(ctx context.Context, config *ScoreConfig, projectID ulid.ULID) error

	// Delete permanently deletes a score config
	Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error

	// ExistsByName checks if a score config with the given name exists in the project
	ExistsByName(ctx context.Context, projectID ulid.ULID, name string) (bool, error)
}
