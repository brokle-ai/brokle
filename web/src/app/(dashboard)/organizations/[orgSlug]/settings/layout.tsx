'use client'

import { Separator } from '@/components/ui/separator'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { OrganizationSettingsNav } from '@/features/organizations'

export default function OrganizationSettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <>
      <DashboardHeader />

      <Main fixed>

        <div className='space-y-0.5'>
          <h1 className='text-xl font-bold tracking-tight md:text-2xl'>
            Organization Settings
          </h1>
          <p className='text-muted-foreground'>
            Manage your organization details, members, and security settings.
          </p>
        </div>
        <Separator className='my-4 lg:my-6' />

        {/* Two-column layout: OrganizationSettingsNav on left, content on right */}
        <div className='flex flex-1 flex-col space-y-2 overflow-hidden md:space-y-2 lg:flex-row lg:space-y-0 lg:space-x-12'>
          <aside className='top-0 lg:sticky lg:w-1/5'>
            <OrganizationSettingsNav />
          </aside>
          <div className='flex w-full overflow-y-hidden p-1'>
            {children}
          </div>
        </div>
      </Main>
    </>
  )
}
