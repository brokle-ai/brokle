import { useLocation } from '@tanstack/react-router'
import { Database, History, Settings } from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDatasetDetail } from '../../context/dataset-detail-context'

const TABS = [
  { value: 'items', label: 'Items', suffix: '' },
  { value: 'versions', label: 'Versions', suffix: '/versions' },
  { value: 'settings', label: 'Settings', suffix: '/settings' },
] as const

const ICONS = {
  items: Database,
  versions: History,
  settings: Settings,
} as const

export function DatasetDetailTabs() {
  const location = useLocation()
  const { orgId, projectId, datasetId, dataset } = useDatasetDetail()

  if (!dataset) {
    return null
  }

  const basePath = `/o/${orgId}/p/${projectId}/datasets/${datasetId}`
  const pathname = location.pathname

  const activeTab =
    TABS.find((tab) => {
      if (tab.suffix === '') return pathname === basePath
      return pathname.endsWith(tab.suffix)
    })?.value ?? 'items'

  return (
    <Tabs value={activeTab} className="w-full">
      <TabsList>
        {TABS.map((tab) => {
          const Icon = ICONS[tab.value]
          return (
            <TabsTrigger key={tab.value} value={tab.value} asChild>
              {/* Raw <a> to sidestep TanStack's typed routing for the
                  three sibling tab URLs — we know the paths are
                  co-located under the same $datasetId segment and
                  TanStack Router will still intercept same-origin
                  clicks. */}
              <a href={`${basePath}${tab.suffix}`} className="gap-1.5">
                <Icon className="h-4 w-4" />
                {tab.label}
              </a>
            </TabsTrigger>
          )
        })}
      </TabsList>
    </Tabs>
  )
}
