package auth

import (
	"log/slog"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	authService "brokle/internal/core/services/auth"
	"brokle/internal/core/services/registration"
	userService "brokle/internal/core/services/user"
)

// Handler bundles every service a dashboard-plane auth operation needs.
// SDK-plane operations have their own lighter-weight SDKHandler.
type Handler struct {
	authSvc       *authService.AuthService
	userSvc       *userService.UserService
	profileSvc    *userService.ProfileService
	regSvc        *registration.RegistrationService
	sessionSvc    *authService.SessionService
	oauthProvider *authService.OAuthProviderService
	apiKeySvc     *authService.APIKeyService
	cfg           *config.Config
	logger        *slog.Logger
}

// New constructs a Handler with all services any auth Register* function
// might need (Public, Protected, SDK).
func New(
	authSvc *authService.AuthService,
	userSvc *userService.UserService,
	profileSvc *userService.ProfileService,
	regSvc *registration.RegistrationService,
	sessionSvc *authService.SessionService,
	oauthProvider *authService.OAuthProviderService,
	apiKeySvc *authService.APIKeyService,
	cfg *config.Config,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		authSvc:       authSvc,
		userSvc:       userSvc,
		profileSvc:    profileSvc,
		regSvc:        regSvc,
		sessionSvc:    sessionSvc,
		oauthProvider: oauthProvider,
		apiKeySvc:     apiKeySvc,
		cfg:           cfg,
		logger:        logger,
	}
}

// SDKHandler is a lightweight handler bundling only what SDK-plane
// auth operations need (API key validation). Distinct from the
// dashboard-plane Handler so we don't pollute the SDK surface with
// services it shouldn't be able to reach.
type SDKHandler struct {
	apiKeySvc *authService.APIKeyService
	logger    *slog.Logger
}

// NewSDK constructs an SDKHandler.
func NewSDK(apiKeySvc *authService.APIKeyService, logger *slog.Logger) *SDKHandler {
	return &SDKHandler{apiKeySvc: apiKeySvc, logger: logger}
}

// ---- request bodies -------------------------------------------------

type loginBody struct {
	Email      string         `json:"email"              validate:"required,email"`
	Password   string         `json:"password"           validate:"required,min=1"`
	DeviceInfo map[string]any `json:"device_info,omitempty"`
}

type signupBody struct {
	Email            string  `json:"email"                        validate:"required,email"`
	Password         string  `json:"password"                     validate:"required,min=8"`
	FirstName        string  `json:"first_name"                   validate:"required,min=1,max=100"`
	LastName         string  `json:"last_name"                    validate:"required,min=1,max=100"`
	Role             string  `json:"role"                         validate:"required,oneof=engineer product designer executive other"`
	OrganizationName *string `json:"organization_name,omitempty"`
	InvitationToken  *string `json:"invitation_token,omitempty"`
	ReferralSource   *string `json:"referral_source,omitempty"`
}

type updateAuthProfileBody struct {
	Bio         *string `json:"bio,omitempty"           validate:"omitempty,max=500"`
	Location    *string `json:"location,omitempty"      validate:"omitempty,max=100"`
	Website     *string `json:"website,omitempty"       validate:"omitempty,url"`
	TwitterURL  *string `json:"twitter_url,omitempty"   validate:"omitempty,url"`
	LinkedInURL *string `json:"linkedin_url,omitempty"  validate:"omitempty,url"`
	GitHubURL   *string `json:"github_url,omitempty"    validate:"omitempty,url"`
	AvatarURL   *string `json:"avatar_url,omitempty"    validate:"omitempty,url"`
	Phone       *string `json:"phone,omitempty"         validate:"omitempty,max=50"`
	Timezone    *string `json:"timezone,omitempty"`
	Language    *string `json:"language,omitempty"      validate:"omitempty,min=2,max=2"`
	Theme       *string `json:"theme,omitempty"         validate:"omitempty,oneof=light dark auto"`

	EmailNotifications    *bool `json:"email_notifications,omitempty"`
	PushNotifications     *bool `json:"push_notifications,omitempty"`
	MarketingEmails       *bool `json:"marketing_emails,omitempty"`
	WeeklyReports         *bool `json:"weekly_reports,omitempty"`
	MonthlyReports        *bool `json:"monthly_reports,omitempty"`
	SecurityAlerts        *bool `json:"security_alerts,omitempty"`
	BillingAlerts         *bool `json:"billing_alerts,omitempty"`
	UsageThresholdPercent *int  `json:"usage_threshold_percent,omitempty" validate:"omitempty,min=0,max=100"`
}

type forgotPasswordBody struct {
	Email string `json:"email" validate:"required,email"`
}

type resetPasswordBody struct {
	Token       string `json:"token"        validate:"required,min=1"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type changePasswordBody struct {
	CurrentPassword string `json:"current_password" validate:"required,min=1"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type completeOAuthSignupBody struct {
	SessionID        string  `json:"session_id"                   validate:"required,min=1"`
	Role             string  `json:"role"                         validate:"required,oneof=engineer product designer executive other"`
	OrganizationName *string `json:"organization_name,omitempty"`
	ReferralSource   *string `json:"referral_source,omitempty"`
}

// ---- response bodies ------------------------------------------------

type loginResponse struct {
	User      any   `json:"user"`
	ExpiresAt int64 `json:"expires_at"`
	ExpiresIn int64 `json:"expires_in"`
}

type refreshResponse struct {
	ExpiresAt int64 `json:"expires_at"`
	ExpiresIn int64 `json:"expires_in"`
}

type completeOAuthSignupResponse struct {
	User         any   `json:"user"`
	Organization any   `json:"organization"`
	ExpiresAt    int64 `json:"expires_at"`
	ExpiresIn    int64 `json:"expires_in"`
}

type listAuthSessionsResponse struct {
	Sessions []*authDomain.UserSession `json:"sessions"`
}

type validateAPIKeyResponse struct {
	AuthContext    *authDomain.AuthContext `json:"auth_context"`
	ProjectID      string                  `json:"project_id"`
	OrganizationID string                  `json:"organization_id"`
}
