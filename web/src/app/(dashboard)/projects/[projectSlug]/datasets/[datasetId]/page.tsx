'use client'

import { use } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { DatasetDetail } from '@/features/datasets'

interface DatasetDetailPageProps {
  params: Promise<{ projectSlug: string; datasetId: string }>
}

export default function DatasetDetailPage({ params }: DatasetDetailPageProps) {
  const { projectSlug, datasetId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main>
        <DatasetDetail projectSlug={projectSlug} datasetId={datasetId} />
      </Main>
    </>
  )
}
