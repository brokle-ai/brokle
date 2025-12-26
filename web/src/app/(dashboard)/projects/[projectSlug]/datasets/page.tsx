'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Datasets } from '@/features/datasets'

export default function DatasetsPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Datasets projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
