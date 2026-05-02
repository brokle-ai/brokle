package prompt

import (
	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/pkg/response"
)

// ---- list-prompts response shape ------------------------------------

type listPromptsResponse struct {
	Data       []*promptDomain.PromptListItem `json:"data"`
	Pagination *response.Pagination           `json:"pagination"`
}

// ---- labels ---------------------------------------------------------

type labelsResponse struct {
	Labels []string `json:"labels"`
}

type protectedLabelsResponse struct {
	ProtectedLabels []string `json:"protected_labels"`
}

// ---- compiler helpers -----------------------------------------------

type ValidateTemplateRequest struct {
	Template any                          `json:"template"`
	Type     promptDomain.PromptType      `json:"type"`
	Dialect  promptDomain.TemplateDialect `json:"dialect,omitempty"`
}

type ValidateTemplateResponse struct {
	Valid     bool                         `json:"valid"`
	Dialect   promptDomain.TemplateDialect `json:"dialect"`
	Variables []string                     `json:"variables"`
	Errors    []promptDomain.SyntaxError   `json:"errors"`
	Warnings  []promptDomain.SyntaxWarning `json:"warnings"`
}

type PreviewTemplateRequest struct {
	Template  any                          `json:"template"`
	Type      promptDomain.PromptType      `json:"type"`
	Variables map[string]any               `json:"variables"`
	Dialect   promptDomain.TemplateDialect `json:"dialect,omitempty"`
}

type PreviewTemplateResponse struct {
	Compiled any                          `json:"compiled"`
	Dialect  promptDomain.TemplateDialect `json:"dialect"`
}

type DetectDialectRequest struct {
	Template any                     `json:"template"`
	Type     promptDomain.PromptType `json:"type"`
}

type DetectDialectResponse struct {
	Dialect promptDomain.TemplateDialect `json:"dialect"`
}
