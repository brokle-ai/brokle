import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { cache } from 'react'
import { AUTH_CONSTANTS } from './constants'
import type { User } from '@/types/auth'

export interface ServerSession {
  user: User | null
  isAuthenticated: boolean
  expiresAt: number | null
}

/**
 * Get server-side session from cookies/headers (cached per request)
 * This prevents auth flashing by validating tokens server-side
 */
export const getServerSession = cache(async (): Promise<ServerSession> => {
  try {
    const cookieStore = await cookies()
    
    // Try to get tokens from cookies (backup) or headers
    const accessTokenCookie = cookieStore.get(AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE)
    const refreshTokenCookie = cookieStore.get(AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE)
    
    if (!accessTokenCookie?.value) {
      return {
        user: null,
        isAuthenticated: false,
        expiresAt: null
      }
    }

    // Basic token validation (decode JWT to check expiry)
    const tokenData = parseJWTPayload(accessTokenCookie.value)
    if (!tokenData || !tokenData.exp || Date.now() / 1000 > tokenData.exp) {
      return {
        user: null,
        isAuthenticated: false,
        expiresAt: null
      }
    }

    // Extract user info from token
    const user: User = {
      id: tokenData.user_id || tokenData.sub,
      email: tokenData.email || '',
      firstName: tokenData.first_name || '',
      lastName: tokenData.last_name || '',
      name: tokenData.name || `${tokenData.first_name || ''} ${tokenData.last_name || ''}`.trim() || tokenData.email || '',
      role: tokenData.role || 'user',
      organizationId: tokenData.org_id || '',
      projects: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      isEmailVerified: tokenData.email_verified || false
    }

    return {
      user,
      isAuthenticated: true,
      expiresAt: tokenData.exp * 1000 // Convert to milliseconds
    }
  } catch (error) {
    console.error('[ServerAuth] Failed to validate server session:', error)
    return {
      user: null,
      isAuthenticated: false,
      expiresAt: null
    }
  }
})

/**
 * Validate server-side tokens without API calls
 * Returns true if tokens exist and are not expired
 */
export async function validateServerTokens(): Promise<boolean> {
  const session = await getServerSession()
  return session.isAuthenticated
}

/**
 * Redirect to login if not authenticated (server-side)
 * Use this in page components to prevent auth flash
 */
export async function requireAuth(): Promise<ServerSession> {
  const session = await getServerSession()
  
  if (!session.isAuthenticated) {
    redirect(AUTH_CONSTANTS.LOGIN_ROUTE)
  }
  
  return session
}

/**
 * Redirect to dashboard if already authenticated (server-side)
 * Use this on login/signup pages
 */
export async function requireGuest(): Promise<void> {
  const session = await getServerSession()
  
  if (session.isAuthenticated) {
    redirect(AUTH_CONSTANTS.DASHBOARD_ROUTE)
  }
}

/**
 * Get user from server-side session or null if not authenticated
 */
export async function getServerUser(): Promise<User | null> {
  const session = await getServerSession()
  return session.user
}

/**
 * Parse JWT payload without verification (for server-side basic validation)
 * Only use for extracting non-sensitive data like expiry time
 */
function parseJWTPayload(token: string): any {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    
    const payload = parts[1]
    const decoded = Buffer.from(payload, 'base64url').toString('utf-8')
    return JSON.parse(decoded)
  } catch {
    return null
  }
}

/**
 * Check if request has valid authentication headers/cookies
 * Useful for API routes and middleware
 */
export async function isRequestAuthenticated(): Promise<boolean> {
  return validateServerTokens()
}