'use client'

import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'
import type { QueueStats } from '../types'

interface ProgressIndicatorProps {
  /** Queue statistics with item counts */
  stats: QueueStats
  /** Additional CSS classes */
  className?: string
  /** Whether to show detailed breakdown */
  showBreakdown?: boolean
  /** Whether to use compact mode (smaller text) */
  compact?: boolean
}

/**
 * Progress indicator for annotation queue
 *
 * Features:
 * - Animated progress bar (300ms transition via CSS)
 * - Percentage display
 * - Optional status breakdown (pending, in_progress, completed, skipped)
 * - Compact mode for tight spaces
 *
 * Follows Opik pattern for progress visualization
 */
export function ProgressIndicator({
  stats,
  className,
  showBreakdown = true,
  compact = false,
}: ProgressIndicatorProps) {
  // Calculate percentage based on completed + skipped (both are "done")
  const doneCount = stats.completed_items + stats.skipped_items
  const percentage =
    stats.total_items > 0 ? Math.round((doneCount / stats.total_items) * 100) : 0

  return (
    <div className={cn('space-y-2', className)}>
      {/* Header with label and progress */}
      <div className={cn('flex justify-between', compact ? 'text-xs' : 'text-sm')}>
        <span className="text-muted-foreground">Progress</span>
        <span className="font-medium">
          {doneCount}/{stats.total_items}{' '}
          <span className="text-muted-foreground">({percentage}%)</span>
        </span>
      </div>

      {/* Progress bar with smooth animation */}
      <Progress
        value={percentage}
        className={cn(
          // Smooth 300ms animation for progress changes (Opik pattern)
          '[&>div]:transition-all [&>div]:duration-300 [&>div]:ease-out',
          compact ? 'h-1.5' : 'h-2'
        )}
      />

      {/* Status breakdown */}
      {showBreakdown && (
        <div
          className={cn(
            'flex justify-between text-muted-foreground',
            compact ? 'text-[10px]' : 'text-xs'
          )}
        >
          <StatusCount
            label="pending"
            count={stats.pending_items}
            colorClass="text-yellow-600 dark:text-yellow-500"
          />
          {stats.in_progress_items > 0 && (
            <StatusCount
              label="active"
              count={stats.in_progress_items}
              colorClass="text-blue-600 dark:text-blue-500"
            />
          )}
          <StatusCount
            label="done"
            count={stats.completed_items}
            colorClass="text-green-600 dark:text-green-500"
          />
          {stats.skipped_items > 0 && (
            <StatusCount
              label="skipped"
              count={stats.skipped_items}
              colorClass="text-gray-500"
            />
          )}
        </div>
      )}
    </div>
  )
}

/**
 * Individual status count display
 */
function StatusCount({
  label,
  count,
  colorClass,
}: {
  label: string
  count: number
  colorClass?: string
}) {
  return (
    <span>
      <span className={colorClass}>{count}</span>{' '}
      <span className="opacity-75">{label}</span>
    </span>
  )
}

/**
 * Minimal progress display (just percentage)
 * Used in compact UI contexts
 */
export function ProgressBadge({
  stats,
  className,
}: {
  stats: QueueStats
  className?: string
}) {
  const doneCount = stats.completed_items + stats.skipped_items
  const percentage =
    stats.total_items > 0 ? Math.round((doneCount / stats.total_items) * 100) : 0

  return (
    <span className={cn('text-sm font-medium', className)}>
      {percentage}%{' '}
      <span className="text-muted-foreground text-xs">
        ({doneCount}/{stats.total_items})
      </span>
    </span>
  )
}
