package app

import (
	"log/slog"
	"os"

	"brokle/internal/config"
	authService "brokle/internal/core/services/auth"
)

// ProvideAuthServices wires the auth-domain services. Constructor
// failure for the JWT signing service is fatal — there is no useful
// degraded mode for token validation, so we exit rather than panic
// later in the request path.
func ProvideAuthServices(
	cfg *config.Config,
	userRepos *UserRepositories,
	authRepos *AuthRepositories,
	orgRepos *OrganizationRepositories,
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
		orgRepos.Project,
		logger.With("service", "apiKey"),
	)

	authSvc := authService.NewAuthService(
		&cfg.Auth,
		userRepos.User,
		authRepos.UserSession,
		authRepos.AuditLog,
		jwtService,
		roleService,
		authRepos.PasswordResetToken,
		blacklistedTokenService,
		databases.Redis.Client,
		logger,
	)

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
