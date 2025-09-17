package handlers

import (
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	obsServices "brokle/internal/services/observability"
	"brokle/internal/transport/http/handlers/admin"
	"brokle/internal/transport/http/handlers/ai"
	"brokle/internal/transport/http/handlers/analytics"
	"brokle/internal/transport/http/handlers/keypair"
	authHandler "brokle/internal/transport/http/handlers/auth"
	"brokle/internal/transport/http/handlers/billing"
	"brokle/internal/transport/http/handlers/environment"
	"brokle/internal/transport/http/handlers/health"
	"brokle/internal/transport/http/handlers/logs"
	"brokle/internal/transport/http/handlers/metrics"
	"brokle/internal/transport/http/handlers/observability"
	organizationHandler "brokle/internal/transport/http/handlers/organization"
	"brokle/internal/transport/http/handlers/project"
	"brokle/internal/transport/http/handlers/rbac"
	userHandler "brokle/internal/transport/http/handlers/user"
	"brokle/internal/transport/http/handlers/websocket"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	Health       *health.Handler
	Metrics      *metrics.Handler
	Auth         *authHandler.Handler
	User         *userHandler.Handler
	Organization *organizationHandler.Handler
	Project      *project.Handler
	Environment  *environment.Handler
	KeyPair      *keypair.Handler
	Analytics     *analytics.Handler
	Logs          *logs.Handler
	Billing       *billing.Handler
	AI            *ai.Handler
	WebSocket     *websocket.Handler
	Admin         *admin.TokenAdminHandler
	RBAC          *rbac.Handler
	Observability *observability.Handler
}

// NewHandlers creates a new handlers instance with all dependencies
func NewHandlers(
	cfg *config.Config,
	logger *logrus.Logger,
	authService auth.AuthService,
	blacklistedTokens auth.BlacklistedTokenService,
	userService user.UserService,
	profileService user.ProfileService,
	onboardingService user.OnboardingService,
	organizationService organization.OrganizationService,
	memberService organization.MemberService,
	projectService organization.ProjectService,
	environmentService organization.EnvironmentService,
	invitationService organization.InvitationService,
	settingsService organization.OrganizationSettingsService,
	roleService auth.RoleService,
	permissionService auth.PermissionService,
	organizationMemberService auth.OrganizationMemberService,
	keyPairService auth.KeyPairService,
	observabilityServices *obsServices.ServiceRegistry,
	// Add other service dependencies as they're implemented
) *Handlers {
	return &Handlers{
		Health:       health.NewHandler(cfg, logger),
		Metrics:      metrics.NewHandler(cfg, logger),
		Auth:         authHandler.NewHandler(cfg, logger, authService, userService),
		User:         userHandler.NewHandler(cfg, logger, userService, profileService, onboardingService),
		Organization: organizationHandler.NewHandler(cfg, logger, organizationService, memberService, projectService, environmentService, invitationService, settingsService, userService, roleService),
		Project:      project.NewHandler(cfg, logger, projectService, organizationService, memberService, environmentService),
		Environment:  environment.NewHandler(cfg, logger),
		KeyPair:      keypair.NewHandler(cfg, logger, keyPairService),
		Analytics:    analytics.NewHandler(cfg, logger),
		Logs:         logs.NewHandler(cfg, logger),
		Billing:      billing.NewHandler(cfg, logger),
		AI:           ai.NewHandler(cfg, logger),
		WebSocket:    websocket.NewHandler(cfg, logger),
		Admin:         admin.NewTokenAdminHandler(authService, blacklistedTokens, logger),
		RBAC:          rbac.NewHandler(cfg, logger, roleService, permissionService, organizationMemberService),
		Observability: observability.NewHandler(cfg, logger, observabilityServices),
	}
}