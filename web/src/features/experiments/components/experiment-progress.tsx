'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { CheckCircle, XCircle, Clock, Timer } from 'lucide-react'
import { useExperimentProgressQuery } from '../hooks/use-experiments'
import { ExperimentStatusBadge } from './experiment-status-badge'

interface ExperimentProgressProps {
  projectId: string
  experimentId: string
}

/**
 * Format duration from seconds to human-readable string
 */
function formatDurationSeconds(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.round(seconds % 60)
  if (minutes < 60) return `${minutes}m ${remainingSeconds}s`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

export function ExperimentProgress({
  projectId,
  experimentId,
}: ExperimentProgressProps) {
  const { data: progress, isLoading } = useExperimentProgressQuery(
    projectId,
    experimentId
  )

  if (isLoading) return <ExperimentProgressSkeleton />
  if (!progress) return null

  const {
    status,
    total_items,
    completed_items,
    failed_items,
    pending_items,
    progress_pct,
    elapsed_seconds,
    eta_seconds,
  } = progress

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium">Progress</CardTitle>
          <ExperimentStatusBadge status={status} />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Progress bar */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">
              {completed_items + failed_items} of {total_items} items
            </span>
            <span className="font-medium">{Math.round(progress_pct)}%</span>
          </div>
          <Progress value={progress_pct} className="h-2" />
        </div>

        {/* Stats grid */}
        <div className="grid grid-cols-3 gap-4">
          <div className="flex items-center gap-2">
            <CheckCircle className="h-4 w-4 text-green-500" />
            <div>
              <p className="text-sm font-medium">{completed_items}</p>
              <p className="text-xs text-muted-foreground">Completed</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <XCircle className="h-4 w-4 text-red-500" />
            <div>
              <p className="text-sm font-medium">{failed_items}</p>
              <p className="text-xs text-muted-foreground">Failed</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-sm font-medium">{pending_items}</p>
              <p className="text-xs text-muted-foreground">Pending</p>
            </div>
          </div>
        </div>

        {/* Time info for running experiments */}
        {status === 'running' && (elapsed_seconds || eta_seconds) && (
          <div className="flex items-center justify-between text-sm border-t pt-3">
            {elapsed_seconds !== undefined && (
              <div className="flex items-center gap-1.5 text-muted-foreground">
                <Timer className="h-3.5 w-3.5" />
                <span>Elapsed: {formatDurationSeconds(elapsed_seconds)}</span>
              </div>
            )}
            {eta_seconds !== undefined && (
              <div className="text-muted-foreground">
                ETA: {formatDurationSeconds(eta_seconds)}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ExperimentProgressSkeleton() {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-5 w-20" />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-4 w-8" />
          </div>
          <Skeleton className="h-2 w-full" />
        </div>
        <div className="grid grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
