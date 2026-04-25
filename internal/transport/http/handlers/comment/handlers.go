// Package comment exposes /api/v1/traces/{id}/comments chi routes —
// trace-attached discussion threads with reactions. Dashboard plane;
// every op requires RequireAuth and a project_id query param for
// tenant scoping.
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

type handler struct {
	svc    *commentService.CommentService
	logger *slog.Logger
}

// RegisterRoutes mounts the comment routes on r. Expected mount
// context: the authed dashboard chi group (RequireAuth + LimitByUser).
func RegisterRoutes(r chi.Router, svc *commentService.CommentService, logger *slog.Logger) {
	h := &handler{svc: svc, logger: logger}

	r.Route("/api/v1/traces/{id}/comments", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/count", h.count)
		r.Put("/{comment_id}", h.update)
		r.Delete("/{comment_id}", h.delete)
		r.Post("/{comment_id}/reactions", h.toggleReaction)
		r.Post("/{comment_id}/replies", h.createReply)
	})
}

// parseScope validates the trace-ID path segment + the project_id
// query param (tenant scoping). Returns a typed AppError on failure.
func parseScope(r *http.Request) (traceID string, projectID uuid.UUID, err error) {
	traceID = chi.URLParam(r, "id")
	if traceID == "" {
		return "", uuid.Nil, appErrors.NewValidationError(
			"Missing trace ID", "id is required",
			appErrors.WithParam("id"),
		)
	}
	pidStr := r.URL.Query().Get("project_id")
	projectID, perr := uuid.Parse(pidStr)
	if perr != nil {
		return "", uuid.Nil, appErrors.NewValidationError(
			"Invalid project ID",
			"project_id must be a valid UUID",
			appErrors.WithParam("project_id"),
		)
	}
	return traceID, projectID, nil
}

// ----- create-comment --------------------------------------------------

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) count(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) toggleReaction(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) createReply(w http.ResponseWriter, r *http.Request) {
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
