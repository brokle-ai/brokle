'use client'

import * as React from 'react'
import { usePathname } from 'next/navigation'
import { OrganizationSelector } from './organization-selector'
import { ProjectSelector } from './project-selector'
import { cn } from '@/lib/utils'

interface ContextNavbarProps {
  className?: string
}

export function ContextNavbar({ className }: ContextNavbarProps) {
  const pathname = usePathname()
  
  // Always show organization selector
  // Show project selector on project pages
  const showProjectSelector = pathname.startsWith('/projects/')

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <OrganizationSelector />
      {showProjectSelector && <ProjectSelector />}
    </div>
  )
}