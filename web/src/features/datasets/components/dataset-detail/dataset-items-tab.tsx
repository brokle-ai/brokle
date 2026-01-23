'use client'

import { useDatasetDetail } from '../../context/dataset-detail-context'
import { DatasetItemTable } from '../dataset-item-table'
import { AddDatasetItemDialog } from '../add-dataset-item-dialog'
import { ImportJsonDialog } from '../import-json-dialog'
import { ImportCsvDialog } from '../import-csv-dialog'
import { Skeleton } from '@/components/ui/skeleton'

export function DatasetItemsTab() {
  const { dataset, isLoading, projectId, datasetId } = useDatasetDetail()

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Skeleton className="h-6 w-16" />
          <div className="flex gap-2">
            <Skeleton className="h-9 w-24" />
            <Skeleton className="h-9 w-24" />
            <Skeleton className="h-9 w-24" />
          </div>
        </div>
        <Skeleton className="h-[400px]" />
      </div>
    )
  }

  if (!projectId || !dataset) {
    return null
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium">Items</h2>
        <div className="flex items-center gap-2">
          <ImportJsonDialog projectId={projectId} datasetId={datasetId} />
          <ImportCsvDialog projectId={projectId} datasetId={datasetId} />
          <AddDatasetItemDialog projectId={projectId} datasetId={datasetId} />
        </div>
      </div>
      <DatasetItemTable projectId={projectId} datasetId={datasetId} />
    </div>
  )
}
