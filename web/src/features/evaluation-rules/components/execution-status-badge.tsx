'use client'

import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { CheckCircle, XCircle, Clock, Loader2, Ban } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ExecutionStatus } from '../types'

interface ExecutionStatusBadgeProps {
  status: ExecutionStatus
  className?: string
  showTooltip?: boolean
  timestamp?: string
}

const statusConfig: Record<
  ExecutionStatus,
  {
    label: string
    variant: 'default' | 'secondary' | 'destructive' | 'outline'
    icon: typeof CheckCircle
    iconClassName: string
    description: string
  }
> = {
  completed: {
    label: 'Completed',
    variant: 'default',
    icon: CheckCircle,
    iconClassName: 'text-green-500',
    description: 'Execution finished successfully',
  },
  running: {
    label: 'Running',
    variant: 'secondary',
    icon: Loader2,
    iconClassName: 'text-yellow-500 animate-spin',
    description: 'Execution is currently in progress',
  },
  pending: {
    label: 'Pending',
    variant: 'outline',
    icon: Clock,
    iconClassName: 'text-muted-foreground',
    description: 'Execution is queued and waiting to start',
  },
  failed: {
    label: 'Failed',
    variant: 'destructive',
    icon: XCircle,
    iconClassName: 'text-red-500',
    description: 'Execution encountered an error',
  },
  cancelled: {
    label: 'Cancelled',
    variant: 'outline',
    icon: Ban,
    iconClassName: 'text-muted-foreground',
    description: 'Execution was cancelled',
  },
}

export function ExecutionStatusBadge({
  status,
  className,
  showTooltip = true,
  timestamp,
}: ExecutionStatusBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon

  const badge = (
    <Badge variant={config.variant} className={cn('gap-1', className)}>
      <Icon className={cn('h-3 w-3', config.iconClassName)} />
      {config.label}
    </Badge>
  )

  if (!showTooltip) {
    return badge
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>{badge}</TooltipTrigger>
      <TooltipContent>
        <p>{config.description}</p>
        {timestamp && (
          <p className="text-xs text-muted-foreground mt-1">
            {new Date(timestamp).toLocaleString()}
          </p>
        )}
      </TooltipContent>
    </Tooltip>
  )
}

/**
 * Check if an execution status is terminal (no longer changing)
 */
export function isTerminalStatus(status: ExecutionStatus): boolean {
  return status === 'completed' || status === 'failed' || status === 'cancelled'
}
