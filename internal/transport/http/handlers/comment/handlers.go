// Package comment exposes /api/v1/traces/{id}/comments Huma
// operations — trace-attached discussion threads with reactions.
// Dashboard plane (apiAdmin); every op requires RequireAuth and
// takes a project_id query param for tenant scoping.
package comment

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	commentDomain "brokle/internal/core/domain/comment"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	svc    commentDomain.Service
	logger *slog.Logger
}

// RegisterRoutes registers every comment operation on apiAdmin.
func RegisterRoutes(api huma.API, svc commentDomain.Service, logger *slog.Logger) {
	h := &handler{svc: svc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID:   "create-comment",
		Method:        http.MethodPost,
		Path:          "/api/v1/traces/{id}/comments",
		Tags:          []string{"comments"},
		Summary:       "Create a top-level comment on a trace",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.create)

	huma.Register(api, huma.Operation{
		OperationID: "list-comments",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/{id}/comments",
		Tags:        []string{"comments"},
		Summary:     "List all comments on a trace",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-comment-count",
		Method:      http.MethodGet,
		Path:        "/api/v1/traces/{id}/comments/count",
		Tags:        []string{"comments"},
		Summary:     "Get the comment count for a trace",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.count)

	huma.Register(api, huma.Operation{
		OperationID: "update-comment",
		Method:      http.MethodPut,
		Path:        "/api/v1/traces/{id}/comments/{comment_id}",
		Tags:        []string{"comments"},
		Summary:     "Update a comment (owner only)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.update)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-comment",
		Method:        http.MethodDelete,
		Path:          "/api/v1/traces/{id}/comments/{comment_id}",
		Tags:          []string{"comments"},
		Summary:       "Delete a comment (owner only)",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.delete)

	huma.Register(api, huma.Operation{
		OperationID: "toggle-reaction",
		Method:      http.MethodPost,
		Path:        "/api/v1/traces/{id}/comments/{comment_id}/reactions",
		Tags:        []string{"comments"},
		Summary:     "Toggle an emoji reaction on a comment",
		Description: "Adds the reaction when the user has not yet reacted with this emoji; removes it when they have. Returns the full summary so clients can re-render counts.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.toggleReaction)

	huma.Register(api, huma.Operation{
		OperationID:   "create-reply",
		Method:        http.MethodPost,
		Path:          "/api/v1/traces/{id}/comments/{comment_id}/replies",
		Tags:          []string{"comments"},
		Summary:       "Reply to a top-level comment",
		Description:   "Threads are one level deep — replies cannot have replies. The service enforces this and returns 400 if the parent is itself a reply.",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createReply)
}

// ----- shared input fields + helpers --------------------------------

type traceScope struct {
	TraceID   string `path:"id" doc:"Trace identifier the comment is attached to"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace (tenant scope)"`
}

type traceCommentScope struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
}

func parseScope(traceID, projectIDStr string) (projectID uuid.UUID, err error) {
	if traceID == "" {
		return uuid.Nil, appErrors.NewValidationError("Missing trace ID", "id is required")
	}
	projectID, err = uuid.Parse(projectIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project ID", "project_id must be a valid UUID")
	}
	return projectID, nil
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

func (h *handler) create(ctx context.Context, in *CreateCommentInput) (*CreateCommentOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	c, err := h.svc.CreateComment(ctx, projectID, in.TraceID, userID, &in.Body)
	if err != nil {
		h.logger.WarnContext(ctx, "comment: create failed", "user_id", userID, "project_id", projectID, "trace_id", in.TraceID, "error", err)
		return nil, err
	}
	return &CreateCommentOutput{Body: c}, nil
}

// ----- list-comments ------------------------------------------------

type ListCommentsInput struct {
	traceScope
}

type ListCommentsOutput struct {
	Body *commentDomain.ListCommentsResponse
}

func (h *handler) list(ctx context.Context, in *ListCommentsInput) (*ListCommentsOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	// Current user — optional, used to decorate reactions with
	// the hasUser flag. RequireAuth guarantees a user is present,
	// so MustGetUserID is safe here; passing a pointer keeps the
	// service layer signature stable with the OptionalAuth-era
	// gin handler.
	userID := httpctx.MustGetUserID(ctx)

	res, err := h.svc.ListComments(ctx, projectID, in.TraceID, &userID)
	if err != nil {
		return nil, err
	}
	return &ListCommentsOutput{Body: res}, nil
}

// ----- get-comment-count --------------------------------------------

type GetCommentCountInput struct {
	traceScope
}

type GetCommentCountOutput struct {
	Body *commentDomain.CommentCountResponse
}

func (h *handler) count(ctx context.Context, in *GetCommentCountInput) (*GetCommentCountOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	res, err := h.svc.GetCommentCount(ctx, projectID, in.TraceID)
	if err != nil {
		return nil, err
	}
	return &GetCommentCountOutput{Body: res}, nil
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

func (h *handler) update(ctx context.Context, in *UpdateCommentInput) (*UpdateCommentOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	commentID, err := uuid.Parse(in.CommentID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid comment ID", "comment_id must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	c, err := h.svc.UpdateComment(ctx, projectID, in.TraceID, commentID, userID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &UpdateCommentOutput{Body: c}, nil
}

// ----- delete-comment -----------------------------------------------

type DeleteCommentInput struct {
	traceCommentScope
}

type DeleteCommentOutput struct{}

func (h *handler) delete(ctx context.Context, in *DeleteCommentInput) (*DeleteCommentOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	commentID, err := uuid.Parse(in.CommentID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid comment ID", "comment_id must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.svc.DeleteComment(ctx, projectID, in.TraceID, commentID, userID); err != nil {
		return nil, err
	}
	return &DeleteCommentOutput{}, nil
}

// ----- toggle-reaction ---------------------------------------------

type ToggleReactionInput struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
	Body      commentDomain.ToggleReactionRequest
}

type ToggleReactionOutput struct {
	Body []commentDomain.ReactionSummary
}

func (h *handler) toggleReaction(ctx context.Context, in *ToggleReactionInput) (*ToggleReactionOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	commentID, err := uuid.Parse(in.CommentID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid comment ID", "comment_id must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	reactions, err := h.svc.ToggleReaction(ctx, projectID, in.TraceID, commentID, userID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &ToggleReactionOutput{Body: reactions}, nil
}

// ----- create-reply ------------------------------------------------

type CreateReplyInput struct {
	TraceID   string `path:"id" doc:"Trace identifier"`
	CommentID string `path:"comment_id" format:"uuid" doc:"Parent comment identifier"`
	ProjectID string `query:"project_id" format:"uuid" doc:"Project that owns the trace"`
	Body      commentDomain.CreateCommentRequest
}

type CreateReplyOutput struct {
	Body *commentDomain.CommentResponse
}

func (h *handler) createReply(ctx context.Context, in *CreateReplyInput) (*CreateReplyOutput, error) {
	projectID, err := parseScope(in.TraceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	parentID, err := uuid.Parse(in.CommentID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid comment ID", "comment_id must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)

	reply, err := h.svc.CreateReply(ctx, projectID, in.TraceID, parentID, userID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &CreateReplyOutput{Body: reply}, nil
}
