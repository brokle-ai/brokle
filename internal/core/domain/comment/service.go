package comment

import (
	"context"

	"brokle/pkg/ulid"
)

type Service interface {
	CreateComment(ctx context.Context, projectID ulid.ULID, traceID string, userID ulid.ULID, req *CreateCommentRequest) (*CommentResponse, error)

	// UpdateComment - only the comment owner can update their comment.
	UpdateComment(ctx context.Context, projectID ulid.ULID, traceID string, commentID, userID ulid.ULID, req *UpdateCommentRequest) (*CommentResponse, error)

	// DeleteComment - only the comment owner can delete their comment.
	DeleteComment(ctx context.Context, projectID ulid.ULID, traceID string, commentID, userID ulid.ULID) error

	// ListComments returns comments ordered by created_at ascending.
	// currentUserID is used to set the HasUser flag on reaction summaries.
	ListComments(ctx context.Context, projectID ulid.ULID, traceID string, currentUserID *ulid.ULID) (*ListCommentsResponse, error)

	GetCommentCount(ctx context.Context, projectID ulid.ULID, traceID string) (*CommentCountResponse, error)

	// ToggleReaction enforces max 6 unique emoji types per comment.
	ToggleReaction(ctx context.Context, projectID ulid.ULID, traceID string, commentID, userID ulid.ULID, req *ToggleReactionRequest) ([]ReactionSummary, error)

	// CreateReply - replies cannot have replies (one level deep only).
	CreateReply(ctx context.Context, projectID ulid.ULID, traceID string, parentID, userID ulid.ULID, req *CreateCommentRequest) (*CommentResponse, error)
}
