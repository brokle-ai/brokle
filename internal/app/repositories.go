package app

// Repository wiring. Each domain has its own Provide<Domain>Repositories
// constructor that returns a populated <Domain>Repositories struct;
// ProvideRepositories aggregates them into RepositoryContainer.
//
// Adding a new repository to one domain touches only its constructor
// and the corresponding container struct in containers.go.

import (
	"log/slog"

	analyticsRepo "brokle/internal/infrastructure/repository/analytics"
	annotationRepo "brokle/internal/infrastructure/repository/annotation"
	authRepo "brokle/internal/infrastructure/repository/auth"
	billingRepo "brokle/internal/infrastructure/repository/billing"
	credentialsRepo "brokle/internal/infrastructure/repository/credentials"
	dashboardRepo "brokle/internal/infrastructure/repository/dashboard"
	evaluationRepo "brokle/internal/infrastructure/repository/evaluation"
	observabilityRepo "brokle/internal/infrastructure/repository/observability"
	orgRepo "brokle/internal/infrastructure/repository/organization"
	playgroundRepo "brokle/internal/infrastructure/repository/playground"
	promptRepo "brokle/internal/infrastructure/repository/prompt"
	storageRepo "brokle/internal/infrastructure/repository/storage"
	userRepo "brokle/internal/infrastructure/repository/user"
	websiteRepo "brokle/internal/infrastructure/repository/website"

	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/db"
)

// ProvideRepositories aggregates the per-domain repository constructors
// into the canonical RepositoryContainer. Called from ProvideCore.
func ProvideRepositories(dbs *DatabaseContainer, logger *slog.Logger) *RepositoryContainer {
	return &RepositoryContainer{
		User:          ProvideUserRepositories(dbs.TxManager),
		Auth:          ProvideAuthRepositories(dbs.TxManager),
		Organization:  ProvideOrganizationRepositories(dbs.TxManager),
		Observability: ProvideObservabilityRepositories(dbs.ClickHouse, dbs.TxManager, dbs.Redis),
		Storage:       ProvideStorageRepositories(dbs.ClickHouse),
		Billing:       ProvideBillingRepositories(dbs.TxManager, dbs.ClickHouse, logger),
		Analytics:     ProvideAnalyticsRepositories(dbs.TxManager, dbs.ClickHouse),
		Prompt:        ProvidePromptRepositories(dbs.TxManager, dbs.Redis),
		Credentials:   ProvideCredentialsRepositories(dbs.TxManager),
		Playground:    ProvidePlaygroundRepositories(dbs.TxManager),
		Evaluation:    ProvideEvaluationRepositories(dbs.TxManager),
		Dashboard:     ProvideDashboardRepositories(dbs.TxManager, dbs.ClickHouse),
		Annotation:    ProvideAnnotationRepositories(dbs.TxManager),
		Website:       ProvideWebsiteRepositories(dbs.TxManager),
	}
}

// ----- Per-domain repository constructors ---------------------------

func ProvideUserRepositories(tm *db.TxManager) *UserRepositories {
	return &UserRepositories{
		User: userRepo.NewUserRepository(tm),
	}
}

func ProvideAuthRepositories(tm *db.TxManager) *AuthRepositories {
	return &AuthRepositories{
		UserSession:        authRepo.NewUserSessionRepository(tm),
		BlacklistedToken:   authRepo.NewBlacklistedTokenRepository(tm),
		PasswordResetToken: authRepo.NewPasswordResetTokenRepository(tm),
		APIKey:             authRepo.NewAPIKeyRepository(tm),
		Role:               authRepo.NewRoleRepository(tm),
		OrganizationMember: authRepo.NewOrganizationMemberRepository(tm),
		Permission:         authRepo.NewPermissionRepository(tm),
		RolePermission:     authRepo.NewRolePermissionRepository(tm),
		AuditLog:           authRepo.NewAuditLogRepository(tm),
	}
}

func ProvideOrganizationRepositories(tm *db.TxManager) *OrganizationRepositories {
	return &OrganizationRepositories{
		Organization: orgRepo.NewOrganizationRepository(tm),
		Member:       orgRepo.NewMemberRepository(tm),
		Project:      orgRepo.NewProjectRepository(tm),
		Invitation:   orgRepo.NewInvitationRepository(tm),
		Settings:     orgRepo.NewOrganizationSettingsRepository(tm),
	}
}

func ProvideObservabilityRepositories(clickhouseDB *database.ClickHouseDB, tm *db.TxManager, redisDB *database.RedisDB) *ObservabilityRepositories {
	return &ObservabilityRepositories{
		Trace:                  observabilityRepo.NewTraceRepository(clickhouseDB.Conn),
		Score:                  observabilityRepo.NewScoreRepository(clickhouseDB.Conn),
		ScoreAnalytics:         observabilityRepo.NewScoreAnalyticsRepository(clickhouseDB.Conn),
		Metrics:                observabilityRepo.NewMetricsRepository(clickhouseDB.Conn),
		Logs:                   observabilityRepo.NewLogsRepository(clickhouseDB.Conn),
		GenAIEvents:            observabilityRepo.NewGenAIEventsRepository(clickhouseDB.Conn),
		TelemetryDeduplication: observabilityRepo.NewTelemetryDeduplicationRepository(redisDB),
		FilterPreset:           observabilityRepo.NewFilterPresetRepository(tm),
	}
}

func ProvideStorageRepositories(clickhouseDB *database.ClickHouseDB) *StorageRepositories {
	return &StorageRepositories{
		BlobStorage: storageRepo.NewBlobStorageRepository(clickhouseDB.Conn),
	}
}

func ProvideBillingRepositories(tm *db.TxManager, clickhouseDB *database.ClickHouseDB, _ *slog.Logger) *BillingRepositories {
	return &BillingRepositories{
		// Usage-based billing repositories
		BillableUsage:       billingRepo.NewBillableUsageRepository(clickhouseDB.Conn),
		Plan:                billingRepo.NewPlanRepository(tm),
		OrganizationBilling: billingRepo.NewOrganizationBillingRepository(tm),
		UsageBudget:         billingRepo.NewUsageBudgetRepository(tm),
		UsageAlert:          billingRepo.NewUsageAlertRepository(tm),
		// Enterprise custom pricing repositories
		Contract:        billingRepo.NewContractRepository(tm),
		VolumeTier:      billingRepo.NewVolumeDiscountTierRepository(tm),
		ContractHistory: billingRepo.NewContractHistoryRepository(tm),
	}
}

func ProvideAnalyticsRepositories(tm *db.TxManager, clickhouseDB *database.ClickHouseDB) *AnalyticsRepositories {
	return &AnalyticsRepositories{
		ProviderModel: analyticsRepo.NewProviderModelRepository(tm),
		Overview:      analyticsRepo.NewOverviewRepository(clickhouseDB.Conn),
	}
}

func ProvidePromptRepositories(tm *db.TxManager, redisDB *database.RedisDB) *PromptRepositories {
	return &PromptRepositories{
		Prompt:         promptRepo.NewPromptRepository(tm),
		Version:        promptRepo.NewVersionRepository(tm),
		Label:          promptRepo.NewLabelRepository(tm),
		ProtectedLabel: promptRepo.NewProtectedLabelRepository(tm),
		Cache:          promptRepo.NewCacheRepository(redisDB),
	}
}

func ProvideCredentialsRepositories(tm *db.TxManager) *CredentialsRepositories {
	return &CredentialsRepositories{
		ProviderCredential: credentialsRepo.NewProviderCredentialRepository(tm),
	}
}

func ProvidePlaygroundRepositories(tm *db.TxManager) *PlaygroundRepositories {
	return &PlaygroundRepositories{
		Session: playgroundRepo.NewSessionRepository(tm),
	}
}

func ProvideEvaluationRepositories(tm *db.TxManager) *EvaluationRepositories {
	return &EvaluationRepositories{
		ScoreConfig:        evaluationRepo.NewScoreConfigRepository(tm),
		Dataset:            evaluationRepo.NewDatasetRepository(tm),
		DatasetItem:        evaluationRepo.NewDatasetItemRepository(tm),
		DatasetVersion:     evaluationRepo.NewDatasetVersionRepository(tm),
		Experiment:         evaluationRepo.NewExperimentRepository(tm),
		ExperimentItem:     evaluationRepo.NewExperimentItemRepository(tm),
		ExperimentConfig:   evaluationRepo.NewExperimentConfigRepository(tm),
		Evaluator:          evaluationRepo.NewEvaluatorRepository(tm),
		EvaluatorExecution: evaluationRepo.NewEvaluatorExecutionRepository(tm),
	}
}

func ProvideDashboardRepositories(tm *db.TxManager, clickhouseDB *database.ClickHouseDB) *DashboardRepositories {
	return &DashboardRepositories{
		Dashboard:   dashboardRepo.NewDashboardRepository(tm),
		WidgetQuery: dashboardRepo.NewWidgetQueryRepository(clickhouseDB.Conn),
		Template:    dashboardRepo.NewTemplateRepository(tm),
	}
}

func ProvideAnnotationRepositories(tm *db.TxManager) *AnnotationRepositories {
	return &AnnotationRepositories{
		Queue:      annotationRepo.NewQueueRepository(tm),
		Item:       annotationRepo.NewItemRepository(tm),
		Assignment: annotationRepo.NewAssignmentRepository(tm),
	}
}

func ProvideWebsiteRepositories(tm *db.TxManager) *WebsiteRepositories {
	return &WebsiteRepositories{
		ContactSubmission: websiteRepo.NewContactSubmissionRepository(tm),
	}
}
