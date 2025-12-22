'use client'

import { useExperimentsQuery } from '../hooks/use-experiments'
import { ExperimentCard } from './experiment-card'
import { Skeleton } from '@/components/ui/skeleton'
import { FlaskConical } from 'lucide-react'
import type { Experiment } from '../types'

interface ExperimentListProps {
  projectId: string
  projectSlug: string
  onEdit?: (experiment: Experiment) => void
}

export function ExperimentList({ projectId, projectSlug, onEdit }: ExperimentListProps) {
  const { data: experiments, isLoading } = useExperimentsQuery(projectId)

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-32" />
        ))}
      </div>
    )
  }

  if (!experiments?.length) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <FlaskConical className="h-12 w-12 text-muted-foreground/50 mb-4" />
        <h3 className="text-lg font-medium">No experiments yet</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Experiments are created via the SDK using <code className="text-xs bg-muted px-1 py-0.5 rounded">brokle.evaluate()</code>
        </p>
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {experiments.map((experiment) => (
        <ExperimentCard
          key={experiment.id}
          experiment={experiment}
          projectId={projectId}
          projectSlug={projectSlug}
          onEdit={onEdit}
        />
      ))}
    </div>
  )
}
