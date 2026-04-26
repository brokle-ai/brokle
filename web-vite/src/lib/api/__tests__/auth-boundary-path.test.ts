import { describe, it, expect } from 'vitest'
import { isAuthBoundaryPath } from '../client'

// `isAuthBoundaryPath` decides whether a 401 from a given URL
// triggers refresh-retry. It must return TRUE only for the
// unauthenticated /api/v1/auth/* endpoints; protected auth endpoints
// (logout, profile, change-password, sessions/*) must return FALSE so
// their 401s rotate via the refresh cookie.
//
// The expected sets mirror the backend's split in
// `internal/transport/http/handlers/auth/routes.go` —
// RegisterPublicRoutes vs RegisterProtectedRoutes.

const PUBLIC_PATHS = [
  '/api/v1/auth/login',
  '/api/v1/auth/signup',
  '/api/v1/auth/refresh',
  '/api/v1/auth/forgot-password',
  '/api/v1/auth/reset-password',
  '/api/v1/auth/google',
  '/api/v1/auth/google/callback',
  '/api/v1/auth/github',
  '/api/v1/auth/github/callback',
  '/api/v1/auth/complete-oauth-signup',
  '/api/v1/auth/exchange-session/abc123',
  '/api/v1/auth/exchange-session/uuid-with-dashes',
]

const PROTECTED_AUTH_PATHS = [
  '/api/v1/auth/me',
  '/api/v1/auth/logout',
  '/api/v1/auth/change-password',
  '/api/v1/auth/profile',
  '/api/v1/auth/sessions',
  '/api/v1/auth/sessions/abc123',
  '/api/v1/auth/sessions/abc123/revoke',
  '/api/v1/auth/sessions/revoke-all',
]

const RESOURCE_PATHS = [
  '/api/v1/users/me',
  '/api/v1/organizations',
  '/api/v1/projects',
]

describe('isAuthBoundaryPath', () => {
  describe('with empty API_URL (relative same-origin)', () => {
    it.each(PUBLIC_PATHS)('returns true for public path %s', (path) => {
      expect(isAuthBoundaryPath(path)).toBe(true)
    })

    it.each(PROTECTED_AUTH_PATHS)(
      'returns false for protected auth path %s (must allow refresh-retry)',
      (path) => {
        expect(isAuthBoundaryPath(path)).toBe(false)
      },
    )

    it.each(RESOURCE_PATHS)(
      'returns false for resource path %s',
      (path) => {
        expect(isAuthBoundaryPath(path)).toBe(false)
      },
    )
  })

  describe('with absolute http URLs', () => {
    it('handles full origin + public path', () => {
      expect(
        isAuthBoundaryPath('http://localhost:8080/api/v1/auth/login'),
      ).toBe(true)
    })

    it('handles full origin + protected path', () => {
      expect(
        isAuthBoundaryPath('http://localhost:8080/api/v1/auth/profile'),
      ).toBe(false)
    })
  })

  describe('rejects unrelated paths that share a prefix', () => {
    it('does not match /api/v1/auth/exchange-sessions (plural, no trailing slash)', () => {
      expect(
        isAuthBoundaryPath('/api/v1/auth/exchange-sessions'),
      ).toBe(false)
    })
  })
})
