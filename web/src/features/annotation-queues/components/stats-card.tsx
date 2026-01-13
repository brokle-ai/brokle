'use client'

import { LucideIcon } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

interface StatsCardProps {
  title: string
  value: number | string
  description?: string
  icon?: LucideIcon
  color?: 'default' | 'green' | 'yellow' | 'red' | 'blue'
  progress?: number
  isLoading?: boolean
  className?: string
}

function getColorClasses(color: StatsCardProps['color']): string {
  switch (color) {
    case 'green':
      return 'text-green-600'
    case 'yellow':
      return 'text-yellow-600'
    case 'red':
      return 'text-red-600'
    case 'blue':
      return 'text-blue-600'
    default:
      return ''
  }
}

export function StatsCard({
  title,
  value,
  description,
  icon: Icon,
  color = 'default',
  progress,
  isLoading = false,
  className,
}: StatsCardProps) {
  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader className="pb-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-8 w-16 mt-2" />
        </CardHeader>
        {progress !== undefined && (
          <CardContent className="pt-0">
            <Skeleton className="h-2 w-full" />
          </CardContent>
        )}
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardDescription className="flex items-center gap-1">
          {Icon && <Icon className="h-4 w-4" />}
          {title}
        </CardDescription>
        <CardTitle className={cn('text-2xl', getColorClasses(color))}>
          {value}
        </CardTitle>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
      </CardHeader>
      {progress !== undefined && (
        <CardContent className="pt-0">
          <Progress value={progress} className="h-2" />
        </CardContent>
      )}
    </Card>
  )
}

// Convenience wrapper for displaying queue stats
interface QueueStatsCardsProps {
  stats: {
    total_items: number
    pending_items: number
    in_progress_items: number
    completed_items: number
    skipped_items: number
  } | null
  isLoading?: boolean
  className?: string
}

export function QueueStatsCards({ stats, isLoading = false, className }: QueueStatsCardsProps) {
  const completionPercentage = stats
    ? stats.total_items > 0
      ? Math.round(((stats.completed_items + stats.skipped_items) / stats.total_items) * 100)
      : 0
    : 0

  return (
    <div className={cn('grid gap-4 md:grid-cols-4', className)}>
      <StatsCard
        title="Total Items"
        value={stats?.total_items ?? 0}
        isLoading={isLoading}
      />
      <StatsCard
        title="Pending"
        value={stats?.pending_items ?? 0}
        color="yellow"
        isLoading={isLoading}
      />
      <StatsCard
        title="Completed"
        value={stats?.completed_items ?? 0}
        color="green"
        isLoading={isLoading}
      />
      <StatsCard
        title="Progress"
        value={`${completionPercentage}%`}
        progress={completionPercentage}
        isLoading={isLoading}
      />
    </div>
  )
}
