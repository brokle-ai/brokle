'use client'

import { ReactNode } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface BaseChartProps {
  title?: string
  description?: string
  children: ReactNode
  loading?: boolean
  error?: string
  height?: number
  className?: string
  actions?: ReactNode
}

export function BaseChart({
  title,
  description,
  children,
  loading = false,
  error,
  height = 300,
  className = '',
  actions
}: BaseChartProps) {
  if (loading) {
    return (
      <Card className={className}>
        {(title || description) && (
          <CardHeader>
            {title && <CardTitle><Skeleton className="h-6 w-32" /></CardTitle>}
            {description && <CardDescription><Skeleton className="h-4 w-48" /></CardDescription>}
          </CardHeader>
        )}
        <CardContent>
          <Skeleton className="w-full" style={{ height: `${height}px` }} />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={className}>
        {(title || description) && (
          <CardHeader>
            {title && <CardTitle>{title}</CardTitle>}
            {description && <CardDescription>{description}</CardDescription>}
          </CardHeader>
        )}
        <CardContent className="flex items-center justify-center" style={{ height: `${height}px` }}>
          <div className="text-center">
            <p className="text-muted-foreground text-sm">Failed to load chart</p>
            <p className="text-destructive text-xs mt-1">{error}</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      {(title || description || actions) && (
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              {title && <CardTitle>{title}</CardTitle>}
              {description && <CardDescription>{description}</CardDescription>}
            </div>
            {actions && <div className="flex items-center space-x-2">{actions}</div>}
          </div>
        </CardHeader>
      )}
      <CardContent>
        <div style={{ height: `${height}px` }}>
          {children}
        </div>
      </CardContent>
    </Card>
  )
}