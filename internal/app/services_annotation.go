package app

import (
	"log/slog"

	commonDomain "brokle/internal/core/domain/common"
	annotationService "brokle/internal/core/services/annotation"
	observabilityService "brokle/internal/core/services/observability"
)

// ProvideAnnotationServices wires the human-in-the-loop annotation
// services. Item creation needs the evaluation domain (score config
// validation), the observability score writer (to persist annotated
// scores), and the project repo (for project-scope auth).
func ProvideAnnotationServices(
	transactor commonDomain.Transactor,
	annotationRepos *AnnotationRepositories,
	evaluationServices *EvaluationServices,
	observabilityServices *observabilityService.ServiceRegistry,
	orgRepos *OrganizationRepositories,
	logger *slog.Logger,
) *AnnotationServices {
	queueSvc := annotationService.NewQueueService(
		annotationRepos.Queue,
		annotationRepos.Item,
		annotationRepos.Assignment,
		logger,
	)

	itemSvc := annotationService.NewItemService(
		annotationRepos.Queue,
		annotationRepos.Item,
		annotationRepos.Assignment,
		evaluationServices.ScoreConfig,
		observabilityServices.ScoreService,
		orgRepos.Project,
		transactor,
		logger,
	)

	assignmentSvc := annotationService.NewAssignmentService(
		annotationRepos.Queue,
		annotationRepos.Assignment,
		logger,
	)

	return &AnnotationServices{
		Queue:      queueSvc,
		Item:       itemSvc,
		Assignment: assignmentSvc,
	}
}
