'use client'

import { EvaluatorCard } from './evaluator-card'
import type { WizardEvaluator } from '../../../types'

interface EvaluatorListProps {
  evaluators: WizardEvaluator[]
  emptyMessage: string
}

export function EvaluatorList({ evaluators, emptyMessage }: EvaluatorListProps) {
  if (evaluators.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-sm text-muted-foreground">{emptyMessage}</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {evaluators.map((evaluator) => (
        <EvaluatorCard key={evaluator.id} evaluator={evaluator} />
      ))}
    </div>
  )
}
