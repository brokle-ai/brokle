package auth

import (
	"log/slog"

	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	"brokle/internal/core/services/registration"
)

// handler bundles every service a dashboard-plane auth operation needs.
// Package-private; the only public surface is RegisterPublicRoutes /
// RegisterProtectedRoutes. SDK-plane operations have their own lighter-
// weight sdkHandler.
type handler struct {
	authSvc       authDomain.AuthService
	userSvc       user.UserService
	profileSvc    user.ProfileService
	regSvc        registration.RegistrationService
	sessionSvc    authDomain.SessionService
	oauthProvider *authService.OAuthProviderService
	cfg           *config.Config
	logger        *slog.Logger
}

// PublicDeps bundles every dependency the public dashboard auth
// operations need.
type PublicDeps struct {
	Auth          authDomain.AuthService
	User          user.UserService
	Registration  registration.RegistrationService
	Session       authDomain.SessionService
	OAuthProvider *authService.OAuthProviderService
	Config        *config.Config
	Logger        *slog.Logger
}

// ProtectedDeps mirrors PublicDeps for the authenticated routes.
// Carries the profile service in addition to the common set; public
// operations never touch profile data.
type ProtectedDeps struct {
	Auth          authDomain.AuthService
	User          user.UserService
	Profile       user.ProfileService
	Registration  registration.RegistrationService
	Session       authDomain.SessionService
	OAuthProvider *authService.OAuthProviderService
	Config        *config.Config
	Logger        *slog.Logger
}

// sdkHandler is a lightweight handler bundling only what SDK-plane
// auth operations need (API key validation). Distinct from the
// dashboard-plane `handler` so we don't pollute the SDK surface with
// services it shouldn't be able to reach.
type sdkHandler struct {
	apiKeySvc authDomain.APIKeyService
	logger    *slog.Logger
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
