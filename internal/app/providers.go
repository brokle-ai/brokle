package app

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	orgService "brokle/internal/core/services/organization"
	"brokle/internal/ee/analytics"
	"brokle/internal/ee/compliance"
	"brokle/internal/ee/rbac"
	"brokle/internal/ee/sso"
	"brokle/internal/infrastructure/database"
	authRepo "brokle/internal/infrastructure/repository/auth"
	orgRepo "brokle/internal/infrastructure/repository/organization"
	userRepo "brokle/internal/infrastructure/repository/user"
	"brokle/internal/migration"
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
	User         *UserRepositories
	Auth         *AuthRepositories
	Organization *OrganizationRepositories
}

// ServiceContainer holds all service instances organized by domain
type ServiceContainer struct {
	User         *UserServices
	Auth         *AuthServices
	Organization *OrganizationServices
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
	Session        auth.SessionRepository
	APIKey         auth.APIKeyRepository
	Role           auth.RoleRepository
	Permission     auth.PermissionRepository
	RolePermission auth.RolePermissionRepository
	AuditLog       auth.AuditLogRepository
}

// OrganizationRepositories contains all organization-related repositories
type OrganizationRepositories struct {
	Organization organization.OrganizationRepository
	Member       organization.MemberRepository
	Project      organization.ProjectRepository
	Environment  organization.EnvironmentRepository
	Invitation   organization.InvitationRepository
}

// Domain-specific service containers

// UserServices contains all user-related services
type UserServices struct {
	User user.Service
}

// AuthServices contains all auth-related services
type AuthServices struct {
	Auth auth.AuthService
	JWT  auth.JWTService
	Role auth.RoleService
}

// OrganizationServices contains all organization-related services
type OrganizationServices struct {
	Organization organization.OrganizationService
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
		Session:        authRepo.NewSessionRepository(db),
		APIKey:         authRepo.NewAPIKeyRepository(db),
		Role:           authRepo.NewRoleRepository(db),
		Permission:     authRepo.NewPermissionRepository(db),
		RolePermission: authRepo.NewRolePermissionRepository(db),
		AuditLog:       authRepo.NewAuditLogRepository(db),
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
	}
}

// ProvideRepositories creates all repository containers
func ProvideRepositories(dbs *DatabaseContainer) *RepositoryContainer {
	return &RepositoryContainer{
		User:         ProvideUserRepositories(dbs.Postgres.DB),
		Auth:         ProvideAuthRepositories(dbs.Postgres.DB),
		Organization: ProvideOrganizationRepositories(dbs.Postgres.DB),
	}
}

// ProvideUserServices creates all user-related services
func ProvideUserServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	logger *logrus.Logger,
) *UserServices {
	// TODO: Implement user service once created
	// userSvc := userService.NewUserService(
	//     userRepos.User,
	//     authRepos.AuditLog,
	//     logger,
	// )

	return &UserServices{
		// User: userSvc,
		User: nil, // Placeholder until user service is implemented
	}
}

// ProvideAuthServices creates all auth-related services with proper dependencies
func ProvideAuthServices(
	cfg *config.Config,
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	logger *logrus.Logger,
) *AuthServices {
	// Create JWT service with config
	jwtService := authService.NewJWTService(&auth.TokenConfig{
		SigningKey:      cfg.JWT.PrivateKey,
		SigningMethod:   cfg.JWT.Algorithm,
		Issuer:          cfg.JWT.Issuer,
		AccessTokenTTL:  cfg.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
		APIKeyTokenTTL:  cfg.JWT.APIKeyTokenTTL,
		ClockSkew:       5 * time.Minute,
		RequireAudience: false,
	})

	// Create role service with comprehensive RBAC
	roleService := authService.NewRoleService(
		authRepos.Role,
		authRepos.Permission,
		authRepos.RolePermission,
	)

	// Create auth service with all dependencies
	authSvc := authService.NewAuthService(
		userRepos.User,
		authRepos.Session,
		authRepos.AuditLog,
		jwtService,
		roleService,
	)

	return &AuthServices{
		Auth: authSvc,
		JWT:  jwtService,
		Role: roleService,
	}
}

// ProvideOrganizationServices creates all organization-related services
func ProvideOrganizationServices(
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	orgRepos *OrganizationRepositories,
	authServices *AuthServices,
	logger *logrus.Logger,
) *OrganizationServices {
	// Create organization service with all dependencies
	orgSvc := orgService.NewOrganizationService(
		orgRepos.Organization,
		orgRepos.Member,
		orgRepos.Project,
		orgRepos.Environment,
		orgRepos.Invitation,
		userRepos.User,
		authServices.Role,
		authRepos.AuditLog,
	)

	return &OrganizationServices{
		Organization: orgSvc,
	}
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
	orgServices := ProvideOrganizationServices(
		repos.User,
		repos.Auth,
		repos.Organization,
		authServices,
		logger,
	)

	return &ServiceContainer{
		User:         userServices,
		Auth:         authServices,
		Organization: orgServices,
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
	UserRepository           user.Repository
	OrganizationRepository   organization.OrganizationRepository
	MemberRepository         organization.MemberRepository
	ProjectRepository        organization.ProjectRepository
	EnvironmentRepository    organization.EnvironmentRepository
	InvitationRepository     organization.InvitationRepository
	SessionRepository        auth.SessionRepository
	APIKeyRepository         auth.APIKeyRepository
	RoleRepository           auth.RoleRepository
	PermissionRepository     auth.PermissionRepository
	RolePermissionRepository auth.RolePermissionRepository
	AuditLogRepository       auth.AuditLogRepository
}

// Services provides a flattened view of all services
type Services struct {
	AuthService         auth.AuthService
	OrganizationService organization.OrganizationService
	ComplianceService   compliance.Compliance
	SSOService          sso.SSOProvider
	RBACService         rbac.RBACManager
	EnterpriseAnalytics analytics.EnterpriseAnalytics
}

// Convenience accessors for backward compatibility

// GetAllRepositories returns a flattened view of all repositories (for backward compatibility)
func (pc *ProviderContainer) GetAllRepositories() *Repositories {
	return &Repositories{
		UserRepository:           pc.Repos.User.User,
		OrganizationRepository:   pc.Repos.Organization.Organization,
		MemberRepository:         pc.Repos.Organization.Member,
		ProjectRepository:        pc.Repos.Organization.Project,
		EnvironmentRepository:    pc.Repos.Organization.Environment,
		InvitationRepository:     pc.Repos.Organization.Invitation,
		SessionRepository:        pc.Repos.Auth.Session,
		APIKeyRepository:         pc.Repos.Auth.APIKey,
		RoleRepository:           pc.Repos.Auth.Role,
		PermissionRepository:     pc.Repos.Auth.Permission,
		RolePermissionRepository: pc.Repos.Auth.RolePermission,
		AuditLogRepository:       pc.Repos.Auth.AuditLog,
	}
}

// GetAllServices returns a flattened view of all services (for backward compatibility)
func (pc *ProviderContainer) GetAllServices() *Services {
	return &Services{
		AuthService:         pc.Services.Auth.Auth,
		OrganizationService: pc.Services.Organization.Organization,
		ComplianceService:   pc.Enterprise.Compliance,
		SSOService:          pc.Enterprise.SSO,
		RBACService:         pc.Enterprise.RBAC,
		EnterpriseAnalytics: pc.Enterprise.Analytics,
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
