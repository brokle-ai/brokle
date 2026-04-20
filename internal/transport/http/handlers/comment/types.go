package comment

import (
	commentDomain "brokle/internal/core/domain/comment"
)

// Huma operation types for the comment package.

// ----- shared input fields ------------------------------------------

type traceScope struct {
	TraceID   string `path:"id" doc:"Trace identifier the comment is attached to"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace (tenant scope)"`
}

type traceCommentScope struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
}

// ----- create-comment -----------------------------------------------

type CreateCommentInput struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
	Body      commentDomain.CreateCommentRequest
}

type CreateCommentOutput struct {
	Body *commentDomain.CommentResponse
}

// ----- list-comments ------------------------------------------------

type ListCommentsInput struct {
	traceScope
}

type ListCommentsOutput struct {
	Body *commentDomain.ListCommentsResponse
}

// ----- get-comment-count --------------------------------------------

type GetCommentCountInput struct {
	traceScope
}

type GetCommentCountOutput struct {
	Body *commentDomain.CommentCountResponse
}

// ----- update-comment -----------------------------------------------

type UpdateCommentInput struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
	Body      commentDomain.UpdateCommentRequest
}

type UpdateCommentOutput struct {
	Body *commentDomain.CommentResponse
}

// ----- delete-comment -----------------------------------------------

type DeleteCommentInput struct {
	traceCommentScope
}

type DeleteCommentOutput struct{}

// ----- toggle-reaction ----------------------------------------------

type ToggleReactionInput struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
	Body      commentDomain.ToggleReactionRequest
}

type ToggleReactionOutput struct {
	Body []commentDomain.ReactionSummary
}

// ----- create-reply -------------------------------------------------

type CreateReplyInput struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Parent comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
	Body      commentDomain.CreateCommentRequest
}

type CreateReplyOutput struct {
	Body *commentDomain.CommentResponse
}
