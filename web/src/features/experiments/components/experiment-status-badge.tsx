'use client'

import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  Clock,
  Loader2,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Ban,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ExperimentStatus } from '../types'

interface ExperimentStatusBadgeProps {
  status: ExperimentStatus
  progressPct?: number
  showProgress?: boolean
  className?: string
}

const statusConfig: Record<
  ExperimentStatus,
  {
    label: string
    variant: 'default' | 'secondary' | 'destructive' | 'outline'
    icon: typeof Clock
  }
> = {
  pending: {
    label: 'Pending',
    variant: 'outline',
    icon: Clock,
  },
  running: {
    label: 'Running',
    variant: 'secondary',
    icon: Loader2,
  },
  completed: {
    label: 'Completed',
    variant: 'default',
    icon: CheckCircle2,
  },
  failed: {
    label: 'Failed',
    variant: 'destructive',
    icon: XCircle,
  },
  partial: {
    label: 'Partial',
    variant: 'outline',
    icon: AlertTriangle,
  },
  cancelled: {
    label: 'Cancelled',
    variant: 'outline',
    icon: Ban,
  },
}

export function ExperimentStatusBadge({
  status,
  progressPct,
  showProgress = false,
  className,
}: ExperimentStatusBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon
  const isRunning = status === 'running'

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Badge variant={config.variant} className="flex items-center gap-1.5">
        <Icon className={cn('h-3 w-3', isRunning && 'animate-spin')} />
        <span>{config.label}</span>
        {isRunning && progressPct !== undefined && (
          <span className="text-xs opacity-80">{Math.round(progressPct)}%</span>
        )}
      </Badge>

      {showProgress && isRunning && progressPct !== undefined && (
        <Progress value={progressPct} className="w-20 h-1.5" />
      )}
    </div>
  )
}
