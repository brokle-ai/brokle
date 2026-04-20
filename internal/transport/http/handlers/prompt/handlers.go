// Package prompt exposes the prompt-management operations on both planes:
//   - Dashboard plane (apiAdmin, RequireAuth) — full CRUD over prompts,
//     versions, labels, protected-label settings, and compiler helpers
//     (validate / preview / detect-dialect) under
//     /api/v1/projects/{projectId}/prompts.
//   - SDK plane (apiPublic, RequireSDKAuth) — upsert, list, and fetch-by-name
//     under /v1/prompts. Project ID is derived from the API key.
package prompt

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

type handler struct {
	promptSvc   promptDomain.PromptService
	compilerSvc promptDomain.CompilerService
	logger      *slog.Logger
}

// RegisterRoutes wires the dashboard-plane prompt operations.
func RegisterRoutes(
	api huma.API,
	promptSvc promptDomain.PromptService,
	compilerSvc promptDomain.CompilerService,
	logger *slog.Logger,
) {
	h := &handler{promptSvc: promptSvc, compilerSvc: compilerSvc, logger: logger}

	// ---- prompts -------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-prompts",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/prompts",
		Tags:        []string{"prompts"},
		Summary:     "List prompts for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listPrompts)

	huma.Register(api, huma.Operation{
		OperationID:   "create-prompt",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/prompts",
		Tags:          []string{"prompts"},
		Summary:       "Create a new prompt with initial version",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createPrompt)

	huma.Register(api, huma.Operation{
		OperationID: "get-prompt",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/prompts/{promptId}",
		Tags:        []string{"prompts"},
		Summary:     "Get a prompt with its latest version",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getPrompt)

	huma.Register(api, huma.Operation{
		OperationID: "update-prompt",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/prompts/{promptId}",
		Tags:        []string{"prompts"},
		Summary:     "Update prompt metadata",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updatePrompt)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-prompt",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/prompts/{promptId}",
		Tags:          []string{"prompts"},
		Summary:       "Soft-delete a prompt and its versions",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deletePrompt)

	// ---- versions ------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-prompt-versions",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/prompts/{promptId}/versions",
		Tags:        []string{"prompt-versions"},
		Summary:     "List all versions of a prompt",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listVersions)

	huma.Register(api, huma.Operation{
		OperationID:   "create-prompt-version",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/prompts/{promptId}/versions",
		Tags:          []string{"prompt-versions"},
		Summary:       "Create a new version of a prompt",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createVersion)

	huma.Register(api, huma.Operation{
		OperationID: "get-prompt-version",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/prompts/{promptId}/versions/{versionId}",
		Tags:        []string{"prompt-versions"},
		Summary:     "Get a specific version by ID or version number",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getVersion)

	huma.Register(api, huma.Operation{
		OperationID: "get-prompt-version-diff",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/prompts/{promptId}/diff",
		Tags:        []string{"prompt-versions"},
		Summary:     "Diff two versions of a prompt",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getVersionDiff)

	// ---- labels --------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "set-prompt-version-labels",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/{projectId}/prompts/{promptId}/versions/{versionId}/labels",
		Tags:        []string{"prompt-labels"},
		Summary:     "Set labels on a specific version",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.setLabels)

	huma.Register(api, huma.Operation{
		OperationID: "get-protected-labels",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/prompts/settings/protected-labels",
		Tags:        []string{"prompt-labels"},
		Summary:     "Get protected labels for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getProtectedLabels)

	huma.Register(api, huma.Operation{
		OperationID: "set-protected-labels",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/prompts/settings/protected-labels",
		Tags:        []string{"prompt-labels"},
		Summary:     "Set protected labels for a project",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.setProtectedLabels)

	// ---- compiler helpers ---------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "validate-prompt-template",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/prompts/validate-template",
		Tags:        []string{"prompt-compiler"},
		Summary:     "Validate template syntax and extract variables",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.validateTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "preview-prompt-template",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/prompts/preview-template",
		Tags:        []string{"prompt-compiler"},
		Summary:     "Compile template with variables without saving",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.previewTemplate)

	huma.Register(api, huma.Operation{
		OperationID: "detect-prompt-template-dialect",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/prompts/detect-dialect",
		Tags:        []string{"prompt-compiler"},
		Summary:     "Auto-detect template dialect from content",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.detectDialect)
}

// RegisterSDKRoutes wires the SDK-plane prompt operations (project derived
// from the API key).
func RegisterSDKRoutes(
	api huma.API,
	promptSvc promptDomain.PromptService,
	logger *slog.Logger,
) {
	h := &handler{promptSvc: promptSvc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID: "sdk-upsert-prompt",
		Method:      http.MethodPost,
		Path:        "/v1/prompts",
		Tags:        []string{"SDK - prompts"},
		Summary:     "Create or add a new version to a prompt (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.upsertPrompt)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-list-prompts",
		Method:      http.MethodGet,
		Path:        "/v1/prompts",
		Tags:        []string{"SDK - prompts"},
		Summary:     "List prompts for the authenticated project (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.listPromptsSDK)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-get-prompt-by-name",
		Method:      http.MethodGet,
		Path:        "/v1/prompts/{name}",
		Tags:        []string{"SDK - prompts"},
		Summary:     "Resolve a prompt by name with optional label or version (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.getPromptByName)
}

// ---- shared parsers --------------------------------------------------

func parseProject(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	return id, nil
}

func parsePrompt(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid prompt ID", "promptId must be a valid UUID")
	}
	return id, nil
}

func parseVersionID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid version ID", "versionId must be a valid UUID")
	}
	return id, nil
}

func userIDPtr(ctx context.Context) *uuid.UUID {
	uid, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &uid
}

// ---- pagination / list helpers --------------------------------------

func buildPromptFilters(typeStr, tagsStr, search string, page, limit int, sortBy, sortDir string) (*promptDomain.PromptFilters, error) {
	filters := &promptDomain.PromptFilters{}

	if typeStr != "" {
		pt := promptDomain.PromptType(typeStr)
		if pt != promptDomain.PromptTypeText && pt != promptDomain.PromptTypeChat {
			return nil, appErrors.NewValidationError("Invalid type", "type must be 'text' or 'chat'")
		}
		filters.Type = &pt
	}
	if tagsStr != "" {
		filters.Tags = strings.Split(tagsStr, ",")
	}
	if search != "" {
		s := search
		filters.Search = &s
	}

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	filters.Params = pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  sortBy,
		SortDir: sortDir,
	}
	return filters, nil
}

// ---- prompts: list --------------------------------------------------

func (h *handler) listPrompts(ctx context.Context, in *ListPromptsInput) (*ListPromptsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	filters, err := buildPromptFilters(in.Type, in.Tags, in.Search, in.Page, in.Limit, in.SortBy, in.SortDir)
	if err != nil {
		return nil, err
	}
	prompts, total, err := h.promptSvc.ListPrompts(ctx, projectID, filters)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: list failed", "project_id", projectID, "error", err)
		return nil, err
	}
	return &ListPromptsOutput{Body: listPromptsResponse{
		Data:  prompts,
		Total: total,
		Page:  filters.Params.Page,
		Limit: filters.Params.Limit,
	}}, nil
}

// ---- prompts: create ------------------------------------------------

func (h *handler) createPrompt(ctx context.Context, in *CreatePromptInput) (*CreatePromptOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	if in.Body.Name == "" {
		return nil, appErrors.NewValidationError("Missing name", "prompt name is required")
	}
	if in.Body.Template == nil {
		return nil, appErrors.NewValidationError("Missing template", "prompt template is required")
	}
	prompt, version, labels, err := h.promptSvc.CreatePrompt(ctx, projectID, userIDPtr(ctx), &in.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: create failed", "name", in.Body.Name, "error", err)
		return nil, err
	}
	return &CreatePromptOutput{Body: buildPromptResponse(prompt, version, labels)}, nil
}

// ---- prompts: get ---------------------------------------------------

func (h *handler) getPrompt(ctx context.Context, in *GetPromptInput) (*GetPromptOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	p, err := h.promptSvc.GetPromptByID(ctx, projectID, promptID)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: get-by-id failed", "prompt_id", promptID, "error", err)
		return nil, err
	}
	full, err := h.promptSvc.GetPrompt(ctx, p.ProjectID, p.Name, &promptDomain.GetPromptOptions{Label: "latest"})
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: get-with-version failed", "prompt_id", promptID, "error", err)
		return nil, err
	}
	return &GetPromptOutput{Body: full}, nil
}

// ---- prompts: update ------------------------------------------------

func (h *handler) updatePrompt(ctx context.Context, in *UpdatePromptInput) (*UpdatePromptOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	p, err := h.promptSvc.UpdatePrompt(ctx, projectID, promptID, &in.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: update failed", "prompt_id", promptID, "error", err)
		return nil, err
	}
	return &UpdatePromptOutput{Body: p}, nil
}

// ---- prompts: delete ------------------------------------------------

func (h *handler) deletePrompt(ctx context.Context, in *DeletePromptInput) (*DeletePromptOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	if err := h.promptSvc.DeletePrompt(ctx, projectID, promptID); err != nil {
		h.logger.ErrorContext(ctx, "prompt: delete failed", "prompt_id", promptID, "error", err)
		return nil, err
	}
	return &DeletePromptOutput{}, nil
}

// ---- versions: list -------------------------------------------------

func (h *handler) listVersions(ctx context.Context, in *ListVersionsInput) (*ListVersionsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	versions, err := h.promptSvc.ListVersions(ctx, projectID, promptID)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: list-versions failed", "prompt_id", promptID, "error", err)
		return nil, err
	}
	return &ListVersionsOutput{Body: versions}, nil
}

// ---- versions: create -----------------------------------------------

func (h *handler) createVersion(ctx context.Context, in *CreateVersionInput) (*CreateVersionOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	if in.Body.Template == nil {
		return nil, appErrors.NewValidationError("Missing template", "template is required")
	}
	version, labels, err := h.promptSvc.CreateVersion(ctx, projectID, promptID, userIDPtr(ctx), &in.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: create-version failed", "prompt_id", promptID, "error", err)
		return nil, err
	}
	return &CreateVersionOutput{Body: buildVersionResponse(version, labels)}, nil
}

// ---- versions: get (ID or version-number) ---------------------------

func (h *handler) getVersion(ctx context.Context, in *GetVersionInput) (*GetVersionOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	if versionNum, cerr := strconv.Atoi(in.VersionID); cerr == nil {
		resp, err := h.promptSvc.GetVersion(ctx, projectID, promptID, versionNum)
		if err != nil {
			h.logger.ErrorContext(ctx, "prompt: get-version failed", "prompt_id", promptID, "version", versionNum, "error", err)
			return nil, err
		}
		return &GetVersionOutput{Body: resp}, nil
	}
	versionID, err := uuid.Parse(in.VersionID)
	if err != nil {
		return nil, appErrors.NewValidationError("Invalid version ID", "versionId must be a valid UUID or integer version number")
	}
	resp, err := h.promptSvc.GetVersionByID(ctx, projectID, promptID, versionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: get-version-by-id failed", "version_id", versionID, "error", err)
		return nil, err
	}
	return &GetVersionOutput{Body: resp}, nil
}

// ---- versions: diff -------------------------------------------------

func (h *handler) getVersionDiff(ctx context.Context, in *GetVersionDiffInput) (*GetVersionDiffOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	diff, err := h.promptSvc.GetVersionDiff(ctx, projectID, promptID, in.From, in.To)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: diff failed", "prompt_id", promptID, "from", in.From, "to", in.To, "error", err)
		return nil, err
	}
	return &GetVersionDiffOutput{Body: diff}, nil
}

// ---- labels: set on version -----------------------------------------

func (h *handler) setLabels(ctx context.Context, in *SetLabelsInput) (*SetLabelsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	promptID, err := parsePrompt(in.PromptID)
	if err != nil {
		return nil, err
	}
	versionID, err := parseVersionID(in.VersionID)
	if err != nil {
		return nil, err
	}
	labels, err := h.promptSvc.SetLabels(ctx, projectID, promptID, versionID, userIDPtr(ctx), in.Body.Labels)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: set-labels failed", "version_id", versionID, "error", err)
		return nil, err
	}
	if labels == nil {
		labels = []string{}
	}
	return &SetLabelsOutput{Body: labelsResponse{Labels: labels}}, nil
}

// ---- labels: protected (get / set) ----------------------------------

func (h *handler) getProtectedLabels(ctx context.Context, in *ProtectedLabelsPathInput) (*GetProtectedLabelsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	labels, err := h.promptSvc.GetProtectedLabels(ctx, projectID)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: get-protected-labels failed", "project_id", projectID, "error", err)
		return nil, err
	}
	if labels == nil {
		labels = []string{}
	}
	return &GetProtectedLabelsOutput{Body: protectedLabelsResponse{ProtectedLabels: labels}}, nil
}

func (h *handler) setProtectedLabels(ctx context.Context, in *SetProtectedLabelsInput) (*SetProtectedLabelsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	labels, err := h.promptSvc.SetProtectedLabels(ctx, projectID, userIDPtr(ctx), in.Body.ProtectedLabels)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: set-protected-labels failed", "project_id", projectID, "error", err)
		return nil, err
	}
	if labels == nil {
		labels = []string{}
	}
	return &SetProtectedLabelsOutput{Body: protectedLabelsResponse{ProtectedLabels: labels}}, nil
}

// ---- compiler helpers: validate / preview / detect-dialect ---------

// ValidateTemplateRequest / PreviewTemplateRequest / DetectDialectRequest
// stay in this package — they are handler-plane shapes that wrap the
// compiler service. Kept exported so the generated OpenAPI schema has
// stable names on the SDK side.

func (h *handler) validateTemplate(ctx context.Context, in *ValidateTemplateInput) (*ValidateTemplateOutput, error) {
	if _, err := parseProject(in.ProjectID); err != nil {
		return nil, err
	}
	if in.Body.Template == nil {
		return nil, appErrors.NewValidationError("Missing template", "template is required")
	}
	if in.Body.Type != promptDomain.PromptTypeText && in.Body.Type != promptDomain.PromptTypeChat {
		return nil, appErrors.NewValidationError("Invalid type", "type must be 'text' or 'chat'")
	}

	dialect := in.Body.Dialect
	if dialect == "" || dialect == promptDomain.DialectAuto {
		detected, err := h.compilerSvc.DetectDialect(in.Body.Template, in.Body.Type)
		if err != nil {
			return nil, err
		}
		dialect = detected
	}
	result, err := h.compilerSvc.ValidateSyntax(in.Body.Template, in.Body.Type, dialect)
	if err != nil {
		return nil, err
	}
	variables, err := h.compilerSvc.ExtractVariablesWithDialect(in.Body.Template, in.Body.Type, dialect)
	if err != nil {
		variables = []string{}
	}
	if variables == nil {
		variables = []string{}
	}
	errs := result.Errors
	if errs == nil {
		errs = []promptDomain.SyntaxError{}
	}
	warns := result.Warnings
	if warns == nil {
		warns = []promptDomain.SyntaxWarning{}
	}
	return &ValidateTemplateOutput{Body: ValidateTemplateResponse{
		Valid:     result.Valid,
		Dialect:   result.Dialect,
		Variables: variables,
		Errors:    errs,
		Warnings:  warns,
	}}, nil
}

func (h *handler) previewTemplate(ctx context.Context, in *PreviewTemplateInput) (*PreviewTemplateOutput, error) {
	if _, err := parseProject(in.ProjectID); err != nil {
		return nil, err
	}
	if in.Body.Template == nil {
		return nil, appErrors.NewValidationError("Missing template", "template is required")
	}
	if in.Body.Variables == nil {
		return nil, appErrors.NewValidationError("Missing variables", "variables is required")
	}
	if in.Body.Type != promptDomain.PromptTypeText && in.Body.Type != promptDomain.PromptTypeChat {
		return nil, appErrors.NewValidationError("Invalid type", "type must be 'text' or 'chat'")
	}

	dialect := in.Body.Dialect
	if dialect == "" || dialect == promptDomain.DialectAuto {
		detected, err := h.compilerSvc.DetectDialect(in.Body.Template, in.Body.Type)
		if err != nil {
			return nil, err
		}
		dialect = detected
	}
	compiled, err := h.compilerSvc.CompileWithDialect(in.Body.Template, in.Body.Type, in.Body.Variables, dialect)
	if err != nil {
		return nil, err
	}

	var wrapped any
	switch in.Body.Type {
	case promptDomain.PromptTypeText:
		content, ok := compiled.(string)
		if !ok {
			return nil, appErrors.NewInternalError("unexpected compilation result type for text template", nil)
		}
		wrapped = promptDomain.TextTemplate{Content: content}
	case promptDomain.PromptTypeChat:
		msgs, ok := compiled.([]promptDomain.ChatMessage)
		if !ok {
			return nil, appErrors.NewInternalError("unexpected compilation result type for chat template", nil)
		}
		wrapped = promptDomain.ChatTemplate{Messages: msgs}
	default:
		wrapped = compiled
	}
	return &PreviewTemplateOutput{Body: PreviewTemplateResponse{Compiled: wrapped, Dialect: dialect}}, nil
}

func (h *handler) detectDialect(ctx context.Context, in *DetectDialectInput) (*DetectDialectOutput, error) {
	if _, err := parseProject(in.ProjectID); err != nil {
		return nil, err
	}
	if in.Body.Template == nil {
		return nil, appErrors.NewValidationError("Missing template", "template is required")
	}
	if in.Body.Type != promptDomain.PromptTypeText && in.Body.Type != promptDomain.PromptTypeChat {
		return nil, appErrors.NewValidationError("Invalid type", "type must be 'text' or 'chat'")
	}
	dialect, err := h.compilerSvc.DetectDialect(in.Body.Template, in.Body.Type)
	if err != nil {
		return nil, err
	}
	return &DetectDialectOutput{Body: DetectDialectResponse{Dialect: dialect}}, nil
}

// ---- SDK: upsert ----------------------------------------------------

func (h *handler) upsertPrompt(ctx context.Context, in *UpsertPromptInput) (*UpsertPromptOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	if in.Body.Name == "" {
		return nil, appErrors.NewValidationError("Missing name", "prompt name is required")
	}
	if in.Body.Template == nil {
		return nil, appErrors.NewValidationError("Missing template", "prompt template is required")
	}
	result, err := h.promptSvc.UpsertPrompt(ctx, projectID, nil, &in.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: upsert failed", "name", in.Body.Name, "error", err)
		return nil, err
	}
	return &UpsertPromptOutput{Body: result}, nil
}

// ---- SDK: list ------------------------------------------------------

func (h *handler) listPromptsSDK(ctx context.Context, in *ListPromptsSDKInput) (*ListPromptsOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	filters, err := buildPromptFilters(in.Type, in.Tags, in.Search, in.Page, in.Limit, in.SortBy, in.SortDir)
	if err != nil {
		return nil, err
	}
	prompts, total, err := h.promptSvc.ListPrompts(ctx, projectID, filters)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: sdk-list failed", "project_id", projectID, "error", err)
		return nil, err
	}
	return &ListPromptsOutput{Body: listPromptsResponse{
		Data:  prompts,
		Total: total,
		Page:  filters.Params.Page,
		Limit: filters.Params.Limit,
	}}, nil
}

// ---- SDK: get by name ----------------------------------------------

func (h *handler) getPromptByName(ctx context.Context, in *GetPromptByNameInput) (*GetPromptByNameOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	if in.Name == "" {
		return nil, appErrors.NewValidationError("Missing name", "prompt name path parameter is required")
	}
	opts := &promptDomain.GetPromptOptions{Label: "latest"}
	if in.Label != "" {
		opts.Label = in.Label
	}
	if in.Version > 0 {
		v := in.Version
		opts.Version = &v
		opts.Label = ""
	}
	if in.CacheTTL > 0 {
		t := in.CacheTTL
		opts.CacheTTL = &t
	}
	resp, err := h.promptSvc.GetPrompt(ctx, projectID, in.Name, opts)
	if err != nil {
		h.logger.ErrorContext(ctx, "prompt: sdk-get-by-name failed", "name", in.Name, "error", err)
		return nil, err
	}
	return &GetPromptByNameOutput{Body: resp}, nil
}

// ---- response builders ---------------------------------------------

func buildPromptResponse(prompt *promptDomain.Prompt, version *promptDomain.Version, labels []string) *promptDomain.PromptResponse {
	if labels == nil {
		labels = []string{}
	}
	return &promptDomain.PromptResponse{
		ID:            prompt.ID,
		ProjectID:     prompt.ProjectID,
		Name:          prompt.Name,
		Type:          prompt.Type,
		Description:   prompt.Description,
		Tags:          []string(prompt.Tags),
		Version:       version.Version,
		VersionID:     version.ID,
		Labels:        labels,
		Template:      version.Template,
		Config:        version.Config,
		Variables:     []string(version.Variables),
		CommitMessage: version.CommitMessage,
		CreatedAt:     version.CreatedAt,
		CreatedBy:     version.CreatedBy,
	}
}

func buildVersionResponse(version *promptDomain.Version, labels []string) *promptDomain.VersionResponse {
	if labels == nil {
		labels = []string{}
	}
	return &promptDomain.VersionResponse{
		ID:            version.ID,
		Version:       version.Version,
		Template:      version.Template,
		Config:        version.Config,
		Variables:     []string(version.Variables),
		CommitMessage: version.CommitMessage,
		Labels:        labels,
		CreatedAt:     version.CreatedAt,
		CreatedBy:     version.CreatedBy,
	}
}
