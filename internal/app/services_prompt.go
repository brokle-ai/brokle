package app

import (
	"log/slog"

	"brokle/internal/config"
	commonDomain "brokle/internal/core/domain/common"
	analyticsService "brokle/internal/core/services/analytics"
	promptService "brokle/internal/core/services/prompt"
)

// ProvidePromptServices wires the prompt-domain services. Depends on
// analytics ProviderPricing because cost-tracking on prompt execution
// needs the pricing catalog.
func ProvidePromptServices(
	transactor commonDomain.Transactor,
	promptRepos *PromptRepositories,
	pricingService *analyticsService.ProviderPricingService,
	cfg *config.Config,
	logger *slog.Logger,
) *PromptServices {
	compilerSvc := promptService.NewCompilerService()
	aiClientConfig := &promptService.AIClientConfig{
		DefaultTimeout: cfg.External.LLMTimeout,
	}

	executionSvc := promptService.NewExecutionService(compilerSvc, pricingService, aiClientConfig)
	promptSvc := promptService.NewPromptService(
		transactor,
		promptRepos.Prompt,
		promptRepos.Version,
		promptRepos.Label,
		promptRepos.ProtectedLabel,
		promptRepos.Cache,
		compilerSvc,
		logger,
	)

	return &PromptServices{
		Prompt:    promptSvc,
		Compiler:  compilerSvc,
		Execution: executionSvc,
	}
}
