// Package auth is the dashboard-plane authentication handler domain.
// Exposes login / logout / password-management / session-management /
// OAuth flows as chi routes. Cookie set/clear helpers in cookies.go;
// OAuth flow in oauth.go; SDK-plane validate-key in sdk.go.
package auth

import (
	"net/http"
	"time"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/internal/core/services/registration"
	"brokle/internal/transport/http/handlers/shared"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// ----- login ---------------------------------------------------------

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	loginResp, err := h.authSvc.Login(r.Context(), &authDomain.LoginRequest{
		Email:      body.Email,
		Password:   body.Password,
		DeviceInfo: body.DeviceInfo,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "login failed", "email", body.Email, "error", err)
		response.WriteError(w, err)
		return
	}

	u, err := h.userSvc.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"login: user fetch failed after successful credentials",
			"email", body.Email, "error", err)
		response.WriteError(w, appErrors.Internal("Failed to complete authentication", err))
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "login: CSRF token generation failed", "error", err)
		response.WriteError(w, appErrors.Internal("Authentication setup failed", err))
		return
	}

	setAuthCookies(w, loginResp.AccessToken, loginResp.RefreshToken, csrfToken, h.cfg.Server.CookieDomain)
	h.logger.InfoContext(r.Context(), "login successful", "email", body.Email)
	response.Success(w, expiryResponse(loginResp, u))
}

// ----- signup --------------------------------------------------------

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var body signupBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if body.InvitationToken == nil && body.OrganizationName == nil {
		response.WriteError(w, appErrors.BadRequest(
			"signup requires either organization_name or invitation_token: provide organization_name for a fresh signup, or invitation_token to join an existing organization",
		))
		return
	}

	regReq := &registration.RegisterRequest{
		Email:            body.Email,
		Password:         body.Password,
		FirstName:        body.FirstName,
		LastName:         body.LastName,
		Role:             body.Role,
		ReferralSource:   body.ReferralSource,
		OrganizationName: body.OrganizationName,
		InvitationToken:  body.InvitationToken,
		IsOAuthUser:      false,
	}

	var (
		regResp *registration.RegistrationResponse
		err     error
	)
	if body.InvitationToken != nil {
		regResp, err = h.regSvc.RegisterWithInvitation(r.Context(), regReq)
	} else {
		regResp, err = h.regSvc.RegisterWithOrganization(r.Context(), regReq)
	}
	if err != nil {
		h.logger.WarnContext(r.Context(), "signup failed", "email", body.Email, "error", err)
		response.WriteError(w, err)
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "signup: CSRF token generation failed", "error", err)
		response.WriteError(w, appErrors.Internal("Authentication setup failed", err))
		return
	}

	setAuthCookies(w, regResp.LoginTokens.AccessToken, regResp.LoginTokens.RefreshToken, csrfToken, h.cfg.Server.CookieDomain)
	h.logger.InfoContext(r.Context(), "signup successful",
		"email", body.Email, "user_id", regResp.User.ID)
	response.Created(w, expiryResponse(regResp.LoginTokens, regResp.User))
}

// ----- get-current-user (/me) ---------------------------------------

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	claims := httpctx.MustGetTokenClaims(r.Context())

	u, err := h.userSvc.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "get-current-user: user fetch failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}

	expiresAtMs := claims.ExpiresAt * 1000
	expiresInMs := expiresAtMs - time.Now().UnixMilli()

	response.Success(w, loginResponse{
		User:      u,
		ExpiresAt: expiresAtMs,
		ExpiresIn: expiresInMs,
	})
}

// ----- get-profile ---------------------------------------------------

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	u, err := h.userSvc.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "get-profile: user fetch failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, u)
}

// ----- update-profile -----------------------------------------------

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body updateAuthProfileBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	req := &user.UpdateUserProfileRequest{
		Bio:         body.Bio,
		Location:    body.Location,
		Website:     body.Website,
		TwitterURL:  body.TwitterURL,
		LinkedInURL: body.LinkedInURL,
		GitHubURL:   body.GitHubURL,
		AvatarURL:   body.AvatarURL,
		Phone:       body.Phone,
		Timezone:    body.Timezone,
		Language:    body.Language,
		Theme:       body.Theme,

		EmailNotifications:    body.EmailNotifications,
		PushNotifications:     body.PushNotifications,
		MarketingEmails:       body.MarketingEmails,
		WeeklyReports:         body.WeeklyReports,
		MonthlyReports:        body.MonthlyReports,
		SecurityAlerts:        body.SecurityAlerts,
		BillingAlerts:         body.BillingAlerts,
		UsageThresholdPercent: body.UsageThresholdPercent,
	}

	profile, err := h.profileSvc.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		h.logger.WarnContext(r.Context(), "update-profile failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, profile)
}

// ----- sessions list -------------------------------------------------

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	sessions, err := h.sessionSvc.GetUserSessions(r.Context(), userID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "list-sessions: fetch failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	response.Success(w, listAuthSessionsResponse{Sessions: sessions})
}

// ----- session get ---------------------------------------------------

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	sessionID, err := request.URLParamUUID(r, "session_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sess, err := h.sessionSvc.GetSession(r.Context(), sessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if sess.UserID != userID {
		h.logger.WarnContext(r.Context(), "get-session: cross-user access attempt",
			"actor", userID, "session_id", sessionID, "owner", sess.UserID)
		response.WriteError(w, appErrors.NotFound("session"))
		return
	}
	response.Success(w, sess)
}

// ----- session revoke -----------------------------------------------

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	sessionID, err := request.URLParamUUID(r, "session_id")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	sess, err := h.sessionSvc.GetSession(r.Context(), sessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if sess.UserID != userID {
		h.logger.WarnContext(r.Context(), "revoke-session: cross-user access attempt",
			"actor", userID, "session_id", sessionID, "owner", sess.UserID)
		response.WriteError(w, appErrors.NotFound("session"))
		return
	}
	if err := h.sessionSvc.RevokeSession(r.Context(), sessionID); err != nil {
		h.logger.WarnContext(r.Context(), "revoke-session: failed",
			"user_id", userID, "session_id", sessionID, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "session revoked",
		"user_id", userID, "session_id", sessionID)
	response.Success(w, shared.MessageResponse{Message: "Session revoked successfully"})
}

// ----- sessions revoke-all ------------------------------------------

func (h *Handler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	if err := h.authSvc.RevokeAllSessions(r.Context(), userID); err != nil {
		h.logger.WarnContext(r.Context(), "revoke-all-sessions: failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "all sessions revoked", "user_id", userID)
	clearAuthCookies(w, h.cfg.Server.CookieDomain)
	response.Success(w, shared.MessageResponse{Message: "All sessions revoked successfully"})
}

// ----- logout --------------------------------------------------------

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := httpctx.MustGetTokenClaims(r.Context())

	if err := h.authSvc.Logout(r.Context(), claims.JWTID, claims.UserID); err != nil {
		h.logger.ErrorContext(r.Context(), "logout: blacklist failed",
			"jti", claims.JWTID, "user_id", claims.UserID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "logout successful", "user_id", claims.UserID)
	clearAuthCookies(w, h.cfg.Server.CookieDomain)
	response.Success(w, shared.MessageResponse{Message: "Logged out successfully"})
}

// ----- refresh-tokens -----------------------------------------------

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieNameRefresh)
	if err != nil || cookie.Value == "" {
		h.logger.WarnContext(r.Context(), "refresh: missing refresh_token cookie")
		clearAuthCookies(w, h.cfg.Server.CookieDomain)
		response.WriteError(w, appErrors.Unauthenticated("Refresh token not found"))
		return
	}

	loginResp, err := h.authSvc.RefreshToken(r.Context(), &authDomain.RefreshTokenRequest{
		RefreshToken: cookie.Value,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "refresh: token validation failed", "error", err)
		clearAuthCookies(w, h.cfg.Server.CookieDomain)
		response.WriteError(w, appErrors.Unauthenticated("Refresh token invalid or expired"))
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "refresh: CSRF token generation failed", "error", err)
		clearAuthCookies(w, h.cfg.Server.CookieDomain)
		response.WriteError(w, appErrors.Internal("Token refresh setup failed", err))
		return
	}

	setAuthCookies(w, loginResp.AccessToken, loginResp.RefreshToken, csrfToken, h.cfg.Server.CookieDomain)
	h.logger.InfoContext(r.Context(), "refresh successful")

	expiresAt := time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)
	response.Success(w, refreshResponse{
		ExpiresAt: expiresAt.UnixMilli(),
		ExpiresIn: loginResp.ExpiresIn * 1000,
	})
}

// ----- forgot-password ----------------------------------------------

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body forgotPasswordBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	// Intentionally ignore the error — we return the same response
	// whether the email is registered or not to prevent account
	// enumeration via the forgot-password form. The authService logs
	// internally.
	if err := h.authSvc.ResetPassword(r.Context(), body.Email); err != nil {
		h.logger.WarnContext(r.Context(),
			"forgot-password: reset initiation failed (swallowed)",
			"email", body.Email, "error", err)
	}

	h.logger.InfoContext(r.Context(), "forgot-password requested", "email", body.Email)

	response.Success(w, shared.MessageResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ----- reset-password -----------------------------------------------

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body resetPasswordBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	prefixLen := len(body.Token)
	if prefixLen > 6 {
		prefixLen = 6
	}
	h.logger.WarnContext(r.Context(),
		"reset-password endpoint hit — service-layer completion not yet implemented",
		"token_prefix", body.Token[:prefixLen])
	response.WriteError(w, appErrors.NotImplemented("", 
		"Password reset completion is not yet implemented in the service layer"))
}

// ----- change-password ----------------------------------------------

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())

	var body changePasswordBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	if err := h.userSvc.ChangePassword(r.Context(), userID, body.CurrentPassword, body.NewPassword); err != nil {
		h.logger.WarnContext(r.Context(), "change-password failed",
			"user_id", userID, "error", err)
		response.WriteError(w, err)
		return
	}

	h.logger.InfoContext(r.Context(), "change-password successful", "user_id", userID)
	response.Success(w, shared.MessageResponse{Message: "Password changed successfully"})
}

// ----- shared helpers ------------------------------------------------

// expiryResponse converts an authService.LoginResponse + user object
// into the loginResponse body shape. Extracted so login and any future
// endpoint that returns the same body reuse one expiry-math site.
func expiryResponse(loginResp *authDomain.LoginResponse, u any) loginResponse {
	expiresAt := time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)
	return loginResponse{
		User:      u,
		ExpiresAt: expiresAt.UnixMilli(),
		ExpiresIn: loginResp.ExpiresIn * 1000,
	}
}
