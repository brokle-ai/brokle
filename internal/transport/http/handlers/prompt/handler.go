package prompt

import (
	"log/slog"

	"brokle/internal/config"
	promptDomain "brokle/internal/core/domain/prompt"
)

// Handler contains all prompt-related HTTP handlers
type Handler struct {
	config           *config.Config
	logger           *slog.Logger
	promptService    promptDomain.PromptService
	compilerService  promptDomain.CompilerService
	executionService promptDomain.ExecutionService
}

// NewHandler creates a new prompt handler
func NewHandler(
	cfg *config.Config,
	logger *slog.Logger,
	promptService promptDomain.PromptService,
	compilerService promptDomain.CompilerService,
	executionService promptDomain.ExecutionService,
) *Handler {
	return &Handler{
		config:           cfg,
		logger:           logger,
		promptService:    promptService,
		compilerService:  compilerService,
		executionService: executionService,
	}
}
