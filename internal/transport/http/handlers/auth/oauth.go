package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/core/services/registration"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// OAuth initiate + callback handlers for Google and GitHub. The two
// providers share identical flow shapes — this file centralises them
// so the provider-specific paths are just a string dispatched to the
// shared logic, not a copy-paste duplication.
//
// Flow overview:
//
//  1. initiate*OAuth → generate state, store in Redis, 302 to provider
//  2. user authenticates on provider, provider redirects back to
//     /api/v1/auth/{google,github}/callback?code=…&state=…
//  3. callback validates state, exchanges code for a profile, and
//     either (a) redirects to the frontend with a login session id
//     for existing OAuth users, (b) redirects to the frontend signup
//     page with an OAuth session id for new users, or (c) redirects
//     to the sign-in page with an error query param on any failure.
//
// IMPORTANT: every callback exit path RETURNS A REDIRECT — never an
// error envelope. Client browsers don't handle JSON error bodies
// from a URL they were redirected to; errors become query params on
// the signin redirect and the frontend renders the message.

// ----- initiate-google-oauth ---------------------------------------

func (h *handler) initiateGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	h.initiateOAuth(w, r, "google", r.URL.Query().Get("invitation_token"))
}

// ----- initiate-github-oauth ---------------------------------------

func (h *handler) initiateGithubOAuth(w http.ResponseWriter, r *http.Request) {
	h.initiateOAuth(w, r, "github", r.URL.Query().Get("invitation_token"))
}

// initiateOAuth is the shared path both provider-specific initiators
// dispatch to.
func (h *handler) initiateOAuth(w http.ResponseWriter, r *http.Request, provider, invitationToken string) {
	var invitePtr *string
	if invitationToken != "" {
		invitePtr = &invitationToken
	}

	state, err := h.oauthProvider.GenerateState(r.Context(), invitePtr)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "oauth initiate: state generation failed",
			"provider", provider, "error", err)
		response.WriteError(w, err)
		return
	}

	authURL, err := h.oauthProvider.GetAuthorizationURL(provider, state)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "oauth initiate: authorization URL build failed",
			"provider", provider, "error", err)
		response.WriteError(w, err)
		return
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// ----- google-oauth-callback -------------------------------------

func (h *handler) googleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.oauthCallback(w, r, "google")
}

// ----- github-oauth-callback -------------------------------------

func (h *handler) githubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.oauthCallback(w, r, "github")
}

// oauthCallback is the shared path both provider-specific callbacks
// dispatch to. All exit paths redirect; errors become query params
// on the signin redirect.
func (h *handler) oauthCallback(w http.ResponseWriter, r *http.Request, provider string) {
	ctx := r.Context()
	frontend := h.cfg.Server.AppURL
	redirect := func(u string) {
		http.Redirect(w, r, u, http.StatusFound)
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		h.logger.WarnContext(ctx, "oauth callback: missing code or state", "provider", provider)
		redirect(frontend + "/auth/signin?error=oauth_failed")
		return
	}

	invitationToken, err := h.oauthProvider.ValidateState(ctx, state)
	if err != nil {
		h.logger.WarnContext(ctx, "oauth callback: invalid state",
			"provider", provider, "error", err)
		redirect(frontend + "/auth/signin?error=invalid_state")
		return
	}

	token, err := h.oauthProvider.ExchangeCode(ctx, provider, code)
	if err != nil {
		h.logger.ErrorContext(ctx, "oauth callback: code exchange failed",
			"provider", provider, "error", err)
		redirect(frontend + "/auth/signin?error=token_exchange_failed")
		return
	}

	userProfile, err := h.oauthProvider.GetUserProfile(ctx, provider, token)
	if err != nil {
		h.logger.ErrorContext(ctx, "oauth callback: profile fetch failed",
			"provider", provider, "error", err)
		redirect(frontend + "/auth/signin?error=profile_fetch_failed")
		return
	}

	existingUser, lookupErr := h.userSvc.GetUserByEmail(ctx, userProfile.Email)
	if lookupErr == nil && existingUser != nil {
		if existingUser.AuthMethod != "oauth" {
			h.logger.WarnContext(ctx, "oauth callback: password account attempted OAuth login",
				"email", userProfile.Email, "auth_method", existingUser.AuthMethod)
			redirect(frontend + "/auth/signin?error=account_exists_use_password")
			return
		}

		if existingUser.OAuthProvider == nil || *existingUser.OAuthProvider != provider {
			storedProvider := "a_different_provider"
			if existingUser.OAuthProvider != nil {
				storedProvider = *existingUser.OAuthProvider
			}
			h.logger.WarnContext(ctx, "oauth callback: wrong provider",
				"email", userProfile.Email, "stored", storedProvider, "attempted", provider)
			redirect(fmt.Sprintf("%s/auth/signin?error=use_%s", frontend, storedProvider))
			return
		}

		if existingUser.OAuthProviderID != nil && *existingUser.OAuthProviderID != userProfile.ProviderID {
			h.logger.ErrorContext(ctx, "oauth callback: provider ID mismatch (possible account takeover)",
				"email", userProfile.Email, "provider", provider)
			redirect(frontend + "/auth/signin?error=authentication_failed")
			return
		}

		loginTokens, tokErr := h.authSvc.GenerateTokensForUser(ctx, existingUser.ID)
		if tokErr != nil {
			h.logger.ErrorContext(ctx, "oauth callback: token generation failed", "error", tokErr)
			redirect(frontend + "/auth/signin?error=login_failed")
			return
		}

		sessionID, sessErr := h.authSvc.CreateLoginTokenSession(ctx,
			loginTokens.AccessToken, loginTokens.RefreshToken,
			loginTokens.ExpiresIn, existingUser.ID)
		if sessErr != nil {
			h.logger.ErrorContext(ctx, "oauth callback: login-session creation failed", "error", sessErr)
			redirect(frontend + "/auth/signin?error=session_failed")
			return
		}

		h.logger.InfoContext(ctx, "oauth callback: existing user login",
			"email", userProfile.Email, "provider", provider)
		redirect(fmt.Sprintf("%s/auth/callback?session=%s&type=login", frontend, sessionID))
		return
	}

	// New user — store profile in OAuth session, redirect to signup-
	// completion page.
	session := &authDomain.OAuthSession{
		Email:           userProfile.Email,
		FirstName:       userProfile.FirstName,
		LastName:        userProfile.LastName,
		Provider:        userProfile.Provider,
		ProviderID:      userProfile.ProviderID,
		InvitationToken: invitationToken,
	}
	sessionID, err := h.authSvc.CreateOAuthSession(ctx, session)
	if err != nil {
		h.logger.ErrorContext(ctx, "oauth callback: OAuth session creation failed", "error", err)
		redirect(frontend + "/auth/signin?error=session_creation_failed")
		return
	}

	h.logger.InfoContext(ctx, "oauth callback: new user signup session created",
		"email", userProfile.Email, "provider", provider)
	redirect(fmt.Sprintf("%s/auth/signup?session=%s", frontend, sessionID))
}

// ----- complete-oauth-signup ---------------------------------------

func (h *handler) completeOAuthSignup(w http.ResponseWriter, r *http.Request) {
	var body completeOAuthSignupBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	session, err := h.authSvc.GetOAuthSession(r.Context(), body.SessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	if session.Email == "" || session.FirstName == "" || session.LastName == "" ||
		session.Provider == "" || session.ProviderID == "" {
		h.logger.ErrorContext(r.Context(),
			"complete-oauth-signup: OAuth session missing required fields",
			"session_id", body.SessionID)
		response.WriteError(w, appErrors.NewValidationError(
			"Invalid OAuth session",
			"session missing one or more required profile fields",
		))
		return
	}

	oauthReq := &registration.OAuthRegistrationRequest{
		Email:            session.Email,
		FirstName:        session.FirstName,
		LastName:         session.LastName,
		Role:             body.Role,
		Provider:         session.Provider,
		ProviderID:       session.ProviderID,
		ReferralSource:   body.ReferralSource,
		OrganizationName: body.OrganizationName,
		InvitationToken:  session.InvitationToken,
	}

	regResp, err := h.regSvc.CompleteOAuthRegistration(r.Context(), oauthReq)
	if err != nil {
		h.logger.WarnContext(r.Context(), "complete-oauth-signup: registration failed",
			"email", session.Email, "error", err)
		response.WriteError(w, err)
		return
	}

	// Best-effort session cleanup.
	_ = h.authSvc.DeleteOAuthSession(r.Context(), body.SessionID)

	u, err := h.userSvc.GetUserByEmail(r.Context(), session.Email)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"complete-oauth-signup: user fetch failed after registration",
			"email", session.Email, "error", err)
		response.WriteError(w, appErrors.NewInternalError("Failed to complete authentication", err))
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"complete-oauth-signup: CSRF token generation failed", "error", err)
		response.WriteError(w, appErrors.NewInternalError("Authentication setup failed", err))
		return
	}

	setAuthCookies(w,
		regResp.LoginTokens.AccessToken,
		regResp.LoginTokens.RefreshToken,
		csrfToken,
		h.cfg.Server.CookieDomain,
	)

	expiresAt := time.Now().Add(time.Duration(regResp.LoginTokens.ExpiresIn) * time.Second)

	h.logger.InfoContext(r.Context(), "complete-oauth-signup successful",
		"email", session.Email, "provider", session.Provider)

	response.Created(w, completeOAuthSignupResponse{
		User:         u,
		Organization: regResp.Organization,
		ExpiresAt:    expiresAt.UnixMilli(),
		ExpiresIn:    regResp.LoginTokens.ExpiresIn * 1000,
	})
}

// ----- exchange-login-session --------------------------------------

func (h *handler) exchangeLoginSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "session_id")

	sessionData, err := h.authSvc.GetLoginTokenSession(r.Context(), sessionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	if sessionData.AccessToken == "" || sessionData.RefreshToken == "" ||
		sessionData.ExpiresIn <= 0 || sessionData.UserID == uuid.Nil {
		h.logger.ErrorContext(r.Context(),
			"exchange-login-session: session missing fields", "session_id", sessionID)
		response.WriteError(w, appErrors.NewValidationError(
			"Invalid session data",
			"login session missing required fields",
		))
		return
	}

	u, err := h.userSvc.GetUser(r.Context(), sessionData.UserID)
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"exchange-login-session: user fetch failed",
			"user_id", sessionData.UserID, "error", err)
		response.WriteError(w, appErrors.NewInternalError("Failed to complete authentication", err))
		return
	}

	csrfToken, err := generateCSRFToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(),
			"exchange-login-session: CSRF token generation failed", "error", err)
		response.WriteError(w, appErrors.NewInternalError("Authentication setup failed", err))
		return
	}

	setAuthCookies(w,
		sessionData.AccessToken,
		sessionData.RefreshToken,
		csrfToken,
		h.cfg.Server.CookieDomain,
	)

	expiresAt := time.Now().Add(time.Duration(sessionData.ExpiresIn) * time.Second)

	h.logger.InfoContext(r.Context(), "exchange-login-session successful",
		"user_id", sessionData.UserID)

	response.Success(w, loginResponse{
		User:      u,
		ExpiresAt: expiresAt.UnixMilli(),
		ExpiresIn: sessionData.ExpiresIn * 1000,
	})
}

