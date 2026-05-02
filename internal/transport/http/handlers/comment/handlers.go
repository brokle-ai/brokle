// Package comment exposes trace-attached comment chi routes —
// discussion threads with reactions. Dashboard plane; mounted under
// the project subgroup so tenant scoping (projectID) is sourced from
// httpctx.MustGetProjectID rather than a query parameter.
package comment

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	commentDomain "brokle/internal/core/domain/comment"
	commentService "brokle/internal/core/services/comment"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	svc    *commentService.CommentService
	logger *slog.Logger
}

// New constructs a Handler with all required services.
func New(svc *commentService.CommentService, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// parseScope validates the trace-ID path segment and pulls the
// projectID from context (set by RequireProjectAccess on the parent
// chi subgroup). Returns a typed AppError on missing trace ID.
func parseScope(r *http.Request) (traceID string, projectID uuid.UUID, err error) {
	traceID = chi.URLParam(r, "id")
	if traceID == "" {
		return "", uuid.Nil, appErrors.InvalidParam("id", "is required")
	}
	projectID = httpctx.MustGetProjectID(r.Context())
	return traceID, projectID, nil
}

// ----- create-comment --------------------------------------------------

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body commentBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	c, err := h.svc.CreateComment(r.Context(), projectID, traceID, userID,
		&commentDomain.CreateCommentRequest{Content: body.Content})
	if err != nil {
		h.logger.WarnContext(r.Context(), "comment: create failed",
			"user_id", userID, "project_id", projectID, "trace_id", traceID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Created(w, c)
}

// ----- list-comments ---------------------------------------------------

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	res, err := h.svc.ListComments(r.Context(), projectID, traceID, &userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

// ----- get-comment-count -----------------------------------------------

func (h *Handler) Count(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := h.svc.GetCommentCount(r.Context(), projectID, traceID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

// ----- update-comment --------------------------------------------------

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	commentID, err := request.URLParamUUID(r, "comment_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body commentBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	c, err := h.svc.UpdateComment(r.Context(), projectID, traceID, commentID, userID,
		&commentDomain.UpdateCommentRequest{Content: body.Content})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, c)
}

// ----- delete-comment --------------------------------------------------

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	commentID, err := request.URLParamUUID(r, "comment_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	if err := h.svc.DeleteComment(r.Context(), projectID, traceID, commentID, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ----- toggle-reaction -------------------------------------------------

func (h *Handler) ToggleReaction(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	commentID, err := request.URLParamUUID(r, "comment_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body toggleReactionBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	reactions, err := h.svc.ToggleReaction(r.Context(), projectID, traceID, commentID, userID,
		&commentDomain.ToggleReactionRequest{Emoji: body.Emoji})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, reactions)
}

// ----- create-reply ----------------------------------------------------

func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
	traceID, projectID, err := parseScope(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	parentID, err := request.URLParamUUID(r, "comment_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body commentBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	reply, err := h.svc.CreateReply(r.Context(), projectID, traceID, parentID, userID,
		&commentDomain.CreateCommentRequest{Content: body.Content})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, reply)
}
