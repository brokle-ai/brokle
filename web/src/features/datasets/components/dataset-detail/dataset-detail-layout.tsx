'use client'

import type { ReactNode } from 'react'
import { DatasetDetailProvider } from '../../context/dataset-detail-context'
import { DatasetDetailHeader } from './dataset-detail-header'
import { DatasetDetailTabs } from './dataset-detail-tabs'
import { DatasetDetailDialogs } from '../dataset-detail-dialogs'

interface DatasetDetailLayoutProps {
  children: ReactNode
  projectSlug: string
  datasetId: string
}

export function DatasetDetailLayout({
  children,
  projectSlug,
  datasetId,
}: DatasetDetailLayoutProps) {
  return (
    <DatasetDetailProvider projectSlug={projectSlug} datasetId={datasetId}>
      <div className="flex h-full flex-col">
        <DatasetDetailHeader />
        <DatasetDetailTabs />
        <div className="flex-1 overflow-auto py-4">
          {children}
        </div>
      </div>
      <DatasetDetailDialogs />
    </DatasetDetailProvider>
  )
}
