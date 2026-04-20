package auth

import (
	"brokle/internal/config"
	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	authService "brokle/internal/core/services/auth"
	"brokle/internal/core/services/registration"
	"brokle/internal/transport/http/handlers/shared"
	appErrors "brokle/pkg/errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// Huma operation types for the auth package.

// handler bundles every service an auth operation needs. Package-
// private; the only public surface is RegisterPublicRoutes /
// RegisterProtectedRoutes (dashboard plane). SDK-plane operations
// have their own lighter-weight sdkHandler in sdk.go.
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

// LoginInput is the POST /api/v1/auth/login request.
type LoginInput struct {
	Body loginBody
}

type loginBody struct {
	Email      string         `json:"email" format:"email" doc:"User email address"`
	Password   string         `json:"password" minLength:"1" doc:"User password"`
	DeviceInfo map[string]any `json:"device_info,omitempty" doc:"Optional device metadata for session tracking"`
}

// LoginOutput emits three Set-Cookie headers alongside the JSON
// body. Huma serialises the SetCookie slice into multiple
// Set-Cookie header lines.
type LoginOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      loginResponse
}

// loginResponse is the on-the-wire success shape. Tokens are NOT
// in the body — they ride the Set-Cookie headers. The dashboard
// uses expires_at / expires_in to schedule its own proactive
// refresh call so it doesn't wait for the next request to bounce
// off a 401.
type loginResponse struct {
	User      any   `json:"user" doc:"Authenticated user object"`
	ExpiresAt int64 `json:"expires_at" doc:"Access-token expiry as Unix milliseconds"`
	ExpiresIn int64 `json:"expires_in" doc:"Access-token TTL in milliseconds"`
}

type SignupInput struct {
	Body signupBody
}

// signupBody validates the "at least one of org-name / invitation-
// token must be provided" rule in the handler because Huma's
// declarative constraints don't express mutual-or-either
// requirements. The gin handler had the same check at auth.go:173.
type signupBody struct {
	Email            string  `json:"email" format:"email" doc:"User email address"`
	Password         string  `json:"password" minLength:"8" doc:"Password (minimum 8 characters)"`
	FirstName        string  `json:"first_name" minLength:"1" maxLength:"100" doc:"User first name"`
	LastName         string  `json:"last_name" minLength:"1" maxLength:"100" doc:"User last name"`
	Role             string  `json:"role" enum:"engineer,product,designer,executive,other" doc:"Self-declared role"`
	OrganizationName *string `json:"organization_name,omitempty" doc:"Organization name — required for fresh signup, omit when InvitationToken is present"`
	InvitationToken  *string `json:"invitation_token,omitempty" doc:"Invitation token — when present, joins the existing organization instead of creating one"`
	ReferralSource   *string `json:"referral_source,omitempty" doc:"Optional referral source for product analytics"`
}

type SignupOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      loginResponse
}

type GetCurrentUserOutput struct {
	Body loginResponse
}

type GetAuthProfileOutput struct {
	Body any `json:"user" doc:"Authenticated user profile record"`
}

type UpdateAuthProfileInput struct {
	Body updateAuthProfileBody
}

// updateAuthProfileBody mirrors user.UpdateUserProfileRequest. All fields
// are pointer-typed because a missing field means "leave this
// field alone" rather than "clear this field" — only fields the
// client explicitly sets are updated. The body covers profile-
// record fields (bio, location, social URLs, preferences); name/
// email changes flow through a separate user-update endpoint
// (not yet converted).
type updateAuthProfileBody struct {
	Bio         *string `json:"bio,omitempty" maxLength:"500" doc:"Free-form profile bio"`
	Location    *string `json:"location,omitempty" maxLength:"100" doc:"City / country string"`
	Website     *string `json:"website,omitempty" format:"uri" doc:"Personal website URL"`
	TwitterURL  *string `json:"twitter_url,omitempty" format:"uri" doc:"Twitter / X profile URL"`
	LinkedInURL *string `json:"linkedin_url,omitempty" format:"uri" doc:"LinkedIn profile URL"`
	GitHubURL   *string `json:"github_url,omitempty" format:"uri" doc:"GitHub profile URL"`
	AvatarURL   *string `json:"avatar_url,omitempty" format:"uri" doc:"Avatar image URL"`
	Phone       *string `json:"phone,omitempty" maxLength:"50" doc:"Phone number"`
	Timezone    *string `json:"timezone,omitempty" doc:"IANA timezone identifier (e.g. 'America/New_York')"`
	Language    *string `json:"language,omitempty" minLength:"2" maxLength:"2" doc:"ISO 639-1 two-letter language code"`
	Theme       *string `json:"theme,omitempty" enum:"light,dark,auto" doc:"Dashboard colour theme preference"`

	EmailNotifications    *bool `json:"email_notifications,omitempty" doc:"Receive email notifications"`
	PushNotifications     *bool `json:"push_notifications,omitempty" doc:"Receive push notifications"`
	MarketingEmails       *bool `json:"marketing_emails,omitempty" doc:"Receive product-marketing emails"`
	WeeklyReports         *bool `json:"weekly_reports,omitempty" doc:"Receive weekly usage reports"`
	MonthlyReports        *bool `json:"monthly_reports,omitempty" doc:"Receive monthly usage reports"`
	SecurityAlerts        *bool `json:"security_alerts,omitempty" doc:"Receive security alerts"`
	BillingAlerts         *bool `json:"billing_alerts,omitempty" doc:"Receive billing alerts"`
	UsageThresholdPercent *int  `json:"usage_threshold_percent,omitempty" minimum:"0" maximum:"100" doc:"Percent-of-quota threshold that triggers a usage alert"`
}

type UpdateAuthProfileOutput struct {
	Body any `json:"profile" doc:"Updated user profile record"`
}

type ListAuthSessionsOutput struct {
	Body listAuthSessionsResponse
}

type listAuthSessionsResponse struct {
	Sessions []*authDomain.UserSession `json:"sessions" doc:"All sessions owned by the authenticated user"`
}

type GetAuthSessionInput struct {
	SessionID uuid.UUID `path:"session_id" doc:"Session identifier"`
}

type GetAuthSessionOutput struct {
	Body *authDomain.UserSession
}

type RevokeSessionInput struct {
	SessionID uuid.UUID `path:"session_id" doc:"Session identifier to revoke"`
}

type RevokeSessionOutput struct {
	Body shared.MessageResponse
}

type RevokeAllSessionsOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      shared.MessageResponse
}

// LogoutOutput returns three cleared cookies + a confirmation body.
type LogoutOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      shared.MessageResponse
}

// RefreshInput reads the refresh_token cookie Huma sees as a
// request header. `cookie:"refresh_token"` is Huma's supported
// input tag for single-cookie extraction.
type RefreshInput struct {
	RefreshToken string `cookie:"refresh_token" doc:"httpOnly refresh token issued by a prior login/signup/refresh"`
}

// RefreshOutput carries the rotated cookies + expiry metadata.
// User data is NOT returned — the dashboard keeps the user object
// from login/signup and only needs the new expiry values to
// schedule the next refresh. The /me endpoint (not yet converted)
// re-fetches the user when needed.
type RefreshOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      refreshResponse
}

type refreshResponse struct {
	ExpiresAt int64 `json:"expires_at" doc:"New access-token expiry as Unix milliseconds"`
	ExpiresIn int64 `json:"expires_in" doc:"New access-token TTL in milliseconds"`
}

type ForgotPasswordInput struct {
	Body forgotPasswordBody
}

type forgotPasswordBody struct {
	Email string `json:"email" format:"email" doc:"Email address for password reset"`
}

type ForgotPasswordOutput struct {
	Body shared.MessageResponse
}

type ResetPasswordInput struct {
	Body resetPasswordBody
}

type resetPasswordBody struct {
	Token       string `json:"token" minLength:"1" doc:"Password-reset token delivered via email"`
	NewPassword string `json:"new_password" minLength:"8" doc:"New password (minimum 8 characters)"`
}

type ResetPasswordOutput struct {
	Body shared.MessageResponse
}

type ChangePasswordInput struct {
	Body changePasswordBody
}

type changePasswordBody struct {
	CurrentPassword string `json:"current_password" minLength:"1" doc:"Current password — re-auth step"`
	NewPassword     string `json:"new_password" minLength:"8" doc:"New password (minimum 8 characters)"`
}

type ChangePasswordOutput struct {
	Body shared.MessageResponse
}

// authClearError wraps an AppError with a "clear cookies on the
// response" intent. Huma's default error pipeline only emits the
// JSON envelope — it doesn't know about our cookie-clear path.
// The huma.NewError override in internal/server/api_error.go
// detects *authClearError and adds the Set-Cookie clear headers
// to the outgoing response before writing the body.
//
// (The override lives in api_error.go rather than here so the
// server package owns the full error-rendering pipeline. This
// type is defined here because only the auth package raises it.)
type authClearError struct {
	err    *appErrors.AppError
	domain string
}

type InitiateOAuthInput struct {
	InvitationToken string `query:"invitation_token" required:"false" doc:"Optional invitation token. When present, the OAuth signup that follows joins an existing organization instead of creating one."`
}

type InitiateOAuthOutput struct {
	Status   int    `json:"-"`
	Location string `header:"Location"`
}

type OAuthCallbackInput struct {
	Code  string `query:"code" doc:"Authorization code returned by the OAuth provider"`
	State string `query:"state" doc:"CSRF state token echoed back by the provider"`
}

// OAuthCallbackOutput is the redirect-only shape every callback exit
// path emits. Huma's header:"Location" tag turns the Location field
// into the Set-Location header; Status is wired via DefaultStatus in
// the operation registration and the runtime override below.
type OAuthCallbackOutput struct {
	Status   int    `json:"-"`
	Location string `header:"Location"`
}

type CompleteOAuthSignupInput struct {
	Body completeOAuthSignupBody
}

type completeOAuthSignupBody struct {
	SessionID        string  `json:"session_id" minLength:"1" doc:"OAuth session ID returned by the callback redirect"`
	Role             string  `json:"role" enum:"engineer,product,designer,executive,other" doc:"Self-declared role"`
	OrganizationName *string `json:"organization_name,omitempty" doc:"Organization name — required for fresh signup, omitted when the session's invitation_token is present"`
	ReferralSource   *string `json:"referral_source,omitempty" doc:"Optional referral source for product analytics"`
}

type CompleteOAuthSignupOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      completeOAuthSignupResponse
}

// completeOAuthSignupResponse is the login-with-org shape. Matches
// the gin handler's responseData: user + organization + expiry in
// milliseconds.
type completeOAuthSignupResponse struct {
	User         any   `json:"user" doc:"Authenticated user object"`
	Organization any   `json:"organization" doc:"Organization created (fresh signup) or joined (invitation signup)"`
	ExpiresAt    int64 `json:"expires_at" doc:"Access-token expiry as Unix milliseconds"`
	ExpiresIn    int64 `json:"expires_in" doc:"Access-token TTL in milliseconds"`
}

type ExchangeLoginSessionInput struct {
	SessionID string `path:"session_id" doc:"One-time login session ID from the OAuth callback redirect"`
}

type ExchangeLoginSessionOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      loginResponse
}

// sdkHandler is a lightweight handler bundling only what SDK-plane
// auth operations need (API key validation). Distinct from the
// dashboard-plane `handler` so we don't pollute the SDK surface
// with services it shouldn't be able to reach.
type sdkHandler struct {
	apiKeySvc authDomain.APIKeyService
	logger    *slog.Logger
}

type ValidateAPIKeyInput struct {
	XAPIKey       string `header:"X-API-Key" required:"false" doc:"Canonical API-key header"`
	Authorization string `header:"Authorization" required:"false" doc:"Fallback for clients that cannot set custom headers; value must be 'Bearer <key>'"`
}

type ValidateAPIKeyOutput struct {
	Body validateAPIKeyResponse
}

// validateAPIKeyResponse mirrors authDomain.ValidateAPIKeyResponse
// so SDK consumers parse a stable shape independent of any future
// internal renames.
type validateAPIKeyResponse struct {
	AuthContext    *authDomain.AuthContext `json:"auth_context" doc:"Resolved user + API-key identifiers"`
	ProjectID      string                  `json:"project_id" doc:"Project the API key belongs to"`
	OrganizationID string                  `json:"organization_id" doc:"Organization that owns the project"`
}

// PublicDeps bundles every dependency the public dashboard auth
// operations need. Grouping them in a struct keeps the
// RegisterPublicRoutes signature stable as we add more operations
// that need additional services — and keeps the call site in
// server/routes.go readable.
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
// Carries the profile service in addition to the common set;
// public operations never touch profile data.
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
