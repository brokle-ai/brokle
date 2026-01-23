'use client'

import { use } from 'react'
import type { ReactNode } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { DatasetDetailLayout } from '@/features/datasets'

interface DatasetDetailLayoutProps {
  children: ReactNode
  params: Promise<{ projectSlug: string; datasetId: string }>
}

export default function DatasetDetailLayoutPage({ children, params }: DatasetDetailLayoutProps) {
  const { projectSlug, datasetId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main fixed>
        <DatasetDetailLayout projectSlug={projectSlug} datasetId={datasetId}>
          {children}
        </DatasetDetailLayout>
      </Main>
    </>
  )
}
