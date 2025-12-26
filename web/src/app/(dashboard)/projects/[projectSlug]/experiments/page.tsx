'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Experiments } from '@/features/experiments'

export default function ExperimentsPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Experiments projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
