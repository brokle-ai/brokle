package app

import (
	"log/slog"

	commonDomain "brokle/internal/core/domain/common"
	evaluationService "brokle/internal/core/services/evaluation"
	observabilityService "brokle/internal/core/services/observability"
	"brokle/internal/infrastructure/database"
)

// ProvideEvaluationServices wires the evaluation domain — score
// configs, datasets/items/versions, experiments, evaluator pipeline.
//
// EvaluatorExecutionService must be built before EvaluatorService
// because the latter consumes it. Evaluator additionally needs the
// observability trace repo (to fetch span context for evaluations)
// and Redis (for the running-execution lock).
func ProvideEvaluationServices(
	transactor commonDomain.Transactor,
	evaluationRepos *EvaluationRepositories,
	observabilityRepos *ObservabilityRepositories,
	observabilityServices *observabilityService.ServiceRegistry,
	promptRepos *PromptRepositories,
	redisDB *database.RedisDB,
	logger *slog.Logger,
) *EvaluationServices {
	scoreConfigSvc := evaluationService.NewScoreConfigService(
		evaluationRepos.ScoreConfig,
		observabilityRepos.Score,
		logger,
	)

	datasetSvc := evaluationService.NewDatasetService(
		evaluationRepos.Dataset,
		logger,
	)

	datasetItemSvc := evaluationService.NewDatasetItemService(
		evaluationRepos.DatasetItem,
		evaluationRepos.Dataset,
		observabilityRepos.Trace,
		logger,
	)

	datasetVersionSvc := evaluationService.NewDatasetVersionService(
		transactor,
		evaluationRepos.DatasetVersion,
		evaluationRepos.Dataset,
		evaluationRepos.DatasetItem,
		logger,
	)

	experimentSvc := evaluationService.NewExperimentService(
		evaluationRepos.Experiment,
		evaluationRepos.Dataset,
		observabilityRepos.Score,
		logger,
	)

	experimentItemSvc := evaluationService.NewExperimentItemService(
		evaluationRepos.ExperimentItem,
		evaluationRepos.Experiment,
		evaluationRepos.DatasetItem,
		observabilityServices.ScoreService,
		logger,
	)

	experimentWizardSvc := evaluationService.NewExperimentWizardService(
		transactor,
		evaluationRepos.Experiment,
		evaluationRepos.ExperimentConfig,
		evaluationRepos.Dataset,
		evaluationRepos.DatasetItem,
		evaluationRepos.DatasetVersion,
		promptRepos.Prompt,
		promptRepos.Version,
		logger,
	)

	evaluatorExecutionSvc := evaluationService.NewEvaluatorExecutionService(
		evaluationRepos.EvaluatorExecution,
		logger,
	)

	evaluatorSvc := evaluationService.NewEvaluatorService(
		evaluationRepos.Evaluator,
		evaluatorExecutionSvc,
		observabilityRepos.Trace,
		redisDB,
		logger,
	)

	return &EvaluationServices{
		ScoreConfig:        scoreConfigSvc,
		Dataset:            datasetSvc,
		DatasetItem:        datasetItemSvc,
		DatasetVersion:     datasetVersionSvc,
		Experiment:         experimentSvc,
		ExperimentItem:     experimentItemSvc,
		ExperimentWizard:   experimentWizardSvc,
		Evaluator:          evaluatorSvc,
		EvaluatorExecution: evaluatorExecutionSvc,
	}
}
