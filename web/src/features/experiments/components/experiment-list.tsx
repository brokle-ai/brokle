'use client'

import { ExperimentCard } from './experiment-card'
import type { Experiment } from '../types'

interface ExperimentListProps {
  data: Experiment[]
}

export function ExperimentList({ data }: ExperimentListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.map((experiment) => (
        <ExperimentCard key={experiment.id} experiment={experiment} />
      ))}
    </div>
  )
}
