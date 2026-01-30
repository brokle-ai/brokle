'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Evaluators } from '@/features/evaluators'

export default function EvaluatorsPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Evaluators projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
