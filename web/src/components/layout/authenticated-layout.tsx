'use client'

import { useState, useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { SearchProvider } from '@/context/search-context'
import { SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { ConfigDrawer } from '@/components/config-drawer'
import SkipToMain from '@/components/skip-to-main'
import type { User } from '@/types/auth'

interface Props {
  children: React.ReactNode
  defaultSidebarOpen?: boolean
  serverUser?: User | null
}

export function AuthenticatedLayout({ 
  children, 
  defaultSidebarOpen = true,
  serverUser: _serverUser // eslint-disable-line @typescript-eslint/no-unused-vars
}: Props) {
  const [mounted, setMounted] = useState(false)
  const pathname = usePathname()
  
  // Auto-detect onboarding routes to hide sidebar
  const shouldHideSidebar = pathname.includes('/onboarding')

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <SearchProvider>
      <SidebarProvider defaultOpen={defaultSidebarOpen}>
        <SkipToMain />
        {!shouldHideSidebar && <AppSidebar />}
        <div
          id='content'
          className={cn(
            'w-full max-w-full',
            !shouldHideSidebar && 'ml-auto',
            !shouldHideSidebar && 'peer-data-[state=collapsed]:w-[calc(100%-var(--sidebar-width-icon))]',
            !shouldHideSidebar && 'peer-data-[state=expanded]:w-[calc(100%-var(--sidebar-width))]',
            !shouldHideSidebar && 'sm:transition-[width] sm:duration-200 sm:ease-linear',
            'flex h-svh flex-col',
            'group-data-[scroll-locked=1]/body:h-full',
            'has-[main.fixed-main]:group-data-[scroll-locked=1]/body:h-svh'
          )}
        >
          {children}
          {mounted && <ConfigDrawer />}
        </div>
      </SidebarProvider>
    </SearchProvider>
  )
}
