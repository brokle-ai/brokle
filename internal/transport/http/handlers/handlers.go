package handlers

import (
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/gateway"
	"brokle/internal/core/domain/organization"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	"brokle/internal/core/services/registration"
	obsServices "brokle/internal/core/services/observability"
	"brokle/internal/transport/http/handlers/admin"
	"brokle/internal/transport/http/handlers/ai"
	"brokle/internal/transport/http/handlers/analytics"
	"brokle/internal/transport/http/handlers/apikey"
	authHandler "brokle/internal/transport/http/handlers/auth"
	"brokle/internal/transport/http/handlers/billing"
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
	Health        *health.Handler
	Metrics       *metrics.Handler
	Auth          *authHandler.Handler
	User          *userHandler.Handler
	Organization  *organizationHandler.Handler
	Project       *project.Handler
	APIKey        *apikey.Handler
	Analytics     *analytics.Handler
	Logs          *logs.Handler
	Billing       *billing.Handler
	AI            *ai.Handler
	WebSocket     *websocket.Handler
	Admin         *admin.TokenAdminHandler
	RBAC          *rbac.Handler
	Observability *observability.Handler
	OTLP          *observability.OTLPHandler
}

// NewHandlers creates a new handlers instance with all dependencies
func NewHandlers(
	cfg *config.Config,
	logger *logrus.Logger,
	authService auth.AuthService,
	apiKeyService auth.APIKeyService,
	blacklistedTokens auth.BlacklistedTokenService,
	registrationService registration.RegistrationService,
	oauthProvider *authService.OAuthProviderService,
	userService user.UserService,
	profileService user.ProfileService,
	onboardingService user.OnboardingService,
	organizationService organization.OrganizationService,
	memberService organization.MemberService,
	projectService organization.ProjectService,
	invitationService organization.InvitationService,
	settingsService organization.OrganizationSettingsService,
	roleService auth.RoleService,
	permissionService auth.PermissionService,
	organizationMemberService auth.OrganizationMemberService,
	scopeService auth.ScopeService,
	observabilityServices *obsServices.ServiceRegistry,
	// Gateway services
	gatewayService gateway.GatewayService,
	routingService gateway.RoutingService,
	costService gateway.CostService,
	// Add other service dependencies as they're implemented
) *Handlers {
	return &Handlers{
		Health:        health.NewHandler(cfg, logger),
		Metrics:       metrics.NewHandler(cfg, logger),
		Auth:          authHandler.NewHandler(cfg, logger, authService, apiKeyService, userService, registrationService, oauthProvider),
		User:          userHandler.NewHandler(cfg, logger, userService, profileService, onboardingService, organizationService),
		Organization:  organizationHandler.NewHandler(cfg, logger, organizationService, memberService, projectService, invitationService, settingsService, userService, roleService),
		Project:       project.NewHandler(cfg, logger, projectService, organizationService, memberService),
		APIKey:        apikey.NewHandler(cfg, logger, apiKeyService),
		Analytics:     analytics.NewHandler(cfg, logger),
		Logs:          logs.NewHandler(cfg, logger),
		Billing:       billing.NewHandler(cfg, logger),
		AI:            ai.NewHandler(cfg, logger, gatewayService, routingService, costService),
		WebSocket:     websocket.NewHandler(cfg, logger),
		Admin:         admin.NewTokenAdminHandler(authService, blacklistedTokens, logger),
		RBAC:          rbac.NewHandler(cfg, logger, roleService, permissionService, organizationMemberService, scopeService),
		Observability: observability.NewHandler(cfg, logger, observabilityServices),
		OTLP:          observability.NewOTLPHandler(observabilityServices.StreamProducer, observabilityServices.DeduplicationService, observabilityServices.OTLPConverterService, logger),
	}
}