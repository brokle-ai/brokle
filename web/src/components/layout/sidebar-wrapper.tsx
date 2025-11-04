'use client'

import { cn } from '@/lib/utils'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'

interface SidebarWrapperProps {
  children: React.ReactNode
  sidebar: React.ReactNode
  defaultOpen?: boolean
}

export function SidebarWrapper({
  children,
  sidebar,
  defaultOpen = true
}: SidebarWrapperProps) {
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