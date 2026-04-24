package auth

import (
	"github.com/go-chi/chi/v5"
)

// RegisterPublicRoutes mounts every unauthenticated auth operation on
// r. Expected mount context: the pre-auth dashboard chi group
// (LimitByIP). These endpoints live under /api/v1/auth/* and carry no
// SDK semantics.
func RegisterPublicRoutes(r chi.Router, d PublicDeps) {
	h := &handler{
		authSvc:       d.Auth,
		userSvc:       d.User,
		regSvc:        d.Registration,
		sessionSvc:    d.Session,
		oauthProvider: d.OAuthProvider,
		cfg:           d.Config,
		logger:        d.Logger,
	}

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/signup", h.signup)
		r.Post("/refresh", h.refresh)
		r.Post("/forgot-password", h.forgotPassword)
		r.Post("/reset-password", h.resetPassword)

		r.Get("/google", h.initiateGoogleOAuth)
		r.Get("/google/callback", h.googleOAuthCallback)
		r.Get("/github", h.initiateGithubOAuth)
		r.Get("/github/callback", h.githubOAuthCallback)
		r.Post("/complete-oauth-signup", h.completeOAuthSignup)
		r.Post("/exchange-session/{session_id}", h.exchangeLoginSession)
	})
}

// RegisterProtectedRoutes mounts every auth operation that requires a
// valid session. Expected mount context: the authed dashboard chi
// group (RequireAuth + LimitByUser).
func RegisterProtectedRoutes(r chi.Router, d ProtectedDeps) {
	h := &handler{
		authSvc:       d.Auth,
		userSvc:       d.User,
		profileSvc:    d.Profile,
		regSvc:        d.Registration,
		sessionSvc:    d.Session,
		oauthProvider: d.OAuthProvider,
		cfg:           d.Config,
		logger:        d.Logger,
	}

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Get("/me", h.getCurrentUser)
		r.Post("/logout", h.logout)
		r.Post("/change-password", h.changePassword)
		r.Get("/profile", h.getProfile)
		r.Patch("/profile", h.updateProfile)
		r.Get("/sessions", h.listSessions)
		r.Get("/sessions/{session_id}", h.getSession)
		r.Post("/sessions/{session_id}/revoke", h.revokeSession)
		r.Post("/sessions/revoke-all", h.revokeAllSessions)
	})
}
