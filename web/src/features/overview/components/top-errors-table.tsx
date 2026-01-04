'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowRight, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { TopError } from '../types'

interface TopErrorsTableProps {
  data: TopError[]
  isLoading?: boolean
  error?: string | null
  className?: string
  projectSlug?: string
}

function formatRelativeTime(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffSecs < 60) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

function formatCount(count: number): string {
  if (count < 1000) return count.toString()
  if (count < 10000) return `${(count / 1000).toFixed(1)}k`
  return `${Math.round(count / 1000)}k`
}

function truncateMessage(message: string, maxLength: number = 50): string {
  if (message.length <= maxLength) return message
  return message.substring(0, maxLength - 3) + '...'
}

export function TopErrorsTable({
  data,
  isLoading,
  error,
  className,
  projectSlug,
}: TopErrorsTableProps) {
  const router = useRouter()

  const handleViewAll = () => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/traces?status=error`)
    }
  }

  const handleErrorClick = (errorMessage: string) => {
    if (projectSlug) {
      const encodedMessage = encodeURIComponent(errorMessage)
      router.push(`/projects/${projectSlug}/traces?status=error&search=${encodedMessage}`)
    }
  }

  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium">Top Errors</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Skeleton className="h-4 w-4" />
                  <Skeleton className="h-4 w-48" />
                </div>
                <div className="flex items-center gap-3">
                  <Skeleton className="h-5 w-12" />
                  <Skeleton className="h-4 w-16" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={cn('border-destructive', className)}>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium text-destructive">
            Top Errors
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">{error}</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base font-medium">Top Errors</CardTitle>
        <Button
          variant="ghost"
          size="sm"
          className="gap-1 text-xs"
          onClick={handleViewAll}
        >
          View All
          <ArrowRight className="h-3 w-3" />
        </Button>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <div className="h-[200px] flex flex-col items-center justify-center text-muted-foreground gap-2">
            <AlertCircle className="h-8 w-8 text-green-500" />
            <span className="text-sm">No errors detected. Great job!</span>
          </div>
        ) : (
          <div className="space-y-2">
            {data.map((err, index) => (
              <div
                key={`${err.message}-${index}`}
                className="flex items-center justify-between py-2 px-2 rounded-md hover:bg-muted/50 cursor-pointer transition-colors"
                onClick={() => handleErrorClick(err.message)}
                role="button"
                tabIndex={0}
              >
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <AlertCircle className="h-4 w-4 text-destructive shrink-0" />
                  <span
                    className="text-sm font-medium truncate text-destructive"
                    title={err.message}
                  >
                    {truncateMessage(err.message)}
                  </span>
                </div>
                <div className="flex items-center gap-4 shrink-0 text-sm text-muted-foreground">
                  <Badge variant="secondary" className="shrink-0">
                    {formatCount(err.count)}x
                  </Badge>
                  <span className="w-16 text-right">
                    {formatRelativeTime(err.last_seen)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
