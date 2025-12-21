package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

// isDevelopment returns true if APP_ENV is empty or "development".
func isDevelopment() bool {
	env := os.Getenv("APP_ENV")
	return env == "" || env == "development"
}

// generateCSRFToken generates a cryptographically secure CSRF token
// Returns error instead of panicking to allow proper error handling
func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// setAuthCookies sets httpOnly authentication cookies with proper security flags
// Uses http.SetCookie for full control over cookie attributes including SameSite
func setAuthCookies(w http.ResponseWriter, access, refresh, csrf string) {
	isSecure := !isDevelopment()

	// Access token cookie (15 minutes)
	// - HttpOnly: prevents XSS attacks
	// - Secure: requires HTTPS (disabled in dev)
	// - SameSite=Lax: allows cross-origin navigation (email links, SSO redirects)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Refresh token cookie (7 days)
	// - Path restricted to refresh endpoint only for security
	// - SameSite=Strict for enhanced security (no cross-site requests needed)
	// - Longer expiry for session persistence
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/api/v1/auth/refresh", // Only sent to refresh endpoint
		MaxAge:   604800,                 // 7 days
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode, // Strict (more secure, path-restricted anyway)
	})

	// CSRF token cookie (15 minutes)
	// - NOT HttpOnly: must be readable by JavaScript to set X-CSRF-Token header
	// - Used for double-submit cookie CSRF protection
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrf,
		Path:     "/",
		MaxAge:   900,   // 15 minutes
		HttpOnly: false, // CRITICAL: Must be readable by JS
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookies clears all authentication cookies
// CRITICAL: Paths must match the original cookie paths exactly for proper clearing
func clearAuthCookies(w http.ResponseWriter) {
	isSecure := !isDevelopment()

	// Clear access token (path: /)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Clear refresh token (path: /api/v1/auth/refresh)
	// MUST match original path and SameSite for proper clearing
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode, // Must match original
	})

	// Clear CSRF token (path: /)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
