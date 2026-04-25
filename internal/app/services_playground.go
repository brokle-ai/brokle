package app

import (
	"log/slog"

	credentialsService "brokle/internal/core/services/credentials"
	playgroundService "brokle/internal/core/services/playground"
	promptService "brokle/internal/core/services/prompt"
)

// ProvidePlaygroundServices wires the playground session service.
// Cross-domain dependencies are explicit constructor params rather
// than reaches into the credentials/prompt service containers — keeps
// the orchestrator's wire order obvious and makes tests easier.
func ProvidePlaygroundServices(
	playgroundRepos *PlaygroundRepositories,
	credentialsService *credentialsService.ProviderCredentialService,
	compilerService *promptService.CompilerService,
	executionService *promptService.ExecutionService,
	logger *slog.Logger,
) *PlaygroundServices {
	playgroundSvc := playgroundService.NewPlaygroundService(
		playgroundRepos.Session,
		credentialsService,
		compilerService,
		executionService,
		logger,
	)

	return &PlaygroundServices{
		Playground: playgroundSvc,
	}
}
