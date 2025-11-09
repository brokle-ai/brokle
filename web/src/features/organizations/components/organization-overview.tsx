'use client'

import { useWorkspace } from '@/context/workspace-context'
import { ProjectGrid } from './project-grid'

export function OrganizationOverview() {
  const { currentOrganization } = useWorkspace()

  if (!currentOrganization) {
    return <div>No organization selected</div>
  }

  return <ProjectGrid />
}
