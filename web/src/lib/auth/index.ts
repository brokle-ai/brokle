// Main auth exports
export { getTokenManager } from './token-manager'
export { getSessionSync } from './session-sync'
export { SecureStorage } from './storage'
export { AUTH_CONSTANTS, AUTH_EVENTS, AUTH_ERRORS } from './constants'

// Types
export type { TokenRefreshCallback } from './token-manager'

// Utility functions
export { validateEmail, validatePassword, isValidUrl } from './validation'

// Re-export auth types
export type {
  AuthTokens,
  LoginCredentials,
  SignUpCredentials,
  AuthResponse,
  User,
  Organization,
} from '@/types/auth'