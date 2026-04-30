package app

// Container types for the application's hierarchical dependency
// graph. The provider functions in repositories.go,
// services_<domain>.go, services.go, workers.go, server.go and
// providers.go return these structs; the orchestrator chains them
// together at boot.

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"brokle/internal/config"
	analyticsDomain "brokle/internal/core/domain/analytics"
	annotationDomain "brokle/internal/core/domain/annotation"
	authDomain "brokle/internal/core/domain/auth"
	billingDomain "brokle/internal/core/domain/billing"
	commonDomain "brokle/internal/core/domain/common"
	credentialsDomain "brokle/internal/core/domain/credentials"
	dashboardDomain "brokle/internal/core/domain/dashboard"
	evaluationDomain "brokle/internal/core/domain/evaluation"
	observabilityDomain "brokle/internal/core/domain/observability"
	organizationDomain "brokle/internal/core/domain/organization"
	playgroundDomain "brokle/internal/core/domain/playground"
	promptDomain "brokle/internal/core/domain/prompt"
	storageDomain "brokle/internal/core/domain/storage"
	userDomain "brokle/internal/core/domain/user"
	websiteDomain "brokle/internal/core/domain/website"
	analyticsService "brokle/internal/core/services/analytics"
	annotationService "brokle/internal/core/services/annotation"
	authService "brokle/internal/core/services/auth"
	billingService "brokle/internal/core/services/billing"
	commentService "brokle/internal/core/services/comment"
	credentialsService "brokle/internal/core/services/credentials"
	dashboardService "brokle/internal/core/services/dashboard"
	evaluationService "brokle/internal/core/services/evaluation"
	observabilityService "brokle/internal/core/services/observability"
	orgService "brokle/internal/core/services/organization"
	playgroundService "brokle/internal/core/services/playground"
	promptService "brokle/internal/core/services/prompt"
	registrationService "brokle/internal/core/services/registration"
	userService "brokle/internal/core/services/user"
	websiteService "brokle/internal/core/services/website"
	eeAnalytics "brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/db"
	"brokle/internal/server"
	grpcTransport "brokle/internal/transport/grpc"
	"brokle/internal/workers"
	annotationWorker "brokle/internal/workers/annotation"
	evaluationWorker "brokle/internal/workers/evaluation"
)

// DeploymentMode signals which mode the binary was launched in.
// ProvideServerServices and ProvideWorkerServices use it to omit
// services that are not needed for the active mode.
type DeploymentMode string

const (
	ModeServer DeploymentMode = "server"
	ModeWorker DeploymentMode = "worker"
)

// ----- Top-level containers -----------------------------------------

type CoreContainer struct {
	Config     *config.Config
	Logger     *slog.Logger
	Databases  *DatabaseContainer
	Repos      *RepositoryContainer
	Transactor commonDomain.Transactor
	Services   *ServiceContainer
	Enterprise *EnterpriseContainer
}

type ServerContainer struct {
	HTTPServer *server.Server
	GRPCServer *grpcTransport.Server
}

type ProviderContainer struct {
	Core    *CoreContainer
	Server  *ServerContainer // nil in worker mode
	Workers *WorkerContainer // nil in server mode
	Mode    DeploymentMode
}

type DatabaseContainer struct {
	// Pool is the process-wide pgx v5 pool owned by this container.
	// All repositories access PostgreSQL through TxManager, which wraps
	// this pool; there is no second handle.
	Pool       *pgxpool.Pool
	TxManager  *db.TxManager
	Redis      *database.RedisDB
	ClickHouse *database.ClickHouseDB
}

type WorkerContainer struct {
	TelemetryConsumer        *workers.TelemetryStreamConsumer
	EvaluatorWorker          *evaluationWorker.EvaluatorWorker
	EvaluationWorker         *evaluationWorker.EvaluationWorker
	ManualTriggerWorker      *evaluationWorker.ManualTriggerWorker
	UsageAggregationWorker   *workers.UsageAggregationWorker
	ContractExpirationWorker *workers.ContractExpirationWorker
	LockExpiryWorker         *annotationWorker.LockExpiryWorker
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
	Annotation    *AnnotationRepositories
	Website       *WebsiteRepositories
}

type ServiceContainer struct {
	User                *UserServices
	Auth                *AuthServices
	Registration        *registrationService.RegistrationService
	OrganizationService *orgService.OrganizationService
	MemberService       *orgService.MemberService
	ProjectService      *orgService.ProjectService
	InvitationService   *orgService.InvitationService
	SettingsService     *orgService.OrganizationSettingsService
	Observability       *observabilityService.ServiceRegistry
	Billing             *BillingServices
	Analytics           *AnalyticsServices
	Prompt              *PromptServices
	Credentials         *CredentialsServices
	Playground          *PlaygroundServices
	Evaluation          *EvaluationServices
	Dashboard           *DashboardServices
	Annotation          *AnnotationServices
	Comment             *commentService.CommentService
	Website             *websiteService.WebsiteService
}

type EnterpriseContainer struct {
	SSO        sso.SSOProvider
	RBAC       rbac.RBACManager
	Compliance compliance.Compliance
	Analytics  eeAnalytics.EnterpriseAnalytics
}

// ----- Per-domain repository containers -----------------------------

type UserRepositories struct {
	User userDomain.Repository
}

type AuthRepositories struct {
	UserSession        authDomain.UserSessionRepository
	BlacklistedToken   authDomain.BlacklistedTokenRepository
	PasswordResetToken authDomain.PasswordResetTokenRepository
	APIKey             authDomain.APIKeyRepository
	Role               authDomain.RoleRepository
	OrganizationMember authDomain.OrganizationMemberRepository
	ProjectMember      authDomain.ProjectMemberRepository
	Permission         authDomain.PermissionRepository
	RolePermission     authDomain.RolePermissionRepository
	AuditLog           authDomain.AuditLogRepository
}

type OrganizationRepositories struct {
	Organization organizationDomain.OrganizationRepository
	Member       organizationDomain.MemberRepository
	Project      organizationDomain.ProjectRepository
	Invitation   organizationDomain.InvitationRepository
	Settings     organizationDomain.OrganizationSettingsRepository
}

type ObservabilityRepositories struct {
	Trace                  observabilityDomain.TraceRepository
	Score                  observabilityDomain.ScoreRepository
	ScoreAnalytics         observabilityDomain.ScoreAnalyticsRepository
	Metrics                observabilityDomain.MetricsRepository
	Logs                   observabilityDomain.LogsRepository
	GenAIEvents            observabilityDomain.GenAIEventsRepository
	TelemetryDeduplication observabilityDomain.TelemetryDeduplicationRepository
	FilterPreset           observabilityDomain.FilterPresetRepository
}

type StorageRepositories struct {
	BlobStorage storageDomain.BlobStorageRepository
}

type BillingRepositories struct {
	BillingRecord billingDomain.BillingRecordRepository
	// Usage-based billing repositories (Spans + GB + Scores)
	BillableUsage       billingDomain.BillableUsageRepository
	Plan                billingDomain.PlanRepository
	OrganizationBilling billingDomain.OrganizationBillingRepository
	UsageBudget         billingDomain.UsageBudgetRepository
	UsageAlert          billingDomain.UsageAlertRepository
	// Enterprise custom pricing repositories
	Contract        billingDomain.ContractRepository
	VolumeTier      billingDomain.VolumeDiscountTierRepository
	ContractHistory billingDomain.ContractHistoryRepository
}

type AnalyticsRepositories struct {
	ProviderModel analyticsDomain.ProviderModelRepository
	Overview      analyticsDomain.OverviewRepository
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
	ScoreConfig        evaluationDomain.ScoreConfigRepository
	Dataset            evaluationDomain.DatasetRepository
	DatasetItem        evaluationDomain.DatasetItemRepository
	DatasetVersion     evaluationDomain.DatasetVersionRepository
	Experiment         evaluationDomain.ExperimentRepository
	ExperimentItem     evaluationDomain.ExperimentItemRepository
	ExperimentConfig   evaluationDomain.ExperimentConfigRepository
	Evaluator          evaluationDomain.EvaluatorRepository
	EvaluatorExecution evaluationDomain.EvaluatorExecutionRepository
}

type DashboardRepositories struct {
	Dashboard   dashboardDomain.DashboardRepository
	WidgetQuery dashboardDomain.WidgetQueryRepository
	Template    dashboardDomain.TemplateRepository
}

type AnnotationRepositories struct {
	Queue      annotationDomain.QueueRepository
	Item       annotationDomain.ItemRepository
	Assignment annotationDomain.AssignmentRepository
}

type WebsiteRepositories struct {
	ContactSubmission websiteDomain.ContactSubmissionRepository
}

// ----- Per-domain service containers --------------------------------

type UserServices struct {
	User    *userService.UserService
	Profile *userService.ProfileService
}

type AuthServices struct {
	Auth                *authService.AuthService
	JWT                 *authService.JWTService
	Sessions            *authService.SessionService
	APIKey              *authService.APIKeyService
	Role                *authService.RoleService
	Permission          *authService.PermissionService
	OrganizationMembers *authService.OrganizationMemberService
	ProjectMembers      *authService.ProjectMemberService
	BlacklistedTokens   *authService.BlacklistedTokenService
	OAuthProvider       *authService.OAuthProviderService
}

type BillingServices struct {
	// Usage-based billing services (Spans + GB + Scores)
	BillableUsage *billingService.BillableUsageService
	Budget        *billingService.BudgetService
	// Enterprise custom pricing services
	Pricing  *billingService.PricingService
	Contract *billingService.ContractService
}

type AnalyticsServices struct {
	ProviderPricing *analyticsService.ProviderPricingService
	Overview        *analyticsService.OverviewService
}

type PromptServices struct {
	Prompt    *promptService.PromptService
	Compiler  *promptService.CompilerService
	Execution *promptService.ExecutionService
}

type CredentialsServices struct {
	ProviderCredential *credentialsService.ProviderCredentialService
	ModelCatalog       *credentialsService.ModelCatalogService
}

type PlaygroundServices struct {
	Playground *playgroundService.PlaygroundService
}

type EvaluationServices struct {
	ScoreConfig        *evaluationService.ScoreConfigService
	Dataset            *evaluationService.DatasetService
	DatasetItem        *evaluationService.DatasetItemService
	DatasetVersion     *evaluationService.DatasetVersionService
	Experiment         *evaluationService.ExperimentService
	ExperimentItem     *evaluationService.ExperimentItemService
	ExperimentWizard   *evaluationService.ExperimentWizardService
	Evaluator          *evaluationService.EvaluatorService
	EvaluatorExecution *evaluationService.EvaluatorExecutionService
}

type DashboardServices struct {
	Dashboard   *dashboardService.DashboardService
	WidgetQuery *dashboardService.WidgetQueryService
	Template    *dashboardService.TemplateService
}

type AnnotationServices struct {
	Queue      *annotationService.QueueService
	Item       *annotationService.ItemService
	Assignment *annotationService.AssignmentService
}
