import { AUTH_CONSTANTS } from './constants'
import { SecureStorage } from './storage'
import { getSessionSync } from './session-sync'
import type { AuthTokens } from '@/types/auth'

export interface TokenRefreshCallback {
  (): Promise<AuthTokens>
}

export interface TokenStorageCallback {
  (): void | Promise<void>
}

export class TokenManager {
  private accessToken: string | null = null
  private expiresAt: number | null = null
  private refreshCallback: TokenRefreshCallback | null = null
  private storageCallback: TokenStorageCallback | null = null
  private refreshPromise: Promise<string | null> | null = null
  private refreshTimer: NodeJS.Timeout | null = null
  private sessionSync = getSessionSync()

  constructor() {
    this.loadFromStorage()
    this.setupAutoRefresh()
    this.setupSessionSync()
  }

  // Public methods
  /**
   * Set new authentication tokens and store them securely
   * Also sets up automatic refresh and broadcasts to other tabs
   * Calls storage callback for coordination with navigation
   */
  async setTokens(tokens: AuthTokens): Promise<void> {
    this.accessToken = tokens.accessToken
    this.expiresAt = Date.now() + (tokens.expiresIn * 1000)
    
    // Store in secure storage
    SecureStorage.setTokens(tokens)
    
    // Setup auto-refresh
    this.setupAutoRefresh()
    
    // Broadcast to other tabs
    this.sessionSync.broadcastTokenRefresh()
    
    // Notify storage completion for navigation coordination
    if (this.storageCallback) {
      try {
        await this.storageCallback()
      } catch (error) {
        console.error('[TokenManager] Storage callback failed:', error)
      }
    }
  }

  setRefreshCallback(callback: TokenRefreshCallback): void {
    this.refreshCallback = callback
  }

  /**
   * Set callback to be executed after tokens are stored
   * Useful for coordinating navigation with cookie availability
   */
  setStorageCallback(callback: TokenStorageCallback): void {
    this.storageCallback = callback
  }

  async getValidAccessToken(): Promise<string | null> {
    // Return cached token if still valid
    if (this.accessToken && this.isTokenValid()) {
      return this.accessToken
    }

    // Try to refresh token
    return this.refreshAccessToken()
  }

  async refreshAccessToken(): Promise<string | null> {
    // Prevent multiple simultaneous refresh attempts
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    this.refreshPromise = this.performTokenRefresh()

    try {
      const token = await this.refreshPromise
      return token
    } finally {
      this.refreshPromise = null
    }
  }

  clearTokens(): void {
    this.accessToken = null
    this.expiresAt = null
    
    SecureStorage.clearTokens()
    
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer)
      this.refreshTimer = null
    }
  }

  isAuthenticated(): boolean {
    return this.accessToken !== null && this.isTokenValid()
  }

  getTokenTimeLeft(): number {
    if (!this.expiresAt) return 0
    return Math.max(0, this.expiresAt - Date.now())
  }

  // Private methods
  private loadFromStorage(): void {
    const storedToken = SecureStorage.getAccessToken()
    const storedExpiresAt = SecureStorage.getExpiresAt()
    
    this.accessToken = storedToken
    this.expiresAt = storedExpiresAt
  }

  private isTokenValid(): boolean {
    if (!this.expiresAt) return false
    return Date.now() < this.expiresAt - AUTH_CONSTANTS.TOKEN_REFRESH_THRESHOLD
  }

  private async performTokenRefresh(): Promise<string | null> {
    if (!this.refreshCallback) {
      console.error('[TokenManager] No refresh callback registered')
      return null
    }

    const refreshToken = SecureStorage.getRefreshToken()
    if (!refreshToken) {
      console.warn('[TokenManager] No refresh token available')
      return null
    }

    try {
      const tokens = await this.refreshCallback()
      await this.setTokens(tokens)
      return this.accessToken
    } catch (error) {
      console.error('[TokenManager] Token refresh failed:', error)
      
      // Clear invalid tokens
      this.clearTokens()
      
      // Broadcast session expired
      this.sessionSync.broadcastSessionExpired()
      
      return null
    }
  }

  private setupAutoRefresh(): void {
    // Clear existing timer
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer)
    }

    if (!this.expiresAt || !this.refreshCallback) {
      return
    }

    const timeUntilRefresh = Math.max(
      0,
      this.expiresAt - Date.now() - AUTH_CONSTANTS.TOKEN_REFRESH_THRESHOLD
    )

    this.refreshTimer = setTimeout(() => {
      this.refreshAccessToken().catch(error => {
        console.error('[TokenManager] Auto-refresh failed:', error)
      })
    }, timeUntilRefresh)
  }

  private setupSessionSync(): void {
    // Listen for token refresh events from other tabs
    this.sessionSync.on('TOKEN_REFRESH', () => {
      this.loadFromStorage()
      this.setupAutoRefresh()
    })

    // Listen for logout events from other tabs
    this.sessionSync.on('LOGOUT', () => {
      this.clearTokens()
    })

    // Listen for session expired events from other tabs
    this.sessionSync.on('SESSION_EXPIRED', () => {
      this.clearTokens()
    })
  }

  // Debug methods
  debug(): void {
    if (process.env.NODE_ENV !== 'development') return

    console.group('🔐 TokenManager Debug')
    console.log('Access Token (memory):', this.accessToken?.substring(0, 20) + '...')
    console.log('Refresh Token (storage):', SecureStorage.getRefreshToken()?.substring(0, 20) + '...')
    console.log('Expires At:', this.expiresAt ? new Date(this.expiresAt).toISOString() : 'null')
    console.log('Is Valid:', this.isTokenValid())
    console.log('Time Left:', Math.round(this.getTokenTimeLeft() / 1000), 'seconds')
    console.log('Is Authenticated:', this.isAuthenticated())
    console.log('Has Refresh Callback:', !!this.refreshCallback)
    console.groupEnd()

    SecureStorage.debugTokens()
  }

  // Cleanup
  destroy(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer)
      this.refreshTimer = null
    }
    
    this.refreshCallback = null
    this.refreshPromise = null
  }
}

// Global singleton instance
let tokenManager: TokenManager | null = null

export function getTokenManager(): TokenManager {
  if (!tokenManager) {
    tokenManager = new TokenManager()
  }
  return tokenManager
}

// Cleanup on page unload
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', () => {
    if (tokenManager) {
      tokenManager.destroy()
      tokenManager = null
    }
  })
}