'use client'

import { WorkspaceError, WorkspaceErrorCode } from '@/context/workspace-errors'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'
import { AlertCircle, Home, RefreshCw } from 'lucide-react'

interface WorkspaceErrorPageProps {
  error: WorkspaceError
}

/**
 * Full-page workspace error component
 *
 * Displays errors without dashboard chrome (no sidebar, no header).
 * Used by DashboardLayoutContent when workspace context has an error.
 *
 * Features:
 * - Full-page centered layout
 * - Error-specific titles and icons
 * - Action buttons (Go Home, Try Again)
 * - Development mode debug info
 * - Accessible and responsive
 *
 * @example
 * ```tsx
 * const { error } = useWorkspace()
 * if (error) return <WorkspaceErrorPage error={error} />
 * ```
 */
export function WorkspaceErrorPage({ error }: WorkspaceErrorPageProps) {
  const router = useRouter()

  const getTitle = () => {
    switch (error.code) {
      case WorkspaceErrorCode.ORG_NOT_FOUND:
        return 'Organization Not Found'
      case WorkspaceErrorCode.PROJECT_NOT_FOUND:
        return 'Project Not Found'
      case WorkspaceErrorCode.INVALID_ORG_SLUG:
      case WorkspaceErrorCode.INVALID_PROJECT_SLUG:
        return 'Invalid Link'
      case WorkspaceErrorCode.ORG_NO_ACCESS:
      case WorkspaceErrorCode.PROJECT_NO_ACCESS:
        return 'Access Denied'
      case WorkspaceErrorCode.NETWORK_ERROR:
        return 'Connection Error'
      case WorkspaceErrorCode.API_FAILED:
        return 'Service Unavailable'
      default:
        return 'Something Went Wrong'
    }
  }

  const getIcon = () => {
    if (error.code === WorkspaceErrorCode.NETWORK_ERROR) {
      return <RefreshCw className="h-16 w-16 text-muted-foreground" />
    }
    return <AlertCircle className="h-16 w-16 text-muted-foreground" />
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="mx-auto max-w-md text-center px-6">
        <div className="mb-6 flex justify-center">
          {getIcon()}
        </div>

        <h1 className="mb-3 text-3xl font-bold tracking-tight">
          {getTitle()}
        </h1>

        <p className="mb-8 text-base text-muted-foreground leading-relaxed">
          {error.userMessage}
        </p>

        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button
            onClick={() => router.push('/')}
            variant="default"
            className="gap-2"
          >
            <Home className="h-4 w-4" />
            Go to Home
          </Button>

          {error.code === WorkspaceErrorCode.NETWORK_ERROR && (
            <Button
              onClick={() => window.location.reload()}
              variant="outline"
              className="gap-2"
            >
              <RefreshCw className="h-4 w-4" />
              Try Again
            </Button>
          )}
        </div>

        {process.env.NODE_ENV === 'development' && (
          <div className="mt-8 rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-left">
            <p className="mb-2 text-sm font-semibold text-destructive">
              Development Info:
            </p>
            <pre className="text-xs text-muted-foreground overflow-auto">
              {JSON.stringify(
                {
                  code: error.code,
                  message: error.message,
                  context: error.context,
                },
                null,
                2
              )}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
