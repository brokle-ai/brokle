'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Database, History, Settings } from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDatasetDetail } from '../../context/dataset-detail-context'

const TABS = [
  { value: 'items', label: 'Items', icon: Database, href: '' },
  { value: 'versions', label: 'Versions', icon: History, href: '/versions' },
  { value: 'settings', label: 'Settings', icon: Settings, href: '/settings' },
] as const

export function DatasetDetailTabs() {
  const pathname = usePathname()
  const { projectSlug, datasetId, dataset } = useDatasetDetail()

  if (!dataset) {
    return null
  }

  const basePath = `/projects/${projectSlug}/datasets/${datasetId}`

  // Determine active tab from pathname
  const activeTab = TABS.find(tab => {
    if (tab.href === '') {
      // Items tab is active when pathname exactly matches base path
      return pathname === basePath
    }
    return pathname.includes(tab.href)
  })?.value || 'items'

  return (
    <Tabs value={activeTab} className="w-full">
      <TabsList>
        {TABS.map((tab) => {
          const Icon = tab.icon
          return (
            <TabsTrigger key={tab.value} value={tab.value} asChild>
              <Link href={`${basePath}${tab.href}`} className="gap-1.5">
                <Icon className="h-4 w-4" />
                {tab.label}
              </Link>
            </TabsTrigger>
          )
        })}
      </TabsList>
    </Tabs>
  )
}
