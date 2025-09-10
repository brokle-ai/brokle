import { NextResponse, NextRequest } from 'next/server'

/**
 * Check if the given pathname is a public route that doesn't require authentication
 */
function isPublicRoute(pathname: string): boolean {
  const publicRoutes = [
    // Next.js internal routes
    '/_next',
    '/api',
    
    // Authentication routes
    '/auth/',
    
    // Static assets
    '/favicon.ico',
    '/robots.txt',
    '/terms',
    '/privacy',
  ]
  
  return publicRoutes.some(route => pathname.startsWith(route))
}

/**
 * Validate JWT token expiration without signature verification
 */
function isTokenValid(token: string): boolean {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return false
    
    // Decode JWT payload (base64url)
    const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')))
    const now = Math.floor(Date.now() / 1000)
    
    // Token is valid if it has exp and isn't expired (with 30s buffer)
    return payload.exp && payload.exp > (now + 30)
  } catch {
    return false
  }
}

/**
 * Simplified middleware: validate JWT, redirect if invalid
 */
export function middleware(request: NextRequest) {
  console.log('[MIDDLEWARE] Running for:', request.nextUrl.pathname)
  const { pathname } = request.nextUrl
  
  // Skip public routes
  if (isPublicRoute(pathname)) {
    return NextResponse.next()
  }
  
  // Get access token from cookie
  const accessToken = request.cookies.get('access_token')?.value
  
  // Protected route without valid token → redirect to login
  if (!accessToken || !isTokenValid(accessToken)) {
    // Keep minimal logging for debugging auth issues
    console.log('[Auth] Redirecting to login:', pathname, !accessToken ? 'no token' : 'invalid token')
    const loginUrl = new URL('/auth/signin', request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }
  
  return NextResponse.next()
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