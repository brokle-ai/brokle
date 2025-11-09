'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Traces } from '@/features/traces'
import type { ProjectParams } from '@/features/organizations'

export default function ProjectTracesPage() {
  const params = useParams() as ProjectParams

  return (
    <>
      <DashboardHeader />
      <Main>
        <Traces projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
