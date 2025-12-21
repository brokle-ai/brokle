'use client'

import { ContentSection } from '@/features/settings'
import { ScoreConfigsSection } from '@/features/evaluations'
import { useWorkspace } from '@/context/workspace-context'

export default function ScoreConfigsSettingsPage() {
  const { currentProject } = useWorkspace()

  if (!currentProject) {
    return null
  }

  return (
    <ContentSection
      title="Score Configs"
      description="Define validation rules and schemas for evaluation scores."
    >
      <ScoreConfigsSection projectId={currentProject.id} />
    </ContentSection>
  )
}
