// Clean API exports - Direct functions with explicit versioning
// Perfect for dashboard application - latest optimal endpoints

// Feature-based API exports
export * from '@/features/authentication/api/auth-api'
export * from '@/features/organizations/api/organizations-api'
export * from '@/features/projects/api/projects-api'

// Remaining services (not yet migrated to features)
export * from './services/users'
export * from './services/public'
export * from './services/rbac'

// Core client and types
export { BrokleAPIClient } from './core/client'
export type { 
  APIClientConfig,
  RequestOptions,
  APIResponse,
  QueryParams,
  PaginatedResponse,
  BrokleAPIError
} from './core/types'

// Re-export auth types for convenience
export type {
  AuthResponse,
  AuthTokens,
  LoginCredentials,
  SignUpCredentials,
  User,
  Organization
} from '@/features/authentication'

// Types are exported directly from feature API files
// No need to re-export here since features handle their own types

// Re-export public API types
export type {
  HealthStatus,
  ServiceHealth,
  PublicStats,
  SystemStatus,
  SystemIncident,
  ContactFormData,
  FeedbackData
} from './services/public'

// Development helper - Make API functions available globally for debugging
if (process.env.NODE_ENV === 'development') {
  if (typeof window !== 'undefined') {
    import('@/features/authentication/api/auth-api').then((auth) => {
      import('@/features/organizations/api/organizations-api').then((orgs) => {
        import('@/features/projects/api/projects-api').then((projects) => {
          (window as any).brokleAPI = {
            auth,
            organizations: orgs,
            projects,
            client: new (require('./core/client').BrokleAPIClient)('/api')
          }
        })
      })
    })
  }
}