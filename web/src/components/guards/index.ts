// Authentication guards - imported from authentication feature
export {
  AuthGuard,
  UnauthorizedFallback,
} from '@/features/authentication'

// TODO: VerifiedGuard - implement if needed
// TODO: Role-based guards removed - implement PermissionGuard with backend integration
// Future: export { PermissionGuard } from './permission-guard'

// Fallback components
export { LoadingSpinner, PageLoadingSpinner } from './loading-spinner'
export { ForbiddenFallback } from './forbidden-fallback'