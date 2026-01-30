'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { FlaskConical } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/layout/page-header'
import { Skeleton } from '@/components/ui/skeleton'
import { useDatasetDetail } from '../../context/dataset-detail-context'
import { DatasetVersionManager } from '../dataset-version-manager'

export function DatasetDetailHeader() {
  const { dataset, isLoading, projectSlug, projectId, datasetId } = useDatasetDetail()

  if (isLoading) {
    return (
      <div className="space-y-2 pb-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-72" />
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
    <PageHeader
      title={dataset.name}
      backHref={`/projects/${projectSlug}/datasets`}
      description={dataset.description}
      metadata={`Created ${formatDistanceToNow(new Date(dataset.created_at), { addSuffix: true })}`}
    >
      <DatasetVersionManager projectId={projectId} datasetId={datasetId} />
      <Button asChild>
        <Link href={`/projects/${projectSlug}/experiments/new?datasetId=${datasetId}`}>
          <FlaskConical className="mr-2 h-4 w-4" />
          Run Experiment
        </Link>
      </Button>
    </PageHeader>
  )
}
