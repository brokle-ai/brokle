'use client'

import { useDatasetsQuery } from '../hooks/use-datasets'
import { DatasetCard } from './dataset-card'
import { Skeleton } from '@/components/ui/skeleton'
import { Database } from 'lucide-react'

interface DatasetListProps {
  projectId: string
  projectSlug: string
}

export function DatasetList({ projectId, projectSlug }: DatasetListProps) {
  const { data: datasets, isLoading } = useDatasetsQuery(projectId)

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-32" />
        ))}
      </div>
    )
  }

  if (!datasets?.length) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <Database className="h-12 w-12 text-muted-foreground/50 mb-4" />
        <h3 className="text-lg font-medium">No datasets yet</h3>
        <p className="text-sm text-muted-foreground mt-1">
          Create a dataset to start batch evaluations
        </p>
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {datasets.map((dataset) => (
        <DatasetCard key={dataset.id} dataset={dataset} projectId={projectId} projectSlug={projectSlug} />
      ))}
    </div>
  )
}
