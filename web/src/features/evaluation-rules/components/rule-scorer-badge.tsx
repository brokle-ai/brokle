'use client'

import { Badge } from '@/components/ui/badge'
import { Bot, Code, Regex } from 'lucide-react'
import type { ScorerType } from '../types'

interface RuleScorerBadgeProps {
  scorerType: ScorerType
  className?: string
}

const scorerConfig: Record<
  ScorerType,
  { label: string; variant: 'default' | 'secondary' | 'outline'; icon: typeof Bot }
> = {
  llm: {
    label: 'LLM',
    variant: 'default',
    icon: Bot,
  },
  builtin: {
    label: 'Builtin',
    variant: 'secondary',
    icon: Code,
  },
  regex: {
    label: 'Regex',
    variant: 'outline',
    icon: Regex,
  },
}

export function RuleScorerBadge({ scorerType, className }: RuleScorerBadgeProps) {
  const config = scorerConfig[scorerType]
  const Icon = config.icon

  return (
    <Badge variant={config.variant} className={className}>
      <Icon className="mr-1 h-3 w-3" />
      {config.label}
    </Badge>
  )
}
