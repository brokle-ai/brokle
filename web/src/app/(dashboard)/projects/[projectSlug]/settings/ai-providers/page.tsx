'use client'

import { ContentSection } from '@/features/settings'
import { AIProvidersSettings } from '@/features/ai-providers'
import { useWorkspace } from '@/context/workspace-context'

export default function ProjectAIProvidersPage() {
  const { currentProject } = useWorkspace()

  if (!currentProject) {
    return null
  }

  return (
    <ContentSection
      title="AI Providers"
      description="Configure API credentials for AI model providers used in the playground."
    >
      <AIProvidersSettings projectId={currentProject.id} />
    </ContentSection>
  )
}
