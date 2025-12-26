'use client'

import { use } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { ExperimentDetail } from '@/features/experiments'

interface ExperimentDetailPageProps {
  params: Promise<{ projectSlug: string; experimentId: string }>
}

export default function ExperimentDetailPage({ params }: ExperimentDetailPageProps) {
  const { projectSlug, experimentId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main>
        <ExperimentDetail projectSlug={projectSlug} experimentId={experimentId} />
      </Main>
    </>
  )
}
