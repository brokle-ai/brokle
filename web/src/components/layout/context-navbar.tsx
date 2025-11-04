'use client'

import * as React from 'react'
import { Slash } from 'lucide-react'
import { useWorkspace } from '@/context/workspace-context'
import { OrganizationSelector } from './organization-selector'
import { ProjectSelector } from './project-selector'
import { cn } from '@/lib/utils'

interface ContextNavbarProps {
  className?: string
}

export function ContextNavbar({ className }: ContextNavbarProps) {
  const { currentProject } = useWorkspace()

  // Show project selector when there's an active project in context
  const showProjectSelector = !!currentProject

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