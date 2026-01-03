'use client'

import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { Prompts } from '@/features/prompts'

export default function ProjectPromptsPage() {
  const params = useParams<{ projectSlug: string }>()

  return (
    <>
      <DashboardHeader />
      <Main>
        <Prompts projectSlug={params.projectSlug} />
      </Main>
    </>
  )
}
