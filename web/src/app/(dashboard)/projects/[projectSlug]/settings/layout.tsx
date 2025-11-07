'use client'

import { Separator } from '@/components/ui/separator'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ProjectSettingsNav } from '@/features/projects'

export default function ProjectSettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <>
      <Header>
        <div className='ml-auto flex items-center space-x-4'>
          <Search />
          <ThemeSwitch className="hidden sm:flex" />
          <ProfileDropdown className="hidden sm:flex" />
        </div>
      </Header>

      <Main fixed>

        <div className='space-y-0.5'>
          <h1 className='text-2xl font-bold tracking-tight md:text-3xl'>
            Project Settings
          </h1>
          <p className='text-muted-foreground'>
            Manage your project configuration, API keys, and security settings.
          </p>
        </div>
        <Separator className='my-4 lg:my-6' />

        {/* Two-column layout: ProjectSettingsNav on left, content on right */}
        <div className='flex flex-1 flex-col space-y-2 overflow-hidden md:space-y-2 lg:flex-row lg:space-y-0 lg:space-x-12'>
          <aside className='top-0 lg:sticky lg:w-1/5'>
            <ProjectSettingsNav />
          </aside>
          <div className='flex w-full overflow-y-hidden p-1'>
            {children}
          </div>
        </div>
      </Main>
    </>
  )
}
