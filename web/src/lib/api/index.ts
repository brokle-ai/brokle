// Clean API exports - Direct functions with explicit versioning
// Perfect for dashboard application - latest optimal endpoints

// Direct service exports - always use optimal backend version per endpoint
export * from './services/auth'
export * from './services/organizations' 
export * from './services/analytics'
export * from './services/users'
export * from './services/onboarding'
export * from './services/dashboard'
export * from './services/public'

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
  Organization,
  ApiKey
} from '@/types/auth'

// Re-export analytics types
export type {
  AnalyticsMetric,
  ModelUsage,
  ProviderUsage,
  DashboardStats,
  AnalyticsQuery,
  CostAnalytics
} from './services/analytics'

// Re-export user types
export type {
  CreateUserData,
  InviteUserData
} from './services/users'

// Re-export dashboard types
export type {
  QuickStat,
  ChartData,
  DashboardOverview,
  DashboardConfig
} from './services/dashboard'

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

// Development helper - Clean API object (optional)
if (process.env.NODE_ENV === 'development') {
  if (typeof window !== 'undefined') {
    // Make API functions available globally for debugging
    import('./services/auth').then((auth) => {
      import('./services/organizations').then((orgs) => {
        import('./services/analytics').then((analytics) => {
          (window as any).brokleAPI = {
            auth,
            organizations: orgs,
            analytics,
            // Easy debugging access
            client: new (require('./core/client').BrokleAPIClient)('/api')
          }
        })
      })
    })
  }
}