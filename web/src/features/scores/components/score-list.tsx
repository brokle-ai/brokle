'use client'

import { ScoreBadge } from './score-badge'
import type { Score } from '../types'

interface ScoreListProps {
  scores: Score[]
  showReasons?: boolean
}

export function ScoreList({ scores, showReasons = false }: ScoreListProps) {
  if (scores.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No scores recorded</p>
    )
  }

  return (
    <div className="flex flex-wrap gap-2">
      {scores.map((score) => (
        <ScoreBadge key={score.id} score={score} showReason={showReasons} />
      ))}
    </div>
  )
}
