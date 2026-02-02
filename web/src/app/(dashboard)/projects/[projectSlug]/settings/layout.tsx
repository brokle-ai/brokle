'use client'

import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { PageHeader } from '@/components/layout/page-header'
import { UnifiedSettingsNav } from '@/features/settings'

export default function ProjectSettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <>
      <DashboardHeader />

      <Main fixed>
        <PageHeader title="Settings" />

        {/* Two-column layout: UnifiedSettingsNav on left, content on right */}
        <div className='flex flex-1 flex-col space-y-2 overflow-hidden md:space-y-2 lg:flex-row lg:space-y-0 lg:space-x-6'>
          <aside className='top-0 lg:sticky lg:w-1/5'>
            <UnifiedSettingsNav />
          </aside>
          <div className='flex w-full overflow-y-hidden p-1'>
            {children}
          </div>
        </div>
      </Main>
    </>
  )
}
