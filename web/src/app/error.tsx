'use client'

import { useEffect } from 'react'
import Link from 'next/link'
import { ServerCrash, RefreshCw, Home } from 'lucide-react'
import { Button } from '@/components/ui/button'

/**
 * Global Error Boundary
 * Catches unhandled errors in the app and provides recovery options
 * Must be a Client Component in Next.js App Router
 *
 * Next.js App Router: app/error.tsx
 */
export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // Log error to monitoring service (e.g., Sentry, LogRocket, etc.)
    console.error('Application error occurred:', error)

    // In production, send to error tracking service:
    // if (process.env.NODE_ENV === 'production') {
    //   Sentry.captureException(error, {
    //     tags: { errorBoundary: 'app-root' },
    //     extra: { digest: error.digest },
    //   })
    // }
  }, [error])

  return (
    <div className="flex min-h-svh flex-col items-center justify-center px-4 py-12">
      <div className="mx-auto max-w-md text-center">
        {/* Icon */}
        <div className="mb-6 flex justify-center">
          <div className="rounded-full bg-muted p-6">
            <ServerCrash className="h-12 w-12 text-muted-foreground" />
          </div>
        </div>

        {/* Status Code */}
        <h1 className="mb-2 text-6xl font-bold tracking-tight text-foreground">
          500
        </h1>

        {/* Title */}
        <h2 className="mb-4 text-2xl font-semibold tracking-tight text-foreground">
          Something Went Wrong
        </h2>

        {/* Description */}
        <p className="mb-8 text-muted-foreground">
          An unexpected error occurred. Our team has been notified. Please try again or return to the dashboard.
        </p>

        {/* Development Mode: Show error details */}
        {process.env.NODE_ENV === 'development' && (
          <div className="mb-6 rounded-lg bg-muted p-4 text-left">
            <p className="mb-2 text-sm font-semibold text-destructive">
              Development Error Details:
            </p>
            <p className="text-xs font-mono text-muted-foreground break-words">
              {error.message}
            </p>
            {error.digest && (
              <p className="mt-2 text-xs text-muted-foreground">
                Error Digest: {error.digest}
              </p>
            )}
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button
            variant="outline"
            onClick={reset}
            className="gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            Try Again
          </Button>

          <Button asChild>
            <Link href="/" className="gap-2">
              <Home className="h-4 w-4" />
              Go to Dashboard
            </Link>
          </Button>
        </div>
      </div>
    </div>
  )
}