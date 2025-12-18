package playground

import (
	"context"

	"brokle/pkg/ulid"
)

// SessionRepository defines the repository interface for playground sessions.
type SessionRepository interface {
	// Create creates a new playground session.
	Create(ctx context.Context, session *Session) error

	// GetByID retrieves a session by its ID.
	// Returns ErrSessionNotFound if not found.
	GetByID(ctx context.Context, id ulid.ULID) (*Session, error)

	// List retrieves sessions for a project (for sidebar).
	// Ordered by last_used_at DESC.
	List(ctx context.Context, projectID ulid.ULID, limit int) ([]*Session, error)

	// ListByTags retrieves sessions filtered by tags.
	// Returns sessions where tags contains any of the provided tags.
	ListByTags(ctx context.Context, projectID ulid.ULID, tags []string, limit int) ([]*Session, error)

	// Update updates an existing session.
	// Updates updated_at automatically.
	Update(ctx context.Context, session *Session) error

	// UpdateLastRun updates only the last_run and last_used_at fields.
	UpdateLastRun(ctx context.Context, id ulid.ULID, lastRun JSON) error

	// UpdateWindows updates only the windows JSONB field.
	UpdateWindows(ctx context.Context, id ulid.ULID, windows JSON) error

	// Delete removes a session by ID.
	Delete(ctx context.Context, id ulid.ULID) error

	// Exists checks if a session exists.
	Exists(ctx context.Context, id ulid.ULID) (bool, error)

	// ExistsByProjectID checks if a session exists for a specific project.
	ExistsByProjectID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (bool, error)
}
