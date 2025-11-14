'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Tasks } from '@/features/tasks'
import type { ProjectParams } from '@/features/organizations'

export default function ProjectTasksPage() {
  const params = useParams() as ProjectParams

  return (
    <>
      <DashboardHeader />
      <Main>
        <Tasks projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
