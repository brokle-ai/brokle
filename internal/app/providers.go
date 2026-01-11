package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gorm.io/gorm"

	"brokle/internal/config"
	"brokle/internal/core/domain/analytics"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/billing"
	"brokle/internal/core/domain/common"
	credentialsDomain "brokle/internal/core/domain/credentials"
	dashboardDomain "brokle/internal/core/domain/dashboard"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	"brokle/internal/core/domain/organization"
	playgroundDomain "brokle/internal/core/domain/playground"
	promptDomain "brokle/internal/core/domain/prompt"
	storageDomain "brokle/internal/core/domain/storage"
	"brokle/internal/core/domain/user"
	analyticsService "brokle/internal/core/services/analytics"
	authService "brokle/internal/core/services/auth"
	billingService "brokle/internal/core/services/billing"
	credentialsService "brokle/internal/core/services/credentials"
	dashboardService "brokle/internal/core/services/dashboard"
	evaluationService "brokle/internal/core/services/evaluation"
	observabilityService "brokle/internal/core/services/observability"
	orgService "brokle/internal/core/services/organization"
	playgroundService "brokle/internal/core/services/playground"
	promptService "brokle/internal/core/services/prompt"
	registrationService "brokle/internal/core/services/registration"
	storageService "brokle/internal/core/services/storage"
	userService "brokle/internal/core/services/user"
	eeAnalytics "brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/infrastructure/database"
	analyticsRepo "brokle/internal/infrastructure/repository/analytics"
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
	"brokle/internal/infrastructure/storage"
	"brokle/internal/infrastructure/streams"
	grpcTransport "brokle/internal/transport/grpc"
	"brokle/internal/transport/http"
	"brokle/internal/transport/http/handlers"
	"brokle/internal/workers"
	evaluationWorker "brokle/internal/workers/evaluation"
	"brokle/pkg/email"
	"brokle/pkg/encryption"
	"brokle/pkg/ulid"
)

type DeploymentMode string

const (
	ModeServer DeploymentMode = "server"
	ModeWorker DeploymentMode = "worker"
)

type CoreContainer struct {
	Config     *config.Config
	Logger     *slog.Logger
	Databases  *DatabaseContainer
	Repos      *RepositoryContainer
	TxManager  common.TransactionManager
	Services   *ServiceContainer
	Enterprise *EnterpriseContainer
}

type ServerContainer struct {
	HTTPServer *http.Server
	GRPCServer *grpcTransport.Server
}

type ProviderContainer struct {
	Core    *CoreContainer
	Server  *ServerContainer // nil in worker mode
	Workers *WorkerContainer // nil in server mode
	Mode    DeploymentMode
}

type DatabaseContainer struct {
	Postgres   *database.PostgresDB
	Redis      *database.RedisDB
	ClickHouse *database.ClickHouseDB
}

type WorkerContainer struct {
	TelemetryConsumer    *workers.TelemetryStreamConsumer
	RuleWorker           *evaluationWorker.RuleWorker
	EvaluationWorker     *evaluationWorker.EvaluationWorker
	ManualTriggerWorker  *evaluationWorker.ManualTriggerWorker
}

type RepositoryContainer struct {
	User          *UserRepositories
	Auth          *AuthRepositories
	Organization  *OrganizationRepositories
	Observability *ObservabilityRepositories
	Storage       *StorageRepositories
	Billing       *BillingRepositories
	Analytics     *AnalyticsRepositories
	Prompt        *PromptRepositories
	Credentials   *CredentialsRepositories
	Playground    *PlaygroundRepositories
	Evaluation    *EvaluationRepositories
	Dashboard     *DashboardRepositories
}

type ServiceContainer struct {
	User                *UserServices
	Auth                *AuthServices
	Registration        registrationService.RegistrationService
	OrganizationService organization.OrganizationService
	MemberService       organization.MemberService
	ProjectService      organization.ProjectService
	InvitationService   organization.InvitationService
	SettingsService     organization.OrganizationSettingsService
	Observability       *observabilityService.ServiceRegistry
	Billing             *BillingServices
	Analytics           *AnalyticsServices
	Prompt              *PromptServices
	Credentials         *CredentialsServices
	Playground          *PlaygroundServices
	Evaluation          *EvaluationServices
	Dashboard           *DashboardServices
}

type EnterpriseContainer struct {
	SSO        sso.SSOProvider
	RBAC       rbac.RBACManager
	Compliance compliance.Compliance
	Analytics  eeAnalytics.EnterpriseAnalytics
}

type UserRepositories struct {
	User user.Repository
}

type AuthRepositories struct {
	UserSession        auth.UserSessionRepository
	BlacklistedToken   auth.BlacklistedTokenRepository
	PasswordResetToken auth.PasswordResetTokenRepository
	APIKey             auth.APIKeyRepository
	Role               auth.RoleRepository
	OrganizationMember auth.OrganizationMemberRepository
	Permission         auth.PermissionRepository
	RolePermission     auth.RolePermissionRepository
	AuditLog           auth.AuditLogRepository
}

type OrganizationRepositories struct {
	Organization organization.OrganizationRepository
	Member       organization.MemberRepository
	Project      organization.ProjectRepository
	Invitation   organization.InvitationRepository
	Settings     organization.OrganizationSettingsRepository
}

type ObservabilityRepositories struct {
	Trace                  observability.TraceRepository
	Score                  observability.ScoreRepository
	ScoreAnalytics         observability.ScoreAnalyticsRepository
	Metrics                observability.MetricsRepository
	Logs                   observability.LogsRepository
	GenAIEvents            observability.GenAIEventsRepository
	TelemetryDeduplication observability.TelemetryDeduplicationRepository
	FilterPreset           observability.FilterPresetRepository
}

type StorageRepositories struct {
	BlobStorage storageDomain.BlobStorageRepository
}

type BillingRepositories struct {
	Usage         billing.UsageRepository
	BillingRecord billing.BillingRecordRepository
	Quota         billing.QuotaRepository
}

type AnalyticsRepositories struct {
	ProviderModel analytics.ProviderModelRepository
	Overview      analytics.OverviewRepository
}

type PromptRepositories struct {
	Prompt         promptDomain.PromptRepository
	Version        promptDomain.VersionRepository
	Label          promptDomain.LabelRepository
	ProtectedLabel promptDomain.ProtectedLabelRepository
	Cache          promptDomain.CacheRepository
}

type CredentialsRepositories struct {
	ProviderCredential credentialsDomain.ProviderCredentialRepository
}

type PlaygroundRepositories struct {
	Session playgroundDomain.SessionRepository
}

type EvaluationRepositories struct {
	ScoreConfig    evaluationDomain.ScoreConfigRepository
	Dataset        evaluationDomain.DatasetRepository
	DatasetItem    evaluationDomain.DatasetItemRepository
	Experiment     evaluationDomain.ExperimentRepository
	ExperimentItem evaluationDomain.ExperimentItemRepository
	Rule           evaluationDomain.RuleRepository
	RuleExecution  evaluationDomain.RuleExecutionRepository
}

type DashboardRepositories struct {
	Dashboard   dashboardDomain.DashboardRepository
	WidgetQuery dashboardDomain.WidgetQueryRepository
	Template    dashboardDomain.TemplateRepository
}

type UserServices struct {
	User    user.UserService
	Profile user.ProfileService
}

type AuthServices struct {
	Auth                auth.AuthService
	JWT                 auth.JWTService
	Sessions            auth.SessionService
	APIKey              auth.APIKeyService
	Role                auth.RoleService
	Permission          auth.PermissionService
	OrganizationMembers auth.OrganizationMemberService
	BlacklistedTokens   auth.BlacklistedTokenService
	Scope               auth.ScopeService
	OAuthProvider       *authService.OAuthProviderService
}

type BillingServices struct {
	Billing *billingService.BillingService
}

type AnalyticsServices struct {
	ProviderPricing analytics.ProviderPricingService
	Overview        analytics.OverviewService
}

type PromptServices struct {
	Prompt    promptDomain.PromptService
	Compiler  promptDomain.CompilerService
	Execution promptDomain.ExecutionService
}

type CredentialsServices struct {
	ProviderCredential credentialsDomain.ProviderCredentialService
	ModelCatalog       credentialsService.ModelCatalogService
}

type PlaygroundServices struct {
	Playground playgroundDomain.PlaygroundService
}

type EvaluationServices struct {
	ScoreConfig    evaluationDomain.ScoreConfigService
	Dataset        evaluationDomain.DatasetService
	DatasetItem    evaluationDomain.DatasetItemService
	Experiment     evaluationDomain.ExperimentService
	ExperimentItem evaluationDomain.ExperimentItemService
	Rule           evaluationDomain.RuleService
	RuleExecution  evaluationDomain.RuleExecutionService
}

type DashboardServices struct {
	Dashboard   dashboardDomain.DashboardService
	WidgetQuery dashboardDomain.WidgetQueryService
	Template    dashboardDomain.TemplateService
}

func ProvideDatabases(cfg *config.Config, logger *slog.Logger) (*DatabaseContainer, error) {
	postgres, err := database.NewPostgresDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	redis, err := database.NewRedisDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	clickhouse, err := database.NewClickHouseDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &DatabaseContainer{
		Postgres:   postgres,
		Redis:      redis,
		ClickHouse: clickhouse,
	}, nil
}

func ProvideWorkers(core *CoreContainer) (*WorkerContainer, error) {
	deduplicationService := observabilityService.NewTelemetryDeduplicationService(
		core.Repos.Observability.TelemetryDeduplication,
	)

	consumerConfig := &workers.TelemetryStreamConsumerConfig{
		ConsumerGroup:     "telemetry-workers",
		ConsumerID:        "worker-" + ulid.New().String(),
		BatchSize:         50,
		BlockDuration:     time.Second,
		MaxRetries:        3,
		RetryBackoff:      500 * time.Millisecond,
		DiscoveryInterval: 30 * time.Second,
		MaxStreamsPerRead: 10,
	}

	telemetryConsumer := workers.NewTelemetryStreamConsumer(
		core.Databases.Redis,
		deduplicationService,
		core.Logger,
		consumerConfig,
		core.Services.Observability.TraceService,
		core.Services.Observability.ScoreService,
		core.Services.Observability.MetricsService,
		core.Services.Observability.LogsService,
		core.Services.Observability.GenAIEventsService,
		core.Services.Observability.ArchiveService, // S3 raw telemetry archival (nil if disabled)
		&core.Config.Archive,                       // Archive config
	)

	// Create evaluation rule worker using config
	discoveryInterval, _ := time.ParseDuration(core.Config.Workers.RuleWorker.DiscoveryInterval)
	if discoveryInterval == 0 {
		discoveryInterval = 30 * time.Second
	}
	ruleCacheTTL, _ := time.ParseDuration(core.Config.Workers.RuleWorker.RuleCacheTTL)
	if ruleCacheTTL == 0 {
		ruleCacheTTL = 30 * time.Second
	}

	ruleWorkerConfig := &evaluationWorker.RuleWorkerConfig{
		ConsumerGroup:     "evaluation-rule-workers",
		ConsumerID:        "rule-worker-" + ulid.New().String(),
		BatchSize:         core.Config.Workers.RuleWorker.BatchSize,
		BlockDuration:     time.Duration(core.Config.Workers.RuleWorker.BlockDurationMs) * time.Millisecond,
		MaxRetries:        core.Config.Workers.RuleWorker.MaxRetries,
		RetryBackoff:      time.Duration(core.Config.Workers.RuleWorker.RetryBackoffMs) * time.Millisecond,
		DiscoveryInterval: discoveryInterval,
		MaxStreamsPerRead: core.Config.Workers.RuleWorker.MaxStreamsPerRead,
		RuleCacheTTL:      ruleCacheTTL,
	}

	ruleWorker := evaluationWorker.NewRuleWorker(
		core.Databases.Redis,
		core.Services.Evaluation.Rule,
		core.Services.Evaluation.RuleExecution,
		core.Logger,
		ruleWorkerConfig,
	)

	// Create scorers for evaluation worker
	builtinScorer := evaluationWorker.NewBuiltinScorer(core.Logger)
	regexScorer := evaluationWorker.NewRegexScorer(core.Logger)

	// LLM scorer requires credentials and execution services
	var llmScorer evaluationWorker.Scorer
	if core.Services.Credentials != nil && core.Services.Prompt != nil {
		llmScorer = evaluationWorker.NewLLMScorer(
			core.Services.Credentials.ProviderCredential,
			core.Services.Prompt.Execution,
			core.Logger,
		)
		core.Logger.Info("LLM scorer initialized for evaluation worker")
	} else {
		core.Logger.Warn("LLM scorer disabled: credentials or prompt services not available")
	}

	// Create evaluation worker
	evalWorkerConfig := &evaluationWorker.EvaluationWorkerConfig{
		ConsumerGroup:  "evaluation-execution-workers",
		ConsumerID:     "eval-worker-" + ulid.New().String(),
		BatchSize:      10,
		BlockDuration:  time.Second,
		MaxRetries:     3,
		RetryBackoff:   500 * time.Millisecond,
		MaxConcurrency: 5,
	}

	evalWorker := evaluationWorker.NewEvaluationWorker(
		core.Databases.Redis,
		core.Services.Observability.ScoreService,
		core.Services.Evaluation.RuleExecution,
		llmScorer,
		builtinScorer,
		regexScorer,
		core.Logger,
		evalWorkerConfig,
	)

	manualTriggerWorkerConfig := &evaluationWorker.ManualTriggerWorkerConfig{
		ConsumerGroup:  "manual-trigger-workers",
		ConsumerID:     "manual-trigger-" + ulid.New().String(),
		BlockDuration:  time.Second,
		MaxRetries:     3,
		RetryBackoff:   500 * time.Millisecond,
		MaxConcurrency: 3,
	}

	manualTriggerWorker := evaluationWorker.NewManualTriggerWorker(
		core.Databases.Redis,
		core.Services.Observability.TraceService,
		core.Services.Evaluation.RuleExecution,
		core.Logger,
		manualTriggerWorkerConfig,
	)

	return &WorkerContainer{
		TelemetryConsumer:   telemetryConsumer,
		RuleWorker:          ruleWorker,
		EvaluationWorker:    evalWorker,
		ManualTriggerWorker: manualTriggerWorker,
	}, nil
}

func ProvideCore(cfg *config.Config, logger *slog.Logger) (*CoreContainer, error) {
	databases, err := ProvideDatabases(cfg, logger)
	if err != nil {
		return nil, err
	}

	repos := ProvideRepositories(databases, logger)

	// Concrete → interface for dependency inversion
	txManager := database.NewTransactionManager(databases.Postgres.DB)

	return &CoreContainer{
		Config:     cfg,
		Logger:     logger,
		Databases:  databases,
		Repos:      repos,
		TxManager:  txManager, // Stored as common.TransactionManager interface
		Services:   nil,       // Populated by mode-specific provider
		Enterprise: nil,       // Populated by mode-specific provider
	}, nil
}

func ProvideServerServices(core *CoreContainer) *ServiceContainer {
	cfg := core.Config
	logger := core.Logger
	repos := core.Repos
	databases := core.Databases

	billingServices := ProvideBillingServices(repos.Billing, logger)
	analyticsServices := ProvideAnalyticsServices(repos.Analytics)
	observabilityServices := ProvideObservabilityServices(repos.Observability, repos.Storage, analyticsServices, databases.Redis, cfg, logger)
	authServices := ProvideAuthServices(cfg, repos.User, repos.Auth, databases, logger)
	userServices := ProvideUserServices(repos.User, repos.Auth, logger)
	orgService, memberService, projectService, invitationService, settingsService :=
		ProvideOrganizationServices(repos.User, repos.Auth, repos.Organization, authServices, cfg, logger)

	// Orchestrates user, org, project creation atomically
	registrationSvc := registrationService.NewRegistrationService(
		core.TxManager,
		repos.User.User,
		orgService,
		projectService,
		memberService,
		invitationService,
		authServices.Role,
		authServices.Auth,
	)

	promptServices := ProvidePromptServices(core.TxManager, repos.Prompt, analyticsServices.ProviderPricing, cfg, logger)

	// Config validation ensures AI_KEY_ENCRYPTION_KEY is valid, so credentials service is guaranteed to initialize
	credentialsServices := ProvideCredentialsServices(repos.Credentials, repos.Analytics, cfg, logger)

	playgroundServices := ProvidePlaygroundServices(
		repos.Playground,
		credentialsServices.ProviderCredential,
		promptServices.Compiler,
		promptServices.Execution,
		logger,
	)

	evaluationServices := ProvideEvaluationServices(repos.Evaluation, repos.Observability, observabilityServices, databases.Redis, logger)

	dashboardServices := ProvideDashboardServices(repos.Dashboard, logger)

	// Overview service needs projectService and credentials repo (created after other services)
	overviewSvc := analyticsService.NewOverviewService(
		repos.Analytics.Overview,
		projectService,
		repos.Credentials.ProviderCredential,
		logger,
	)
	analyticsServices.Overview = overviewSvc

	return &ServiceContainer{
		User:                userServices,
		Auth:                authServices,
		Registration:        registrationSvc,
		OrganizationService: orgService,
		MemberService:       memberService,
		ProjectService:      projectService,
		InvitationService:   invitationService,
		SettingsService:     settingsService,
		Observability:       observabilityServices,
		Billing:             billingServices,
		Analytics:           analyticsServices,
		Prompt:              promptServices,
		Credentials:         credentialsServices,
		Playground:          playgroundServices,
		Evaluation:          evaluationServices,
		Dashboard:           dashboardServices,
	}
}

func ProvideWorkerServices(core *CoreContainer) *ServiceContainer {
	cfg := core.Config
	logger := core.Logger
	repos := core.Repos
	databases := core.Databases

	billingServices := ProvideBillingServices(repos.Billing, logger)
	analyticsServices := ProvideAnalyticsServices(repos.Analytics)
	observabilityServices := ProvideObservabilityServices(repos.Observability, repos.Storage, analyticsServices, databases.Redis, cfg, logger)

	// Prompt services needed for LLM scorer
	promptServices := ProvidePromptServices(core.TxManager, repos.Prompt, analyticsServices.ProviderPricing, cfg, logger)

	// Credentials services needed for LLM scorer (optional - only if encryption key configured)
	var credentialsServices *CredentialsServices
	if cfg.Encryption.AIKeyEncryptionKey != "" {
		credentialsServices = ProvideCredentialsServices(repos.Credentials, repos.Analytics, cfg, logger)
	} else {
		logger.Warn("AI_KEY_ENCRYPTION_KEY not configured, LLM scorer will be disabled")
	}

	// Evaluation services needed for rule worker
	evaluationServices := ProvideEvaluationServices(repos.Evaluation, repos.Observability, observabilityServices, databases.Redis, logger)

	return &ServiceContainer{
		User:                nil, // Worker doesn't need auth/user/org services
		Auth:                nil,
		Registration:        nil,
		OrganizationService: nil,
		MemberService:       nil,
		ProjectService:      nil,
		InvitationService:   nil,
		SettingsService:     nil,
		Prompt:              promptServices,    // Needed for LLM scorer
		Credentials:         credentialsServices, // Needed for LLM scorer
		Playground:          nil,
		Observability:       observabilityServices,
		Billing:             billingServices,
		Analytics:           analyticsServices,
		Evaluation:          evaluationServices, // Needed for rule worker
	}
}

func ProvideServer(core *CoreContainer) (*ServerContainer, error) {
	// Get credentials services (may be nil if encryption key not configured)
	var credentialsSvc credentialsDomain.ProviderCredentialService
	var modelCatalogSvc credentialsService.ModelCatalogService
	if core.Services.Credentials != nil {
		credentialsSvc = core.Services.Credentials.ProviderCredential
		modelCatalogSvc = core.Services.Credentials.ModelCatalog
	}

	// Get playground service
	var playgroundSvc playgroundDomain.PlaygroundService
	if core.Services.Playground != nil {
		playgroundSvc = core.Services.Playground.Playground
	}

	// Get evaluation services
	var scoreConfigSvc evaluationDomain.ScoreConfigService
	var datasetSvc evaluationDomain.DatasetService
	var datasetItemSvc evaluationDomain.DatasetItemService
	var experimentSvc evaluationDomain.ExperimentService
	var experimentItemSvc evaluationDomain.ExperimentItemService
	var ruleSvc evaluationDomain.RuleService
	var ruleExecutionSvc evaluationDomain.RuleExecutionService
	if core.Services.Evaluation != nil {
		scoreConfigSvc = core.Services.Evaluation.ScoreConfig
		datasetSvc = core.Services.Evaluation.Dataset
		datasetItemSvc = core.Services.Evaluation.DatasetItem
		experimentSvc = core.Services.Evaluation.Experiment
		experimentItemSvc = core.Services.Evaluation.ExperimentItem
		ruleSvc = core.Services.Evaluation.Rule
		ruleExecutionSvc = core.Services.Evaluation.RuleExecution
	}

	// Get dashboard services
	var dashboardSvc dashboardDomain.DashboardService
	var widgetQuerySvc dashboardDomain.WidgetQueryService
	var templateSvc dashboardDomain.TemplateService
	if core.Services.Dashboard != nil {
		dashboardSvc = core.Services.Dashboard.Dashboard
		widgetQuerySvc = core.Services.Dashboard.WidgetQuery
		templateSvc = core.Services.Dashboard.Template
	}

	httpHandlers := handlers.NewHandlers(
		core.Config,
		core.Logger,
		core.Services.Auth.Auth,
		core.Services.Auth.APIKey,
		core.Services.Auth.BlacklistedTokens,
		core.Services.Registration,       // Registration service for signup
		core.Services.Auth.OAuthProvider, // OAuth provider for Google/GitHub signup
		core.Services.User.User,
		core.Services.User.Profile,
		core.Services.OrganizationService,
		core.Services.MemberService,
		core.Services.ProjectService,
		core.Services.InvitationService,
		core.Services.SettingsService,
		core.Services.Auth.Role,
		core.Services.Auth.Permission,
		core.Services.Auth.OrganizationMembers,
		core.Services.Auth.Scope,
		core.Services.Observability,
		core.Services.Prompt.Prompt,
		core.Services.Prompt.Compiler,
		credentialsSvc,
		modelCatalogSvc,
		playgroundSvc,
		scoreConfigSvc,
		datasetSvc,
		datasetItemSvc,
		experimentSvc,
		experimentItemSvc,
		ruleSvc,
		ruleExecutionSvc,
		dashboardSvc,
		widgetQuerySvc,
		templateSvc,
		core.Services.Analytics.Overview,
	)

	httpServer := http.NewServer(
		core.Config,
		core.Logger,
		httpHandlers,
		core.Services.Auth.JWT,
		core.Services.Auth.BlacklistedTokens,
		core.Services.Auth.OrganizationMembers,
		core.Services.Auth.APIKey,
		core.Databases.Redis.Client,
	)

	slogLogger := core.Logger

	grpcOTLPHandler := grpcTransport.NewOTLPHandler(
		core.Services.Observability.StreamProducer,
		core.Services.Observability.DeduplicationService,
		core.Services.Observability.OTLPConverterService,
		slogLogger,
	)

	grpcOTLPMetricsHandler := grpcTransport.NewOTLPMetricsHandler(
		core.Services.Observability.StreamProducer,
		core.Services.Observability.OTLPMetricsConverterService,
		slogLogger,
	)

	grpcOTLPLogsHandler := grpcTransport.NewOTLPLogsHandler(
		core.Services.Observability.StreamProducer,
		core.Services.Observability.OTLPLogsConverterService,
		core.Services.Observability.OTLPEventsConverterService,
		slogLogger,
	)

	grpcAuthInterceptor := grpcTransport.NewAuthInterceptor(
		core.Services.Auth.APIKey,
		slogLogger,
	)

	grpcServer, err := grpcTransport.NewServer(
		core.Config.GRPC.Port,
		grpcOTLPHandler,
		grpcOTLPMetricsHandler,
		grpcOTLPLogsHandler,
		grpcAuthInterceptor,
		slogLogger,
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

func ProvideUserRepositories(db *gorm.DB) *UserRepositories {
	return &UserRepositories{
		User: userRepo.NewUserRepository(db),
	}
}

func ProvideAuthRepositories(db *gorm.DB) *AuthRepositories {
	return &AuthRepositories{
		UserSession:        authRepo.NewUserSessionRepository(db),
		BlacklistedToken:   authRepo.NewBlacklistedTokenRepository(db),
		PasswordResetToken: authRepo.NewPasswordResetTokenRepository(db),
		APIKey:             authRepo.NewAPIKeyRepository(db),
		Role:               authRepo.NewRoleRepository(db),
		OrganizationMember: authRepo.NewOrganizationMemberRepository(db),
		Permission:         authRepo.NewPermissionRepository(db),
		RolePermission:     authRepo.NewRolePermissionRepository(db),
		AuditLog:           authRepo.NewAuditLogRepository(db),
	}
}

func ProvideOrganizationRepositories(db *gorm.DB) *OrganizationRepositories {
	return &OrganizationRepositories{
		Organization: orgRepo.NewOrganizationRepository(db),
		Member:       orgRepo.NewMemberRepository(db),
		Project:      orgRepo.NewProjectRepository(db),
		Invitation:   orgRepo.NewInvitationRepository(db),
		Settings:     orgRepo.NewOrganizationSettingsRepository(db),
	}
}

func ProvideObservabilityRepositories(clickhouseDB *database.ClickHouseDB, postgresDB *gorm.DB, redisDB *database.RedisDB) *ObservabilityRepositories {
	return &ObservabilityRepositories{
		Trace:                  observabilityRepo.NewTraceRepository(clickhouseDB.Conn),
		Score:                  observabilityRepo.NewScoreRepository(clickhouseDB.Conn),
		ScoreAnalytics:         observabilityRepo.NewScoreAnalyticsRepository(clickhouseDB.Conn),
		Metrics:                observabilityRepo.NewMetricsRepository(clickhouseDB.Conn),
		Logs:                   observabilityRepo.NewLogsRepository(clickhouseDB.Conn),
		GenAIEvents:            observabilityRepo.NewGenAIEventsRepository(clickhouseDB.Conn),
		TelemetryDeduplication: observabilityRepo.NewTelemetryDeduplicationRepository(redisDB),
		FilterPreset:           observabilityRepo.NewFilterPresetRepository(postgresDB),
	}
}

func ProvideStorageRepositories(clickhouseDB *database.ClickHouseDB) *StorageRepositories {
	return &StorageRepositories{
		BlobStorage: storageRepo.NewBlobStorageRepository(clickhouseDB.Conn),
	}
}

func ProvideBillingRepositories(db *gorm.DB, logger *slog.Logger) *BillingRepositories {
	return &BillingRepositories{
		Usage:         billingRepo.NewUsageRepository(db, logger),
		BillingRecord: billingRepo.NewBillingRecordRepository(db, logger),
		Quota:         billingRepo.NewQuotaRepository(db, logger),
	}
}

func ProvideAnalyticsRepositories(db *gorm.DB, clickhouseDB *database.ClickHouseDB) *AnalyticsRepositories {
	return &AnalyticsRepositories{
		ProviderModel: analyticsRepo.NewProviderModelRepository(db),
		Overview:      analyticsRepo.NewOverviewRepository(clickhouseDB.Conn),
	}
}

func ProvidePromptRepositories(db *gorm.DB, redisDB *database.RedisDB) *PromptRepositories {
	return &PromptRepositories{
		Prompt:         promptRepo.NewPromptRepository(db),
		Version:        promptRepo.NewVersionRepository(db),
		Label:          promptRepo.NewLabelRepository(db),
		ProtectedLabel: promptRepo.NewProtectedLabelRepository(db),
		Cache:          promptRepo.NewCacheRepository(redisDB),
	}
}

func ProvideCredentialsRepositories(db *gorm.DB) *CredentialsRepositories {
	return &CredentialsRepositories{
		ProviderCredential: credentialsRepo.NewProviderCredentialRepository(db),
	}
}

func ProvidePlaygroundRepositories(db *gorm.DB) *PlaygroundRepositories {
	return &PlaygroundRepositories{
		Session: playgroundRepo.NewSessionRepository(db),
	}
}

func ProvideEvaluationRepositories(db *gorm.DB) *EvaluationRepositories {
	return &EvaluationRepositories{
		ScoreConfig:    evaluationRepo.NewScoreConfigRepository(db),
		Dataset:        evaluationRepo.NewDatasetRepository(db),
		DatasetItem:    evaluationRepo.NewDatasetItemRepository(db),
		Experiment:     evaluationRepo.NewExperimentRepository(db),
		ExperimentItem: evaluationRepo.NewExperimentItemRepository(db),
		Rule:           evaluationRepo.NewRuleRepository(db),
		RuleExecution:  evaluationRepo.NewRuleExecutionRepository(db),
	}
}

func ProvideDashboardRepositories(db *gorm.DB, clickhouseDB *database.ClickHouseDB) *DashboardRepositories {
	return &DashboardRepositories{
		Dashboard:   dashboardRepo.NewDashboardRepository(db),
		WidgetQuery: dashboardRepo.NewWidgetQueryRepository(clickhouseDB.Conn),
		Template:    dashboardRepo.NewTemplateRepository(db),
	}
}

func ProvideRepositories(dbs *DatabaseContainer, logger *slog.Logger) *RepositoryContainer {
	return &RepositoryContainer{
		User:          ProvideUserRepositories(dbs.Postgres.DB),
		Auth:          ProvideAuthRepositories(dbs.Postgres.DB),
		Organization:  ProvideOrganizationRepositories(dbs.Postgres.DB),
		Observability: ProvideObservabilityRepositories(dbs.ClickHouse, dbs.Postgres.DB, dbs.Redis),
		Storage:       ProvideStorageRepositories(dbs.ClickHouse),
		Billing:       ProvideBillingRepositories(dbs.Postgres.DB, logger),
		Analytics:     ProvideAnalyticsRepositories(dbs.Postgres.DB, dbs.ClickHouse),
		Prompt:        ProvidePromptRepositories(dbs.Postgres.DB, dbs.Redis),
		Credentials:   ProvideCredentialsRepositories(dbs.Postgres.DB),
		Playground:    ProvidePlaygroundRepositories(dbs.Postgres.DB),
		Evaluation:    ProvideEvaluationRepositories(dbs.Postgres.DB),
		Dashboard:     ProvideDashboardRepositories(dbs.Postgres.DB, dbs.ClickHouse),
	}
}

func ProvideUserServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	logger *slog.Logger,
) *UserServices {
	userSvc := userService.NewUserService(
		userRepos.User,
		nil,
		authRepos.OrganizationMember,
	)

	profileSvc := userService.NewProfileService(
		userRepos.User,
	)

	return &UserServices{
		User:    userSvc,
		Profile: profileSvc,
	}
}

func ProvideAuthServices(
	cfg *config.Config,
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	databases *DatabaseContainer,
	logger *slog.Logger,
) *AuthServices {
	jwtService, err := authService.NewJWTService(&cfg.Auth)
	if err != nil {
		logger.Error("Failed to create JWT service", "error", err)
		os.Exit(1)
	}

	permissionService := authService.NewPermissionService(
		authRepos.Permission,
		authRepos.RolePermission,
	)

	roleService := authService.NewRoleService(
		authRepos.Role,
		authRepos.RolePermission,
	)

	orgMemberService := authService.NewOrganizationMemberService(
		authRepos.OrganizationMember,
		authRepos.Role,
	)

	blacklistedTokenService := authService.NewBlacklistedTokenService(
		authRepos.BlacklistedToken,
	)

	sessionService := authService.NewSessionService(
		&cfg.Auth,
		authRepos.UserSession,
		userRepos.User,
		jwtService,
	)

	apiKeyService := authService.NewAPIKeyService(
		authRepos.APIKey,
		authRepos.OrganizationMember,
	)

	coreAuthSvc := authService.NewAuthService(
		&cfg.Auth,
		userRepos.User,
		authRepos.UserSession,
		jwtService,
		roleService,
		authRepos.PasswordResetToken,
		blacklistedTokenService,
		databases.Redis.Client,
	)

	// Audit decorator for clean separation of concerns
	authSvc := authService.NewAuditDecorator(coreAuthSvc, authRepos.AuditLog, logger)

	scopeService := authService.NewScopeService(
		authRepos.OrganizationMember,
		authRepos.Role,
		authRepos.Permission,
	)

	frontendURL := "http://localhost:3000"
	if url := os.Getenv("NEXT_PUBLIC_APP_URL"); url != "" {
		frontendURL = url
	}
	oauthProvider := authService.NewOAuthProviderService(
		&cfg.Auth,
		databases.Redis.Client,
		frontendURL,
	)

	return &AuthServices{
		Auth:                authSvc,
		JWT:                 jwtService,
		Sessions:            sessionService,
		APIKey:              apiKeyService,
		Role:                roleService,
		Permission:          permissionService,
		OrganizationMembers: orgMemberService,
		BlacklistedTokens:   blacklistedTokenService,
		Scope:               scopeService,
		OAuthProvider:       oauthProvider,
	}
}

func ProvideOrganizationServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	orgRepos *OrganizationRepositories,
	authServices *AuthServices,
	cfg *config.Config,
	logger *slog.Logger,
) (
	organization.OrganizationService,
	organization.MemberService,
	organization.ProjectService,
	organization.InvitationService,
	organization.OrganizationSettingsService,
) {
	memberSvc := orgService.NewMemberService(
		orgRepos.Member,
		orgRepos.Organization,
		userRepos.User,
		authServices.Role,
	)

	projectSvc := orgService.NewProjectService(
		orgRepos.Project,
		orgRepos.Organization,
		orgRepos.Member,
	)

	// Create email sender based on configuration
	emailSender, err := createEmailSender(&cfg.External.Email, logger)
	if err != nil {
		logger.Error("failed to create email sender", "error", err)
		os.Exit(1)
	}

	invitationSvc := orgService.NewInvitationService(
		orgRepos.Invitation,
		orgRepos.Organization,
		orgRepos.Member,
		userRepos.User,
		authServices.Role,
		emailSender,
		orgService.InvitationServiceConfig{
			AppURL: cfg.Server.AppURL,
		},
		logger.With("service", "invitation"),
	)

	orgSvc := orgService.NewOrganizationService(
		orgRepos.Organization,
		userRepos.User,
		memberSvc,
		projectSvc,
		authServices.Role,
	)

	settingsSvc := orgService.NewOrganizationSettingsService(
		orgRepos.Settings,
		orgRepos.Member,
	)

	return orgSvc, memberSvc, projectSvc, invitationSvc, settingsSvc
}

// TelemetryService created without analytics worker - inject via SetAnalyticsWorker() after startup.
func ProvideObservabilityServices(
	observabilityRepos *ObservabilityRepositories,
	storageRepos *StorageRepositories,
	analyticsServices *AnalyticsServices,
	redisDB *database.RedisDB,
	cfg *config.Config,
	logger *slog.Logger,
) *observabilityService.ServiceRegistry {
	deduplicationService := observabilityService.NewTelemetryDeduplicationService(observabilityRepos.TelemetryDeduplication)
	streamProducer := streams.NewTelemetryStreamProducer(redisDB, logger)
	telemetryService := observabilityService.NewTelemetryService(
		deduplicationService,
		streamProducer,
		logger,
	)

	var s3Client *storage.S3Client
	if cfg.BlobStorage.Provider != "" && cfg.BlobStorage.BucketName != "" {
		var err error
		s3Client, err = storage.NewS3Client(&cfg.BlobStorage, logger)
		if err != nil {
			logger.Warn("Failed to initialize S3 client, blob storage will be disabled", "error", err)
		}
	}

	blobStorageSvc := storageService.NewBlobStorageService(
		storageRepos.BlobStorage,
		s3Client,
		&cfg.BlobStorage,
		logger,
	)

	return observabilityService.NewServiceRegistry(
		observabilityRepos.Trace,
		observabilityRepos.Score,
		observabilityRepos.ScoreAnalytics,
		observabilityRepos.Metrics,
		observabilityRepos.Logs,
		observabilityRepos.GenAIEvents,
		observabilityRepos.FilterPreset,
		blobStorageSvc,
		s3Client,
		&cfg.Archive, // Archive config for S3 raw telemetry archival
		streamProducer,
		deduplicationService,
		telemetryService,
		analyticsServices.ProviderPricing,
		&cfg.Observability,
		logger,
	)
}

func ProvideBillingServices(
	billingRepos *BillingRepositories,
	logger *slog.Logger,
) *BillingServices {
	orgService := &simpleBillingOrgService{logger: logger}
	billingServiceImpl := billingService.NewBillingService(
		logger,
		nil,                        // config - use defaults
		billingRepos.Usage,         // UsageRepository
		billingRepos.BillingRecord, // BillingRecordRepository
		billingRepos.Quota,         // QuotaRepository
		orgService,
	)

	return &BillingServices{
		Billing: billingServiceImpl,
	}
}

func ProvideAnalyticsServices(
	analyticsRepos *AnalyticsRepositories,
) *AnalyticsServices {
	providerPricingServiceImpl := analyticsService.NewProviderPricingService(analyticsRepos.ProviderModel)

	return &AnalyticsServices{
		ProviderPricing: providerPricingServiceImpl,
	}
}

func ProvidePromptServices(
	txManager common.TransactionManager,
	promptRepos *PromptRepositories,
	pricingService analytics.ProviderPricingService,
	cfg *config.Config,
	logger *slog.Logger,
) *PromptServices {
	compilerSvc := promptService.NewCompilerService()
	aiClientConfig := &promptService.AIClientConfig{
		DefaultTimeout: cfg.External.LLMTimeout,
	}

	executionSvc := promptService.NewExecutionService(compilerSvc, pricingService, aiClientConfig)
	promptSvc := promptService.NewPromptService(
		txManager,
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

func ProvideCredentialsServices(
	credentialsRepos *CredentialsRepositories,
	analyticsRepos *AnalyticsRepositories,
	cfg *config.Config,
	logger *slog.Logger,
) *CredentialsServices {
	encryptor, err := encryption.NewServiceFromBase64(cfg.Encryption.AIKeyEncryptionKey)
	if err != nil {
		panic(fmt.Sprintf("encryption initialization failed after config validation: %v (this is a bug)", err))
	}

	providerSvc := credentialsService.NewProviderCredentialService(
		credentialsRepos.ProviderCredential,
		encryptor,
		logger,
	)

	// Model catalog combines default models from DB with custom models from credentials
	modelCatalogSvc := credentialsService.NewModelCatalogService(
		credentialsRepos.ProviderCredential,
		analyticsRepos.ProviderModel,
		logger,
	)

	return &CredentialsServices{
		ProviderCredential: providerSvc,
		ModelCatalog:       modelCatalogSvc,
	}
}

func ProvidePlaygroundServices(
	playgroundRepos *PlaygroundRepositories,
	credentialsService credentialsDomain.ProviderCredentialService,
	compilerService promptDomain.CompilerService,
	executionService promptDomain.ExecutionService,
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

func ProvideEvaluationServices(
	evaluationRepos *EvaluationRepositories,
	observabilityRepos *ObservabilityRepositories,
	observabilityServices *observabilityService.ServiceRegistry,
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

	// RuleExecutionService must be created before RuleService since RuleService depends on it
	ruleExecutionSvc := evaluationService.NewRuleExecutionService(
		evaluationRepos.RuleExecution,
		logger,
	)

	ruleSvc := evaluationService.NewRuleService(
		evaluationRepos.Rule,
		ruleExecutionSvc,
		redisDB,
		logger,
	)

	return &EvaluationServices{
		ScoreConfig:    scoreConfigSvc,
		Dataset:        datasetSvc,
		DatasetItem:    datasetItemSvc,
		Experiment:     experimentSvc,
		ExperimentItem: experimentItemSvc,
		Rule:           ruleSvc,
		RuleExecution:  ruleExecutionSvc,
	}
}

func ProvideDashboardServices(
	dashboardRepos *DashboardRepositories,
	logger *slog.Logger,
) *DashboardServices {
	dashboardSvc := dashboardService.NewDashboardService(
		dashboardRepos.Dashboard,
		logger,
	)

	widgetQuerySvc := dashboardService.NewWidgetQueryService(
		dashboardRepos.Dashboard,
		dashboardRepos.WidgetQuery,
		logger,
	)

	templateSvc := dashboardService.NewTemplateService(
		dashboardRepos.Template,
		dashboardRepos.Dashboard,
		logger,
	)

	return &DashboardServices{
		Dashboard:   dashboardSvc,
		WidgetQuery: widgetQuerySvc,
		Template:    templateSvc,
	}
}

type simpleBillingOrgService struct {
	logger *slog.Logger
}

func (s *simpleBillingOrgService) GetBillingTier(ctx context.Context, orgID ulid.ULID) (string, error) {
	// Default to free tier for now - in production this would query the org service
	return "free", nil
}

func (s *simpleBillingOrgService) GetDiscountRate(ctx context.Context, orgID ulid.ULID) (float64, error) {
	// Default to no discount - in production this would query the org service
	return 0.0, nil
}

func (s *simpleBillingOrgService) GetPaymentMethod(ctx context.Context, orgID ulid.ULID) (*billing.PaymentMethod, error) {
	// No payment method by default - in production this would query the org service
	return nil, nil
}

func ProvideEnterpriseServices(cfg *config.Config, logger *slog.Logger) *EnterpriseContainer {
	return &EnterpriseContainer{
		SSO:        sso.New(),         // Uses stub or real based on build tags
		RBAC:       rbac.New(),        // Uses stub or real based on build tags
		Compliance: compliance.New(),  // Uses stub or real based on build tags
		Analytics:  eeAnalytics.New(), // Uses stub or real based on build tags
	}
}

func (pc *ProviderContainer) HealthCheck() map[string]string {
	health := make(map[string]string)

	if pc.Core != nil && pc.Core.Databases != nil {
		if pc.Core.Databases.Postgres != nil {
			if err := pc.Core.Databases.Postgres.Health(); err != nil {
				health["postgres"] = "unhealthy: " + err.Error()
			} else {
				health["postgres"] = "healthy"
			}
		}

		if pc.Core.Databases.Redis != nil {
			if err := pc.Core.Databases.Redis.Health(); err != nil {
				health["redis"] = "unhealthy: " + err.Error()
			} else {
				health["redis"] = "healthy"
			}
		}

		if pc.Core.Databases.ClickHouse != nil {
			if err := pc.Core.Databases.ClickHouse.Health(); err != nil {
				health["clickhouse"] = "unhealthy: " + err.Error()
			} else {
				health["clickhouse"] = "healthy"
			}
		}
	}

	if pc.Workers != nil && pc.Workers.TelemetryConsumer != nil {
		stats := pc.Workers.TelemetryConsumer.GetStats()
		// Consider healthy if: running (has stats) and error rate < 10% of processed batches
		batchesProcessed := stats["batches_processed"]
		errorsCount := stats["errors_count"]

		if batchesProcessed == 0 && errorsCount == 0 {
			// Newly started - healthy
			health["telemetry_stream_consumer"] = "healthy (no activity yet)"
		} else if batchesProcessed > 0 {
			errorRate := float64(errorsCount) / float64(batchesProcessed)
			if errorRate < 0.10 { // Less than 10% error rate
				health["telemetry_stream_consumer"] = fmt.Sprintf("healthy (processed: %d, errors: %d, streams: %d)",
					batchesProcessed, errorsCount, stats["active_streams"])
			} else {
				health["telemetry_stream_consumer"] = fmt.Sprintf("degraded (high error rate: %.1f%%)", errorRate*100)
			}
		} else {
			// Errors but no successful processing
			health["telemetry_stream_consumer"] = fmt.Sprintf("unhealthy (errors: %d, no successful processing)", errorsCount)
		}
	}

	// Evaluation rule worker health
	if pc.Workers != nil && pc.Workers.RuleWorker != nil {
		stats := pc.Workers.RuleWorker.GetStats()
		spansProcessed := stats["spans_processed"]
		errorsCount := stats["errors_count"]

		if spansProcessed == 0 && errorsCount == 0 {
			health["evaluation_rule_worker"] = "healthy (no activity yet)"
		} else if spansProcessed > 0 {
			health["evaluation_rule_worker"] = fmt.Sprintf("healthy (spans_processed: %d, jobs_emitted: %d, errors: %d)",
				spansProcessed, stats["jobs_emitted"], errorsCount)
		} else {
			health["evaluation_rule_worker"] = fmt.Sprintf("unhealthy (errors: %d)", errorsCount)
		}
	}

	// Evaluation worker health
	if pc.Workers != nil && pc.Workers.EvaluationWorker != nil {
		stats := pc.Workers.EvaluationWorker.GetStats()
		jobsProcessed := stats["jobs_processed"]
		errorsCount := stats["errors_count"]

		if jobsProcessed == 0 && errorsCount == 0 {
			health["evaluation_worker"] = "healthy (no activity yet)"
		} else if jobsProcessed > 0 {
			health["evaluation_worker"] = fmt.Sprintf("healthy (processed: %d, scores: %d, llm: %d, builtin: %d, regex: %d)",
				jobsProcessed, stats["scores_created"], stats["llm_calls"], stats["builtin_calls"], stats["regex_calls"])
		} else {
			health["evaluation_worker"] = fmt.Sprintf("unhealthy (errors: %d)", errorsCount)
		}
	}

	health["mode"] = string(pc.Mode)

	return health
}

func (pc *ProviderContainer) Shutdown() error {
	var lastErr error
	logger := pc.Core.Logger

	if pc.Workers != nil {
		if pc.Workers.TelemetryConsumer != nil {
			logger.Info("Stopping telemetry stream consumer...")
			pc.Workers.TelemetryConsumer.Stop()
			logger.Info("Telemetry stream consumer stopped")
		}

		if pc.Workers.RuleWorker != nil {
			logger.Info("Stopping evaluation rule worker...")
			pc.Workers.RuleWorker.Stop()
			logger.Info("Evaluation rule worker stopped")
		}

		if pc.Workers.EvaluationWorker != nil {
			logger.Info("Stopping evaluation worker...")
			pc.Workers.EvaluationWorker.Stop()
			logger.Info("Evaluation worker stopped")
		}
	}

	if pc.Core != nil && pc.Core.Databases != nil {
		if pc.Core.Databases.Postgres != nil {
			if err := pc.Core.Databases.Postgres.Close(); err != nil {
				logger.Error("Failed to close PostgreSQL connection", "error", err)
				lastErr = err
			}
		}

		if pc.Core.Databases.Redis != nil {
			if err := pc.Core.Databases.Redis.Close(); err != nil {
				logger.Error("Failed to close Redis connection", "error", err)
				lastErr = err
			}
		}

		if pc.Core.Databases.ClickHouse != nil {
			if err := pc.Core.Databases.ClickHouse.Close(); err != nil {
				logger.Error("Failed to close ClickHouse connection", "error", err)
				lastErr = err
			}
		}
	}

	return lastErr
}

// createEmailSender creates an email sender based on the configured provider.
// Returns NoOpEmailSender if no provider is configured (email disabled).
func createEmailSender(cfg *config.EmailConfig, logger *slog.Logger) (email.EmailSender, error) {
	if cfg.Provider == "" {
		logger.Warn("email sender not configured, invitations will not be sent via email")
		return &email.NoOpEmailSender{}, nil
	}

	logger.Info("initializing email sender", "provider", cfg.Provider)

	switch cfg.Provider {
	case "resend":
		return email.NewResendClient(email.ResendConfig{
			APIKey:    cfg.ResendAPIKey,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
		}), nil

	case "smtp":
		return email.NewSMTPClient(email.SMTPConfig{
			Host:      cfg.SMTPHost,
			Port:      cfg.SMTPPort,
			Username:  cfg.SMTPUsername,
			Password:  cfg.SMTPPassword,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
			UseTLS:    cfg.SMTPUseTLS,
		}), nil

	case "ses":
		client, err := email.NewSESClient(email.SESConfig{
			Region:    cfg.SESRegion,
			AccessKey: cfg.SESAccessKey,
			SecretKey: cfg.SESSecretKey,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create SES client: %w", err)
		}
		return client, nil

	case "sendgrid":
		return email.NewSendGridClient(email.SendGridConfig{
			APIKey:    cfg.SendGridAPIKey,
			FromEmail: cfg.FromEmail,
			FromName:  cfg.FromName,
		}), nil

	default:
		return nil, fmt.Errorf("unknown email provider: %s", cfg.Provider)
	}
}
