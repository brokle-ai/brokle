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
	"brokle/pkg/encryption"
	"brokle/pkg/ulid"
)

// DeploymentMode tracks which mode the app is running in
type DeploymentMode string

const (
	ModeServer DeploymentMode = "server"
	ModeWorker DeploymentMode = "worker"
)

// CoreContainer holds shared infrastructure (databases, repos, services)
type CoreContainer struct {
	Config     *config.Config
	Logger     *slog.Logger
	Databases  *DatabaseContainer
	Repos      *RepositoryContainer
	TxManager  common.TransactionManager
	Services   *ServiceContainer
	Enterprise *EnterpriseContainer
}

// ServerContainer holds HTTP and gRPC server components
type ServerContainer struct {
	HTTPServer *http.Server
	GRPCServer *grpcTransport.Server
}

// ProviderContainer holds all provider instances for dependency injection
type ProviderContainer struct {
	Core    *CoreContainer
	Server  *ServerContainer // nil in worker mode
	Workers *WorkerContainer // nil in server mode
	Mode    DeploymentMode
}

// DatabaseContainer holds all database connections
type DatabaseContainer struct {
	Postgres   *database.PostgresDB
	Redis      *database.RedisDB
	ClickHouse *database.ClickHouseDB
}

// WorkerContainer holds all background worker instances
type WorkerContainer struct {
	TelemetryConsumer *workers.TelemetryStreamConsumer
}

// RepositoryContainer holds all repository instances organized by domain
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
}

// ServiceContainer holds all service instances organized by domain
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
}

// EnterpriseContainer holds all enterprise service instances
type EnterpriseContainer struct {
	SSO        sso.SSOProvider
	RBAC       rbac.RBACManager
	Compliance compliance.Compliance
	Analytics  eeAnalytics.EnterpriseAnalytics
}

// Domain-specific repository containers

// UserRepositories contains all user-related repositories
type UserRepositories struct {
	User user.Repository
}

// AuthRepositories contains all auth-related repositories
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

// OrganizationRepositories contains all organization-related repositories
type OrganizationRepositories struct {
	Organization organization.OrganizationRepository
	Member       organization.MemberRepository
	Project      organization.ProjectRepository
	Invitation   organization.InvitationRepository
	Settings     organization.OrganizationSettingsRepository
}

// ObservabilityRepositories contains all observability-related repositories
type ObservabilityRepositories struct {
	Trace                  observability.TraceRepository
	Score                  observability.ScoreRepository
	Metrics                observability.MetricsRepository
	Logs                   observability.LogsRepository
	GenAIEvents            observability.GenAIEventsRepository
	TelemetryDeduplication observability.TelemetryDeduplicationRepository
}

// StorageRepositories contains all storage-related repositories
type StorageRepositories struct {
	BlobStorage storageDomain.BlobStorageRepository
}

// BillingRepositories contains all billing-related repositories
type BillingRepositories struct {
	Usage         billing.UsageRepository
	BillingRecord billing.BillingRecordRepository
	Quota         billing.QuotaRepository
}

// AnalyticsRepositories contains all analytics-related repositories
type AnalyticsRepositories struct {
	ProviderModel analytics.ProviderModelRepository
}

// PromptRepositories contains all prompt-related repositories
type PromptRepositories struct {
	Prompt         promptDomain.PromptRepository
	Version        promptDomain.VersionRepository
	Label          promptDomain.LabelRepository
	ProtectedLabel promptDomain.ProtectedLabelRepository
	Cache          promptDomain.CacheRepository
}

// CredentialsRepositories contains all credentials-related repositories
type CredentialsRepositories struct {
	ProviderCredential credentialsDomain.ProviderCredentialRepository
}

// PlaygroundRepositories contains all playground-related repositories
type PlaygroundRepositories struct {
	Session playgroundDomain.SessionRepository
}

// EvaluationRepositories contains all evaluation-related repositories
type EvaluationRepositories struct {
	ScoreConfig evaluationDomain.ScoreConfigRepository
}

// Domain-specific service containers

// UserServices contains all user-related services
type UserServices struct {
	User    user.UserService
	Profile user.ProfileService
}

// AuthServices contains all auth-related services
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

// BillingServices contains all billing-related services
type BillingServices struct {
	Billing *billingService.BillingService
}

// AnalyticsServices contains all analytics-related services
type AnalyticsServices struct {
	ProviderPricing analytics.ProviderPricingService
}

// PromptServices contains all prompt-related services
type PromptServices struct {
	Prompt    promptDomain.PromptService
	Compiler  promptDomain.CompilerService
	Execution promptDomain.ExecutionService
}

// CredentialsServices contains all credentials-related services
type CredentialsServices struct {
	ProviderCredential credentialsDomain.ProviderCredentialService
	ModelCatalog       credentialsService.ModelCatalogService
}

// PlaygroundServices contains all playground-related services
type PlaygroundServices struct {
	Playground playgroundDomain.PlaygroundService
}

// EvaluationServices contains all evaluation-related services
type EvaluationServices struct {
	ScoreConfig evaluationDomain.ScoreConfigService
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

// ProvideWorkers creates workers using shared core infrastructure
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

	return &WorkerContainer{
		TelemetryConsumer: telemetryConsumer,
	}, nil
}

// ProvideCore creates shared infrastructure (used by both server and worker)
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

// ProvideServerServices creates ALL services for server mode
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
		ProvideOrganizationServices(repos.User, repos.Auth, repos.Organization, authServices, logger)

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
	credentialsServices, err := ProvideCredentialsServices(repos.Credentials, repos.Analytics, cfg, logger)
	if err != nil {
		logger.Error("failed to initialize credentials services", "error", err)
		// Playground will fail without credentials - no env fallback
		credentialsServices = nil
	}

	// Extract credentials service (may be nil if credentials initialization failed)
	var credSvc credentialsDomain.ProviderCredentialService
	if credentialsServices != nil {
		credSvc = credentialsServices.ProviderCredential
	}

	playgroundServices := ProvidePlaygroundServices(
		repos.Playground,
		credSvc,
		promptServices.Compiler,
		promptServices.Execution,
		logger,
	)

	evaluationServices := ProvideEvaluationServices(repos.Evaluation, logger)

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
	}
}

// ProvideWorkerServices creates ONLY services worker needs (no auth)
func ProvideWorkerServices(core *CoreContainer) *ServiceContainer {
	cfg := core.Config
	logger := core.Logger
	repos := core.Repos
	databases := core.Databases

	billingServices := ProvideBillingServices(repos.Billing, logger)
	analyticsServices := ProvideAnalyticsServices(repos.Analytics)
	observabilityServices := ProvideObservabilityServices(repos.Observability, repos.Storage, analyticsServices, databases.Redis, cfg, logger)

	return &ServiceContainer{
		User:                nil, // Worker doesn't need auth/user/org/prompt/credentials/playground services
		Auth:                nil,
		Registration:        nil,
		OrganizationService: nil,
		MemberService:       nil,
		ProjectService:      nil,
		InvitationService:   nil,
		SettingsService:     nil,
		Prompt:              nil,
		Credentials:         nil,
		Playground:          nil,
		Observability:       observabilityServices,
		Billing:             billingServices,
		Analytics:           analyticsServices,
	}
}

// ProvideServer creates HTTP server using shared core
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

	// Get evaluation service
	var scoreConfigSvc evaluationDomain.ScoreConfigService
	if core.Services.Evaluation != nil {
		scoreConfigSvc = core.Services.Evaluation.ScoreConfig
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

// ProvideUserRepositories creates all user-related repositories
func ProvideUserRepositories(db *gorm.DB) *UserRepositories {
	return &UserRepositories{
		User: userRepo.NewUserRepository(db),
	}
}

// ProvideAuthRepositories creates all auth-related repositories
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

// ProvideOrganizationRepositories creates all organization-related repositories
func ProvideOrganizationRepositories(db *gorm.DB) *OrganizationRepositories {
	return &OrganizationRepositories{
		Organization: orgRepo.NewOrganizationRepository(db),
		Member:       orgRepo.NewMemberRepository(db),
		Project:      orgRepo.NewProjectRepository(db),
		Invitation:   orgRepo.NewInvitationRepository(db),
		Settings:     orgRepo.NewOrganizationSettingsRepository(db),
	}
}

// ProvideObservabilityRepositories creates all observability-related repositories
func ProvideObservabilityRepositories(clickhouseDB *database.ClickHouseDB, postgresDB *gorm.DB, redisDB *database.RedisDB) *ObservabilityRepositories {
	return &ObservabilityRepositories{
		Trace:                  observabilityRepo.NewTraceRepository(clickhouseDB.Conn),
		Score:                  observabilityRepo.NewScoreRepository(clickhouseDB.Conn),
		Metrics:                observabilityRepo.NewMetricsRepository(clickhouseDB.Conn),
		Logs:                   observabilityRepo.NewLogsRepository(clickhouseDB.Conn),
		GenAIEvents:            observabilityRepo.NewGenAIEventsRepository(clickhouseDB.Conn),
		TelemetryDeduplication: observabilityRepo.NewTelemetryDeduplicationRepository(redisDB),
	}
}

// ProvideStorageRepositories creates all storage-related repositories
func ProvideStorageRepositories(clickhouseDB *database.ClickHouseDB) *StorageRepositories {
	return &StorageRepositories{
		BlobStorage: storageRepo.NewBlobStorageRepository(clickhouseDB.Conn),
	}
}

// ProvideBillingRepositories creates all billing-related repositories
func ProvideBillingRepositories(db *gorm.DB, logger *slog.Logger) *BillingRepositories {
	return &BillingRepositories{
		Usage:         billingRepo.NewUsageRepository(db, logger),
		BillingRecord: billingRepo.NewBillingRecordRepository(db, logger),
		Quota:         billingRepo.NewQuotaRepository(db, logger),
	}
}

// ProvideAnalyticsRepositories creates all analytics-related repositories
func ProvideAnalyticsRepositories(db *gorm.DB) *AnalyticsRepositories {
	return &AnalyticsRepositories{
		ProviderModel: analyticsRepo.NewProviderModelRepository(db),
	}
}

// ProvidePromptRepositories creates all prompt-related repositories
func ProvidePromptRepositories(db *gorm.DB, redisDB *database.RedisDB) *PromptRepositories {
	return &PromptRepositories{
		Prompt:         promptRepo.NewPromptRepository(db),
		Version:        promptRepo.NewVersionRepository(db),
		Label:          promptRepo.NewLabelRepository(db),
		ProtectedLabel: promptRepo.NewProtectedLabelRepository(db),
		Cache:          promptRepo.NewCacheRepository(redisDB),
	}
}

// ProvideCredentialsRepositories creates credentials repository container
func ProvideCredentialsRepositories(db *gorm.DB) *CredentialsRepositories {
	return &CredentialsRepositories{
		ProviderCredential: credentialsRepo.NewProviderCredentialRepository(db),
	}
}

// ProvidePlaygroundRepositories creates playground repository container
func ProvidePlaygroundRepositories(db *gorm.DB) *PlaygroundRepositories {
	return &PlaygroundRepositories{
		Session: playgroundRepo.NewSessionRepository(db),
	}
}

// ProvideEvaluationRepositories creates evaluation repository container
func ProvideEvaluationRepositories(db *gorm.DB) *EvaluationRepositories {
	return &EvaluationRepositories{
		ScoreConfig: evaluationRepo.NewScoreConfigRepository(db),
	}
}

// ProvideRepositories creates all repository containers
func ProvideRepositories(dbs *DatabaseContainer, logger *slog.Logger) *RepositoryContainer {
	return &RepositoryContainer{
		User:          ProvideUserRepositories(dbs.Postgres.DB),
		Auth:          ProvideAuthRepositories(dbs.Postgres.DB),
		Organization:  ProvideOrganizationRepositories(dbs.Postgres.DB),
		Observability: ProvideObservabilityRepositories(dbs.ClickHouse, dbs.Postgres.DB, dbs.Redis),
		Storage:       ProvideStorageRepositories(dbs.ClickHouse),
		Billing:       ProvideBillingRepositories(dbs.Postgres.DB, logger),
		Analytics:     ProvideAnalyticsRepositories(dbs.Postgres.DB),
		Prompt:        ProvidePromptRepositories(dbs.Postgres.DB, dbs.Redis),
		Credentials:   ProvideCredentialsRepositories(dbs.Postgres.DB),
		Playground:    ProvidePlaygroundRepositories(dbs.Postgres.DB),
		Evaluation:    ProvideEvaluationRepositories(dbs.Postgres.DB),
	}
}

// ProvideUserServices creates all user-related services
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

// ProvideAuthServices creates all auth-related services with proper dependencies
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

// ProvideOrganizationServices creates all organization-related services
func ProvideOrganizationServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	orgRepos *OrganizationRepositories,
	authServices *AuthServices,
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

	invitationSvc := orgService.NewInvitationService(
		orgRepos.Invitation,
		orgRepos.Organization,
		orgRepos.Member,
		userRepos.User,
		authServices.Role,
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

// ProvideObservabilityServices creates all observability-related services.
// TelemetryService is created without analytics worker - inject via SetAnalyticsWorker() after startup.
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
		observabilityRepos.Metrics,
		observabilityRepos.Logs,
		observabilityRepos.GenAIEvents,
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

// ProvideBillingServices creates all billing-related services
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

// ProvideAnalyticsServices creates all analytics-related services
func ProvideAnalyticsServices(
	analyticsRepos *AnalyticsRepositories,
) *AnalyticsServices {
	providerPricingServiceImpl := analyticsService.NewProviderPricingService(analyticsRepos.ProviderModel)

	return &AnalyticsServices{
		ProviderPricing: providerPricingServiceImpl,
	}
}

// ProvidePromptServices creates all prompt-related services
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

// ProvideCredentialsServices creates all credentials-related services
func ProvideCredentialsServices(
	credentialsRepos *CredentialsRepositories,
	analyticsRepos *AnalyticsRepositories,
	cfg *config.Config,
	logger *slog.Logger,
) (*CredentialsServices, error) {
	// Create encryption service from config
	encryptor, err := encryption.NewServiceFromBase64(cfg.Encryption.AIKeyEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
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
	}, nil
}

// ProvidePlaygroundServices creates all playground-related services
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

// ProvideEvaluationServices creates all evaluation-related services
func ProvideEvaluationServices(
	evaluationRepos *EvaluationRepositories,
	logger *slog.Logger,
) *EvaluationServices {
	scoreConfigSvc := evaluationService.NewScoreConfigService(
		evaluationRepos.ScoreConfig,
		logger,
	)

	return &EvaluationServices{
		ScoreConfig: scoreConfigSvc,
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

// ProvideEnterpriseServices creates all enterprise services using build tags
func ProvideEnterpriseServices(cfg *config.Config, logger *slog.Logger) *EnterpriseContainer {
	return &EnterpriseContainer{
		SSO:        sso.New(),         // Uses stub or real based on build tags
		RBAC:       rbac.New(),        // Uses stub or real based on build tags
		Compliance: compliance.New(),  // Uses stub or real based on build tags
		Analytics:  eeAnalytics.New(), // Uses stub or real based on build tags
	}
}

// HealthCheck checks health of all providers (mode-aware)
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

	health["mode"] = string(pc.Mode)

	return health
}

// Shutdown gracefully shuts down all providers
func (pc *ProviderContainer) Shutdown() error {
	var lastErr error
	logger := pc.Core.Logger

	if pc.Workers != nil && pc.Workers.TelemetryConsumer != nil {
		logger.Info("Stopping telemetry stream consumer...")
		pc.Workers.TelemetryConsumer.Stop()
		logger.Info("Telemetry stream consumer stopped")
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
