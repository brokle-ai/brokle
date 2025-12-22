'use client'

import { Badge } from '@/components/ui/badge'
import { Clock, Play, CheckCircle2, XCircle } from 'lucide-react'
import type { ExperimentStatus } from '../types'

interface ExperimentStatusBadgeProps {
  status: ExperimentStatus
  className?: string
}

const statusConfig: Record<
  ExperimentStatus,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; icon: typeof Clock }
> = {
  pending: {
    label: 'Pending',
    variant: 'outline',
    icon: Clock,
  },
  running: {
    label: 'Running',
    variant: 'secondary',
    icon: Play,
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
}

export function ExperimentStatusBadge({ status, className }: ExperimentStatusBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon

  return (
    <Badge variant={config.variant} className={className}>
      <Icon className="mr-1 h-3 w-3" />
      {config.label}
    </Badge>
  )
}
