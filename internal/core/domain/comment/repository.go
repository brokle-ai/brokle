package comment

import (
	"context"

	"brokle/pkg/ulid"
)

type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByID(ctx context.Context, id ulid.ULID) (*Comment, error)
	GetByIDWithUser(ctx context.Context, id ulid.ULID) (*CommentWithUser, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id ulid.ULID) error

	// ListByEntity returns top-level comments (parent_id IS NULL) ordered by created_at ascending.
	ListByEntity(ctx context.Context, entityType EntityType, entityID string, projectID ulid.ULID) ([]*CommentWithUser, error)

	// ListReplies returns a map of parent_id -> replies.
	ListReplies(ctx context.Context, parentIDs []ulid.ULID) (map[string][]*CommentWithUser, error)

	// CountReplies returns a map of parent_id -> reply count.
	CountReplies(ctx context.Context, parentIDs []ulid.ULID) (map[string]int, error)

	// CountByEntity returns count of non-deleted comments (including replies).
	CountByEntity(ctx context.Context, entityType EntityType, entityID string, projectID ulid.ULID) (int64, error)

	HasActiveReplies(ctx context.Context, parentID ulid.ULID) (bool, error)
}
