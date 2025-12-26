'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Scores } from '@/features/scores'

export default function ScoresPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Scores projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
