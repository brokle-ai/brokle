'use client'

import { use } from 'react'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { RuleDetail } from '@/features/evaluation-rules'

interface RuleDetailPageProps {
  params: Promise<{ projectSlug: string; ruleId: string }>
}

export default function RuleDetailPage({ params }: RuleDetailPageProps) {
  const { projectSlug, ruleId } = use(params)

  return (
    <>
      <DashboardHeader />
      <Main>
        <RuleDetail projectSlug={projectSlug} ruleId={ruleId} />
      </Main>
    </>
  )
}
