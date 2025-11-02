package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"brokle/internal/config"
	"brokle/internal/core/domain/common"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/gateway"
	"brokle/internal/core/domain/observability"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	orgService "brokle/internal/core/services/organization"
	registrationService "brokle/internal/core/services/registration"
	userService "brokle/internal/core/services/user"
	"brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/infrastructure/database"
	"brokle/internal/infrastructure/providers"
	_ "brokle/internal/infrastructure/providers/openai" // Import for side-effect (provider registration)
	authRepo "brokle/internal/infrastructure/repository/auth"
	billingRepo "brokle/internal/infrastructure/repository/billing"
	gatewayRepo "brokle/internal/infrastructure/repository/gateway"
	observabilityRepo "brokle/internal/infrastructure/repository/observability"
	orgRepo "brokle/internal/infrastructure/repository/organization"
	userRepo "brokle/internal/infrastructure/repository/user"
	"brokle/internal/infrastructure/storage"
	"brokle/internal/infrastructure/streams"
	billingService "brokle/internal/services/billing"
	gatewayService "brokle/internal/services/gateway"
	observabilityService "brokle/internal/services/observability"
	"brokle/internal/transport/http"
	"brokle/internal/transport/http/handlers"
	"brokle/internal/workers"
	gatewayAnalytics "brokle/internal/workers/analytics"
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
	Logger     *logrus.Logger
	Databases  *DatabaseContainer
	Repos      *RepositoryContainer
	TxManager  common.TransactionManager
	Services   *ServiceContainer
	Enterprise *EnterpriseContainer
}

// ServerContainer holds HTTP server components
type ServerContainer struct {
	HTTPServer *http.Server
}

// ProviderContainer holds all provider instances for dependency injection
type ProviderContainer struct {
	Core    *CoreContainer
	Server  *ServerContainer  // nil in worker mode
	Workers *WorkerContainer  // nil in server mode
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
	TelemetryConsumer   *workers.TelemetryStreamConsumer
	GatewayAnalytics    *gatewayAnalytics.GatewayAnalyticsWorker
}

// RepositoryContainer holds all repository instances organized by domain
type RepositoryContainer struct {
	User          *UserRepositories
	Auth          *AuthRepositories
	Organization  *OrganizationRepositories
	Observability *ObservabilityRepositories
	Gateway       *GatewayRepositories
	Billing       *BillingRepositories
}

// ServiceContainer holds all service instances organized by domain
type ServiceContainer struct {
	User               *UserServices
	Auth               *AuthServices
	Registration       registrationService.RegistrationService // Registration orchestrator
	// Direct organization services - no wrapper
	OrganizationService    organization.OrganizationService
	MemberService         organization.MemberService
	ProjectService        organization.ProjectService
	InvitationService     organization.InvitationService
	SettingsService       organization.OrganizationSettingsService
	// Observability services
	Observability         *observabilityService.ServiceRegistry
	// Gateway services
	Gateway               *GatewayServices
	// Billing services
	Billing               *BillingServices
}

// EnterpriseContainer holds all enterprise service instances
type EnterpriseContainer struct {
	SSO        sso.SSOProvider
	RBAC       rbac.RBACManager
	Compliance compliance.Compliance
	Analytics  analytics.EnterpriseAnalytics
}

// Domain-specific repository containers

// UserRepositories contains all user-related repositories
type UserRepositories struct {
	User user.Repository
}

// AuthRepositories contains all auth-related repositories
type AuthRepositories struct {
	UserSession         auth.UserSessionRepository
	BlacklistedToken    auth.BlacklistedTokenRepository
	PasswordResetToken  auth.PasswordResetTokenRepository
	APIKey              auth.APIKeyRepository
	Role                auth.RoleRepository
	OrganizationMember  auth.OrganizationMemberRepository
	Permission          auth.PermissionRepository
	RolePermission      auth.RolePermissionRepository
	AuditLog            auth.AuditLogRepository
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
	Observation            observability.ObservationRepository
	Score                  observability.ScoreRepository
	BlobStorage            observability.BlobStorageRepository
	TelemetryDeduplication observability.TelemetryDeduplicationRepository
}

// GatewayRepositories contains all gateway-related repositories
type GatewayRepositories struct {
	Provider       gateway.ProviderRepository
	Model          gateway.ModelRepository
	ProviderConfig gateway.ProviderConfigRepository
	Analytics      *gatewayRepo.Repository
}

// BillingRepositories contains all billing-related repositories
type BillingRepositories struct {
	Billing *billingRepo.Repository
}

// Domain-specific service containers

// UserServices contains all user-related services
type UserServices struct {
	User        user.UserService
	Profile     user.ProfileService
	Onboarding  user.OnboardingService
}

// AuthServices contains all auth-related services
type AuthServices struct {
	Auth                   auth.AuthService
	JWT                    auth.JWTService
	Sessions               auth.SessionService
	APIKey                 auth.APIKeyService
	Role                   auth.RoleService
	Permission             auth.PermissionService
	OrganizationMembers    auth.OrganizationMemberService
	BlacklistedTokens      auth.BlacklistedTokenService
	Scope                  auth.ScopeService
	OAuthProvider          *authService.OAuthProviderService
}

// GatewayServices contains all gateway-related services
type GatewayServices struct {
	Gateway gateway.GatewayService
	Routing gateway.RoutingService
	Cost    gateway.CostService
}

// BillingServices contains all billing-related services
type BillingServices struct {
	Billing *billingService.BillingService
}


// Provider functions for modular DI

// ProvideDatabases initializes all database connections
func ProvideDatabases(cfg *config.Config, logger *logrus.Logger) (*DatabaseContainer, error) {
	// Initialize PostgreSQL
	postgres, err := database.NewPostgresDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize Redis
	redis, err := database.NewRedisDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize ClickHouse
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
	// Create deduplication service
	deduplicationService := observabilityService.NewTelemetryDeduplicationService(
		core.Repos.Observability.TelemetryDeduplication,
	)

	// Create telemetry stream consumer
	consumerConfig := &workers.TelemetryStreamConsumerConfig{
		ConsumerGroup:     "telemetry-workers",
		ConsumerID:        fmt.Sprintf("worker-%s", ulid.New().String()),
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
		core.Services.Observability.ObservationService,
		core.Services.Observability.ScoreService,
	)

	// Create gateway analytics worker
	gatewayAnalyticsWorker := gatewayAnalytics.NewGatewayAnalyticsWorker(
		core.Logger,
		nil,
		core.Repos.Gateway.Analytics,
		core.Services.Billing.Billing,
	)

	return &WorkerContainer{
		TelemetryConsumer: telemetryConsumer,
		GatewayAnalytics:  gatewayAnalyticsWorker,
	}, nil
}

// ProvideCore creates shared infrastructure (used by both server and worker)
func ProvideCore(cfg *config.Config, logger *logrus.Logger) (*CoreContainer, error) {
	// Initialize databases
	databases, err := ProvideDatabases(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	repos := ProvideRepositories(databases, logger)

	// Initialize transaction manager (concrete → interface for dependency inversion)
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

	// Create gateway and billing services
	gatewayServices := ProvideGatewayServices(repos.Gateway, logger)
	billingServices := ProvideBillingServices(repos.Billing, repos.Gateway, logger)

	// Create observability services
	observabilityServices := ProvideObservabilityServices(repos.Observability, databases.Redis, cfg, logger)

	// Create auth services
	authServices := ProvideAuthServices(cfg, repos.User, repos.Auth, databases, logger)

	// Create user services
	userServices := ProvideUserServices(repos.User, repos.Auth, logger)

	// Create organization services
	orgService, memberService, projectService, invitationService, settingsService :=
		ProvideOrganizationServices(repos.User, repos.Auth, repos.Organization, authServices, logger)

	// Create registration service (orchestrates user, org, project creation)
	registrationSvc := registrationService.NewRegistrationService(
		core.TxManager, // Transaction manager for atomic multi-repository operations
		repos.User.User,
		orgService,
		projectService,
		memberService,
		invitationService,
		authServices.Role,
		authServices.Auth,
	)

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
		Gateway:             gatewayServices,
		Billing:             billingServices,
	}
}

// ProvideWorkerServices creates ONLY services worker needs (no auth)
func ProvideWorkerServices(core *CoreContainer) *ServiceContainer {
	cfg := core.Config
	logger := core.Logger
	repos := core.Repos
	databases := core.Databases

	// Only services worker uses for processing
	gatewayServices := ProvideGatewayServices(repos.Gateway, logger)
	billingServices := ProvideBillingServices(repos.Billing, repos.Gateway, logger)
	observabilityServices := ProvideObservabilityServices(repos.Observability, databases.Redis, cfg, logger)

	return &ServiceContainer{
		// Worker doesn't need auth/user/org services
		User:                nil,
		Auth:                nil,
		Registration:        nil,
		OrganizationService: nil,
		MemberService:       nil,
		ProjectService:      nil,
		InvitationService:   nil,
		SettingsService:     nil,
		// Worker only needs these
		Observability:       observabilityServices,
		Gateway:             gatewayServices,
		Billing:             billingServices,
	}
}

// ProvideServer creates HTTP server using shared core
func ProvideServer(core *CoreContainer) (*ServerContainer, error) {
	// Initialize HTTP handlers
	httpHandlers := handlers.NewHandlers(
		core.Config,
		core.Logger,
		core.Services.Auth.Auth,
		core.Services.Auth.APIKey,
		core.Services.Auth.BlacklistedTokens,
		core.Services.Registration, // Registration service for signup
		core.Services.Auth.OAuthProvider, // OAuth provider for Google/GitHub signup
		core.Services.User.User,
		core.Services.User.Profile,
		core.Services.User.Onboarding,
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
		core.Services.Gateway.Gateway,
		core.Services.Gateway.Routing,
		core.Services.Gateway.Cost,
	)

	// Initialize HTTP server
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

	return &ServerContainer{
		HTTPServer: httpServer,
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
func ProvideObservabilityRepositories(clickhouseDB *database.ClickHouseDB, redisDB *database.RedisDB) *ObservabilityRepositories {
	return &ObservabilityRepositories{
		Trace:                  observabilityRepo.NewTraceRepository(clickhouseDB.Conn),
		Observation:            observabilityRepo.NewObservationRepository(clickhouseDB.Conn),
		Score:                  observabilityRepo.NewScoreRepository(clickhouseDB.Conn),
		BlobStorage:            observabilityRepo.NewBlobStorageRepository(clickhouseDB.Conn),
		TelemetryDeduplication: observabilityRepo.NewTelemetryDeduplicationRepository(redisDB),
	}
}

// ProvideGatewayRepositories creates all gateway-related repositories
func ProvideGatewayRepositories(db *gorm.DB, conn clickhouse.Conn, logger *logrus.Logger) *GatewayRepositories {
	return &GatewayRepositories{
		Provider:       gatewayRepo.NewProviderRepository(db),
		Model:          gatewayRepo.NewModelRepository(db),
		ProviderConfig: gatewayRepo.NewProviderConfigRepository(db),
		Analytics:      gatewayRepo.NewRepository(conn, logger),
	}
}

// ProvideBillingRepositories creates all billing-related repositories
func ProvideBillingRepositories(db *gorm.DB, logger *logrus.Logger) *BillingRepositories {
	return &BillingRepositories{
		Billing: billingRepo.NewRepository(db, logger),
	}
}

// ProvideRepositories creates all repository containers
func ProvideRepositories(dbs *DatabaseContainer, logger *logrus.Logger) *RepositoryContainer {
	return &RepositoryContainer{
		User:          ProvideUserRepositories(dbs.Postgres.DB),
		Auth:          ProvideAuthRepositories(dbs.Postgres.DB),
		Organization:  ProvideOrganizationRepositories(dbs.Postgres.DB),
		Observability: ProvideObservabilityRepositories(dbs.ClickHouse, dbs.Redis),
		Gateway:       ProvideGatewayRepositories(dbs.Postgres.DB, dbs.ClickHouse.Conn, logger),
		Billing:       ProvideBillingRepositories(dbs.Postgres.DB, logger),
	}
}

// ProvideUserServices creates all user-related services
func ProvideUserServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	logger *logrus.Logger,
) *UserServices {
	// Import the actual user service implementations
	userSvc := userService.NewUserService(
		userRepos.User,
		nil, // AuthService - would need to be injected if needed
	)
	
	profileSvc := userService.NewProfileService(
		userRepos.User,
	)
	
	onboardingSvc := userService.NewOnboardingService(
		userRepos.User,
	)

	return &UserServices{
		User:        userSvc,
		Profile:     profileSvc,
		Onboarding:  onboardingSvc,
	}
}

// ProvideAuthServices creates all auth-related services with proper dependencies
func ProvideAuthServices(
	cfg *config.Config,
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	databases *DatabaseContainer,
	logger *logrus.Logger,
) *AuthServices {
	// Create JWT service with auth config
	jwtService, err := authService.NewJWTService(&cfg.Auth)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create JWT service")
	}

	// Create permission service with comprehensive permission management
	permissionService := authService.NewPermissionService(
		authRepos.Permission,
		authRepos.RolePermission,
	)

	// Create role service with clean RBAC (template roles only)
	roleService := authService.NewRoleService(
		authRepos.Role,
		authRepos.RolePermission,
	)

	// Create organization member service for RBAC membership management
	orgMemberService := authService.NewOrganizationMemberService(
		authRepos.OrganizationMember,
		authRepos.Role,
	)

	// Create blacklisted token service for immediate revocation
	blacklistedTokenService := authService.NewBlacklistedTokenService(
		authRepos.BlacklistedToken,
	)

	// Create session service for session management
	sessionService := authService.NewSessionService(
		&cfg.Auth,
		authRepos.UserSession,
		userRepos.User,
		jwtService,
	)

	// Create API key service for programmatic authentication
	apiKeyService := authService.NewAPIKeyService(
		authRepos.APIKey,
		authRepos.OrganizationMember,
	)

	// Create core auth service (without audit logging)
	coreAuthSvc := authService.NewAuthService(
		&cfg.Auth,
		userRepos.User,
		authRepos.UserSession,
		jwtService,
		roleService,
		authRepos.PasswordResetToken,
		blacklistedTokenService,
		databases.Redis.Client, // Redis client for OAuth session storage
	)

	// Wrap with audit decorator for clean separation of concerns
	authSvc := authService.NewAuditDecorator(coreAuthSvc, authRepos.AuditLog, logger)

	// Create scope service for scope-based authorization
	scopeService := authService.NewScopeService(
		authRepos.OrganizationMember,
		authRepos.Role,
		authRepos.Permission,
	)

	// Create OAuth provider service (Google/GitHub signup)
	frontendURL := "http://localhost:3000" // Default for development
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
	logger *logrus.Logger,
) (
	organization.OrganizationService,
	organization.MemberService,
	organization.ProjectService,
	organization.InvitationService,
	organization.OrganizationSettingsService,
) {
	// Create individual services
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

	// Create organization service with dependencies on other services
	orgSvc := orgService.NewOrganizationService(
		orgRepos.Organization,
		userRepos.User,
		memberSvc,
		projectSvc,
		authServices.Role,
	)

	// Create settings service
	settingsSvc := orgService.NewOrganizationSettingsService(
		orgRepos.Settings,
		orgRepos.Member,
	)

	return orgSvc, memberSvc, projectSvc, invitationSvc, settingsSvc
}

// ProvideObservabilityServices creates all observability-related services
// Note: TelemetryService is created without analytics worker (nil).
// The worker must be injected later via SetAnalyticsWorker() after it's started.
func ProvideObservabilityServices(
	observabilityRepos *ObservabilityRepositories,
	redisDB *database.RedisDB,
	cfg *config.Config,
	logger *logrus.Logger,
) *observabilityService.ServiceRegistry {
	// Create deduplication service for telemetry
	deduplicationService := observabilityService.NewTelemetryDeduplicationService(observabilityRepos.TelemetryDeduplication)

	// Create Redis Streams producer for telemetry
	streamProducer := streams.NewTelemetryStreamProducer(redisDB, logger)

	// Create telemetry service (health/metrics monitoring only)
	telemetryService := observabilityService.NewTelemetryService(
		deduplicationService,
		streamProducer,
		logger,
	)

	// Create S3 client for blob storage (pass nil if not configured)
	var s3Client *storage.S3Client
	if cfg.BlobStorage.Provider != "" && cfg.BlobStorage.BucketName != "" {
		var err error
		s3Client, err = storage.NewS3Client(&cfg.BlobStorage, logger)
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize S3 client, blob storage will be disabled")
		}
	}

	return observabilityService.NewServiceRegistry(
		observabilityRepos.Trace,
		observabilityRepos.Observation,
		observabilityRepos.Score,
		observabilityRepos.BlobStorage,
		s3Client,
		&cfg.BlobStorage,
		streamProducer,
		deduplicationService,
		telemetryService,
		logger,
	)
}

// ProvideGatewayServices creates all gateway-related services
func ProvideGatewayServices(
	gatewayRepos *GatewayRepositories,
	logger *logrus.Logger,
) *GatewayServices {
	// Create provider factory (OpenAI provider auto-registers via init())
	providerFactory := providers.NewProviderFactory()

	// Create cost service with repository dependencies
	costService := gatewayService.NewCostService(
		gatewayRepos.Model,
		gatewayRepos.Provider,
		gatewayRepos.ProviderConfig,
		logger,
	)

	// Create routing service with cost service dependency
	routingService := gatewayService.NewRoutingService(
		gatewayRepos.Provider,
		gatewayRepos.Model,
		gatewayRepos.ProviderConfig,
		costService,
		logger,
	)

	// Create gateway service with all dependencies
	gatewayServiceImpl := gatewayService.NewGatewayService(
		gatewayRepos.Provider,
		gatewayRepos.Model,
		gatewayRepos.ProviderConfig,
		routingService,
		costService,
		providerFactory,
		logger,
	)

	return &GatewayServices{
		Gateway: gatewayServiceImpl,
		Routing: routingService,
		Cost:    costService,
	}
}

// ProvideBillingServices creates all billing-related services
func ProvideBillingServices(
	billingRepos *BillingRepositories,
	gatewayRepos *GatewayRepositories,
	logger *logrus.Logger,
) *BillingServices {
	// Create organization service interface implementation
	// This is a simple implementation - in production you'd inject the real org service
	orgService := &simpleBillingOrgService{logger: logger}
	
	// Create billing service
	billingServiceImpl := billingService.NewBillingService(
		logger,
		nil, // config - use defaults
		billingRepos.Billing,
		orgService,
	)
	
	return &BillingServices{
		Billing: billingServiceImpl,
	}
}

// Event publisher removed - events system deleted (not part of OTEL architecture)

// simpleBillingOrgService is a simple implementation of BillingOrganizationService
type simpleBillingOrgService struct {
	logger *logrus.Logger
}

func (s *simpleBillingOrgService) GetBillingTier(ctx context.Context, orgID ulid.ULID) (string, error) {
	// Default to free tier for now - in production this would query the org service
	return "free", nil
}

func (s *simpleBillingOrgService) GetDiscountRate(ctx context.Context, orgID ulid.ULID) (float64, error) {
	// Default to no discount - in production this would query the org service
	return 0.0, nil
}

func (s *simpleBillingOrgService) GetPaymentMethod(ctx context.Context, orgID ulid.ULID) (*billingService.PaymentMethod, error) {
	// No payment method by default - in production this would query the org service
	return nil, nil
}

// ProvideEnterpriseServices creates all enterprise services using build tags
func ProvideEnterpriseServices(cfg *config.Config, logger *logrus.Logger) *EnterpriseContainer {
	return &EnterpriseContainer{
		SSO:        sso.New(),        // Uses stub or real based on build tags
		RBAC:       rbac.New(),       // Uses stub or real based on build tags
		Compliance: compliance.New(), // Uses stub or real based on build tags
		Analytics:  analytics.New(),  // Uses stub or real based on build tags
	}
}

// Health checking for all providers (mode-aware)
func (pc *ProviderContainer) HealthCheck() map[string]string {
	health := make(map[string]string)

	// Check core databases
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

	// Check worker health
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

	if pc.Workers != nil && pc.Workers.GatewayAnalytics != nil {
		if pc.Workers.GatewayAnalytics.IsHealthy() {
			health["gateway_analytics_worker"] = "healthy"
		} else {
			health["gateway_analytics_worker"] = "unhealthy: queue depth exceeded or processing failed"
		}
	}

	// Add deployment mode
	health["mode"] = string(pc.Mode)

	return health
}

// Graceful shutdown of all providers
func (pc *ProviderContainer) Shutdown() error {
	var lastErr error

	logger := pc.Core.Logger

	// Stop background workers first (if present)
	if pc.Workers != nil && pc.Workers.TelemetryConsumer != nil {
		logger.Info("Stopping telemetry stream consumer...")
		pc.Workers.TelemetryConsumer.Stop()
		logger.Info("Telemetry stream consumer stopped")
	}

	if pc.Workers != nil && pc.Workers.GatewayAnalytics != nil {
		logger.Info("Stopping gateway analytics worker...")
		pc.Workers.GatewayAnalytics.Stop()
		logger.Info("Gateway analytics worker stopped")
	}

	// Close database connections
	if pc.Core != nil && pc.Core.Databases != nil {
		if pc.Core.Databases.Postgres != nil {
			if err := pc.Core.Databases.Postgres.Close(); err != nil {
				logger.WithError(err).Error("Failed to close PostgreSQL connection")
				lastErr = err
			}
		}

		if pc.Core.Databases.Redis != nil {
			if err := pc.Core.Databases.Redis.Close(); err != nil {
				logger.WithError(err).Error("Failed to close Redis connection")
				lastErr = err
			}
		}

		if pc.Core.Databases.ClickHouse != nil {
			if err := pc.Core.Databases.ClickHouse.Close(); err != nil {
				logger.WithError(err).Error("Failed to close ClickHouse connection")
				lastErr = err
			}
		}
	}

	return lastErr
}
