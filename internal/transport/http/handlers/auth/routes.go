package auth

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"brokle/internal/config"
	authService "brokle/internal/core/services/auth"
	"brokle/internal/core/services/registration"
	userService "brokle/internal/core/services/user"
)

// RegisterPublicRoutes mounts every unauthenticated auth operation on
// r. Expected mount context: the pre-auth dashboard chi group
// (LimitByIP). These endpoints live under /api/v1/auth/* and carry no
// SDK semantics.
//
// Concrete paths (r.Post / r.Get) rather than an r.Route wrapper:
// RegisterProtectedRoutes also registers under /api/v1/auth/* but from
// a sibling chi.Group with a different middleware stack. Two sibling
// groups share the parent mux's routing tree, so two r.Route calls at
// the same pattern panic at boot on chi's Mount invariant
// (chi/v5/mux.go Mount → findPattern). Direct path registrations
// write concrete leaves to the trie and never collide, even across
// sibling groups. See CLAUDE.md Transport gotcha #35.
func RegisterPublicRoutes(
	r chi.Router,
	authSvc *authService.AuthService,
	userSvc *userService.UserService,
	regSvc *registration.RegistrationService,
	sessionSvc *authService.SessionService,
	oauthProvider *authService.OAuthProviderService,
	cfg *config.Config,
	logger *slog.Logger,
) {
	h := &handler{
		authSvc:       authSvc,
		userSvc:       userSvc,
		regSvc:        regSvc,
		sessionSvc:    sessionSvc,
		oauthProvider: oauthProvider,
		cfg:           cfg,
		logger:        logger,
	}

	r.Post("/api/v1/auth/login", h.login)
	r.Post("/api/v1/auth/signup", h.signup)
	r.Post("/api/v1/auth/refresh", h.refresh)
	r.Post("/api/v1/auth/forgot-password", h.forgotPassword)
	r.Post("/api/v1/auth/reset-password", h.resetPassword)

	r.Get("/api/v1/auth/google", h.initiateGoogleOAuth)
	r.Get("/api/v1/auth/google/callback", h.googleOAuthCallback)
	r.Get("/api/v1/auth/github", h.initiateGithubOAuth)
	r.Get("/api/v1/auth/github/callback", h.githubOAuthCallback)
	r.Post("/api/v1/auth/complete-oauth-signup", h.completeOAuthSignup)
	r.Post("/api/v1/auth/exchange-session/{session_id}", h.exchangeLoginSession)
}

// RegisterProtectedRoutes mounts every auth operation that requires a
// valid session. Expected mount context: the authed dashboard chi
// group (RequireAuth + LimitByUser).
//
// Concrete paths not r.Route — see RegisterPublicRoutes docstring.
func RegisterProtectedRoutes(
	r chi.Router,
	authSvc *authService.AuthService,
	userSvc *userService.UserService,
	profileSvc *userService.ProfileService,
	regSvc *registration.RegistrationService,
	sessionSvc *authService.SessionService,
	oauthProvider *authService.OAuthProviderService,
	cfg *config.Config,
	logger *slog.Logger,
) {
	h := &handler{
		authSvc:       authSvc,
		userSvc:       userSvc,
		profileSvc:    profileSvc,
		regSvc:        regSvc,
		sessionSvc:    sessionSvc,
		oauthProvider: oauthProvider,
		cfg:           cfg,
		logger:        logger,
	}

	r.Get("/api/v1/auth/me", h.getCurrentUser)
	r.Post("/api/v1/auth/logout", h.logout)
	r.Post("/api/v1/auth/change-password", h.changePassword)
	r.Get("/api/v1/auth/profile", h.getProfile)
	r.Patch("/api/v1/auth/profile", h.updateProfile)
	r.Get("/api/v1/auth/sessions", h.listSessions)
	r.Get("/api/v1/auth/sessions/{session_id}", h.getSession)
	r.Post("/api/v1/auth/sessions/{session_id}/revoke", h.revokeSession)
	r.Post("/api/v1/auth/sessions/revoke-all", h.revokeAllSessions)
}
