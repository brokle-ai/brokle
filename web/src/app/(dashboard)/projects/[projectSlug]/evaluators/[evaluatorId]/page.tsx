'use client'

import { use } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { EvaluatorDetail } from '@/features/evaluators'

interface EvaluatorDetailPageProps {
  params: Promise<{ projectSlug: string; evaluatorId: string }>
}

export default function EvaluatorDetailPage({ params }: EvaluatorDetailPageProps) {
  const { projectSlug, evaluatorId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main>
        <EvaluatorDetail projectSlug={projectSlug} evaluatorId={evaluatorId} />
      </Main>
    </>
  )
}
