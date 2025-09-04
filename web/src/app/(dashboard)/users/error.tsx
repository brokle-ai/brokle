'use client'

import { useEffect } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface UsersErrorProps {
  error: Error & { digest?: string }
  reset: () => void
}

export default function UsersError({ error, reset }: UsersErrorProps) {
  useEffect(() => {
    // Log the error to your error reporting service
    console.error('Users page error:', error)
  }, [error])

  return (
    <div className="flex items-center justify-center min-h-[400px] p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
            <AlertTriangle className="h-6 w-6 text-destructive" />
          </div>
          <CardTitle className="text-lg">Something went wrong</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-sm text-muted-foreground">
            We encountered an error while loading the users page. This might be a temporary issue.
          </p>
          {process.env.NODE_ENV === 'development' && (
            <details className="rounded-lg bg-muted p-3">
              <summary className="cursor-pointer text-sm font-medium">
                Error details (dev only)
              </summary>
              <pre className="mt-2 text-xs text-muted-foreground whitespace-pre-wrap">
                {error.message}
              </pre>
            </details>
          )}
          <div className="flex justify-center">
            <Button onClick={reset} className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4" />
              Try again
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}