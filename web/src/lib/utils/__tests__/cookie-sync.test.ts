/**
 * Test suite for cookie synchronization utilities
 */

import { waitForCookieSync, verifyCookieAuth, verifyTokenCookieStructure } from '../cookie-sync'
import { AUTH_CONSTANTS } from '@/lib/auth/constants'

// Mock document.cookie
Object.defineProperty(document, 'cookie', {
  writable: true,
  value: ''
})

describe('Cookie Sync Utilities', () => {
  beforeEach(() => {
    // Clear cookies before each test
    document.cookie = ''
  })

  describe('verifyCookieAuth', () => {
    it('should return false when no cookies are present', () => {
      expect(verifyCookieAuth()).toBe(false)
    })

    it('should return false when only access token cookie is present', () => {
      document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=test-token`
      expect(verifyCookieAuth()).toBe(false)
    })

    it('should return false when only refresh token cookie is present', () => {
      document.cookie = `${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=test-refresh-token`
      expect(verifyCookieAuth()).toBe(false)
    })

    it('should return true when both valid tokens are present', () => {
      document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=test-token; ${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=test-refresh-token`
      expect(verifyCookieAuth()).toBe(true)
    })

    it('should return false when tokens are null strings', () => {
      document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=null; ${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=null`
      expect(verifyCookieAuth()).toBe(false)
    })
  })

  describe('verifyTokenCookieStructure', () => {
    it('should return false when no cookies are present', () => {
      expect(verifyTokenCookieStructure()).toBe(false)
    })

    it('should return false when access token is not a valid JWT structure', () => {
      document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=invalid-token; ${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=test-refresh-token`
      expect(verifyTokenCookieStructure()).toBe(false)
    })

    it('should return true when access token has valid JWT structure (3 parts)', () => {
      document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=header.payload.signature; ${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=test-refresh-token`
      expect(verifyTokenCookieStructure()).toBe(true)
    })
  })

  describe('waitForCookieSync', () => {
    it('should resolve immediately when cookies are already present', async () => {
      document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=test-token; ${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=test-refresh-token`
      
      const result = await waitForCookieSync({ timeout: 1000 })
      expect(result).toBe(true)
    })

    it('should timeout when cookies never appear', async () => {
      const result = await waitForCookieSync({ timeout: 100 })
      expect(result).toBe(false)
    })

    it('should resolve when cookies appear during polling', async () => {
      // Simulate cookies appearing after a delay
      setTimeout(() => {
        document.cookie = `${AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE}=test-token; ${AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE}=test-refresh-token`
      }, 50)

      const result = await waitForCookieSync({ timeout: 200, pollInterval: 25 })
      expect(result).toBe(true)
    })
  })
})