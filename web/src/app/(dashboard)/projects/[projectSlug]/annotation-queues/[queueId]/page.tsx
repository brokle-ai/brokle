'use client'

import { use } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { QueueDetail } from '@/features/annotation-queues'

interface QueueDetailPageProps {
  params: Promise<{ projectSlug: string; queueId: string }>
}

export default function QueueDetailPage({ params }: QueueDetailPageProps) {
  const { projectSlug, queueId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main>
        <QueueDetail projectSlug={projectSlug} queueId={queueId} />
      </Main>
    </>
  )
}
