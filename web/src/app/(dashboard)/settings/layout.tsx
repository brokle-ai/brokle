'use client'

import { Separator } from '@/components/ui/separator'
import { SidebarProvider } from '@/components/ui/sidebar'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { TopNav } from '@/components/layout/top-nav'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { SettingsNav } from '@/components/settings/settings-nav'

const topNav = [
  {
    title: 'Dashboard',
    href: '/',
    isActive: false,
    disabled: false,
  },
  {
    title: 'Settings',
    href: '/settings',
    isActive: true,
    disabled: false,
  },
]

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <SidebarProvider>
      {/* ===== Top Heading ===== */}
      <Header>
        <TopNav links={topNav} />
        <div className='ml-auto flex items-center space-x-4'>
          <Search />
          <ThemeSwitch className="hidden sm:flex" />
          <ProfileDropdown className="hidden sm:flex" />
        </div>
      </Header>

      {/* ===== Main ===== */}
      <Main fixed>
        <div className='space-y-0.5'>
          <h1 className='text-2xl font-bold tracking-tight md:text-3xl'>
            Settings
          </h1>
          <p className='text-muted-foreground'>
            Manage your account settings and set e-mail preferences.
          </p>
        </div>
        <Separator className='my-4 lg:my-6' />
        <div className='flex flex-1 flex-col space-y-2 overflow-hidden md:space-y-2 lg:flex-row lg:space-y-0 lg:space-x-12'>
          <aside className='top-0 lg:sticky lg:w-1/5'>
            <SettingsNav />
          </aside>
          <div className='flex w-full overflow-y-hidden p-1'>
            {children}
          </div>
        </div>
      </Main>
    </SidebarProvider>
  )
}