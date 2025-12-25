'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { DashboardHeader } from '@/components/layout/dashboard-header'
import { Main } from '@/components/layout/main'
import { useProjectOnly } from '@/features/projects'
import {
  RuleList,
  CreateRuleDialog,
  EditRuleDialog,
} from '@/features/evaluation-rules'
import type { EvaluationRule } from '@/features/evaluation-rules'
import { Skeleton } from '@/components/ui/skeleton'

export default function EvaluationRulesPage() {
  const params = useParams<{ projectSlug: string }>()
  const { currentProject, hasProject, isLoading } = useProjectOnly()
  const [editingRule, setEditingRule] = useState<EvaluationRule | null>(null)

  if (isLoading) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <Skeleton className="h-8 w-32" />
              <Skeleton className="h-10 w-40" />
            </div>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-40" />
              ))}
            </div>
          </div>
        </Main>
      </>
    )
  }

  if (!hasProject || !currentProject) {
    return (
      <>
        <DashboardHeader />
        <Main>
          <div className="flex items-center justify-center py-12">
            <p className="text-muted-foreground">No project selected</p>
          </div>
        </Main>
      </>
    )
  }

  return (
    <>
      <DashboardHeader />
      <Main>
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold tracking-tight">Evaluation Rules</h1>
              <p className="text-muted-foreground">
                Automatically score incoming spans using LLM, built-in scorers, or regex patterns
              </p>
            </div>
            <CreateRuleDialog projectId={currentProject.id} />
          </div>
          <RuleList
            projectId={currentProject.id}
            projectSlug={params.projectSlug}
            onEdit={setEditingRule}
          />
        </div>
      </Main>

      <EditRuleDialog
        projectId={currentProject.id}
        rule={editingRule}
        open={!!editingRule}
        onOpenChange={(open) => !open && setEditingRule(null)}
      />
    </>
  )
}
