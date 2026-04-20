package playground

import (
	playgroundDomain "brokle/internal/core/domain/playground"
	prompt "brokle/internal/core/domain/prompt"
	"encoding/json"
)

// Huma operation types for the playground package.

type executePlaygroundBody struct {
	Template        any                 `json:"template" doc:"Prompt template (string for text, []prompt.PlaygroundMessage for chat)"`
	PromptType      prompt.PromptType   `json:"prompt_type" enum:"text,chat" doc:"Prompt variant"`
	Variables       map[string]string   `json:"variables,omitempty" doc:"Template variable substitutions"`
	ConfigOverrides *prompt.ModelConfig `json:"config_overrides,omitempty" doc:"Per-request model config overrides"`
	SessionID       *string             `json:"session_id,omitempty" format:"uuid" doc:"Optional: updates session's last_run"`
	ProjectID       *string             `json:"project_id" format:"uuid" doc:"Project whose credentials + org will be used"`
}

type ExecutePlaygroundInput struct {
	Body executePlaygroundBody
}

type ExecutePlaygroundOutput struct {
	Body *playgroundDomain.ExecuteResponse
}

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

type ListPlaygroundSessionsInput struct {
	ProjectID string   `path:"projectId" format:"uuid"`
	Limit     int      `query:"limit" required:"false" minimum:"1" maximum:"100" doc:"Max sessions to return (default 20)"`
	Tags      []string `query:"tags" required:"false" doc:"Filter by tags (any match)"`
}

type ListPlaygroundSessionsOutput struct {
	Body []*playgroundDomain.PlaygroundSessionSummary
}

type GetPlaygroundSessionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	SessionID string `path:"sessionId" format:"uuid"`
}

type GetPlaygroundSessionOutput struct {
	Body *playgroundDomain.SessionResponse
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

type DeleteSessionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	SessionID string `path:"sessionId" format:"uuid"`
}

type DeleteSessionOutput struct{}

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
