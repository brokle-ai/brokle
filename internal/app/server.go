package app

import (
	"fmt"

	credentialsService "brokle/internal/core/services/credentials"
	dashboardService "brokle/internal/core/services/dashboard"
	evaluationService "brokle/internal/core/services/evaluation"
	playgroundService "brokle/internal/core/services/playground"
	"brokle/internal/server"
	grpcTransport "brokle/internal/transport/grpc"
)

// ProvideServer wires the HTTP and gRPC servers from the populated
// CoreContainer. Optional service domains (Credentials, Playground,
// Evaluation, Dashboard) appear as nil in server.Deps when their
// service container is not built — handler RegisterRoutes calls in
// internal/server/routes.go are responsible for skipping nil deps
// (CLAUDE.md scaffolded-but-unreachable rule).
func ProvideServer(core *CoreContainer) (*ServerContainer, error) {
	var credentialsSvc *credentialsService.ProviderCredentialService
	var modelCatalogSvc *credentialsService.ModelCatalogService
	if core.Services.Credentials != nil {
		credentialsSvc = core.Services.Credentials.ProviderCredential
		modelCatalogSvc = core.Services.Credentials.ModelCatalog
	}

	var playgroundSvc *playgroundService.PlaygroundService
	if core.Services.Playground != nil {
		playgroundSvc = core.Services.Playground.Playground
	}

	var scoreConfigSvc *evaluationService.ScoreConfigService
	var datasetSvc *evaluationService.DatasetService
	var datasetItemSvc *evaluationService.DatasetItemService
	var datasetVersionSvc *evaluationService.DatasetVersionService
	var experimentSvc *evaluationService.ExperimentService
	var experimentItemSvc *evaluationService.ExperimentItemService
	var experimentWizardSvc *evaluationService.ExperimentWizardService
	var evaluatorSvc *evaluationService.EvaluatorService
	var evaluatorExecutionSvc *evaluationService.EvaluatorExecutionService
	if core.Services.Evaluation != nil {
		scoreConfigSvc = core.Services.Evaluation.ScoreConfig
		datasetSvc = core.Services.Evaluation.Dataset
		datasetItemSvc = core.Services.Evaluation.DatasetItem
		datasetVersionSvc = core.Services.Evaluation.DatasetVersion
		experimentSvc = core.Services.Evaluation.Experiment
		experimentItemSvc = core.Services.Evaluation.ExperimentItem
		experimentWizardSvc = core.Services.Evaluation.ExperimentWizard
		evaluatorSvc = core.Services.Evaluation.Evaluator
		evaluatorExecutionSvc = core.Services.Evaluation.EvaluatorExecution
	}

	var dashboardSvc *dashboardService.DashboardService
	var widgetQuerySvc *dashboardService.WidgetQueryService
	var templateSvc *dashboardService.TemplateService
	if core.Services.Dashboard != nil {
		dashboardSvc = core.Services.Dashboard.Dashboard
		widgetQuerySvc = core.Services.Dashboard.WidgetQuery
		templateSvc = core.Services.Dashboard.Template
	}

	httpServer, err := server.New(server.Deps{
		Config:                 core.Config,
		Logger:                 core.Logger,
		DB:                     core.Databases.Pool,
		Redis:                  core.Databases.Redis.Client,
		ClickHouse:             core.Databases.ClickHouse.Conn,
		JWT:                    core.Services.Auth.JWT,
		Blacklist:              core.Services.Auth.BlacklistedTokens,
		OrgMember:              core.Services.Auth.OrganizationMembers,
		ProjectMember:          core.Services.Auth.ProjectMembers,
		APIKey:                 core.Services.Auth.APIKey,
		Project:                core.Services.ProjectService,
		OrgMemberOrg:           core.Services.MemberService,
		Auth:                   core.Services.Auth.Auth,
		User:                   core.Services.User.User,
		Profile:                core.Services.User.Profile,
		Registration:           core.Services.Registration,
		Session:                core.Services.Auth.Sessions,
		OAuthProvider:          core.Services.Auth.OAuthProvider,
		Organization:           core.Services.OrganizationService,
		Comment:                core.Services.Comment,
		Overview:               core.Services.Analytics.Overview,
		Credential:             credentialsSvc,
		CredentialModelCatalog: modelCatalogSvc,
		Website:                core.Services.Website,
		Dashboard:              dashboardSvc,
		DashboardQuery:         widgetQuerySvc,
		DashboardTemplate:      templateSvc,
		AnnotationQueue:        core.Services.Annotation.Queue,
		AnnotationItem:         core.Services.Annotation.Item,
		AnnotationAssignment:   core.Services.Annotation.Assignment,

		BillingUsage:    core.Services.Billing.BillableUsage,
		BillingBudget:   core.Services.Billing.Budget,
		BillingContract: core.Services.Billing.Contract,
		BillingPricing:  core.Services.Billing.Pricing,

		Role:       core.Services.Auth.Role,
		Permission: core.Services.Auth.Permission,
		Scope:      core.Services.Auth.Scope,

		Invitation:  core.Services.InvitationService,
		OrgSettings: core.Services.SettingsService,

		Prompt:         core.Services.Prompt.Prompt,
		PromptCompiler: core.Services.Prompt.Compiler,

		Playground: playgroundSvc,

		EvalScoreConfig:        scoreConfigSvc,
		EvalDataset:            datasetSvc,
		EvalDatasetItem:        datasetItemSvc,
		EvalDatasetVersion:     datasetVersionSvc,
		EvalExperiment:         experimentSvc,
		EvalExperimentItem:     experimentItemSvc,
		EvalExperimentWizard:   experimentWizardSvc,
		EvalEvaluator:          evaluatorSvc,
		EvalEvaluatorExecution: evaluatorExecutionSvc,

		Observability: core.Services.Observability,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	grpcOTLPHandler := grpcTransport.NewOTLPHandler(
		core.Services.Observability.StreamProducer,
		core.Services.Observability.DeduplicationService,
		core.Services.Observability.OTLPConverterService,
		core.Logger,
	)

	grpcOTLPMetricsHandler := grpcTransport.NewOTLPMetricsHandler(
		core.Services.Observability.StreamProducer,
		core.Services.Observability.OTLPMetricsConverterService,
		core.Logger,
	)

	grpcOTLPLogsHandler := grpcTransport.NewOTLPLogsHandler(
		core.Services.Observability.StreamProducer,
		core.Services.Observability.OTLPLogsConverterService,
		core.Services.Observability.OTLPEventsConverterService,
		core.Logger,
	)

	grpcAuthInterceptor := grpcTransport.NewAuthInterceptor(
		core.Services.Auth.APIKey,
		core.Logger,
	)

	grpcServer, err := grpcTransport.NewServer(
		core.Config.GRPC.Port,
		grpcOTLPHandler,
		grpcOTLPMetricsHandler,
		grpcOTLPLogsHandler,
		grpcAuthInterceptor,
		core.Logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC server: %w", err)
	}

	core.Logger.Info("gRPC OTLP server initialized", "port", core.Config.GRPC.Port)

	return &ServerContainer{
		HTTPServer: httpServer,
		GRPCServer: grpcServer,
	}, nil
}
