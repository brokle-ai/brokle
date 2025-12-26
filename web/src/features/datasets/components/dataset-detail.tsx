'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/layout/page-header'
import { DatasetDetailProvider, useDatasetDetail } from '../context/dataset-detail-context'
import { DatasetDetailDialogs } from './dataset-detail-dialogs'
import { DatasetDetailSkeleton } from './dataset-detail-skeleton'
import { DatasetItemTable } from './dataset-item-table'
import { AddDatasetItemDialog } from './add-dataset-item-dialog'

interface DatasetDetailProps {
  projectSlug: string
  datasetId: string
}

export function DatasetDetail({ projectSlug, datasetId }: DatasetDetailProps) {
  return (
    <DatasetDetailProvider projectSlug={projectSlug} datasetId={datasetId}>
      <DatasetDetailContent />
    </DatasetDetailProvider>
  )
}

function DatasetDetailContent() {
  const { dataset, isLoading, projectSlug, projectId, datasetId, setOpen } = useDatasetDetail()

  if (isLoading) {
    return <DatasetDetailSkeleton />
  }

  if (!projectId) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  if (!dataset) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-lg font-medium">Dataset not found</p>
        <Link
          href={`/projects/${projectSlug}/datasets`}
          className="text-sm text-muted-foreground hover:underline mt-2"
        >
          Back to datasets
        </Link>
      </div>
    )
  }

  return (
    <>
      <div className="space-y-6">
        <PageHeader
          title={dataset.name}
          backHref={`/projects/${projectSlug}/datasets`}
          description={dataset.description}
          metadata={`Created ${formatDistanceToNow(new Date(dataset.created_at), { addSuffix: true })}`}
        >
          <Button variant="outline" onClick={() => setOpen('edit')}>
            <Pencil className="mr-2 h-4 w-4" />
            Edit
          </Button>
          <Button
            variant="outline"
            className="text-destructive hover:text-destructive hover:bg-destructive/10"
            onClick={() => setOpen('delete')}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
        </PageHeader>

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-medium">Items</h2>
            <AddDatasetItemDialog projectId={projectId} datasetId={datasetId} />
          </div>
          <DatasetItemTable projectId={projectId} datasetId={datasetId} />
        </div>
      </div>

      <DatasetDetailDialogs />
    </>
  )
}
