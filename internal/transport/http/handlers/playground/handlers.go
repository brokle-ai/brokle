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

	"github.com/google/uuid"

	playgroundDomain "brokle/internal/core/domain/playground"
	organizationService "brokle/internal/core/services/organization"
	playgroundService "brokle/internal/core/services/playground"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	svc            *playgroundService.PlaygroundService
	projectService *organizationService.ProjectService
	logger         *slog.Logger
}

// New constructs a Handler with all services any playground Register*
// function might need (dashboard + SDK).
func New(
	svc *playgroundService.PlaygroundService,
	projectService *organizationService.ProjectService,
	logger *slog.Logger,
) *Handler {
	return &Handler{svc: svc, projectService: projectService, logger: logger}
}

// ---- shared helpers ---------------------------------------------------

// validateSessionAccess verifies the supplied session_id belongs to the
// pinned projectID from context. Returns the parsed session ID, or nil
// if no session_id was supplied.
func (h *Handler) validateSessionAccess(ctx context.Context, sessionIDStr *string, projectID uuid.UUID) (*uuid.UUID, error) {
	if sessionIDStr == nil {
		return nil, nil
	}
	sessionID, err := uuid.Parse(*sessionIDStr)
	if err != nil {
		return nil, appErrors.InvalidParam("session_id", "must be a valid UUID")
	}
	session, err := h.svc.GetSession(ctx, sessionID)
	if err != nil {
		return nil, appErrors.NotFound("session")
	}
	if session.ProjectID != projectID {
		h.logger.Warn("Session does not belong to pinned project",
			"session_id", sessionID.String(),
			"session_project", session.ProjectID.String(),
			"pinned_project", projectID.String(),
		)
		return nil, appErrors.PermissionDenied("session", "You don't have access to this session")
	}
	return &sessionID, nil
}

// ==============================================================
// Execute (dashboard)
// ==============================================================

func (h *Handler) Execute(w http.ResponseWriter, r *http.Request) {
	var body executePlaygroundBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	projectID := httpctx.MustGetProjectID(r.Context())
	sessionID, err := h.validateSessionAccess(r.Context(), body.SessionID, projectID)
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
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.WriteError(w, appErrors.Internal(
			"Streaming unsupported", fmt.Errorf("ResponseWriter does not implement http.Flusher"),
		))
		return
	}

	var body streamBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	projectID := httpctx.MustGetProjectID(r.Context())
	sessionID, err := h.validateSessionAccess(r.Context(), body.SessionID, projectID)
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

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

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

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
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

func (h *Handler) SdkExecute(w http.ResponseWriter, r *http.Request) {
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
