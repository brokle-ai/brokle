'use client'

import { Badge } from '@/components/ui/badge'
import { Play, Pause, CircleOff } from 'lucide-react'
import type { EvaluatorStatus } from '../types'

interface EvaluatorStatusBadgeProps {
  status: EvaluatorStatus
  className?: string
}

const statusConfig: Record<
  EvaluatorStatus,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; icon: typeof Play }
> = {
  active: {
    label: 'Active',
    variant: 'default',
    icon: Play,
  },
  inactive: {
    label: 'Inactive',
    variant: 'outline',
    icon: CircleOff,
  },
  paused: {
    label: 'Paused',
    variant: 'secondary',
    icon: Pause,
  },
}

export function EvaluatorStatusBadge({ status, className }: EvaluatorStatusBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon

  return (
    <Badge variant={config.variant} className={className}>
      <Icon className="mr-1 h-3 w-3" />
      {config.label}
    </Badge>
  )
}
