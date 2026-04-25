// Package playground exposes prompt playground operations on two surfaces:
//   - Dashboard plane (RequireAuth) — session CRUD under
//     /api/v1/projects/{projectId}/playground/sessions plus execute + stream
//     endpoints keyed by project_id in the request body.
//   - SDK plane (RequireSDKAuth) — /v1/playground/execute with the project
//     derived from the API key.
//
// Stream is the SSE endpoint. Wire format preserved 1:1 from the
// pre-migration implementation: each chunk is a single SSE `message`
// event whose JSON body carries a discriminator `type` field
// ("start" | "content" | "end" | "error" | "metrics").
package playground

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	playgroundDomain "brokle/internal/core/domain/playground"
	organizationService "brokle/internal/core/services/organization"
	playgroundService "brokle/internal/core/services/playground"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type handler struct {
	svc            *playgroundService.PlaygroundService
	projectService *organizationService.ProjectService
	logger         *slog.Logger
}

// RegisterRoutes mounts the dashboard-plane playground routes on r.
// Expected mount context: the authed dashboard chi group.
func RegisterRoutes(
	r chi.Router,
	svc *playgroundService.PlaygroundService,
	projectService *organizationService.ProjectService,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, projectService: projectService, logger: logger}

	r.Post("/api/v1/playground/execute", h.execute)
	r.Post("/api/v1/playground/stream", h.stream)

	r.Route("/api/v1/projects/{projectId}/playground/sessions", func(r chi.Router) {
		r.Post("/", h.createSession)
		r.Get("/", h.listSessions)
		r.Get("/{sessionId}", h.getSession)
		r.Put("/{sessionId}", h.updateSession)
		r.Delete("/{sessionId}", h.deleteSession)
	})
}

// RegisterSDKRoutes mounts the SDK-plane playground route on r.
// Expected mount context: the SDK-authed chi group. Project ID is
// taken from the authenticated API key, not the request body.
func RegisterSDKRoutes(
	r chi.Router,
	svc *playgroundService.PlaygroundService,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, logger: logger}
	r.Post("/v1/playground/execute", h.sdkExecute)
}

// ---- shared helpers ---------------------------------------------------

// validateProjectAccess verifies the caller may use the given project's
// credentials. Required on execute + stream because those endpoints
// take project_id from the request body.
func (h *handler) validateProjectAccess(ctx context.Context, projectIDStr *string) (uuid.UUID, error) {
	if projectIDStr == nil {
		return uuid.Nil, appErrors.NewValidationError(
			"project_id is required",
			"playground execution requires project_id",
			appErrors.WithParam("project_id"),
		)
	}
	projectID, err := uuid.Parse(*projectIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError(
			"Invalid project_id",
			"project_id must be a valid UUID",
			appErrors.WithParam("project_id"),
		)
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.projectService.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return uuid.Nil, appErrors.NewNotFoundError("Project not found")
		}
		h.logger.Warn("User attempted to access project credentials without permission",
			"user_id", userID.String(),
			"project_id", projectID.String(),
			"error", err,
		)
		return uuid.Nil, appErrors.NewForbiddenError("You don't have access to this project")
	}
	return projectID, nil
}

// validateSessionAccess verifies the caller may update the given session.
// Returns the parsed session ID, or nil if no session_id was supplied.
func (h *handler) validateSessionAccess(ctx context.Context, sessionIDStr *string) (*uuid.UUID, error) {
	if sessionIDStr == nil {
		return nil, nil
	}
	sessionID, err := uuid.Parse(*sessionIDStr)
	if err != nil {
		return nil, appErrors.NewValidationError(
			"Invalid session_id",
			"session_id must be a valid UUID",
			appErrors.WithParam("session_id"),
		)
	}
	session, err := h.svc.GetSession(ctx, sessionID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Session not found")
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.projectService.ValidateProjectAccess(ctx, userID, session.ProjectID); err != nil {
		h.logger.Warn("User attempted to access session without project permission",
			"user_id", userID.String(),
			"session_id", sessionID.String(),
			"project_id", session.ProjectID.String(),
		)
		return nil, appErrors.NewForbiddenError("You don't have access to this session")
	}
	return &sessionID, nil
}

// ==============================================================
// Execute (dashboard)
// ==============================================================

func (h *handler) execute(w http.ResponseWriter, r *http.Request) {
	var body executePlaygroundBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	projectID, err := h.validateProjectAccess(r.Context(), body.ProjectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sessionID, err := h.validateSessionAccess(r.Context(), body.SessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	// Derive org from project — never trust a client-supplied org ID.
	project, err := h.projectService.GetProject(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	resp, err := h.svc.ExecutePrompt(r.Context(), &playgroundDomain.ExecuteRequest{
		ProjectID:       projectID,
		OrganizationID:  project.OrganizationID,
		SessionID:       sessionID,
		Template:        body.Template,
		PromptType:      body.PromptType,
		Variables:       body.Variables,
		ConfigOverrides: body.ConfigOverrides,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// ==============================================================
// Stream (dashboard, SSE)
// ==============================================================

// stream is the SSE handler. Validation failures emitted BEFORE we
// commit to the stream go through the normal error envelope; once
// the stream has started, further failures arrive as inline
// `error`-typed StreamChunk events since we can't switch wire shapes
// mid-response.
func (h *handler) stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.WriteError(w, appErrors.NewInternalError(
			"Streaming unsupported", fmt.Errorf("ResponseWriter does not implement http.Flusher"),
		))
		return
	}

	var body streamBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	projectID, err := h.validateProjectAccess(r.Context(), body.ProjectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sessionID, err := h.validateSessionAccess(r.Context(), body.SessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	project, err := h.projectService.GetProject(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	// Commit to the SSE response.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sendEvent := func(chunk StreamChunk) bool {
		data, jerr := json.Marshal(chunk)
		if jerr != nil {
			h.logger.Warn("SSE marshal failed", "error", jerr)
			return false
		}
		if _, werr := fmt.Fprintf(w, "event: message\ndata: %s\n\n", data); werr != nil {
			h.logger.Warn("SSE write failed, client likely disconnected", "error", werr)
			return false
		}
		flusher.Flush()
		return true
	}

	streamResp, err := h.svc.StreamPrompt(r.Context(), &playgroundDomain.StreamRequest{
		ProjectID:       projectID,
		OrganizationID:  project.OrganizationID,
		SessionID:       sessionID,
		Template:        body.Template,
		PromptType:      body.PromptType,
		Variables:       body.Variables,
		ConfigOverrides: body.ConfigOverrides,
	})
	if err != nil {
		sendEvent(StreamChunk{Type: "error", Error: err.Error()})
		return
	}

	for event := range streamResp.EventChan {
		select {
		case <-r.Context().Done():
			return
		default:
			if !sendEvent(StreamChunk{
				Type:         string(event.Type),
				Content:      event.Content,
				Error:        event.Error,
				FinishReason: event.FinishReason,
			}) {
				return
			}
		}
	}

	// Terminal metrics chunk.
	if result, ok := <-streamResp.ResultChan; ok && result != nil {
		var metrics *StreamMetrics
		if result.Usage != nil {
			metrics = &StreamMetrics{
				Model:            result.Model,
				PromptTokens:     result.Usage.PromptTokens,
				CompletionTokens: result.Usage.CompletionTokens,
				TotalTokens:      result.Usage.TotalTokens,
				Cost:             result.Cost,
				TTFTMs:           result.TTFTMs,
				TotalDuration:    result.TotalDuration,
			}
		}
		sendEvent(StreamChunk{Type: "metrics", Metrics: metrics})
	}
}

// ==============================================================
// Sessions (dashboard)
// ==============================================================

func (h *handler) createSession(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body createSessionBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	session, err := h.svc.CreateSession(r.Context(), &playgroundDomain.CreatePlaygroundSessionRequest{
		ProjectID:   projectID,
		CreatedBy:   &userID,
		Name:        body.Name,
		Description: body.Description,
		Tags:        body.Tags,
		Variables:   body.Variables,
		Config:      body.Config,
		Windows:     body.Windows,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, session)
}

func (h *handler) listSessions(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	limit, err := request.QueryInt(r, "limit", 20)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	tags := r.URL.Query()["tags"]

	sessions, err := h.svc.ListSessions(r.Context(), &playgroundDomain.ListSessionsRequest{
		ProjectID: projectID,
		Limit:     limit,
		Tags:      tags,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, sessions)
}

func (h *handler) getSession(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sessionID, err := request.URLParamUUID(r, "sessionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.svc.ValidateProjectAccess(r.Context(), sessionID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	session, err := h.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, session)
}

func (h *handler) updateSession(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sessionID, err := request.URLParamUUID(r, "sessionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.svc.ValidateProjectAccess(r.Context(), sessionID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}

	var body updateSessionBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	session, err := h.svc.UpdateSession(r.Context(), &playgroundDomain.UpdateSessionRequest{
		SessionID:   sessionID,
		Name:        body.Name,
		Description: body.Description,
		Tags:        body.Tags,
		Variables:   body.Variables,
		Config:      body.Config,
		Windows:     body.Windows,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, session)
}

func (h *handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sessionID, err := request.URLParamUUID(r, "sessionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.svc.ValidateProjectAccess(r.Context(), sessionID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.svc.DeleteSession(r.Context(), sessionID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ==============================================================
// SDK execute
// ==============================================================

func (h *handler) sdkExecute(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	var body sdkExecuteBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	resp, err := h.svc.ExecutePrompt(r.Context(), &playgroundDomain.ExecuteRequest{
		ProjectID:       projectID,
		SessionID:       nil, // SDK doesn't use sessions
		Template:        body.Template,
		PromptType:      body.PromptType,
		Variables:       body.Variables,
		ConfigOverrides: body.ConfigOverrides,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}

	h.logger.Info("Prompt executed via SDK",
		"project_id", projectID,
		"prompt_type", body.PromptType,
	)
	response.Success(w, resp)
}
