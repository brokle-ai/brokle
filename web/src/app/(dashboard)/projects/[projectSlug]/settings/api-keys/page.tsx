'use client'

import { ContentSection } from '@/features/settings'
import { ProjectAPIKeysSection } from '@/features/projects'

export default function ProjectAPIKeysPage() {
  return (
    <ContentSection
      title="API Keys"
      description="Manage API keys for accessing your project programmatically."
    >
      <ProjectAPIKeysSection />
    </ContentSection>
  )
}
