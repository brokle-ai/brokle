import { useNavigate } from '@tanstack/react-router'
import { AlertTriangle, Home, RefreshCw } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { AuthenticationError, BrokleError } from '@/lib/api/errors'

// Global error fallback rendered by `__root.tsx`'s `errorComponent`.
// Catches any uncaught error from `beforeLoad` / `loader` / render
// across the whole route tree (TanStack Router re-throws those errors
// into the component tree per the docs). Mirrors the single-global-
// boundary pattern used by Opik (`SentryErrorBoundary` wrapping
// `<RouterProvider />`) and SigNoz (`Sentry.ErrorBoundary` wrapping
// `Router`).
//
// Per-route `errorComponent` declarations on leaf routes still take
// precedence; this fallback only runs when no closer boundary
// handles the error.

interface RootErrorFallbackProps {
  error: Error
  onRetry: () => void
}

function deriveMessage(error: Error): string {
  if (error instanceof BrokleError) {
    return error.body.error.message
  }
  return error.message || 'An unexpected error occurred.'
}

function deriveTitle(error: Error): string {
  if (error instanceof AuthenticationError) {
    // beforeLoad's catch path normally handles this via
    // `throw redirect({ to: '/signin' })`. Reaching here means the
    // redirect itself failed — surface the auth context anyway.
    return 'Your session has ended'
  }
  if (error instanceof BrokleError && error.status >= 500) {
    return 'The server is having trouble'
  }
  return 'Something went wrong'
}

export function RootErrorFallback({ error, onRetry }: RootErrorFallbackProps) {
  const navigate = useNavigate()

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Alert variant="destructive" className="max-w-md">
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>{deriveTitle(error)}</AlertTitle>
        <AlertDescription className="mt-2 space-y-4">
          <p>{deriveMessage(error)}</p>
          <div className="flex flex-wrap gap-2">
            <Button size="sm" onClick={onRetry}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Try again
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => void navigate({ to: '/signin' })}
            >
              Sign in again
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => window.location.reload()}
            >
              <Home className="mr-2 h-4 w-4" />
              Reload page
            </Button>
          </div>
        </AlertDescription>
      </Alert>
    </div>
  )
}
