'use client'

import { Component, type ReactNode } from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface WidgetErrorBoundaryProps {
  children: ReactNode
  widgetId: string
  widgetTitle?: string
  className?: string
  onRetry?: () => void
}

interface WidgetErrorBoundaryState {
  hasError: boolean
  error: Error | null
}

/**
 * Error boundary component that isolates widget failures.
 * When a widget crashes, this component displays a fallback UI
 * instead of breaking the entire dashboard.
 */
export class WidgetErrorBoundary extends Component<
  WidgetErrorBoundaryProps,
  WidgetErrorBoundaryState
> {
  constructor(props: WidgetErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): WidgetErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log to console for debugging
    console.error(
      `Widget Error [${this.props.widgetId}]:`,
      error,
      errorInfo.componentStack
    )
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null })
    this.props.onRetry?.()
  }

  render() {
    if (this.state.hasError) {
      return (
        <WidgetErrorFallback
          widgetTitle={this.props.widgetTitle}
          error={this.state.error}
          onRetry={this.handleRetry}
          className={this.props.className}
        />
      )
    }

    return this.props.children
  }
}

interface WidgetErrorFallbackProps {
  widgetTitle?: string
  error: Error | null
  onRetry?: () => void
  className?: string
}

/**
 * Fallback UI displayed when a widget crashes.
 */
export function WidgetErrorFallback({
  widgetTitle,
  error,
  onRetry,
  className,
}: WidgetErrorFallbackProps) {
  return (
    <div
      className={cn(
        'flex h-full flex-col items-center justify-center gap-3 rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-center',
        className
      )}
    >
      <div className="flex items-center gap-2 text-destructive">
        <AlertCircle className="h-5 w-5" />
        <span className="font-medium">Widget Error</span>
      </div>

      {widgetTitle && (
        <p className="text-sm text-muted-foreground">
          Failed to render &quot;{widgetTitle}&quot;
        </p>
      )}

      {error?.message && (
        <p className="max-w-[200px] truncate text-xs text-muted-foreground">
          {error.message}
        </p>
      )}

      {onRetry && (
        <Button
          variant="outline"
          size="sm"
          onClick={onRetry}
          className="mt-2 gap-1.5"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          Retry
        </Button>
      )}
    </div>
  )
}
