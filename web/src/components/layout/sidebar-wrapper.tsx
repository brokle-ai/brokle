'use client'

import { cn } from '@/lib/utils'
import { getCookie } from '@/lib/utils/cookies'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'

interface SidebarWrapperProps {
  children: React.ReactNode
  sidebar: React.ReactNode
}

export function SidebarWrapper({ children, sidebar }: SidebarWrapperProps) {
  // Read sidebar state from cookie to persist user preference
  const defaultOpen = getCookie('sidebar_state') !== 'false'

  return (
    <SidebarProvider defaultOpen={defaultOpen}>
      {sidebar}
      <SidebarInset
        id="content"
        className={cn(
          '@container/content',
          'has-[[data-layout=fixed]]:h-svh'
        )}
      >
        {children}
      </SidebarInset>
    </SidebarProvider>
  )
}