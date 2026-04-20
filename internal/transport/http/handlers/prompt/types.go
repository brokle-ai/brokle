package prompt

import (
	promptDomain "brokle/internal/core/domain/prompt"
)

// Huma operation types for the prompt package.

type listPromptsResponse struct {
	Data  []*promptDomain.PromptListItem `json:"data"`
	Total int64                          `json:"total"`
	Page  int                            `json:"page"`
	Limit int                            `json:"limit"`
}

type ListPromptsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Type      string `query:"type" required:"false" enum:"text,chat"`
	Tags      string `query:"tags" required:"false" doc:"Comma-separated tag list"`
	Search    string `query:"search" required:"false"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	SortBy    string `query:"sort_by" required:"false"`
	SortDir   string `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type ListPromptsOutput struct {
	Body listPromptsResponse
}

type CreatePromptInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      promptDomain.CreatePromptRequest
}

type CreatePromptOutput struct {
	Body *promptDomain.PromptResponse
}

type GetPromptInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	PromptID  string `path:"promptId" format:"uuid"`
}

type GetPromptOutput struct {
	Body *promptDomain.PromptResponse
}

type UpdatePromptInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	PromptID  string `path:"promptId" format:"uuid"`
	Body      promptDomain.UpdatePromptRequest
}

type UpdatePromptOutput struct {
	Body *promptDomain.Prompt
}

type DeletePromptInput = GetPromptInput

type DeletePromptOutput struct{}

type ListVersionsInput = GetPromptInput

type ListVersionsOutput struct {
	Body []*promptDomain.VersionResponse
}

type CreateVersionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	PromptID  string `path:"promptId" format:"uuid"`
	Body      promptDomain.CreateVersionRequest
}

type CreateVersionOutput struct {
	Body *promptDomain.VersionResponse
}

type GetVersionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	PromptID  string `path:"promptId" format:"uuid"`
	// VersionID may be either a UUID or an integer version number, so we
	// deliberately do NOT tag it with format:"uuid".
	VersionID string `path:"versionId" doc:"Version UUID or integer version number"`
}

type GetVersionOutput struct {
	Body *promptDomain.VersionResponse
}

type GetVersionDiffInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	PromptID  string `path:"promptId" format:"uuid"`
	From      int    `query:"from" required:"true" doc:"From version number"`
	To        int    `query:"to" required:"true" doc:"To version number"`
}

type GetVersionDiffOutput struct {
	Body *promptDomain.VersionDiffResponse
}

type SetLabelsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	PromptID  string `path:"promptId" format:"uuid"`
	VersionID string `path:"versionId" format:"uuid"`
	Body      promptDomain.SetLabelsRequest
}

type labelsResponse struct {
	Labels []string `json:"labels"`
}

type SetLabelsOutput struct {
	Body labelsResponse
}

type ProtectedLabelsPathInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
}

type protectedLabelsResponse struct {
	ProtectedLabels []string `json:"protected_labels"`
}

type GetProtectedLabelsOutput struct {
	Body protectedLabelsResponse
}

type SetProtectedLabelsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      promptDomain.ProtectedLabelsRequest
}

type SetProtectedLabelsOutput struct {
	Body protectedLabelsResponse
}

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

type ValidateTemplateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      ValidateTemplateRequest
}

type ValidateTemplateOutput struct {
	Body ValidateTemplateResponse
}

type PreviewTemplateInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      PreviewTemplateRequest
}

type PreviewTemplateOutput struct {
	Body PreviewTemplateResponse
}

type DetectDialectInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      DetectDialectRequest
}

type DetectDialectOutput struct {
	Body DetectDialectResponse
}

type UpsertPromptInput struct {
	Body promptDomain.UpsertPromptRequest
}

type UpsertPromptOutput struct {
	Body *promptDomain.UpsertResponse
}

type ListPromptsSDKInput struct {
	Type    string `query:"type" required:"false" enum:"text,chat"`
	Tags    string `query:"tags" required:"false" doc:"Comma-separated tag list"`
	Search  string `query:"search" required:"false"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
	SortBy  string `query:"sort_by" required:"false"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc"`
}

type GetPromptByNameInput struct {
	Name     string `path:"name" doc:"Prompt name"`
	Label    string `query:"label" required:"false" doc:"Label to resolve (default: latest)"`
	Version  int    `query:"version" required:"false" doc:"Specific version number (takes precedence over label)"`
	CacheTTL int    `query:"cache_ttl" required:"false" doc:"Cache TTL in seconds"`
}

type GetPromptByNameOutput struct {
	Body *promptDomain.PromptResponse
}
