package playground

import (
	"encoding/json"

	prompt "brokle/internal/core/domain/prompt"
)

// executePlaygroundBody — POST /api/v1/playground/execute
type executePlaygroundBody struct {
	Template        any                 `json:"template"`
	PromptType      prompt.PromptType   `json:"prompt_type"        validate:"required,oneof=text chat"`
	Variables       map[string]string   `json:"variables,omitempty"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty"`
	SessionID       *string             `json:"session_id,omitempty"`
	ProjectID       *string             `json:"project_id"`
}

// StreamChunk is the single event variant emitted on /stream. The
// gin implementation emitted a JSON blob with a `type` discriminator
// under a single SSE `message` event — preserved verbatim here.
type StreamChunk struct {
	Type         string         `json:"type"`
	Content      string         `json:"content,omitempty"`
	Error        string         `json:"error,omitempty"`
	FinishReason string         `json:"finish_reason,omitempty"`
	Metrics      *StreamMetrics `json:"metrics,omitempty"`
}

// StreamMetrics carries the terminal usage + latency metrics emitted
// after the underlying stream closes.
type StreamMetrics struct {
	Model            string   `json:"model,omitempty"`
	PromptTokens     int      `json:"prompt_tokens,omitempty"`
	CompletionTokens int      `json:"completion_tokens,omitempty"`
	TotalTokens      int      `json:"total_tokens,omitempty"`
	Cost             *float64 `json:"cost,omitempty"`
	TTFTMs           *float64 `json:"ttft_ms,omitempty"`
	TotalDuration    int64    `json:"total_duration_ms,omitempty"`
}

// streamBody — POST /api/v1/playground/stream
type streamBody struct {
	Template        any                 `json:"template"`
	PromptType      prompt.PromptType   `json:"prompt_type"        validate:"required,oneof=text chat"`
	Variables       map[string]string   `json:"variables,omitempty"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty"`
	SessionID       *string             `json:"session_id,omitempty"`
	ProjectID       *string             `json:"project_id"`
}

// createSessionBody — POST /api/v1/projects/{projectId}/playground/sessions
type createSessionBody struct {
	Name        string          `json:"name"                   validate:"required,min=1,max=200"`
	Description *string         `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"         validate:"omitempty,max=10"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Windows     json.RawMessage `json:"windows"`
}

// updateSessionBody — PUT /api/v1/projects/{projectId}/playground/sessions/{sessionId}
type updateSessionBody struct {
	Name        *string         `json:"name,omitempty"         validate:"omitempty,min=1,max=200"`
	Description *string         `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"         validate:"omitempty,max=10"`
	Variables   json.RawMessage `json:"variables,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Windows     json.RawMessage `json:"windows,omitempty"`
}

// sdkExecuteBody — POST /v1/playground/execute
type sdkExecuteBody struct {
	Template        any                 `json:"template"`
	PromptType      prompt.PromptType   `json:"prompt_type"        validate:"required,oneof=text chat"`
	Variables       map[string]string   `json:"variables,omitempty"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty"`
}
