// Package playground provides service implementations for playground session management.
package playground

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	credentialsDomain "brokle/internal/core/domain/credentials"
	playgroundDomain "brokle/internal/core/domain/playground"
	promptDomain "brokle/internal/core/domain/prompt"
	appErrors "brokle/pkg/errors"
	credentials "brokle/internal/core/services/credentials"
	promptService "brokle/internal/core/services/prompt"
	"brokle/pkg/uid"
)

type PlaygroundService struct {
	repo               playgroundDomain.SessionRepository
	credentials *credentials.ProviderCredentialService
	compiler    *promptService.CompilerService
	exec   *promptService.ExecutionService
	logger             *slog.Logger
}

func NewPlaygroundService(
	repo playgroundDomain.SessionRepository,
	credentials *credentials.ProviderCredentialService,
	compiler *promptService.CompilerService,
	exec *promptService.ExecutionService,
	logger *slog.Logger,
) *PlaygroundService {
	return &PlaygroundService{
		repo:               repo,
		credentials: credentials,
		compiler:    compiler,
		exec:   exec,
		logger:             logger,
	}
}

// All sessions are saved (no ephemeral sessions).
func (s *PlaygroundService) CreateSession(ctx context.Context, req *playgroundDomain.CreatePlaygroundSessionRequest) (*playgroundDomain.SessionResponse, error) {
	if req.Name == "" {
		return nil, appErrors.InvalidParam("name", "required", appErrors.WithDetails("name is required"))
	}
	if len(req.Name) > playgroundDomain.MaxNameLength {
		return nil, appErrors.InvalidParam("name", "too long", appErrors.WithDetails("name must be 200 characters or less"))
	}

	if len(req.Windows) == 0 {
		return nil, appErrors.InvalidParam("windows", "required", appErrors.WithDetails("windows must be provided"))
	}

	if len(req.Tags) > playgroundDomain.MaxTagsCount {
		return nil, appErrors.InvalidParam("tags", "too many", appErrors.WithDetails("maximum 10 tags allowed"))
	}

	now := time.Now()

	var variables playgroundDomain.JSON
	if len(req.Variables) > 0 {
		variables = playgroundDomain.JSON(req.Variables)
	} else {
		variables = playgroundDomain.JSON([]byte("{}"))
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	name := req.Name
	session := &playgroundDomain.Session{
		ID:          uid.New(),
		ProjectID:   req.ProjectID,
		Name:        &name,
		Description: req.Description,
		Tags:        tags,
		Variables:   variables,
		Config:      playgroundDomain.JSON(req.Config),
		Windows:     playgroundDomain.JSON(req.Windows),
		CreatedBy:   req.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastUsedAt:  now,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create playground session",
			"error", err,
			"project_id", req.ProjectID,
		)
		return nil, appErrors.Internal("failed to create session", err)
	}

	s.logger.Info("playground session created",
		"session_id", session.ID,
		"project_id", req.ProjectID,
		"name", req.Name,
	)

	return session.ToResponse(), nil
}

func (s *PlaygroundService) GetSession(ctx context.Context, sessionID uuid.UUID) (*playgroundDomain.SessionResponse, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return session.ToResponse(), nil
}

func (s *PlaygroundService) ListSessions(ctx context.Context, req *playgroundDomain.ListSessionsRequest) ([]*playgroundDomain.PlaygroundSessionSummary, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var sessions []*playgroundDomain.Session
	var err error

	if len(req.Tags) > 0 {
		sessions, err = s.repo.ListByTags(ctx, req.ProjectID, req.Tags, limit)
	} else {
		sessions, err = s.repo.List(ctx, req.ProjectID, limit)
	}

	if err != nil {
		return nil, err
	}

	summaries := make([]*playgroundDomain.PlaygroundSessionSummary, len(sessions))
	for i, session := range sessions {
		summaries[i] = session.ToSummary()
	}

	return summaries, nil
}

func (s *PlaygroundService) UpdateSession(ctx context.Context, req *playgroundDomain.UpdateSessionRequest) (*playgroundDomain.SessionResponse, error) {
	session, err := s.repo.GetByID(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if len(*req.Name) > playgroundDomain.MaxNameLength {
			return nil, appErrors.InvalidParam("name", "too long", appErrors.WithDetails("name must be 200 characters or less"))
		}
		session.Name = req.Name
	}
	if req.Description != nil {
		session.Description = req.Description
	}
	if req.Tags != nil {
		if len(req.Tags) > playgroundDomain.MaxTagsCount {
			return nil, appErrors.InvalidParam("tags", "too many", appErrors.WithDetails("maximum 10 tags allowed"))
		}
		session.Tags = req.Tags
	}

	if len(req.Variables) > 0 {
		session.Variables = playgroundDomain.JSON(req.Variables)
	}
	if len(req.Config) > 0 {
		session.Config = playgroundDomain.JSON(req.Config)
	}
	if len(req.Windows) > 0 {
		session.Windows = playgroundDomain.JSON(req.Windows)
	}

	session.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, session); err != nil {
		s.logger.Error("failed to update playground session",
			"error", err,
			"session_id", req.SessionID,
		)
		return nil, err
	}

	return session.ToResponse(), nil
}

func (s *PlaygroundService) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.repo.Delete(ctx, sessionID); err != nil {
		return err
	}

	s.logger.Info("playground session deleted",
		"session_id", sessionID,
	)

	return nil
}

func (s *PlaygroundService) UpdateLastRun(ctx context.Context, req *playgroundDomain.UpdateLastRunRequest) error {
	if req.LastRun == nil {
		return appErrors.InvalidParam("last_run", "is required", appErrors.WithDetails("last_run cannot be empty"))
	}

	lastRunJSON, err := json.Marshal(req.LastRun)
	if err != nil {
		return appErrors.Internal("failed to serialize last run", err)
	}

	if err := s.repo.UpdateLastRun(ctx, req.SessionID, playgroundDomain.JSON(lastRunJSON)); err != nil {
		if !appErrors.IsNotFound(err) {
			s.logger.Error("failed to update last run",
				"error", err,
				"session_id", req.SessionID,
			)
		}
		return err
	}

	return nil
}

func (s *PlaygroundService) UpdateWindows(ctx context.Context, sessionID uuid.UUID, windows json.RawMessage) error {
	if err := s.repo.UpdateWindows(ctx, sessionID, playgroundDomain.JSON(windows)); err != nil {
		if !appErrors.IsNotFound(err) {
			s.logger.Error("failed to update windows",
				"error", err,
				"session_id", sessionID,
			)
		}
		return err
	}

	return nil
}

func (s *PlaygroundService) ValidateProjectAccess(ctx context.Context, sessionID uuid.UUID, projectID uuid.UUID) error {
	exists, err := s.repo.ExistsByProjectID(ctx, sessionID, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return appErrors.NotFound("session", appErrors.WithOp("service.playground.validate_project_access"))
	}
	return nil
}

// ExecutePrompt executes a prompt with full orchestration:
// credential resolution → variable extraction → execution → session update
func (s *PlaygroundService) ExecutePrompt(ctx context.Context, req *playgroundDomain.ExecuteRequest) (*playgroundDomain.ExecuteResponse, error) {
	startTime := time.Now()

	variables, err := s.compiler.ExtractVariables(req.Template, req.PromptType)
	if err != nil {
		return nil, appErrors.InvalidParam("template", "invalid template", appErrors.WithDetails(err.Error()))
	}

	resolvedConfig, err := s.resolveCredentials(ctx, req.OrganizationID, req.ConfigOverrides)
	if err != nil {
		return nil, err
	}

	promptResp := &promptDomain.PromptResponse{
		Type:      req.PromptType,
		Template:  req.Template,
		Variables: variables,
	}

	execResp, err := s.exec.Execute(ctx, promptResp, req.Variables, resolvedConfig)
	if err != nil {
		s.logger.Error("playground execution failed",
			"error", err,
			"project_id", req.ProjectID.String(),
			"organization_id", req.OrganizationID.String(),
		)
		return nil, appErrors.Internal("execution failed", err)
	}

	// Update session last_run (async, non-blocking)
	if req.SessionID != nil {
		go s.updateSessionLastRun(context.WithoutCancel(ctx), *req.SessionID, execResp, startTime)
	}

	s.logger.Debug("playground execution completed",
		"project_id", req.ProjectID.String(),
		"latency_ms", execResp.LatencyMs,
	)

	return &playgroundDomain.ExecuteResponse{
		CompiledPrompt: execResp.CompiledPrompt,
		Response:       execResp.Response,
		LatencyMs:      execResp.LatencyMs,
		Error:          execResp.Error,
	}, nil
}

func (s *PlaygroundService) StreamPrompt(ctx context.Context, req *playgroundDomain.StreamRequest) (*playgroundDomain.StreamResponse, error) {
	startTime := time.Now()

	variables, err := s.compiler.ExtractVariables(req.Template, req.PromptType)
	if err != nil {
		return nil, appErrors.InvalidParam("template", "invalid template", appErrors.WithDetails(err.Error()))
	}

	resolvedConfig, err := s.resolveCredentials(ctx, req.OrganizationID, req.ConfigOverrides)
	if err != nil {
		return nil, err
	}

	promptResp := &promptDomain.PromptResponse{
		Type:      req.PromptType,
		Template:  req.Template,
		Variables: variables,
	}

	eventChan, resultChan, err := s.exec.ExecuteStream(ctx, promptResp, req.Variables, resolvedConfig)
	if err != nil {
		s.logger.Error("playground stream execution failed",
			"error", err,
			"project_id", req.ProjectID.String(),
			"organization_id", req.OrganizationID.String(),
		)
		return nil, appErrors.Internal("stream execution failed", err)
	}

	// Wrap result channel to intercept for session update
	wrappedResultChan := s.wrapResultForSessionUpdate(ctx, req.SessionID, resultChan, startTime)

	return &playgroundDomain.StreamResponse{
		EventChan:  eventChan,
		ResultChan: wrappedResultChan,
	}, nil
}

// resolveCredentials resolves organization-scoped credentials for execution.
// Requires both provider and credential_id to be specified.
func (s *PlaygroundService) resolveCredentials(ctx context.Context, orgID uuid.UUID, overrides *promptDomain.ModelConfig) (*promptDomain.ModelConfig, error) {
	if overrides == nil {
		overrides = &promptDomain.ModelConfig{}
	}

	// Provider must be explicitly specified
	if overrides.Provider == "" {
		return nil, appErrors.InvalidParam("provider", "required", appErrors.WithDetails("provider must be specified"))
	}

	// Credential ID is required (no fallback to adapter-based lookup)
	if overrides.CredentialID == nil || *overrides.CredentialID == uuid.Nil {
		return nil, appErrors.InvalidParam("credential_id", "is required", appErrors.WithDetails("credential_id must be specified"))
	}

	if s.credentials == nil {
		return nil, appErrors.Internal("credentials service not configured", nil)
	}

	credID := *overrides.CredentialID

	keyConfig, err := s.credentials.GetExecutionConfig(ctx, orgID, credID, credentialsDomain.Provider(overrides.Provider))
	if err != nil {
		// Pass typed *Error through — credentials service already
		// emits canonical wire shapes (NotFound("credential") /
		// Conflict("credential", "provider adapter mismatch: ...")).
		// Downgrading Conflict → InvalidParam would hide the
		// state-machine signal and turn a 409 into a 422 with a
		// synthetic field-level param. Internal wrap is reserved
		// for raw (non-typed) errors.
		if appErrors.As(err) != nil {
			return nil, err
		}
		return nil, appErrors.Internal("failed to resolve credentials", err)
	}

	overrides.APIKey = keyConfig.APIKey
	if keyConfig.BaseURL != "" {
		overrides.ResolvedBaseURL = &keyConfig.BaseURL
	}

	// Pass provider-specific config (Azure deployment_id, api_version) and custom headers
	overrides.ProviderConfig = keyConfig.Config
	overrides.CustomHeaders = keyConfig.Headers

	s.logger.Debug("credentials resolved",
		"organization_id", orgID.String(),
		"provider", overrides.Provider,
		"credential_id", overrides.CredentialID,
	)

	return overrides, nil
}

// wrapResultForSessionUpdate intercepts the result channel to update session.
func (s *PlaygroundService) wrapResultForSessionUpdate(
	ctx context.Context,
	sessionID *uuid.UUID,
	resultChan <-chan *promptDomain.StreamResult,
	startTime time.Time,
) <-chan *promptDomain.StreamResult {
	wrappedChan := make(chan *promptDomain.StreamResult, 1)

	go func() {
		defer close(wrappedChan)

		for result := range resultChan {
			wrappedChan <- result

			if sessionID != nil && result != nil {
				go s.updateStreamSessionLastRun(context.WithoutCancel(ctx), *sessionID, result, startTime)
			}
		}
	}()

	return wrappedChan
}

func (s *PlaygroundService) updateSessionLastRun(
	ctx context.Context,
	sessionID uuid.UUID,
	execResp *promptDomain.ExecutePromptResponse,
	startTime time.Time,
) {
	lastRun := &playgroundDomain.LastRun{
		Timestamp: startTime,
		Metrics: &playgroundDomain.RunMetrics{
			LatencyMs: execResp.LatencyMs,
		},
	}

	if execResp.Response != nil {
		lastRun.Content = execResp.Response.Content
		if execResp.Response.Usage != nil {
			lastRun.Metrics.PromptTokens = execResp.Response.Usage.PromptTokens
			lastRun.Metrics.CompletionTokens = execResp.Response.Usage.CompletionTokens
			lastRun.Metrics.TotalTokens = execResp.Response.Usage.TotalTokens
		}
		if execResp.Response.Cost != nil {
			lastRun.Metrics.Cost = *execResp.Response.Cost
		}
		lastRun.Metrics.Model = execResp.Response.Model
	}

	if execResp.Error != "" {
		errStr := execResp.Error
		lastRun.Error = &errStr
	}

	req := &playgroundDomain.UpdateLastRunRequest{
		SessionID: sessionID,
		LastRun:   lastRun,
	}

	if err := s.UpdateLastRun(ctx, req); err != nil {
		s.logger.Warn("failed to update session last_run",
			"session_id", sessionID.String(),
			"error", err,
		)
	}
}

func (s *PlaygroundService) updateStreamSessionLastRun(
	ctx context.Context,
	sessionID uuid.UUID,
	result *promptDomain.StreamResult,
	startTime time.Time,
) {
	lastRun := &playgroundDomain.LastRun{
		Timestamp: startTime,
		Content:   result.Content,
		Metrics: &playgroundDomain.RunMetrics{
			LatencyMs: result.TotalDuration,
			Model:     result.Model,
		},
	}

	if result.Usage != nil {
		lastRun.Metrics.PromptTokens = result.Usage.PromptTokens
		lastRun.Metrics.CompletionTokens = result.Usage.CompletionTokens
		lastRun.Metrics.TotalTokens = result.Usage.TotalTokens
	}

	if result.Cost != nil {
		lastRun.Metrics.Cost = *result.Cost
	}

	if result.TTFTMs != nil {
		lastRun.Metrics.TTFTMs = int64(*result.TTFTMs)
	}

	req := &playgroundDomain.UpdateLastRunRequest{
		SessionID: sessionID,
		LastRun:   lastRun,
	}

	if err := s.UpdateLastRun(ctx, req); err != nil {
		s.logger.Warn("failed to update session last_run",
			"session_id", sessionID.String(),
			"error", err,
		)
	}
}
