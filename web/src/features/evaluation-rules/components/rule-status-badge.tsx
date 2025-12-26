'use client'

import { Badge } from '@/components/ui/badge'
import { Play, Pause, CircleOff } from 'lucide-react'
import type { RuleStatus } from '../types'

interface RuleStatusBadgeProps {
  status: RuleStatus
  className?: string
}

const statusConfig: Record<
  RuleStatus,
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

export function RuleStatusBadge({ status, className }: RuleStatusBadgeProps) {
  const config = statusConfig[status]
  const Icon = config.icon

  return (
    <Badge variant={config.variant} className={className}>
      <Icon className="mr-1 h-3 w-3" />
      {config.label}
    </Badge>
  )
}
