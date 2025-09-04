'use client'

import React from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

interface AuthErrorBoundaryState {
  hasError: boolean
  error?: Error
  errorInfo?: React.ErrorInfo
}

interface AuthErrorBoundaryProps {
  children: React.ReactNode
  fallbackTitle?: string
  fallbackDescription?: string
}

/**
 * Error boundary specifically for authentication-related errors
 * Provides user-friendly error messages and recovery options
 */
export class AuthErrorBoundary extends React.Component<
  AuthErrorBoundaryProps,
  AuthErrorBoundaryState
> {
  constructor(props: AuthErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): AuthErrorBoundaryState {
    return {
      hasError: true,
      error,
    }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    })

    // Log error for monitoring in production
    if (process.env.NODE_ENV === 'production') {
      console.error('[AuthErrorBoundary] Authentication error caught:', {
        error: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
      })
    }
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  handleReload = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      const { error } = this.state
      const {
        fallbackTitle = 'Authentication Error',
        fallbackDescription = 'Something went wrong with the authentication system. Please try again.',
      } = this.props

      // Determine error type for better user messaging
      const isNetworkError = error?.message?.includes('fetch') || error?.message?.includes('network')
      const isTokenError = error?.message?.includes('token') || error?.message?.includes('auth')

      let userMessage = fallbackDescription
      let suggestedAction = 'Try again'

      if (isNetworkError) {
        userMessage = 'Unable to connect to authentication servers. Please check your internet connection.'
        suggestedAction = 'Reload page'
      } else if (isTokenError) {
        userMessage = 'Your session may have expired or become invalid. Please sign in again.'
        suggestedAction = 'Sign in'
      }

      return (
        <Card className="w-full max-w-md mx-auto">
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              <AlertTriangle className="h-12 w-12 text-destructive" />
            </div>
            <CardTitle className="text-lg">{fallbackTitle}</CardTitle>
            <CardDescription>{userMessage}</CardDescription>
          </CardHeader>
          
          <CardContent>
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Error Details</AlertTitle>
              <AlertDescription className="mt-2">
                {process.env.NODE_ENV === 'development' ? (
                  <details className="text-xs">
                    <summary className="cursor-pointer font-medium">Technical Details</summary>
                    <pre className="mt-2 whitespace-pre-wrap break-words">
                      {error?.message || 'Unknown error'}
                      {error?.stack && `\n\nStack trace:\n${error.stack}`}
                    </pre>
                  </details>
                ) : (
                  'An unexpected error occurred. Please try again or contact support if the problem persists.'
                )}
              </AlertDescription>
            </Alert>
          </CardContent>

          <CardFooter className="flex gap-2 justify-center">
            <Button variant="outline" onClick={this.handleRetry} className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4" />
              {suggestedAction}
            </Button>
            <Button onClick={this.handleReload} className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4" />
              Reload Page
            </Button>
          </CardFooter>
        </Card>
      )
    }

    return this.props.children
  }
}

/**
 * Hook version for functional components
 */
export function useAuthErrorHandler() {
  const [error, setError] = React.useState<string | null>(null)
  
  const handleError = React.useCallback((error: unknown) => {
    if (error instanceof Error) {
      setError(error.message)
    } else if (typeof error === 'string') {
      setError(error)
    } else {
      setError('An unexpected error occurred')
    }
  }, [])

  const clearError = React.useCallback(() => {
    setError(null)
  }, [])

  return {
    error,
    hasError: !!error,
    handleError,
    clearError,
  }
}