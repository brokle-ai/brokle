package prompt

import (
	"log/slog"

	"brokle/internal/config"
	promptDomain "brokle/internal/core/domain/prompt"
)

type Handler struct {
	config          *config.Config
	logger          *slog.Logger
	promptService   promptDomain.PromptService
	compilerService promptDomain.CompilerService
}

func NewHandler(
	cfg *config.Config,
	logger *slog.Logger,
	promptService promptDomain.PromptService,
	compilerService promptDomain.CompilerService,
) *Handler {
	return &Handler{
		config:          cfg,
		logger:          logger,
		promptService:   promptService,
		compilerService: compilerService,
	}
}
