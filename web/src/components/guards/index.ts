// Authentication guards — imported from authentication feature.
// Auth enforcement itself lives server-side (proxy.ts + lib/auth/dal.ts);
// the legacy <AuthGuard> client-side wrapper was removed because it
// implemented the closure-stale "redirect on cold load" anti-pattern
// the Next.js 16 auth guide warns against.
export { UnauthorizedFallback } from '@/features/authentication'

// Fallback components
export { LoadingSpinner, PageLoadingSpinner } from './loading-spinner'
export { ForbiddenFallback } from './forbidden-fallback'