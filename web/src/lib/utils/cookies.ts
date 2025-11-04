/**
 * Cookie utility functions for client-side cookie management
 * Based on shadcn-admin patterns for consistent state persistence
 */

const DEFAULT_MAX_AGE = 60 * 60 * 24 * 7 // 7 days in seconds

/**
 * Get a cookie value by name
 * @param name - Cookie name
 * @returns Cookie value or undefined if not found
 */
export function getCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined

  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)

  if (parts.length === 2) {
    return parts.pop()?.split(';').shift()
  }

  return undefined
}

/**
 * Set a cookie with optional max age
 * @param name - Cookie name
 * @param value - Cookie value
 * @param maxAge - Max age in seconds (default: 7 days)
 */
export function setCookie(
  name: string,
  value: string,
  maxAge: number = DEFAULT_MAX_AGE
): void {
  if (typeof document === 'undefined') return

  // Add Secure flag for HTTPS connections
  const isSecure = typeof window !== 'undefined' && window.location.protocol === 'https:'
  const secureFlag = isSecure ? '; Secure' : ''

  document.cookie = `${name}=${value}; path=/; max-age=${maxAge}; SameSite=Lax${secureFlag}`
}

/**
 * Remove a cookie by setting max-age to 0
 * @param name - Cookie name
 */
export function removeCookie(name: string): void {
  if (typeof document === 'undefined') return

  document.cookie = `${name}=; path=/; max-age=0`
}
