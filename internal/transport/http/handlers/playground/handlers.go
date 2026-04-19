// Package playground exposes prompt playground operations on two surfaces:
//   - Dashboard plane (apiAdmin, RequireAuth) — session CRUD under
//     /api/v1/projects/{projectId}/playground/sessions plus execute + stream
//     endpoints keyed by project_id in the request body.
//   - SDK plane (apiPublic, RequireSDKAuth) — /v1/playground/execute with
//     the project derived from the API key.
//
// Stream is the SSE endpoint. Event wire format is preserved 1:1 from the
// pre-migration gin implementation: each chunk is a single SSE `message`
// event whose JSON body carries a discriminator `type` field
// ("start" | "content" | "end" | "error" | "metrics"). This is why we
// register a single event type on the sse map rather than one per variant
// — adopting a per-variant event map would be a breaking wire change for
// any client already streaming from this endpoint.
package playground

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"

	"brokle/internal/core/domain/organization"
	playgroundDomain "brokle/internal/core/domain/playground"
	prompt "brokle/internal/core/domain/prompt"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	svc            playgroundDomain.PlaygroundService
	projectService organization.ProjectService
	logger         *slog.Logger
}

// RegisterRoutes wires the dashboard-plane playground operations on apiAdmin.
func RegisterRoutes(
	api huma.API,
	svc playgroundDomain.PlaygroundService,
	projectService organization.ProjectService,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, projectService: projectService, logger: logger}

	// ---- execute (dashboard) --------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "playground-execute",
		Method:      http.MethodPost,
		Path:        "/api/v1/playground/execute",
		Tags:        []string{"playground"},
		Summary:     "Execute a prompt in the playground",
		Description: "Executes a prompt template with variables. Optionally updates the session's last_run when session_id is provided.",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.execute)

	// ---- stream (dashboard, SSE) ----------------------------------
	sse.Register(api, huma.Operation{
		OperationID: "playground-stream",
		Method:      http.MethodPost,
		Path:        "/api/v1/playground/stream",
		Tags:        []string{"playground"},
		Summary:     "Stream a playground prompt execution (Server-Sent Events)",
		Description: "Streams prompt execution results as SSE messages. Each event is a single `message` with a JSON body carrying a `type` discriminator (start | content | end | error | metrics).",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, map[string]any{
		// Single event name mirrors the pre-migration wire format.
		"message": StreamChunk{},
	}, h.stream)

	// ---- sessions -------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "create-playground-session",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/playground/sessions",
		Tags:          []string{"playground"},
		Summary:       "Create a playground session",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createSession)

	huma.Register(api, huma.Operation{
		OperationID: "list-playground-sessions",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/playground/sessions",
		Tags:        []string{"playground"},
		Summary:     "List playground sessions for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listSessions)

	huma.Register(api, huma.Operation{
		OperationID: "get-playground-session",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/playground/sessions/{sessionId}",
		Tags:        []string{"playground"},
		Summary:     "Get a playground session",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getSession)

	huma.Register(api, huma.Operation{
		OperationID: "update-playground-session",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/playground/sessions/{sessionId}",
		Tags:        []string{"playground"},
		Summary:     "Update a playground session",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateSession)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-playground-session",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/playground/sessions/{sessionId}",
		Tags:          []string{"playground"},
		Summary:       "Delete a playground session",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteSession)
}

// RegisterSDKRoutes wires the SDK-plane playground operations on apiPublic.
// Project ID is taken from the authenticated API key, not the request body.
func RegisterSDKRoutes(
	api huma.API,
	svc playgroundDomain.PlaygroundService,
	logger *slog.Logger,
) {
	h := &handler{svc: svc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "sdk-playground-execute",
		Method:      http.MethodPost,
		Path:        "/v1/playground/execute",
		Tags:        []string{"SDK - playground"},
		Summary:     "Execute a prompt via SDK (project derived from API key)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkExecute)
}

// ---- shared helpers ---------------------------------------------------

func parseProjectID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	return id, nil
}

func parseSessionID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid session ID", "sessionId must be a valid UUID")
	}
	return id, nil
}

// validateProjectAccess verifies the caller may use the given project's
// credentials. Required on both execute + stream because those endpoints
// take project_id from the request body.
func (h *handler) validateProjectAccess(ctx context.Context, projectIDStr *string) (uuid.UUID, error) {
	if projectIDStr == nil {
		return uuid.Nil, appErrors.NewValidationError("project_id is required", "playground execution requires project_id")
	}
	projectID, err := uuid.Parse(*projectIDStr)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project_id", "project_id must be a valid UUID")
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.projectService.ValidateProjectAccess(ctx, userID, projectID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return uuid.Nil, appErrors.NewNotFoundError("Project not found")
		}
		h.logger.Warn("user attempted to access project credentials without permission",
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
		return nil, appErrors.NewValidationError("Invalid session_id", "session_id must be a valid UUID")
	}
	session, err := h.svc.GetSession(ctx, sessionID)
	if err != nil {
		return nil, appErrors.NewNotFoundError("Session not found")
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.projectService.ValidateProjectAccess(ctx, userID, session.ProjectID); err != nil {
		h.logger.Warn("user attempted to access session without project permission",
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

type executeBody struct {
	Template        any                 `json:"template" doc:"Prompt template (string for text, []prompt.ChatMessage for chat)"`
	PromptType      prompt.PromptType   `json:"prompt_type" enum:"text,chat" doc:"Prompt variant"`
	Variables       map[string]string   `json:"variables,omitempty" doc:"Template variable substitutions"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty" doc:"Per-request model config overrides"`
	SessionID       *string             `json:"session_id,omitempty" format:"uuid" doc:"Optional: updates session's last_run"`
	ProjectID       *string             `json:"project_id" format:"uuid" doc:"Project whose credentials + org will be used"`
}

type ExecuteInput struct {
	Body executeBody
}

type ExecuteOutput struct {
	Body *playgroundDomain.ExecuteResponse
}

func (h *handler) execute(ctx context.Context, in *ExecuteInput) (*ExecuteOutput, error) {
	projectID, err := h.validateProjectAccess(ctx, in.Body.ProjectID)
	if err != nil {
		return nil, err
	}
	sessionID, err := h.validateSessionAccess(ctx, in.Body.SessionID)
	if err != nil {
		return nil, err
	}

	// Derive org from project — never trust a client-supplied org ID.
	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	resp, err := h.svc.ExecutePrompt(ctx, &playgroundDomain.ExecuteRequest{
		ProjectID:       projectID,
		OrganizationID:  project.OrganizationID,
		SessionID:       sessionID,
		Template:        in.Body.Template,
		PromptType:      in.Body.PromptType,
		Variables:       in.Body.Variables,
		ConfigOverrides: in.Body.ConfigOverrides,
	})
	if err != nil {
		return nil, err
	}
	return &ExecuteOutput{Body: resp}, nil
}

// ==============================================================
// Stream (dashboard, SSE)
// ==============================================================

// StreamChunk is the single event variant emitted on the /stream endpoint.
// The gin implementation emitted a JSON blob with a `type` discriminator
// under a single SSE `message` event — preserved verbatim here.
type StreamChunk struct {
	Type         string         `json:"type" enum:"start,content,end,error,metrics" doc:"Event discriminator"`
	Content      string         `json:"content,omitempty"`
	Error        string         `json:"error,omitempty"`
	FinishReason string         `json:"finish_reason,omitempty"`
	Metrics      *StreamMetrics `json:"metrics,omitempty"`
}

// StreamMetrics carries the terminal usage + latency metrics emitted after
// the underlying stream closes.
type StreamMetrics struct {
	Model            string   `json:"model,omitempty"`
	PromptTokens     int      `json:"prompt_tokens,omitempty"`
	CompletionTokens int      `json:"completion_tokens,omitempty"`
	TotalTokens      int      `json:"total_tokens,omitempty"`
	Cost             *float64 `json:"cost,omitempty"`
	TTFTMs           *float64 `json:"ttft_ms,omitempty"`
	TotalDuration    int64    `json:"total_duration_ms,omitempty"`
}

type streamBody struct {
	Template        any                 `json:"template"`
	PromptType      prompt.PromptType   `json:"prompt_type" enum:"text,chat"`
	Variables       map[string]string   `json:"variables,omitempty"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty"`
	SessionID       *string             `json:"session_id,omitempty" format:"uuid"`
	ProjectID       *string             `json:"project_id" format:"uuid"`
}

type StreamInput struct {
	Body streamBody
}

// stream is the SSE handler. Per huma/v2/sse semantics this function has
// no error return: validation failures must be surfaced as an inline
// `error`-typed StreamChunk. That's the pre-migration behaviour too — the
// gin implementation used `response.Error` which wrote the shared JSON
// envelope, but once we've committed to the sse.Register shape we cannot
// toggle between an error envelope and an SSE stream mid-handler. The
// trade-off: the HTTP status is already 200 by the time we know the body
// is invalid, so clients must inspect the first event.
func (h *handler) stream(ctx context.Context, in *StreamInput, send sse.Sender) {
	projectID, err := h.validateProjectAccess(ctx, in.Body.ProjectID)
	if err != nil {
		h.sendError(send, err.Error())
		return
	}
	sessionID, err := h.validateSessionAccess(ctx, in.Body.SessionID)
	if err != nil {
		h.sendError(send, err.Error())
		return
	}
	project, err := h.projectService.GetProject(ctx, projectID)
	if err != nil {
		h.sendError(send, err.Error())
		return
	}

	streamResp, err := h.svc.StreamPrompt(ctx, &playgroundDomain.StreamRequest{
		ProjectID:       projectID,
		OrganizationID:  project.OrganizationID,
		SessionID:       sessionID,
		Template:        in.Body.Template,
		PromptType:      in.Body.PromptType,
		Variables:       in.Body.Variables,
		ConfigOverrides: in.Body.ConfigOverrides,
	})
	if err != nil {
		h.sendError(send, err.Error())
		return
	}

	// Pump events until the service closes EventChan, respecting client
	// disconnect via ctx. Identical semantics to the gin loop.
	for event := range streamResp.EventChan {
		select {
		case <-ctx.Done():
			return
		default:
			if err := send.Data(StreamChunk{
				Type:         string(event.Type),
				Content:      event.Content,
				Error:        event.Error,
				FinishReason: event.FinishReason,
			}); err != nil {
				h.logger.Warn("sse send failed, client likely disconnected", "error", err)
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
		if err := send.Data(StreamChunk{Type: "metrics", Metrics: metrics}); err != nil {
			h.logger.Warn("sse metrics send failed", "error", err)
		}
	}
}

func (h *handler) sendError(send sse.Sender, msg string) {
	_ = send.Data(StreamChunk{Type: "error", Error: msg})
}

// ==============================================================
// Sessions (dashboard)
// ==============================================================

type createSessionBody struct {
	Name        string          `json:"name" minLength:"1" maxLength:"200" doc:"Session name"`
	Description *string         `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty" maxItems:"10" doc:"Up to 10 tags, each ≤50 chars"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Windows     json.RawMessage `json:"windows" doc:"Multi-window comparison state (required)"`
}

type CreateSessionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      createSessionBody
}

type CreateSessionOutput struct {
	Body *playgroundDomain.SessionResponse
}

func (h *handler) createSession(ctx context.Context, in *CreateSessionInput) (*CreateSessionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	session, err := h.svc.CreateSession(ctx, &playgroundDomain.CreateSessionRequest{
		ProjectID:   projectID,
		CreatedBy:   &userID,
		Name:        in.Body.Name,
		Description: in.Body.Description,
		Tags:        in.Body.Tags,
		Variables:   in.Body.Variables,
		Config:      in.Body.Config,
		Windows:     in.Body.Windows,
	})
	if err != nil {
		return nil, err
	}
	return &CreateSessionOutput{Body: session}, nil
}

type ListSessionsInput struct {
	ProjectID string   `path:"projectId" format:"uuid"`
	Limit     int      `query:"limit" required:"false" minimum:"1" maximum:"100" doc:"Max sessions to return (default 20)"`
	Tags      []string `query:"tags" required:"false" doc:"Filter by tags (any match)"`
}

type ListSessionsOutput struct {
	Body []*playgroundDomain.SessionSummary
}

func (h *handler) listSessions(ctx context.Context, in *ListSessionsInput) (*ListSessionsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	sessions, err := h.svc.ListSessions(ctx, &playgroundDomain.ListSessionsRequest{
		ProjectID: projectID,
		Limit:     limit,
		Tags:      in.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &ListSessionsOutput{Body: sessions}, nil
}

type GetSessionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	SessionID string `path:"sessionId" format:"uuid"`
}

type GetSessionOutput struct {
	Body *playgroundDomain.SessionResponse
}

func (h *handler) getSession(ctx context.Context, in *GetSessionInput) (*GetSessionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	sessionID, err := parseSessionID(in.SessionID)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ValidateProjectAccess(ctx, sessionID, projectID); err != nil {
		return nil, err
	}
	session, err := h.svc.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &GetSessionOutput{Body: session}, nil
}

type updateSessionBody struct {
	Name        *string         `json:"name,omitempty" minLength:"1" maxLength:"200"`
	Description *string         `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty" maxItems:"10"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Windows     json.RawMessage `json:"windows,omitempty"`
}

type UpdateSessionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	SessionID string `path:"sessionId" format:"uuid"`
	Body      updateSessionBody
}

type UpdateSessionOutput struct {
	Body *playgroundDomain.SessionResponse
}

func (h *handler) updateSession(ctx context.Context, in *UpdateSessionInput) (*UpdateSessionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	sessionID, err := parseSessionID(in.SessionID)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ValidateProjectAccess(ctx, sessionID, projectID); err != nil {
		return nil, err
	}
	session, err := h.svc.UpdateSession(ctx, &playgroundDomain.UpdateSessionRequest{
		SessionID:   sessionID,
		Name:        in.Body.Name,
		Description: in.Body.Description,
		Tags:        in.Body.Tags,
		Variables:   in.Body.Variables,
		Config:      in.Body.Config,
		Windows:     in.Body.Windows,
	})
	if err != nil {
		return nil, err
	}
	return &UpdateSessionOutput{Body: session}, nil
}

type DeleteSessionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	SessionID string `path:"sessionId" format:"uuid"`
}

type DeleteSessionOutput struct{}

func (h *handler) deleteSession(ctx context.Context, in *DeleteSessionInput) (*DeleteSessionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	sessionID, err := parseSessionID(in.SessionID)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ValidateProjectAccess(ctx, sessionID, projectID); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteSession(ctx, sessionID); err != nil {
		return nil, err
	}
	return &DeleteSessionOutput{}, nil
}

// ==============================================================
// SDK execute
// ==============================================================

type sdkExecuteBody struct {
	Template        any                 `json:"template" doc:"Prompt template"`
	PromptType      prompt.PromptType   `json:"prompt_type" enum:"text,chat"`
	Variables       map[string]string   `json:"variables,omitempty"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty"`
}

type SDKExecuteInput struct {
	Body sdkExecuteBody
}

type SDKExecuteOutput struct {
	Body *playgroundDomain.ExecuteResponse
}

func (h *handler) sdkExecute(ctx context.Context, in *SDKExecuteInput) (*SDKExecuteOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)

	resp, err := h.svc.ExecutePrompt(ctx, &playgroundDomain.ExecuteRequest{
		ProjectID:       projectID,
		SessionID:       nil, // SDK doesn't use sessions
		Template:        in.Body.Template,
		PromptType:      in.Body.PromptType,
		Variables:       in.Body.Variables,
		ConfigOverrides: in.Body.ConfigOverrides,
	})
	if err != nil {
		return nil, err
	}

	h.logger.Info("prompt executed via SDK",
		"project_id", projectID,
		"prompt_type", in.Body.PromptType,
	)
	return &SDKExecuteOutput{Body: resp}, nil
}
