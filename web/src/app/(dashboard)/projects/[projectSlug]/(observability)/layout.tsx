'use client'

import { usePathname, useParams } from 'next/navigation'
import Link from 'next/link'
import { Activity, Layers, MessageSquare } from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import type { ReactNode } from 'react'

const TABS = [
  { value: 'traces', label: 'Traces', icon: Activity, href: 'traces' },
  { value: 'spans', label: 'Spans', icon: Layers, href: 'spans' },
  { value: 'sessions', label: 'Sessions', icon: MessageSquare, href: 'sessions' },
] as const

// Pattern to detect detail pages (e.g., /traces/[traceId])
const DETAIL_PAGE_PATTERN = /\/(traces|spans|sessions)\/[^/]+$/

export default function ObservabilityLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const params = useParams()
  const projectSlug = params.projectSlug as string

  // Check if we're on a detail page - if so, don't render the shared layout
  // Detail pages have their own DashboardHeader and Main
  const isDetailPage = DETAIL_PAGE_PATTERN.test(pathname)

  if (isDetailPage) {
    return <>{children}</>
  }

  // Determine active tab from pathname
  const activeTab = TABS.find(tab => pathname.includes(`/${tab.href}`))?.value || 'traces'

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="flex items-center justify-between mt-2 mb-2">
          <h1 className="text-lg font-semibold">Traces</h1>
        </div>
        <Tabs value={activeTab} className="w-full">
          <TabsList className="mb-4">
            {TABS.map((tab) => {
              const Icon = tab.icon
              return (
                <TabsTrigger key={tab.value} value={tab.value} asChild>
                  <Link href={`/projects/${projectSlug}/${tab.href}`} className="gap-1.5">
                    <Icon className="h-4 w-4" />
                    {tab.label}
                  </Link>
                </TabsTrigger>
              )
            })}
          </TabsList>
        </Tabs>
        {children}
      </Main>
    </>
  )
}
