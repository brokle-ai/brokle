'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Prompts } from '@/features/prompts'
import { useOrganizationOnly } from '@/features/organizations'

export default function ProjectPromptsPage() {
  const params = useParams<{ projectSlug: string }>()
  const { currentOrganization } = useOrganizationOnly()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Prompts
          projectSlug={params.projectSlug}
          orgSlug={currentOrganization?.slug || ''}
        />
      </Main>
    </>
  )
}
