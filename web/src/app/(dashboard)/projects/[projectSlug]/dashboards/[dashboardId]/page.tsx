'use client'

import { use } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { DashboardDetail } from '@/features/dashboards'

interface DashboardDetailPageProps {
  params: Promise<{ projectSlug: string; dashboardId: string }>
}

export default function DashboardDetailPage({ params }: DashboardDetailPageProps) {
  const { projectSlug, dashboardId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main>
        <DashboardDetail projectSlug={projectSlug} dashboardId={dashboardId} />
      </Main>
    </>
  )
}
