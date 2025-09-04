'use client'

import * as React from 'react'
import { MetricCard } from './metric-card'
import { cn } from '@/lib/utils'

interface Metric {
  id: string
  title: string
  value: string | number
  description?: string
  icon?: React.ComponentType<{ className?: string }>
  trend?: {
    value: number
    label: string
    direction: 'up' | 'down'
  }
}

interface StatsGridProps extends React.HTMLAttributes<HTMLDivElement> {
  metrics?: Metric[]
  loading?: boolean
  error?: string
  columns?: 1 | 2 | 3 | 4
  gap?: 'sm' | 'md' | 'lg'
  children?: React.ReactNode
}

const gridColumnsMap = {
  1: 'grid-cols-1',
  2: 'grid-cols-1 sm:grid-cols-2',
  3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
  4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
}

const gapMap = {
  sm: 'gap-2',
  md: 'gap-4',
  lg: 'gap-6',
}

export function StatsGrid({
  metrics,
  loading = false,
  error,
  columns = 4,
  gap = 'md',
  className,
  children,
  ...props
}: StatsGridProps) {
  if (error) {
    return (
      <div className={cn('text-center p-8', className)} {...props}>
        <div className='text-destructive text-sm'>Failed to load metrics</div>
        <p className='text-muted-foreground text-xs mt-1'>{error}</p>
      </div>
    )
  }

  return (
    <div
      className={cn(
        'grid',
        gridColumnsMap[columns],
        gapMap[gap],
        className
      )}
      {...props}
    >
      {loading ? (
        // Show loading skeleton
        Array.from({ length: columns }, (_, index) => (
          <MetricCard
            key={index}
            title=''
            value=''
            loading={true}
          />
        ))
      ) : children ? (
        // Use children if provided
        children
      ) : (
        // Use metrics prop if provided
        metrics?.map((metric) => (
          <MetricCard
            key={metric.id}
            title={metric.title}
            value={metric.value}
            description={metric.description}
            icon={metric.icon}
            trend={metric.trend}
          />
        )) || []
      )}
    </div>
  )
}