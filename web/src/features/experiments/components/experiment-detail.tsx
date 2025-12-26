'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/layout/page-header'
import { ExperimentDetailProvider, useExperimentDetail } from '../context/experiment-detail-context'
import { ExperimentDetailDialogs } from './experiment-detail-dialogs'
import { ExperimentDetailSkeleton } from './experiment-detail-skeleton'
import { ExperimentItemTable } from './experiment-item-table'
import { ExperimentStatusBadge } from './experiment-status-badge'

interface ExperimentDetailProps {
  projectSlug: string
  experimentId: string
}

export function ExperimentDetail({ projectSlug, experimentId }: ExperimentDetailProps) {
  return (
    <ExperimentDetailProvider projectSlug={projectSlug} experimentId={experimentId}>
      <ExperimentDetailContent />
    </ExperimentDetailProvider>
  )
}

function ExperimentDetailContent() {
  const { experiment, isLoading, projectSlug, projectId, experimentId, setOpen } = useExperimentDetail()

  if (isLoading) {
    return <ExperimentDetailSkeleton />
  }

  if (!projectId) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">No project selected</p>
      </div>
    )
  }

  if (!experiment) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <p className="text-lg font-medium">Experiment not found</p>
        <Link
          href={`/projects/${projectSlug}/experiments`}
          className="text-sm text-muted-foreground hover:underline mt-2"
        >
          Back to experiments
        </Link>
      </div>
    )
  }

  // Build metadata string
  const metadataParts = [
    `Created ${formatDistanceToNow(new Date(experiment.created_at), { addSuffix: true })}`,
  ]
  if (experiment.started_at) {
    metadataParts.push(`Started ${formatDistanceToNow(new Date(experiment.started_at), { addSuffix: true })}`)
  }
  if (experiment.completed_at) {
    metadataParts.push(`Completed ${formatDistanceToNow(new Date(experiment.completed_at), { addSuffix: true })}`)
  }

  return (
    <>
      <div className="space-y-6">
        <PageHeader
          title={experiment.name}
          backHref={`/projects/${projectSlug}/experiments`}
          description={experiment.description}
          metadata={metadataParts.join(' • ')}
          badges={<ExperimentStatusBadge status={experiment.status} />}
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
            <h2 className="text-lg font-medium">Experiment Items</h2>
          </div>
          <ExperimentItemTable
            projectId={projectId}
            projectSlug={projectSlug}
            experimentId={experimentId}
          />
        </div>
      </div>

      <ExperimentDetailDialogs />
    </>
  )
}
