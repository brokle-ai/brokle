import { NextResponse, NextRequest } from 'next/server'
import { AUTH_CONSTANTS } from '@/lib/auth/constants'

/**
 * Production-ready middleware with precise route patterns and proper error handling
 * Uses simple cookie-based validation without JWT verification to avoid RSA key issues
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  
  try {
    // Get auth tokens from cookies with validation
    const accessTokenCookie = request.cookies.get(AUTH_CONSTANTS.ACCESS_TOKEN_BACKUP_COOKIE)
    const refreshTokenCookie = request.cookies.get(AUTH_CONSTANTS.REFRESH_TOKEN_COOKIE)
    
    // Simple token presence validation (no JWT verification in middleware)
    // JWT verification happens in the server-side auth utilities
    const hasValidTokens = !!(
      accessTokenCookie?.value && 
      refreshTokenCookie?.value &&
      accessTokenCookie.value !== 'null' && 
      refreshTokenCookie.value !== 'null' &&
      accessTokenCookie.value.length > 20 && // Reasonable minimum token length
      refreshTokenCookie.value.length > 20 &&
      // Basic JWT structure check (3 parts separated by dots)
      accessTokenCookie.value.split('.').length === 3
    )
    
    // Define precise route patterns
    const isDashboardRoute = pathname.startsWith('/dashboard')
    const isAuthRoute = pathname.startsWith('/auth/')
    
    // Organization routes pattern: /org-slug or /org-slug/project-slug
    const isOrganizationRoute = (
      !pathname.startsWith('/_next') && 
      !pathname.startsWith('/api') &&
      !pathname.startsWith('/auth') &&
      !pathname.startsWith('/dashboard') &&
      pathname !== '/' &&
      pathname.split('/').length >= 2 &&
      pathname.split('/')[1].length > 0 // Has org slug
    )
    
    const isPublicRoute = (
      pathname === '/' || 
      pathname.startsWith('/_next') || 
      pathname.startsWith('/api') ||
      pathname.startsWith('/favicon') ||
      pathname.startsWith('/robots') ||
      pathname.startsWith('/sitemap') ||
      ['/terms', '/privacy', '/about', '/contact'].includes(pathname)
    )
    
    // Skip middleware for public routes
    if (isPublicRoute) {
      return NextResponse.next()
    }
    
    // Protected route access without authentication
    if ((isDashboardRoute || isOrganizationRoute) && !hasValidTokens) {
      const loginUrl = new URL(AUTH_CONSTANTS.LOGIN_ROUTE, request.url)
      // Preserve the attempted URL for redirect after login
      loginUrl.searchParams.set('redirect', pathname)
      return NextResponse.redirect(loginUrl)
    }
    
    // Authenticated user trying to access auth pages
    if (isAuthRoute && hasValidTokens) {
      const dashboardUrl = new URL(AUTH_CONSTANTS.DASHBOARD_ROUTE, request.url)
      return NextResponse.redirect(dashboardUrl)
    }
    
    // Allow all other requests
    return NextResponse.next()
    
  } catch (error) {
    // Log error but don't block request - fail safely
    console.error('[Middleware] Error processing request:', error)
    return NextResponse.next()
  }
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}