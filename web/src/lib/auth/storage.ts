import { AUTH_CONSTANTS } from './constants'
import type { AuthTokens, User } from '@/types/auth'

interface SecureStorageOptions {
  httpOnly?: boolean
  secure?: boolean
  sameSite?: 'strict' | 'lax' | 'none'
  maxAge?: number
  path?: string
}

export class SecureStorage {
  private static isServer = typeof window === 'undefined'

  // Token storage methods
  /**
   * Store authentication tokens in localStorage and secure cookies
   * Uses backup cookies for server-side access and localStorage for client-side persistence
   */
  static setTokens(tokens: AuthTokens): void {
    if (this.isServer) return

    const expiresAt = Date.now() + (tokens.expiresIn * 1000)
    
    // Store tokens in localStorage for persistence
    try {
      localStorage.setItem(AUTH_CONSTANTS.ACCESS_TOKEN_KEY, tokens.accessToken)
      localStorage.setItem(AUTH_CONSTANTS.REFRESH_TOKEN_KEY, tokens.refreshToken)
      localStorage.setItem(AUTH_CONSTANTS.EXPIRES_AT_KEY, String(expiresAt))
    } catch (error) {
      console.error('[SecureStorage] Failed to store tokens in localStorage:', error)
    }

    // Set secure cookie for refresh token (will be httpOnly in production)
    this.setCookie(AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE, tokens.refreshToken, {
      httpOnly: false, // Will be true in production middleware
      secure: location.protocol === 'https:',
      sameSite: 'strict',
      maxAge: AUTH_CONSTANTS.COOKIE_MAX_AGE,
      path: '/',
    })

    // Set backup cookie for access token (for SSR/middleware)
    this.setCookie(AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE, tokens.accessToken, {
      httpOnly: false,
      secure: location.protocol === 'https:',
      sameSite: 'strict',
      maxAge: AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_MAX_AGE,
      path: '/',
    })

    // Debug logging in development
    if (process.env.NODE_ENV === 'development') {
      console.log('[SecureStorage] Tokens stored, cookies set:', {
        refreshCookieName: AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE,
        accessCookieName: AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE,
        allCookies: document.cookie
      })
    }
  }

  static getAccessToken(): string | null {
    if (this.isServer) return null
    return localStorage.getItem(AUTH_CONSTANTS.ACCESS_TOKEN_KEY)
  }

  static getExpiresAt(): number | null {
    if (this.isServer) return null
    const expiresAt = localStorage.getItem(AUTH_CONSTANTS.EXPIRES_AT_KEY)
    return expiresAt ? parseInt(expiresAt, 10) : null
  }

  static getRefreshToken(): string | null {
    if (this.isServer) return null
    
    // Try localStorage first for faster client-side access
    const localStorageToken = localStorage.getItem(AUTH_CONSTANTS.REFRESH_TOKEN_KEY)
    if (localStorageToken) {
      return localStorageToken
    }
    
    // Fallback to cookie
    return this.getCookie(AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE)
  }

  static clearTokens(): void {
    if (this.isServer) return

    // Clear localStorage
    localStorage.removeItem(AUTH_CONSTANTS.ACCESS_TOKEN_KEY)
    localStorage.removeItem(AUTH_CONSTANTS.REFRESH_TOKEN_KEY)
    localStorage.removeItem(AUTH_CONSTANTS.EXPIRES_AT_KEY)
    localStorage.removeItem(AUTH_CONSTANTS.USER_KEY)

    // Clear cookies
    this.removeCookie(AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE)
    this.removeCookie(AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE)
  }

  // User storage methods
  static setUser(user: User): void {
    if (this.isServer) return
    localStorage.setItem(AUTH_CONSTANTS.USER_KEY, JSON.stringify(user))
  }

  static getUser(): User | null {
    if (this.isServer) return null
    
    try {
      const userStr = localStorage.getItem(AUTH_CONSTANTS.USER_KEY)
      return userStr ? JSON.parse(userStr) : null
    } catch (error) {
      console.error('Failed to parse user from storage:', error)
      return null
    }
  }

  static clearUser(): void {
    if (this.isServer) return
    localStorage.removeItem(AUTH_CONSTANTS.USER_KEY)
  }

  // Cookie utilities
  private static setCookie(
    name: string, 
    value: string, 
    options: SecureStorageOptions = {}
  ): void {
    if (this.isServer) return

    const {
      httpOnly = false,
      secure = false,
      sameSite = 'strict',
      maxAge,
      path = '/',
    } = options

    let cookieString = `${name}=${encodeURIComponent(value)}`
    
    if (maxAge) {
      cookieString += `; Max-Age=${maxAge}`
    }
    
    cookieString += `; Path=${path}`
    cookieString += `; SameSite=${sameSite}`
    
    if (secure) {
      cookieString += '; Secure'
    }
    
    if (httpOnly) {
      cookieString += '; HttpOnly'
    }

    document.cookie = cookieString
  }

  private static getCookie(name: string): string | null {
    if (this.isServer) return null

    const value = `; ${document.cookie}`
    const parts = value.split(`; ${name}=`)
    
    if (parts.length === 2) {
      const cookieValue = parts.pop()?.split(';').shift()
      return cookieValue ? decodeURIComponent(cookieValue) : null
    }
    
    return null
  }

  private static removeCookie(name: string, path: string = '/'): void {
    if (this.isServer) return
    document.cookie = `${name}=; Max-Age=0; Path=${path}; SameSite=strict`
  }

  // Utility methods
  static isTokenExpired(expiresAt: number | null): boolean {
    if (!expiresAt) return true
    return Date.now() >= expiresAt - AUTH_CONSTANTS.TOKEN_REFRESH_THRESHOLD
  }

  static getTokenTimeLeft(expiresAt: number | null): number {
    if (!expiresAt) return 0
    return Math.max(0, expiresAt - Date.now())
  }

  static clearAllAuthData(): void {
    this.clearTokens()
    this.clearUser()
  }

  // Server-side token reading methods
  static getServerSideTokens(): { accessToken: string | null; refreshToken: string | null } {
    if (typeof window !== 'undefined') {
      // Client-side fallback - shouldn't be used on server
      return {
        accessToken: this.getAccessToken(),
        refreshToken: this.getRefreshToken()
      }
    }

    // Server-side: read from cookies headers (would be implemented with next/headers)
    try {
      // This would need to be implemented with proper server context
      // For now, return null on server-side
      return {
        accessToken: null,
        refreshToken: null
      }
    } catch {
      return {
        accessToken: null,
        refreshToken: null
      }
    }
  }

  // Development helpers
  static debugTokens(): void {
    if (process.env.NODE_ENV !== 'development') return

    console.group('🔐 Auth Storage Debug')
    console.log('Access Token (localStorage):', this.getAccessToken()?.substring(0, 20) + '...')
    console.log('Refresh Token (localStorage):', localStorage.getItem(AUTH_CONSTANTS.REFRESH_TOKEN_KEY)?.substring(0, 20) + '...')
    console.log('Refresh Token (cookie):', this.getCookie(AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE)?.substring(0, 20) + '...')
    console.log('Expires At:', new Date(this.getExpiresAt() || 0).toISOString())
    console.log('User:', this.getUser())
    console.log('Is Expired:', this.isTokenExpired(this.getExpiresAt()))
    console.groupEnd()
  }
}