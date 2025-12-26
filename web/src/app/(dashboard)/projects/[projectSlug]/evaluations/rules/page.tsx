'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Rules } from '@/features/evaluation-rules'

export default function EvaluationRulesPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Rules projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
