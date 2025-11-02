'use client'

import * as React from 'react'
import { usePathname } from 'next/navigation'
import { Slash } from 'lucide-react'
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
    <nav
      className={cn("flex items-center gap-1.5", className)}
      aria-label="Context navigation"
    >
      <OrganizationSelector showPlanBadge />

      {showProjectSelector && (
        <>
          <span className="text-muted-foreground" aria-hidden="true">
            <Slash className="h-4 w-4" />
          </span>

          <ProjectSelector />
        </>
      )}
    </nav>
  )
}