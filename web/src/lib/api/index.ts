// Clean API exports - New axios-based implementation

// Core client and types
export { BrokleAPIClient } from './core/client'
export type { 
  APIClientConfig,
  RequestOptions,
  APIResponse,
  BrokleAPIError as APIError,
  QueryParams,
  PaginatedResponse
} from './core/types'

// Import service clients for internal use
import { AuthAPIClient } from './services/auth'
import { PublicAPIClient } from './services/public'
import { AnalyticsAPIClient } from './services/analytics'
import { DashboardAPIClient } from './services/dashboard'
import { OnboardingAPIClient } from './services/onboarding'
import { OrganizationAPIClient } from './services/organizations'
import { UsersAPIClient } from './services/users'

// Re-export service clients
export { AuthAPIClient } from './services/auth'
export { PublicAPIClient } from './services/public' 
export { AnalyticsAPIClient } from './services/analytics'
export { DashboardAPIClient } from './services/dashboard'
export { OnboardingAPIClient } from './services/onboarding'
export { OrganizationAPIClient } from './services/organizations'
export { UsersAPIClient } from './services/users'

// Service client instances (singletons)
export const api = {
  auth: new AuthAPIClient(),
  public: new PublicAPIClient(),
  analytics: new AnalyticsAPIClient(),
  dashboard: new DashboardAPIClient(),
  onboarding: new OnboardingAPIClient(),
  organizations: new OrganizationAPIClient(),
  users: new UsersAPIClient(),
}

// Re-export auth types for convenience
export type {
  AuthResponse,
  AuthTokens,
  LoginCredentials,
  SignUpCredentials,
  User,
  Organization,
  ApiKey
} from '@/types/auth'

// Development helper
if (process.env.NODE_ENV === 'development') {
  // Make API clients available globally for debugging
  if (typeof window !== 'undefined') {
    (window as any).api = api
  }
}