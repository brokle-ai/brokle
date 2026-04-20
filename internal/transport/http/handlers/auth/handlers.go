// Package auth is the dashboard-plane authentication handler domain.
// It exposes login / logout / password-management / session-management
// operations as Huma v2 operations mounted on the apiAdmin surface.
//
// The package is the second vertical slice of the gin → huma handler
// conversion. It establishes the cookie-set / cookie-read / cookie-
// clear patterns every other dashboard domain reuses when it needs
// to touch auth cookies (which is rare — most domains just read the
// authenticated user via httpctx.MustGetUserID(ctx)).
//
// Operations currently registered (follow-up sessions add the rest):
//
//   - POST /api/v1/auth/login             — issue cookies, return user
//   - POST /api/v1/auth/logout            — blacklist JWT, clear cookies (authed)
//   - POST /api/v1/auth/refresh           — rotate cookies via refresh_token cookie
//   - POST /api/v1/auth/forgot-password   — email a reset token (user enumeration-safe)
//   - POST /api/v1/auth/reset-password    — exchange reset token for new password
//   - POST /api/v1/auth/change-password   — authenticated password change
//
// Not yet converted (follow-up sessions):
//
//   - signup / complete-oauth-signup / exchange-session
//   - /me (get current user) / profile get / profile update
//   - sessions list/get/revoke/revoke-all
//   - OAuth initiate / callback (Google, GitHub)
//   - POST /v1/auth/validate-key (SDK plane)
package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/internal/core/services/registration"
	"brokle/internal/transport/http/handlers/shared"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

// ----- login ---------------------------------------------------------

func (h *handler) login(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
	loginResp, err := h.authSvc.Login(ctx, &authDomain.LoginRequest{
		Email:      in.Body.Email,
		Password:   in.Body.Password,
		DeviceInfo: in.Body.DeviceInfo,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "login failed", "email", in.Body.Email, "error", err)
		return nil, err
	}

	// Fetch user data BEFORE building cookies so a user-fetch
	// failure aborts authentication — avoids a half-authenticated
	// session where the client has valid cookies but no user.
	u, err := h.userSvc.GetUserByEmail(ctx, in.Body.Email)
	if err != nil {
		h.logger.ErrorContext(ctx, "login: user fetch failed after successful credentials", "email", in.Body.Email, "error", err)
		return nil, appErrors.NewInternalError("Failed to complete authentication", err)
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(ctx, "login: CSRF token generation failed", "error", err)
		return nil, appErrors.NewInternalError("Authentication setup failed", err)
	}

	cookies := buildAuthCookies(
		loginResp.AccessToken,
		loginResp.RefreshToken,
		csrfToken,
		h.cfg.Server.CookieDomain,
	)

	h.logger.InfoContext(ctx, "login successful", "email", in.Body.Email)

	return &LoginOutput{
		SetCookie: cookies,
		Body:      expiryResponse(loginResp, u),
	}, nil
}

// ----- signup --------------------------------------------------------
//
// Fresh signup (creates a new organization) and invitation-based
// signup (joins an existing org via token) share this one endpoint.
// The service layer routes on the presence of InvitationToken. On
// success the handler issues the same three cookies as login so the
// user is authenticated immediately — no "check your email" step for
// the password flow.

func (h *handler) signup(ctx context.Context, in *SignupInput) (*SignupOutput, error) {
	if in.Body.InvitationToken == nil && in.Body.OrganizationName == nil {
		return nil, appErrors.NewValidationError(
			"Signup requires either organization_name or invitation_token",
			"Provide organization_name for a fresh signup, or invitation_token to join an existing organization",
		)
	}

	regReq := &registration.RegisterRequest{
		Email:            in.Body.Email,
		Password:         in.Body.Password,
		FirstName:        in.Body.FirstName,
		LastName:         in.Body.LastName,
		Role:             in.Body.Role,
		ReferralSource:   in.Body.ReferralSource,
		OrganizationName: in.Body.OrganizationName,
		InvitationToken:  in.Body.InvitationToken,
		IsOAuthUser:      false,
	}

	var (
		regResp *registration.RegistrationResponse
		err     error
	)
	if in.Body.InvitationToken != nil {
		regResp, err = h.regSvc.RegisterWithInvitation(ctx, regReq)
	} else {
		regResp, err = h.regSvc.RegisterWithOrganization(ctx, regReq)
	}
	if err != nil {
		h.logger.WarnContext(ctx, "signup failed", "email", in.Body.Email, "error", err)
		return nil, err
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(ctx, "signup: CSRF token generation failed", "error", err)
		return nil, appErrors.NewInternalError("Authentication setup failed", err)
	}

	cookies := buildAuthCookies(
		regResp.LoginTokens.AccessToken,
		regResp.LoginTokens.RefreshToken,
		csrfToken,
		h.cfg.Server.CookieDomain,
	)

	h.logger.InfoContext(ctx, "signup successful", "email", in.Body.Email, "user_id", regResp.User.ID)

	return &SignupOutput{
		SetCookie: cookies,
		Body:      expiryResponse(regResp.LoginTokens, regResp.User),
	}, nil
}

// ----- get-current-user (/me) ---------------------------------------
//
// Returns the authenticated user object plus access-token expiry
// metadata so the dashboard can schedule a proactive refresh before
// the next request bounces off a 401. This is the most-called
// dashboard endpoint.

func (h *handler) getCurrentUser(ctx context.Context, _ *struct{}) (*GetCurrentUserOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	claims := httpctx.MustGetTokenClaims(ctx)

	u, err := h.userSvc.GetUser(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "get-current-user: user fetch failed", "user_id", userID, "error", err)
		return nil, err
	}

	// Expiry math: JWT exp is Unix seconds; the response uses
	// milliseconds to match the format login/refresh emit, so the
	// dashboard's schedule code has one consistent timestamp type
	// to consume.
	expiresAtMs := claims.ExpiresAt * 1000
	expiresInMs := expiresAtMs - time.Now().UnixMilli()

	return &GetCurrentUserOutput{
		Body: loginResponse{
			User:      u,
			ExpiresAt: expiresAtMs,
			ExpiresIn: expiresInMs,
		},
	}, nil
}

// ----- get-profile ---------------------------------------------------
//
// Returns the authenticated user record. Distinct from
// get-current-user (/me) because /me includes access-token expiry
// metadata the dashboard uses for refresh scheduling, while
// /profile is the "user record for rendering a profile page"
// endpoint — no session metadata, just the persistent fields.

func (h *handler) getProfile(ctx context.Context, _ *struct{}) (*GetAuthProfileOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	u, err := h.userSvc.GetUser(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "get-profile: user fetch failed", "user_id", userID, "error", err)
		return nil, err
	}
	return &GetAuthProfileOutput{Body: u}, nil
}

// ----- update-profile -----------------------------------------------

func (h *handler) updateProfile(ctx context.Context, in *UpdateAuthProfileInput) (*UpdateAuthProfileOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	req := &user.UpdateUserProfileRequest{
		Bio:         in.Body.Bio,
		Location:    in.Body.Location,
		Website:     in.Body.Website,
		TwitterURL:  in.Body.TwitterURL,
		LinkedInURL: in.Body.LinkedInURL,
		GitHubURL:   in.Body.GitHubURL,
		AvatarURL:   in.Body.AvatarURL,
		Phone:       in.Body.Phone,
		Timezone:    in.Body.Timezone,
		Language:    in.Body.Language,
		Theme:       in.Body.Theme,

		EmailNotifications:    in.Body.EmailNotifications,
		PushNotifications:     in.Body.PushNotifications,
		MarketingEmails:       in.Body.MarketingEmails,
		WeeklyReports:         in.Body.WeeklyReports,
		MonthlyReports:        in.Body.MonthlyReports,
		SecurityAlerts:        in.Body.SecurityAlerts,
		BillingAlerts:         in.Body.BillingAlerts,
		UsageThresholdPercent: in.Body.UsageThresholdPercent,
	}

	profile, err := h.profileSvc.UpdateProfile(ctx, userID, req)
	if err != nil {
		h.logger.WarnContext(ctx, "update-profile failed", "user_id", userID, "error", err)
		return nil, err
	}

	return &UpdateAuthProfileOutput{Body: profile}, nil
}

// ----- sessions list -------------------------------------------------

func (h *handler) listSessions(ctx context.Context, _ *struct{}) (*ListAuthSessionsOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	sessions, err := h.sessionSvc.GetUserSessions(ctx, userID)
	if err != nil {
		h.logger.WarnContext(ctx, "list-sessions: fetch failed", "user_id", userID, "error", err)
		return nil, err
	}
	return &ListAuthSessionsOutput{Body: listAuthSessionsResponse{Sessions: sessions}}, nil
}

// ----- session get ---------------------------------------------------

func (h *handler) getSession(ctx context.Context, in *GetAuthSessionInput) (*GetAuthSessionOutput, error) {
	// The session service returns the session regardless of owner;
	// enforce owner-match at the handler layer so a compromised or
	// predicted UUID can't read another user's session.
	userID := httpctx.MustGetUserID(ctx)
	sess, err := h.sessionSvc.GetSession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		h.logger.WarnContext(ctx, "get-session: cross-user access attempt", "actor", userID, "session_id", in.SessionID, "owner", sess.UserID)
		return nil, appErrors.NewNotFoundError("Session")
	}
	return &GetAuthSessionOutput{Body: sess}, nil
}

// ----- session revoke -----------------------------------------------
//
// Revokes a single session by ID. Owner-enforcement mirrors
// get-session — a user can only revoke their own sessions.

func (h *handler) revokeSession(ctx context.Context, in *RevokeSessionInput) (*RevokeSessionOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	sess, err := h.sessionSvc.GetSession(ctx, in.SessionID)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		h.logger.WarnContext(ctx, "revoke-session: cross-user access attempt", "actor", userID, "session_id", in.SessionID, "owner", sess.UserID)
		return nil, appErrors.NewNotFoundError("Session")
	}
	if err := h.sessionSvc.RevokeSession(ctx, in.SessionID); err != nil {
		h.logger.WarnContext(ctx, "revoke-session: failed", "user_id", userID, "session_id", in.SessionID, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "session revoked", "user_id", userID, "session_id", in.SessionID)
	return &RevokeSessionOutput{Body: shared.MessageResponse{Message: "Session revoked successfully"}}, nil
}

// ----- sessions revoke-all ------------------------------------------
//
// Bulk revokes every session the authenticated user owns AND
// clears the caller's own auth cookies so the current browser
// drops its client-side session state alongside the server-side
// invalidation. GDPR/SOC2 "log me out everywhere" flow — the
// backend writes a user-wide timestamp blacklist internally so all
// tokens issued before this call are rejected even if they haven't
// hit the per-JTI blacklist yet.

func (h *handler) revokeAllSessions(ctx context.Context, _ *struct{}) (*RevokeAllSessionsOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	if err := h.authSvc.RevokeAllSessions(ctx, userID); err != nil {
		h.logger.WarnContext(ctx, "revoke-all-sessions: failed", "user_id", userID, "error", err)
		return nil, err
	}
	h.logger.InfoContext(ctx, "all sessions revoked", "user_id", userID)
	return &RevokeAllSessionsOutput{
		SetCookie: buildClearAuthCookies(h.cfg.Server.CookieDomain),
		Body:      shared.MessageResponse{Message: "All sessions revoked successfully"},
	}, nil
}

// ----- logout --------------------------------------------------------

func (h *handler) logout(ctx context.Context, _ *struct{}) (*LogoutOutput, error) {
	// Must* — the route registers RequireAuth upstream, so a
	// misconfiguration panics into the recoverer rather than
	// silently returning 401 (CLAUDE.md gotcha #6).
	claims := httpctx.MustGetTokenClaims(ctx)

	if err := h.authSvc.Logout(ctx, claims.JWTID, claims.UserID); err != nil {
		h.logger.ErrorContext(ctx, "logout: blacklist failed", "jti", claims.JWTID, "user_id", claims.UserID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "logout successful", "user_id", claims.UserID)

	return &LogoutOutput{
		SetCookie: buildClearAuthCookies(h.cfg.Server.CookieDomain),
		Body:      shared.MessageResponse{Message: "Logged out successfully"},
	}, nil
}

// ----- refresh-tokens -----------------------------------------------

func (h *handler) refresh(ctx context.Context, in *RefreshInput) (*RefreshOutput, error) {
	if in.RefreshToken == "" {
		// No cookie — surface as 401 and clear any stale cookies
		// the client might still hold. Huma doesn't support "return
		// an error AND set Set-Cookie", so the clear path lives on
		// the response-shaping side via the wrapped error type
		// below.
		h.logger.WarnContext(ctx, "refresh: missing refresh_token cookie")
		return nil, newAuthClearError(h.cfg.Server.CookieDomain,
			appErrors.NewUnauthorizedError("Refresh token not found"))
	}

	loginResp, err := h.authSvc.RefreshToken(ctx, &authDomain.RefreshTokenRequest{
		RefreshToken: in.RefreshToken,
	})
	if err != nil {
		h.logger.WarnContext(ctx, "refresh: token validation failed", "error", err)
		return nil, newAuthClearError(h.cfg.Server.CookieDomain,
			appErrors.NewUnauthorizedError("Refresh token invalid or expired"))
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(ctx, "refresh: CSRF token generation failed", "error", err)
		return nil, newAuthClearError(h.cfg.Server.CookieDomain,
			appErrors.NewInternalError("Token refresh setup failed", err))
	}

	cookies := buildAuthCookies(
		loginResp.AccessToken,
		loginResp.RefreshToken,
		csrfToken,
		h.cfg.Server.CookieDomain,
	)

	h.logger.InfoContext(ctx, "refresh successful")

	expiresAt := time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)
	return &RefreshOutput{
		SetCookie: cookies,
		Body: refreshResponse{
			ExpiresAt: expiresAt.UnixMilli(),
			ExpiresIn: loginResp.ExpiresIn * 1000,
		},
	}, nil
}

// ----- forgot-password ----------------------------------------------

func (h *handler) forgotPassword(ctx context.Context, in *ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	// Intentionally ignore the error — we return the same response
	// whether the email is registered or not to prevent account
	// enumeration via the forgot-password form. The authService
	// logs internally.
	if err := h.authSvc.ResetPassword(ctx, in.Body.Email); err != nil {
		h.logger.WarnContext(ctx, "forgot-password: reset initiation failed (swallowed)", "email", in.Body.Email, "error", err)
	}

	h.logger.InfoContext(ctx, "forgot-password requested", "email", in.Body.Email)

	return &ForgotPasswordOutput{
		Body: shared.MessageResponse{Message: "If the email exists, a password reset link has been sent"},
	}, nil
}

// ----- reset-password -----------------------------------------------
//
// Two-step flow: (1) forgot-password emails a token, (2) this
// endpoint exchanges the token + new password for an updated user
// credential row. No cookies set here — the user logs in again
// after reset so device-info / session tracking is preserved.

func (h *handler) resetPassword(ctx context.Context, in *ResetPasswordInput) (*ResetPasswordOutput, error) {
	// The current auth.Service interface exposes ResetPassword as
	// a single-argument (email) "send email" flow; the actual
	// token-exchange happens inside the service layer's password
	// reset completion routine. For the chi+huma migration we
	// surface it as NotImplemented until the service layer grows
	// the two-step API explicitly — the gin handler had the same
	// gap (it called h.authService.ResetPassword(ctx, req.Email)
	// with `req.Email` NEVER set — effectively a no-op). Log and
	// return 501 so the dashboard surfaces the gap instead of
	// silently "succeeding".
	h.logger.WarnContext(ctx, "reset-password endpoint hit — service-layer completion not yet implemented", "token_prefix", in.Body.Token[:min(len(in.Body.Token), 6)])
	return nil, appErrors.NewNotImplementedError("Password reset completion is not yet implemented in the service layer")
}

// ----- change-password ----------------------------------------------

func (h *handler) changePassword(ctx context.Context, in *ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID := httpctx.MustGetUserID(ctx)

	if err := h.userSvc.ChangePassword(ctx, userID, in.Body.CurrentPassword, in.Body.NewPassword); err != nil {
		h.logger.WarnContext(ctx, "change-password failed", "user_id", userID, "error", err)
		return nil, err
	}

	h.logger.InfoContext(ctx, "change-password successful", "user_id", userID)

	return &ChangePasswordOutput{
		Body: shared.MessageResponse{Message: "Password changed successfully"},
	}, nil
}

// ----- shared helpers ------------------------------------------------

// expiryResponse converts an authService.LoginResponse + user
// object into the loginResponse body shape. Extracted so login
// and any future endpoint that returns the same body (e.g.
// complete-oauth-signup when it lands) reuse one expiry-math
// site — drift between the two has bitten the gin handler
// twice (see git blame for the old expires_at / expires_in
// field tweaks).
func expiryResponse(loginResp *authDomain.LoginResponse, u any) loginResponse {
	expiresAt := time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)
	return loginResponse{
		User:      u,
		ExpiresAt: expiresAt.UnixMilli(),
		ExpiresIn: loginResp.ExpiresIn * 1000,
	}
}

func newAuthClearError(domain string, err *appErrors.AppError) *authClearError {
	return &authClearError{err: err, domain: domain}
}

func (e *authClearError) Error() string                 { return e.err.Error() }
func (e *authClearError) Unwrap() error                 { return e.err }
func (e *authClearError) GetStatus() int                { return e.err.HTTPStatus() }
func (e *authClearError) AppError() *appErrors.AppError { return e.err }
func (e *authClearError) Cookies() []http.Cookie {
	return buildClearAuthCookies(e.domain)
}

// Silence unused-import warnings for uuid — left in for the
// follow-up operations (session management, etc.) that will need
// it.
var _ = uuid.Nil
