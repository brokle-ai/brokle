import { NextRequest, NextResponse } from 'next/server'
import { jwtVerify, importSPKI } from 'jose'

// Configuration
const JWT_PUBLIC_KEY = process.env.JWT_PUBLIC_KEY || ''
const AUTH_ROUTES = ['/auth/signin', '/auth/signup', '/auth/forgot-password', '/auth/verify-email']
const PUBLIC_ROUTES = [...AUTH_ROUTES, '/api', '/_next', '/favicon.ico', '/images', '/public']
const DASHBOARD_ROUTE = '/dashboard'

interface JWTPayload {
  sub: string // user id
  email: string
  organizationId: string
  role: string
  exp: number
  iat: number
}

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  
  // Skip middleware for public routes
  if (isPublicRoute(pathname)) {
    return NextResponse.next()
  }

  // Handle auth routes - redirect if already authenticated
  if (isAuthRoute(pathname)) {
    const isAuthenticated = await checkAuthentication(request)
    
    if (isAuthenticated) {
      // User is authenticated, redirect to dashboard
      return NextResponse.redirect(new URL(DASHBOARD_ROUTE, request.url))
    }
    
    // User not authenticated, allow access to auth routes
    return NextResponse.next()
  }

  // Handle protected routes - require authentication
  const authResult = await checkAuthentication(request)
  
  if (!authResult.isAuthenticated) {
    // User not authenticated, redirect to signin with return URL
    const signInUrl = new URL('/auth/signin', request.url)
    signInUrl.searchParams.set('redirect', pathname)
    
    return NextResponse.redirect(signInUrl)
  }

  // User is authenticated, add user context to headers for SSR
  const response = NextResponse.next()
  
  if (authResult.payload) {
    response.headers.set('x-user-id', authResult.payload.sub)
    response.headers.set('x-user-email', authResult.payload.email)
    response.headers.set('x-organization-id', authResult.payload.organizationId)
    response.headers.set('x-user-role', authResult.payload.role)
  }

  return response
}

async function checkAuthentication(request: NextRequest): Promise<{
  isAuthenticated: boolean
  payload?: JWTPayload
}> {
  try {
    // Try to get token from multiple sources
    const token = getTokenFromRequest(request)
    
    if (!token) {
      return { isAuthenticated: false }
    }

    if (!JWT_PUBLIC_KEY) {
      console.error('[Middleware] JWT_PUBLIC_KEY not configured')
      return { isAuthenticated: false }
    }

    // Import RSA public key for RS256 verification
    const publicKey = await importSPKI(JWT_PUBLIC_KEY, 'RS256')
    
    // Verify JWT token with RSA public key
    const { payload } = await jwtVerify(token, publicKey, {
      issuer: 'brokle-platform' // Match the issuer from auth service
    })
    
    // Type assertion for our payload structure
    const jwtPayload = payload as unknown as JWTPayload
    
    // Check if token is expired (jwtVerify should handle this, but double-check)
    const now = Math.floor(Date.now() / 1000)
    if (jwtPayload.exp && jwtPayload.exp < now) {
      return { isAuthenticated: false }
    }

    return {
      isAuthenticated: true,
      payload: jwtPayload,
    }
  } catch (error) {
    // Token verification failed
    console.error('[Middleware] Token verification failed:', error)
    return { isAuthenticated: false }
  }
}

function getTokenFromRequest(request: NextRequest): string | null {
  // 1. Try Authorization header
  const authHeader = request.headers.get('authorization')
  if (authHeader?.startsWith('Bearer ')) {
    return authHeader.substring(7)
  }

  // 2. Try backup cookie (set by client for SSR)
  const backupTokenCookie = request.cookies.get('access_token_backup')
  if (backupTokenCookie?.value) {
    return backupTokenCookie.value
  }

  // 3. Try custom header (for API requests)
  const customAuthHeader = request.headers.get('x-auth-token')
  if (customAuthHeader) {
    return customAuthHeader
  }

  return null
}

function isPublicRoute(pathname: string): boolean {
  return PUBLIC_ROUTES.some(route => {
    if (route.endsWith('*')) {
      return pathname.startsWith(route.slice(0, -1))
    }
    return pathname.startsWith(route)
  })
}

function isAuthRoute(pathname: string): boolean {
  return AUTH_ROUTES.some(route => pathname.startsWith(route))
}

// Configure which routes this middleware should run on
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder
     */
    '/((?!api|_next/static|_next/image|favicon.ico|public).*)',
  ],
}