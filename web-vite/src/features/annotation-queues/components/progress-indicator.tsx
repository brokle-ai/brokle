import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'
import type { QueueStats } from '../api/types'

interface ProgressIndicatorProps {
  stats: QueueStats
  className?: string
  showBreakdown?: boolean
  compact?: boolean
}

/**
 * Progress indicator for annotation queue review sessions. Both
 * completed AND skipped items count as "done" — skipping is a valid
 * disposition and should reflect in the percentage so reviewers see
 * forward motion when they pass on items that don't fit the criteria.
 */
export function ProgressIndicator({
  stats,
  className,
  showBreakdown = true,
  compact = false,
}: ProgressIndicatorProps) {
  const doneCount = stats.completed_items + stats.skipped_items
  const percentage =
    stats.total_items > 0 ? Math.round((doneCount / stats.total_items) * 100) : 0

  return (
    <div className={cn('space-y-2', className)}>
      <div className={cn('flex justify-between', compact ? 'text-xs' : 'text-sm')}>
        <span className="text-muted-foreground">Progress</span>
        <span className="font-medium">
          {doneCount}/{stats.total_items}{' '}
          <span className="text-muted-foreground">({percentage}%)</span>
        </span>
      </div>

      <Progress
        value={percentage}
        className={cn(
          '[&>div]:transition-all [&>div]:duration-300 [&>div]:ease-out',
          compact ? 'h-1.5' : 'h-2',
        )}
      />

      {showBreakdown && (
        <div
          className={cn(
            'flex justify-between text-muted-foreground',
            compact ? 'text-[10px]' : 'text-xs',
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

interface StatusCountProps {
  label: string
  count: number
  colorClass?: string
}

function StatusCount({ label, count, colorClass }: StatusCountProps) {
  return (
    <span>
      <span className={colorClass}>{count}</span>{' '}
      <span className="opacity-75">{label}</span>
    </span>
  )
}
