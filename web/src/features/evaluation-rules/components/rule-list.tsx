'use client'

import { RuleCard } from './rule-card'
import type { EvaluationRule } from '../types'

interface RuleListProps {
  data: EvaluationRule[]
}

export function RuleList({ data }: RuleListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.map((rule) => (
        <RuleCard key={rule.id} rule={rule} />
      ))}
    </div>
  )
}
