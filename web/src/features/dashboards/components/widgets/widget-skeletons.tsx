'use client'

import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { WidgetType } from '../../types'

interface WidgetSkeletonProps {
  className?: string
}

/**
 * Skeleton for stat/metric widgets
 */
export function StatSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center gap-2 p-4', className)}>
      <Skeleton className="h-8 w-24" />
      <Skeleton className="h-4 w-16" />
    </div>
  )
}

/**
 * Skeleton for time series/line chart widgets
 */
export function TimeSeriesSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-3 p-4', className)}>
      <div className="flex items-center justify-between">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-16" />
      </div>
      <div className="flex h-full items-end gap-1">
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton
            key={i}
            className="w-full"
            style={{ height: `${Math.random() * 60 + 20}%` }}
          />
        ))}
      </div>
      <div className="flex justify-between">
        <Skeleton className="h-3 w-10" />
        <Skeleton className="h-3 w-10" />
      </div>
    </div>
  )
}

/**
 * Skeleton for bar chart widgets
 */
export function BarSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-3 p-4', className)}>
      <Skeleton className="h-4 w-32" />
      <div className="flex h-full items-end gap-2">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="flex flex-1 flex-col items-center gap-1">
            <Skeleton
              className="w-full"
              style={{ height: `${Math.random() * 60 + 30}%` }}
            />
            <Skeleton className="h-3 w-8" />
          </div>
        ))}
      </div>
    </div>
  )
}

/**
 * Skeleton for pie/donut chart widgets
 */
export function PieSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex items-center justify-center gap-6 p-4', className)}>
      <Skeleton className="h-32 w-32 rounded-full" />
      <div className="flex flex-col gap-2">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="flex items-center gap-2">
            <Skeleton className="h-3 w-3 rounded-sm" />
            <Skeleton className="h-3 w-16" />
          </div>
        ))}
      </div>
    </div>
  )
}

/**
 * Skeleton for table widgets
 */
export function TableSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-2 p-4', className)}>
      {/* Header row */}
      <div className="flex gap-4 border-b pb-2">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-4 w-20" />
      </div>
      {/* Data rows */}
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex gap-4 py-1">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-4 w-20" />
        </div>
      ))}
    </div>
  )
}

/**
 * Skeleton for heatmap widgets
 */
export function HeatmapSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-2 p-4', className)}>
      <Skeleton className="h-4 w-24" />
      <div className="grid grid-cols-7 gap-1">
        {Array.from({ length: 35 }).map((_, i) => (
          <Skeleton
            key={i}
            className="aspect-square rounded-sm"
            style={{ opacity: Math.random() * 0.7 + 0.3 }}
          />
        ))}
      </div>
      <div className="flex justify-between">
        <Skeleton className="h-3 w-12" />
        <Skeleton className="h-3 w-12" />
      </div>
    </div>
  )
}

/**
 * Skeleton for histogram widgets
 */
export function HistogramSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-3 p-4', className)}>
      <div className="flex justify-between">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-16" />
      </div>
      <div className="flex h-full items-end gap-0.5">
        {Array.from({ length: 20 }).map((_, i) => {
          // Create a bell-curve-like distribution
          const distance = Math.abs(i - 10) / 10
          const height = Math.max(20, (1 - distance * 0.8) * 100)
          return (
            <Skeleton
              key={i}
              className="w-full"
              style={{ height: `${height}%` }}
            />
          )
        })}
      </div>
    </div>
  )
}

/**
 * Skeleton for trace list widgets
 */
export function TraceListSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-3 p-4', className)}>
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="flex flex-col gap-1.5 border-b pb-3 last:border-0">
          <div className="flex items-center justify-between">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-3 w-20" />
            <Skeleton className="h-3 w-24" />
          </div>
        </div>
      ))}
    </div>
  )
}

/**
 * Skeleton for text widgets
 */
export function TextSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-2 p-4', className)}>
      <Skeleton className="h-5 w-32" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-3/4" />
      <Skeleton className="h-4 w-5/6" />
    </div>
  )
}

/**
 * Generic widget skeleton (used as default)
 */
export function GenericWidgetSkeleton({ className }: WidgetSkeletonProps) {
  return (
    <div className={cn('flex flex-col gap-3 p-4', className)}>
      <Skeleton className="h-5 w-32" />
      <Skeleton className="h-24 w-full" />
      <div className="flex gap-2">
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-4 w-16" />
      </div>
    </div>
  )
}

/**
 * Get the appropriate skeleton component for a widget type
 */
export function getWidgetSkeleton(type: WidgetType): React.ComponentType<WidgetSkeletonProps> {
  switch (type) {
    case 'stat':
      return StatSkeleton
    case 'time_series':
      return TimeSeriesSkeleton
    case 'bar':
      return BarSkeleton
    case 'pie':
      return PieSkeleton
    case 'table':
      return TableSkeleton
    case 'heatmap':
      return HeatmapSkeleton
    case 'histogram':
      return HistogramSkeleton
    case 'trace_list':
      return TraceListSkeleton
    case 'text':
      return TextSkeleton
    default:
      return GenericWidgetSkeleton
  }
}

interface WidgetSkeletonRendererProps extends WidgetSkeletonProps {
  type: WidgetType
}

/**
 * Renders the appropriate skeleton based on widget type
 */
export function WidgetSkeletonRenderer({ type, className }: WidgetSkeletonRendererProps) {
  const SkeletonComponent = getWidgetSkeleton(type)
  return <SkeletonComponent className={className} />
}
