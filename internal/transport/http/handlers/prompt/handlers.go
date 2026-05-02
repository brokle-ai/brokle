// Package prompt exposes the prompt-management operations on both planes:
//   - Dashboard plane (RequireAuth) — full CRUD over prompts, versions,
//     labels, protected-label settings, and compiler helpers
//     (validate / preview / detect-dialect) under
//     /api/v1/projects/{projectId}/prompts.
//   - SDK plane (RequireSDKAuth) — upsert, list, and fetch-by-name under
//     /v1/prompts. Project ID is derived from the API key.
package prompt

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	promptDomain "brokle/internal/core/domain/prompt"
	promptService "brokle/internal/core/services/prompt"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	promptSvc   *promptService.PromptService
	compilerSvc *promptService.CompilerService
	logger      *slog.Logger
}

// New constructs a Handler with all services any prompt Register* function
// might need (dashboard + SDK).
func New(
	promptSvc *promptService.PromptService,
	compilerSvc *promptService.CompilerService,
	logger *slog.Logger,
) *Handler {
	return &Handler{promptSvc: promptSvc, compilerSvc: compilerSvc, logger: logger}
}

// ---- shared helpers --------------------------------------------------

func userIDPtr(ctx context.Context) *uuid.UUID {
	uid, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &uid
}

// buildPromptFilters translates raw query-string values into a
// PromptFilters struct. Called from both the dashboard and SDK list
// endpoints; returns a typed validation error on bad input.
func buildPromptFilters(r *http.Request) (*promptDomain.PromptFilters, error) {
	q := r.URL.Query()
	filters := &promptDomain.PromptFilters{}

	if t := q.Get("type"); t != "" {
		pt := promptDomain.PromptType(t)
		if pt != promptDomain.PromptTypeText && pt != promptDomain.PromptTypeChat {
			return nil, appErrors.InvalidParam("type", "must be 'text' or 'chat'")
		}
		filters.Type = &pt
	}
	if tags := q.Get("tags"); tags != "" {
		filters.Tags = strings.Split(tags, ",")
	}
	if search := q.Get("search"); search != "" {
		s := search
		filters.Search = &s
	}

	page, err := request.QueryInt(r, "page", 1)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	limit, err := request.QueryInt(r, "limit", 50)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	filters.Params = pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}
	return filters, nil
}

// ---- prompts: list ---------------------------------------------------

func (h *Handler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	filters, err := buildPromptFilters(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	prompts, total, err := h.promptSvc.ListPrompts(r.Context(), projectID, filters)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: list failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, listPromptsResponse{
		Data:       prompts,
		Pagination: response.BuildPagination(filters.Params.Page, filters.Params.Limit, total),
	})
}

// ---- prompts: create -------------------------------------------------

func (h *Handler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body promptDomain.CreatePromptRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Name == "" {
		response.WriteError(w, appErrors.InvalidParam("name", "prompt name is required"))
		return
	}
	if body.Template == nil {
		response.WriteError(w, appErrors.InvalidParam("template", "prompt template is required"))
		return
	}
	prompt, version, labels, err := h.promptSvc.CreatePrompt(r.Context(), projectID, userIDPtr(r.Context()), &body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: create failed", "name", body.Name, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Created(w, buildPromptResponse(prompt, version, labels))
}

// ---- prompts: get ----------------------------------------------------

func (h *Handler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	p, err := h.promptSvc.GetPromptByID(r.Context(), projectID, promptID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: get-by-id failed",
			"prompt_id", promptID, "error", err)
		response.WriteError(w, err)
		return
	}
	full, err := h.promptSvc.GetPrompt(r.Context(), p.ProjectID, p.Name,
		&promptDomain.GetPromptOptions{Label: "latest"})
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: get-with-version failed",
			"prompt_id", promptID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, full)
}

// ---- prompts: update -------------------------------------------------

func (h *Handler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body promptDomain.UpdatePromptRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	p, err := h.promptSvc.UpdatePrompt(r.Context(), projectID, promptID, &body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: update failed",
			"prompt_id", promptID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, p)
}

// ---- prompts: delete -------------------------------------------------

func (h *Handler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.promptSvc.DeletePrompt(r.Context(), projectID, promptID); err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: delete failed",
			"prompt_id", promptID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ---- versions: list --------------------------------------------------

func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versions, err := h.promptSvc.ListVersions(r.Context(), projectID, promptID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: list-versions failed",
			"prompt_id", promptID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, versions)
}

// ---- versions: create ------------------------------------------------

func (h *Handler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body promptDomain.CreateVersionRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Template == nil {
		response.WriteError(w, appErrors.InvalidParam("template", "template is required"))
		return
	}
	version, labels, err := h.promptSvc.CreateVersion(r.Context(), projectID, promptID, userIDPtr(r.Context()), &body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: create-version failed",
			"prompt_id", promptID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Created(w, buildVersionResponse(version, labels))
}

// ---- versions: get (ID or version-number) ---------------------------

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	raw := chi.URLParam(r, "versionId")
	if versionNum, cerr := strconv.Atoi(raw); cerr == nil {
		resp, err := h.promptSvc.GetVersion(r.Context(), projectID, promptID, versionNum)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "prompt: get-version failed",
				"prompt_id", promptID, "version", versionNum, "error", err)
			response.WriteError(w, err)
			return
		}
		response.Success(w, resp)
		return
	}
	versionID, err := uuid.Parse(raw)
	if err != nil {
		response.WriteError(w, appErrors.InvalidParam("versionId", "must be a valid UUID or integer version number"))
		return
	}
	resp, err := h.promptSvc.GetVersionByID(r.Context(), projectID, promptID, versionID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: get-version-by-id failed",
			"version_id", versionID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// ---- versions: diff --------------------------------------------------

func (h *Handler) GetVersionDiff(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	from, err := request.QueryInt(r, "from", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	to, err := request.QueryInt(r, "to", 0)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	diff, err := h.promptSvc.GetVersionDiff(r.Context(), projectID, promptID, from, to)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: diff failed",
			"prompt_id", promptID, "from", from, "to", to, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, diff)
}

// ---- labels: set on version -----------------------------------------

func (h *Handler) SetLabels(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	promptID, err := request.URLParamUUID(r, "promptId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versionID, err := request.URLParamUUID(r, "versionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body promptDomain.SetLabelsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	labels, err := h.promptSvc.SetLabels(r.Context(), projectID, promptID, versionID, userIDPtr(r.Context()), body.Labels)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: set-labels failed",
			"version_id", versionID, "error", err)
		response.WriteError(w, err)
		return
	}
	if labels == nil {
		labels = []string{}
	}
	response.Success(w, labelsResponse{Labels: labels})
}

// ---- labels: protected (get / set) ----------------------------------

func (h *Handler) GetProtectedLabels(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	labels, err := h.promptSvc.GetProtectedLabels(r.Context(), projectID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: get-protected-labels failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}
	if labels == nil {
		labels = []string{}
	}
	response.Success(w, protectedLabelsResponse{ProtectedLabels: labels})
}

func (h *Handler) SetProtectedLabels(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body promptDomain.ProtectedLabelsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	labels, err := h.promptSvc.SetProtectedLabels(r.Context(), projectID, userIDPtr(r.Context()), body.ProtectedLabels)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: set-protected-labels failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}
	if labels == nil {
		labels = []string{}
	}
	response.Success(w, protectedLabelsResponse{ProtectedLabels: labels})
}

// ---- compiler helpers: validate / preview / detect-dialect ---------

func (h *Handler) ValidateTemplate(w http.ResponseWriter, r *http.Request) {
	var body ValidateTemplateRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Template == nil {
		response.WriteError(w, appErrors.InvalidParam("template", "template is required"))
		return
	}
	if body.Type != promptDomain.PromptTypeText && body.Type != promptDomain.PromptTypeChat {
		response.WriteError(w, appErrors.InvalidParam("type", "must be 'text' or 'chat'"))
		return
	}

	dialect := body.Dialect
	if dialect == "" || dialect == promptDomain.DialectAuto {
		detected, err := h.compilerSvc.DetectDialect(body.Template, body.Type)
		if err != nil {
			response.WriteError(w, err)
			return
		}
		dialect = detected
	}
	result, err := h.compilerSvc.ValidateSyntax(body.Template, body.Type, dialect)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	variables, err := h.compilerSvc.ExtractVariablesWithDialect(body.Template, body.Type, dialect)
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
	response.Success(w, ValidateTemplateResponse{
		Valid:     result.Valid,
		Dialect:   result.Dialect,
		Variables: variables,
		Errors:    errs,
		Warnings:  warns,
	})
}

func (h *Handler) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	var body PreviewTemplateRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Template == nil {
		response.WriteError(w, appErrors.InvalidParam("template", "template is required"))
		return
	}
	if body.Variables == nil {
		response.WriteError(w, appErrors.InvalidParam("variables", "variables is required"))
		return
	}
	if body.Type != promptDomain.PromptTypeText && body.Type != promptDomain.PromptTypeChat {
		response.WriteError(w, appErrors.InvalidParam("type", "must be 'text' or 'chat'"))
		return
	}

	dialect := body.Dialect
	if dialect == "" || dialect == promptDomain.DialectAuto {
		detected, err := h.compilerSvc.DetectDialect(body.Template, body.Type)
		if err != nil {
			response.WriteError(w, err)
			return
		}
		dialect = detected
	}
	compiled, err := h.compilerSvc.CompileWithDialect(body.Template, body.Type, body.Variables, dialect)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var wrapped any
	switch body.Type {
	case promptDomain.PromptTypeText:
		content, ok := compiled.(string)
		if !ok {
			response.WriteError(w, appErrors.Internal("unexpected compilation result type for text template", nil))
			return
		}
		wrapped = promptDomain.TextTemplate{Content: content}
	case promptDomain.PromptTypeChat:
		msgs, ok := compiled.([]promptDomain.ChatMessage)
		if !ok {
			response.WriteError(w, appErrors.Internal("unexpected compilation result type for chat template", nil))
			return
		}
		wrapped = promptDomain.ChatTemplate{Messages: msgs}
	default:
		wrapped = compiled
	}
	response.Success(w, PreviewTemplateResponse{Compiled: wrapped, Dialect: dialect})
}

func (h *Handler) DetectDialect(w http.ResponseWriter, r *http.Request) {
	var body DetectDialectRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Template == nil {
		response.WriteError(w, appErrors.InvalidParam("template", "template is required"))
		return
	}
	if body.Type != promptDomain.PromptTypeText && body.Type != promptDomain.PromptTypeChat {
		response.WriteError(w, appErrors.InvalidParam("type", "must be 'text' or 'chat'"))
		return
	}
	dialect, err := h.compilerSvc.DetectDialect(body.Template, body.Type)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, DetectDialectResponse{Dialect: dialect})
}

// ---- SDK: upsert ----------------------------------------------------

func (h *Handler) UpsertPrompt(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	var body promptDomain.UpsertPromptRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	if body.Name == "" {
		response.WriteError(w, appErrors.InvalidParam("name", "prompt name is required"))
		return
	}
	if body.Template == nil {
		response.WriteError(w, appErrors.InvalidParam("template", "prompt template is required"))
		return
	}
	result, err := h.promptSvc.UpsertPrompt(r.Context(), projectID, nil, &body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: upsert failed",
			"name", body.Name, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, result)
}

// ---- SDK: list ------------------------------------------------------

func (h *Handler) ListPromptsSDK(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	filters, err := buildPromptFilters(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	prompts, total, err := h.promptSvc.ListPrompts(r.Context(), projectID, filters)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: sdk-list failed",
			"project_id", projectID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, listPromptsResponse{
		Data:       prompts,
		Pagination: response.BuildPagination(filters.Params.Page, filters.Params.Limit, total),
	})
}

// ---- SDK: get by name ----------------------------------------------

func (h *Handler) GetPromptByName(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	name := chi.URLParam(r, "name")
	if name == "" {
		response.WriteError(w, appErrors.InvalidParam("name", "prompt name path parameter is required"))
		return
	}

	opts := &promptDomain.GetPromptOptions{Label: "latest"}
	if label := r.URL.Query().Get("label"); label != "" {
		opts.Label = label
	}
	if versionStr := r.URL.Query().Get("version"); versionStr != "" {
		v, err := strconv.Atoi(versionStr)
		if err != nil {
			response.WriteError(w, appErrors.InvalidParam("version", "must be an integer"))
			return
		}
		if v > 0 {
			opts.Version = &v
			opts.Label = ""
		}
	}
	if cacheTTLStr := r.URL.Query().Get("cache_ttl"); cacheTTLStr != "" {
		t, err := strconv.Atoi(cacheTTLStr)
		if err != nil {
			response.WriteError(w, appErrors.InvalidParam("cache_ttl", "must be an integer"))
			return
		}
		if t > 0 {
			opts.CacheTTL = &t
		}
	}
	resp, err := h.promptSvc.GetPrompt(r.Context(), projectID, name, opts)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "prompt: sdk-get-by-name failed",
			"name", name, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
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
