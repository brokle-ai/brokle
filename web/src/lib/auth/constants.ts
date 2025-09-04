export const AUTH_CONSTANTS = {
  // Storage keys
  ACCESS_TOKEN_KEY: 'brokle_access_token',
  REFRESH_TOKEN_KEY: 'brokle_refresh_token',
  EXPIRES_AT_KEY: 'brokle_expires_at',
  USER_KEY: 'brokle_user',
  
  // Cookie names
  ACCESS_TOKEN_BACKUP_COOKIE: 'access_token_backup',
  REFRESH_TOKEN_COOKIE: 'refresh_token',
  
  // Timing constants
  TOKEN_REFRESH_THRESHOLD: 30000, // 30 seconds before expiry
  TOKEN_EXPIRY_BUFFER: 30000, // 30 seconds buffer for token expiry checks
  TOKEN_REFRESH_INTERVAL: 60000, // Check every minute
  SESSION_CHECK_INTERVAL: 5000, // Check session every 5 seconds
  
  // Security constants
  COOKIE_MAX_AGE: 7 * 24 * 60 * 60, // 7 days for refresh token
  ACCESS_TOKEN_BACKUP_MAX_AGE: 15 * 60, // 15 minutes for access token backup
  
  // Broadcast channel
  BROADCAST_CHANNEL: 'brokle_auth_sync',
  
  // Auth routes
  LOGIN_ROUTE: '/auth/signin',
  LOGOUT_ROUTE: '/auth/signout',
  DASHBOARD_ROUTE: '/dashboard',
  
  // API endpoints
  LOGIN_ENDPOINT: '/auth/login',
  REFRESH_ENDPOINT: '/v1/sessions/refresh',
  LOGOUT_ENDPOINT: '/auth/logout',
  ME_ENDPOINT: '/auth/me',
} as const

export const AUTH_EVENTS = {
  LOGIN: 'LOGIN',
  LOGOUT: 'LOGOUT',
  TOKEN_REFRESH: 'TOKEN_REFRESH',
  SESSION_EXPIRED: 'SESSION_EXPIRED',
  USER_UPDATED: 'USER_UPDATED',
} as const

export const AUTH_ERRORS = {
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',
  REFRESH_FAILED: 'REFRESH_FAILED',
  NETWORK_ERROR: 'NETWORK_ERROR',
  UNAUTHORIZED: 'UNAUTHORIZED',
} as const