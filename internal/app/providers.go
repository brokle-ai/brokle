package app

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/observability"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	orgService "brokle/internal/core/services/organization"
	userService "brokle/internal/core/services/user"
	"brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/infrastructure/database"
	authRepo "brokle/internal/infrastructure/repository/auth"
	observabilityRepo "brokle/internal/infrastructure/repository/observability"
	orgRepo "brokle/internal/infrastructure/repository/organization"
	userRepo "brokle/internal/infrastructure/repository/user"
	"brokle/internal/migration"
	observabilityService "brokle/internal/services/observability"
)

// ProviderContainer holds all provider instances for dependency injection
type ProviderContainer struct {
	Config     *config.Config
	Logger     *logrus.Logger
	Databases  *DatabaseContainer
	Repos      *RepositoryContainer
	Services   *ServiceContainer
	Enterprise *EnterpriseContainer
}

// DatabaseContainer holds all database connections
type DatabaseContainer struct {
	Postgres   *database.PostgresDB
	Redis      *database.RedisDB
	ClickHouse *database.ClickHouseDB
}

// RepositoryContainer holds all repository instances organized by domain
type RepositoryContainer struct {
	User          *UserRepositories
	Auth          *AuthRepositories
	Organization  *OrganizationRepositories
	Observability *ObservabilityRepositories
}

// ServiceContainer holds all service instances organized by domain
type ServiceContainer struct {
	User               *UserServices
	Auth               *AuthServices
	// Direct organization services - no wrapper
	OrganizationService    organization.OrganizationService
	MemberService         organization.MemberService
	ProjectService        organization.ProjectService
	EnvironmentService    organization.EnvironmentService
	InvitationService     organization.InvitationService
	SettingsService       organization.OrganizationSettingsService
	// Observability services
	Observability         *observabilityService.ServiceRegistry
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
	KeyPair             auth.KeyPairRepository
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
	Environment  organization.EnvironmentRepository
	Invitation   organization.InvitationRepository
	Settings     organization.OrganizationSettingsRepository
}

// ObservabilityRepositories contains all observability-related repositories
type ObservabilityRepositories struct {
	Trace        observability.TraceRepository
	Observation  observability.ObservationRepository
	QualityScore observability.QualityScoreRepository
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
	KeyPairs               auth.KeyPairService
	Role                   auth.RoleService
	Permission             auth.PermissionService
	OrganizationMembers    auth.OrganizationMemberService
	BlacklistedTokens      auth.BlacklistedTokenService
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
		KeyPair:            authRepo.NewKeyPairRepository(db),
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
		Environment:  orgRepo.NewEnvironmentRepository(db),
		Invitation:   orgRepo.NewInvitationRepository(db),
		Settings:     orgRepo.NewOrganizationSettingsRepository(db),
	}
}

// ProvideObservabilityRepositories creates all observability-related repositories
func ProvideObservabilityRepositories(postgresDB *gorm.DB, clickhouseDB *database.ClickHouseDB) *ObservabilityRepositories {
	return &ObservabilityRepositories{
		Trace:        observabilityRepo.NewTraceRepository(postgresDB),
		Observation:  observabilityRepo.NewObservationRepository(postgresDB),
		QualityScore: observabilityRepo.NewQualityScoreRepository(postgresDB),
	}
}

// ProvideRepositories creates all repository containers
func ProvideRepositories(dbs *DatabaseContainer) *RepositoryContainer {
	return &RepositoryContainer{
		User:          ProvideUserRepositories(dbs.Postgres.DB),
		Auth:          ProvideAuthRepositories(dbs.Postgres.DB),
		Organization:  ProvideOrganizationRepositories(dbs.Postgres.DB),
		Observability: ProvideObservabilityRepositories(dbs.Postgres.DB, dbs.ClickHouse),
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

	// Create core auth service (without audit logging)
	coreAuthSvc := authService.NewAuthService(
		&cfg.Auth,
		userRepos.User,
		authRepos.UserSession,
		jwtService,
		roleService,
		authRepos.PasswordResetToken,
		blacklistedTokenService,
	)

	// Wrap with audit decorator for clean separation of concerns
	authSvc := authService.NewAuditDecorator(coreAuthSvc, authRepos.AuditLog, logger)

	// Create key pair service for public+secret key authentication
	keyPairService := authService.NewKeyPairService(
		authRepos.KeyPair,
		userRepos.User,
	)

	return &AuthServices{
		Auth:                authSvc,
		JWT:                 jwtService,
		Sessions:            sessionService,
		KeyPairs:            keyPairService,
		Role:                roleService,
		Permission:          permissionService,
		OrganizationMembers: orgMemberService,
		BlacklistedTokens: blacklistedTokenService,
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
	organization.EnvironmentService,
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

	environmentSvc := orgService.NewEnvironmentService(
		orgRepos.Environment,
		orgRepos.Project,
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
		environmentSvc,
		authServices.Role,
	)

	// Create settings service
	settingsSvc := orgService.NewOrganizationSettingsService(
		orgRepos.Settings,
		orgRepos.Member,
	)

	return orgSvc, memberSvc, projectSvc, environmentSvc, invitationSvc, settingsSvc
}

// ProvideObservabilityServices creates all observability-related services
func ProvideObservabilityServices(
	observabilityRepos *ObservabilityRepositories,
	logger *logrus.Logger,
) *observabilityService.ServiceRegistry {
	// Create a simple event publisher for now
	eventPublisher := &simpleEventPublisher{logger: logger}

	return observabilityService.NewServiceRegistry(
		observabilityRepos.Trace,
		observabilityRepos.Observation,
		observabilityRepos.QualityScore,
		eventPublisher,
	)
}

// simpleEventPublisher is a simple implementation of EventPublisher for initial integration
type simpleEventPublisher struct {
	logger *logrus.Logger
}

func (p *simpleEventPublisher) Publish(ctx context.Context, event *observability.Event) error {
	p.logger.WithFields(logrus.Fields{
		"event_type": event.Type,
		"subject":    event.Subject,
		"project_id": event.ProjectID.String(),
	}).Debug("publishing event")
	return nil
}

func (p *simpleEventPublisher) PublishBatch(ctx context.Context, events []*observability.Event) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// ProvideServices creates all service containers with proper dependency resolution
func ProvideServices(
	cfg *config.Config,
	repos *RepositoryContainer,
	logger *logrus.Logger,
) *ServiceContainer {
	// Create auth services first (other services depend on them)
	authServices := ProvideAuthServices(cfg, repos.User, repos.Auth, logger)

	// Create user services
	userServices := ProvideUserServices(repos.User, repos.Auth, logger)

	// Create organization services (depends on auth services)
	orgService, memberService, projectService, environmentService, invitationService, settingsService := ProvideOrganizationServices(
		repos.User,
		repos.Auth,
		repos.Organization,
		authServices,
		logger,
	)

	// Create observability services
	observabilityServices := ProvideObservabilityServices(repos.Observability, logger)

	return &ServiceContainer{
		User:               userServices,
		Auth:               authServices,
		OrganizationService: orgService,
		MemberService:      memberService,
		ProjectService:     projectService,
		EnvironmentService: environmentService,
		InvitationService:  invitationService,
		SettingsService:    settingsService,
		Observability:      observabilityServices,
	}
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

// ProvideAll creates the complete provider container with all dependencies
func ProvideAll(cfg *config.Config, logger *logrus.Logger) (*ProviderContainer, error) {
	// Initialize databases
	databases, err := ProvideDatabases(cfg, logger)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	repos := ProvideRepositories(databases)

	// Initialize services
	services := ProvideServices(cfg, repos, logger)

	// Initialize enterprise services
	enterprise := ProvideEnterpriseServices(cfg, logger)

	// Initialize and run auto-migration if enabled
	if cfg.Database.AutoMigrate {
		logger.Info("Auto-migration is enabled, running database migrations...")

		migrationManager, err := migration.NewManager(cfg)
		if err != nil {
			logger.WithError(err).Error("Failed to initialize migration manager for auto-migration")
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := migrationManager.AutoMigrate(ctx); err != nil {
				logger.WithError(err).Error("Auto-migration failed - continuing with startup")
			} else {
				logger.Info("Auto-migration completed successfully")
			}

			// Close migration manager after use
			if err := migrationManager.Shutdown(); err != nil {
				logger.WithError(err).Warn("Failed to shutdown migration manager after auto-migration")
			}
		}
	} else {
		logger.Debug("Auto-migration is disabled")
	}

	return &ProviderContainer{
		Config:     cfg,
		Logger:     logger,
		Databases:  databases,
		Repos:      repos,
		Services:   services,
		Enterprise: enterprise,
	}, nil
}

// Backward compatibility types

// Repositories provides a flattened view of all repositories
type Repositories struct {
	UserRepository              user.Repository
	OrganizationRepository      organization.OrganizationRepository
	MemberRepository            organization.MemberRepository
	ProjectRepository           organization.ProjectRepository
	EnvironmentRepository       organization.EnvironmentRepository
	InvitationRepository        organization.InvitationRepository
	OrganizationSettingsRepository organization.OrganizationSettingsRepository
	UserSessionRepository       auth.UserSessionRepository
	PasswordResetTokenRepository auth.PasswordResetTokenRepository
	KeyPairRepository           auth.KeyPairRepository
	RoleRepository              auth.RoleRepository
	PermissionRepository        auth.PermissionRepository
	RolePermissionRepository    auth.RolePermissionRepository
	AuditLogRepository          auth.AuditLogRepository
}

// Services provides a flattened view of all services
type Services struct {
	AuthService                   auth.AuthService
	OrganizationService          organization.OrganizationService
	OrganizationSettingsService  organization.OrganizationSettingsService
	ComplianceService            compliance.Compliance
	SSOService                   sso.SSOProvider
	RBACService                  rbac.RBACManager
	EnterpriseAnalytics          analytics.EnterpriseAnalytics
}

// Convenience accessors for backward compatibility

// GetAllRepositories returns a flattened view of all repositories (for backward compatibility)
func (pc *ProviderContainer) GetAllRepositories() *Repositories {
	return &Repositories{
		UserRepository:                 pc.Repos.User.User,
		OrganizationRepository:         pc.Repos.Organization.Organization,
		MemberRepository:               pc.Repos.Organization.Member,
		ProjectRepository:              pc.Repos.Organization.Project,
		EnvironmentRepository:          pc.Repos.Organization.Environment,
		InvitationRepository:           pc.Repos.Organization.Invitation,
		OrganizationSettingsRepository: pc.Repos.Organization.Settings,
		UserSessionRepository:          pc.Repos.Auth.UserSession,
		PasswordResetTokenRepository:   pc.Repos.Auth.PasswordResetToken,
		KeyPairRepository:              pc.Repos.Auth.KeyPair,
		RoleRepository:                 pc.Repos.Auth.Role,
		PermissionRepository:           pc.Repos.Auth.Permission,
		RolePermissionRepository:       pc.Repos.Auth.RolePermission,
		AuditLogRepository:             pc.Repos.Auth.AuditLog,
	}
}

// GetAllServices returns a flattened view of all services (for backward compatibility)
func (pc *ProviderContainer) GetAllServices() *Services {
	return &Services{
		AuthService:                 pc.Services.Auth.Auth,
		OrganizationService:        pc.Services.OrganizationService,
		OrganizationSettingsService: pc.Services.SettingsService,
		ComplianceService:          pc.Enterprise.Compliance,
		SSOService:                 pc.Enterprise.SSO,
		RBACService:                pc.Enterprise.RBAC,
		EnterpriseAnalytics:        pc.Enterprise.Analytics,
	}
}

// Health checking for all providers
func (pc *ProviderContainer) HealthCheck() map[string]string {
	health := make(map[string]string)

	// Check database connections
	if pc.Databases.Postgres != nil {
		if err := pc.Databases.Postgres.Health(); err != nil {
			health["postgres"] = "unhealthy: " + err.Error()
		} else {
			health["postgres"] = "healthy"
		}
	}

	if pc.Databases.Redis != nil {
		if err := pc.Databases.Redis.Health(); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
		} else {
			health["redis"] = "healthy"
		}
	}

	if pc.Databases.ClickHouse != nil {
		if err := pc.Databases.ClickHouse.Health(); err != nil {
			health["clickhouse"] = "unhealthy: " + err.Error()
		} else {
			health["clickhouse"] = "healthy"
		}
	}

	return health
}

// Graceful shutdown of all providers
func (pc *ProviderContainer) Shutdown() error {
	var lastErr error

	// Close database connections
	if pc.Databases.Postgres != nil {
		if err := pc.Databases.Postgres.Close(); err != nil {
			pc.Logger.WithError(err).Error("Failed to close PostgreSQL connection")
			lastErr = err
		}
	}

	if pc.Databases.Redis != nil {
		if err := pc.Databases.Redis.Close(); err != nil {
			pc.Logger.WithError(err).Error("Failed to close Redis connection")
			lastErr = err
		}
	}

	if pc.Databases.ClickHouse != nil {
		if err := pc.Databases.ClickHouse.Close(); err != nil {
			pc.Logger.WithError(err).Error("Failed to close ClickHouse connection")
			lastErr = err
		}
	}

	return lastErr
}
