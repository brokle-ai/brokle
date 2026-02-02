/**
 * Application page routes (client-side navigation)
 *
 * Use these constants instead of hardcoded path strings.
 * This provides a single source of truth for all routes.
 */

export const ROUTES = {
  // Auth routes (public)
  SIGNIN: '/signin',
  SIGNUP: '/signup',
  FORGOT_PASSWORD: '/forgot-password',
  VERIFY_EMAIL: '/verify-email',
  CALLBACK: '/callback',
  ACCEPT_INVITE: '/accept-invite',

  // Dashboard routes (protected)
  HOME: '/',
  PROJECTS: '/projects',

  // Error routes
  UNAUTHORIZED: '/401',
  NOT_FOUND: '/404',
} as const

export type AppRoute = (typeof ROUTES)[keyof typeof ROUTES]

/**
 * Helper to build signin URL with redirect parameter
 */
export function signinWithRedirect(redirectTo: string): string {
  return `${ROUTES.SIGNIN}?redirect=${encodeURIComponent(redirectTo)}`
}

/**
 * Helper to build signin URL with session status
 */
export function signinWithStatus(
  status: 'expired' | 'ended' | 'logout_success' | 'logout_error'
): string {
  const params: Record<string, string> = {
    expired: 'session=expired',
    ended: 'session=ended',
    logout_success: 'logout=success',
    logout_error: 'logout=error',
  }
  return `${ROUTES.SIGNIN}?${params[status]}`
}
