'use client'

import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { SearchProvider } from '@/context/search-context'
import { OrgProvider } from '@/context/org-context'
import { ConfigDrawer } from '@/components/config-drawer'
import SkipToMain from '@/components/skip-to-main'
import type { User } from '@/types/auth'

interface Props {
  children: React.ReactNode
  serverUser?: User | null
}

export function AuthenticatedLayout({ 
  children, 
  serverUser: _serverUser // eslint-disable-line @typescript-eslint/no-unused-vars
}: Props) {
  const [mounted, setMounted] = useState(false)


  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <SearchProvider>
      <OrgProvider>
        <SkipToMain />
        <div
          id='content'
          className={cn(
            'w-full max-w-full',
            'flex h-svh flex-col',
            'group-data-[scroll-locked=1]/body:h-full',
            'has-[main.fixed-main]:group-data-[scroll-locked=1]/body:h-svh'
          )}
        >
          {children}
          {mounted && <ConfigDrawer />}
        </div>
      </OrgProvider>
    </SearchProvider>
  )
}
