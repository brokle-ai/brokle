'use client'

import { useEvaluationRulesQuery } from '../hooks/use-evaluation-rules'
import { RuleCard } from './rule-card'
import { Skeleton } from '@/components/ui/skeleton'
import { Scale } from 'lucide-react'
import type { EvaluationRule, RuleListParams } from '../types'

interface RuleListProps {
  projectId: string
  projectSlug: string
  params?: RuleListParams
  onEdit?: (rule: EvaluationRule) => void
}

export function RuleList({ projectId, projectSlug, params, onEdit }: RuleListProps) {
  const { data, isLoading } = useEvaluationRulesQuery(projectId, params)

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-40" />
        ))}
      </div>
    )
  }

  if (!data?.rules?.length) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Scale className="h-12 w-12 text-muted-foreground/50 mb-4" />
        <h3 className="text-lg font-medium">No evaluation rules yet</h3>
        <p className="text-sm text-muted-foreground mt-1 max-w-sm">
          Create evaluation rules to automatically score incoming spans using LLM,
          built-in scorers, or regex patterns.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data.rules.map((rule) => (
          <RuleCard
            key={rule.id}
            rule={rule}
            projectId={projectId}
            projectSlug={projectSlug}
            onEdit={onEdit}
          />
        ))}
      </div>
      {data.total > data.limit && (
        <div className="text-center text-sm text-muted-foreground">
          Showing {data.rules.length} of {data.total} rules
        </div>
      )}
    </div>
  )
}
