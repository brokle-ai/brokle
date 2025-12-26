'use client'

import { DatasetCard } from './dataset-card'
import type { Dataset } from '../types'

interface DatasetListProps {
  data: Dataset[]
}

export function DatasetList({ data }: DatasetListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.map((dataset) => (
        <DatasetCard key={dataset.id} dataset={dataset} />
      ))}
    </div>
  )
}
