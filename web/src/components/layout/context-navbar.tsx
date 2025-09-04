'use client'

import * as React from 'react'
import { useOrganization } from '@/context/organization-context'
import { OrganizationSelector } from './organization-selector'
import { ProjectSelector } from './project-selector'
import { cn } from '@/lib/utils'

interface ContextNavbarProps {
  className?: string
}

export function ContextNavbar({ className }: ContextNavbarProps) {
  const { currentProject } = useOrganization()

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <OrganizationSelector />
      {currentProject && <ProjectSelector />}
    </div>
  )
}