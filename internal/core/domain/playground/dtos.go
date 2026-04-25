package playground

import (
	"encoding/json"

	"github.com/google/uuid"

	prompt "brokle/internal/core/domain/prompt"
)

// ----------------------------
// Request DTOs
// ----------------------------

// CreatePlaygroundSessionRequest represents a request to create a new session.
// Name is required for all sessions (no ephemeral sessions).
type CreatePlaygroundSessionRequest struct {
	// ProjectID is set from the URL path parameter
	ProjectID uuid.UUID `json:"-"`

	// CreatedBy is set from the auth context
	CreatedBy *uuid.UUID `json:"-"`

	// Session metadata
	Name        string   `json:"name" validate:"required,max=200"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`

	// Session content
	Variables json.RawMessage `json:"variables,omitempty"`
	Config    json.RawMessage `json:"config,omitempty"`
	Windows   json.RawMessage `json:"windows" validate:"required"`
}

// UpdateSessionRequest represents a request to update a session
type UpdateSessionRequest struct {
	// SessionID is set from the URL path parameter
	SessionID uuid.UUID `json:"-"`

	// Session metadata (optional updates)
	Name        *string  `json:"name,omitempty" validate:"omitempty,max=200"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`

	// Session content (optional updates)
	Variables json.RawMessage `json:"variables,omitempty"`
	Config    json.RawMessage `json:"config,omitempty"`
	Windows   json.RawMessage `json:"windows,omitempty"`
}

// UpdateLastRunRequest represents a request to update the last run
type UpdateLastRunRequest struct {
	// SessionID is set from the URL path parameter
	SessionID uuid.UUID `json:"-"`

	// LastRun is the execution result to store
	LastRun *LastRun `json:"last_run" validate:"required"`
}

// ListSessionsRequest represents a request to list sessions
type ListSessionsRequest struct {
	// ProjectID is set from the URL path parameter
	ProjectID uuid.UUID `json:"-"`

	// Limit is the maximum number of sessions to return (default 20, max 100)
	Limit int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`

	// Tags filters by tags (optional, any match)
	Tags []string `json:"tags,omitempty"`
}

// ----------------------------
// Execution Request/Response DTOs
// ----------------------------

// ExecuteRequest represents a playground execution request.
type ExecuteRequest struct {
	ProjectID       uuid.UUID
	OrganizationID  uuid.UUID // Required for organization-scoped credential resolution
	SessionID       *uuid.UUID
	Template        any
	PromptType      prompt.PromptType
	Variables       map[string]string
	ConfigOverrides *prompt.ModelConfig
}

// ExecuteResponse wraps execution result with metadata.
type ExecuteResponse struct {
	CompiledPrompt any                 `json:"compiled_prompt"`
	Response       *prompt.LLMResponse `json:"response,omitempty"`
	LatencyMs      int64               `json:"latency_ms"`
	Error          string              `json:"error,omitempty"`
}

// StreamRequest represents a streaming execution request.
type StreamRequest struct {
	ProjectID       uuid.UUID
	OrganizationID  uuid.UUID // Required for organization-scoped credential resolution
	SessionID       *uuid.UUID
	Template        any
	PromptType      prompt.PromptType
	Variables       map[string]string
	ConfigOverrides *prompt.ModelConfig
}

// StreamResponse contains channels for streaming execution.
type StreamResponse struct {
	EventChan  <-chan prompt.StreamEvent
	ResultChan <-chan *prompt.StreamResult
}
