// Authentication guards - simple auth verification only
export { 
  AuthGuard,
  VerifiedGuard,
} from './auth-guard'

// TODO: Role-based guards removed - implement PermissionGuard with backend integration
// Future: export { PermissionGuard } from './permission-guard'

// Fallback components
export { LoadingSpinner, PageLoadingSpinner } from './loading-spinner'
export { UnauthorizedFallback } from './unauthorized-fallback'
export { ForbiddenFallback } from './forbidden-fallback'