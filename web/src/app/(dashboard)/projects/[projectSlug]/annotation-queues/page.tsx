'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { AnnotationQueues } from '@/features/annotation-queues'

export default function AnnotationQueuesPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <AnnotationQueues projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
