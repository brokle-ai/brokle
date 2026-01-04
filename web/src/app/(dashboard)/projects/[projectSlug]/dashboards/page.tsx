'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Dashboards } from '@/features/dashboards'

export default function DashboardsPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Dashboards projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
