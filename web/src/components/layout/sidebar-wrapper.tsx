'use client'

import { SidebarProvider } from '@/components/ui/sidebar'

interface SidebarWrapperProps {
  children: React.ReactNode
  sidebar: React.ReactNode
}

export function SidebarWrapper({ children, sidebar }: SidebarWrapperProps) {
  return (
    <SidebarProvider>
      {sidebar}
      <main className="flex-1 overflow-hidden peer-data-[state=collapsed]:w-[calc(100%-var(--sidebar-width-icon))] peer-data-[state=expanded]:w-[calc(100%-var(--sidebar-width))] transition-[width] duration-200 ease-linear">
        {children}
      </main>
    </SidebarProvider>
  )
}