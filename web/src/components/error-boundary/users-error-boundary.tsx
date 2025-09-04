'use client'

import React from 'react'
import { Button } from '@/components/ui/button'
import { AlertCircle, RefreshCw } from 'lucide-react'

interface UsersErrorBoundaryState {
  hasError: boolean
  error?: Error
}

interface UsersErrorBoundaryProps {
  children: React.ReactNode
  fallback?: React.ComponentType<{ error?: Error; reset: () => void }>
}

function DefaultErrorFallback({ 
  error, 
  reset 
}: { 
  error?: Error
  reset: () => void 
}) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4 p-6">
      <div className="flex items-center space-x-2 text-destructive">
        <AlertCircle className="h-5 w-5" />
        <h3 className="text-lg font-semibold">Something went wrong</h3>
      </div>
      
      <div className="text-center space-y-2">
        <p className="text-sm text-muted-foreground">
          An error occurred while loading the users section.
        </p>
        {error && (
          <details className="text-xs text-muted-foreground">
            <summary className="cursor-pointer hover:text-foreground">
              Error details
            </summary>
            <pre className="mt-2 p-2 bg-muted rounded text-left overflow-auto max-w-md">
              {error.message}
            </pre>
          </details>
        )}
      </div>

      <Button onClick={reset} variant="outline" className="space-x-2">
        <RefreshCw className="h-4 w-4" />
        <span>Try again</span>
      </Button>
    </div>
  )
}

export class UsersErrorBoundary extends React.Component<
  UsersErrorBoundaryProps,
  UsersErrorBoundaryState
> {
  constructor(props: UsersErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): UsersErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Users section error boundary caught an error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      const ErrorFallback = this.props.fallback || DefaultErrorFallback
      
      return (
        <ErrorFallback
          error={this.state.error}
          reset={() => this.setState({ hasError: false, error: undefined })}
        />
      )
    }

    return this.props.children
  }
}