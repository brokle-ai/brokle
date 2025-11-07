'use client'

import { ContentSection } from '@/features/settings'
import { ProjectIntegrationsSection } from '@/features/projects'

export default function ProjectIntegrationsPage() {
  return (
    <ContentSection
      title="Integrations"
      description="Connect external services to enhance your project capabilities."
    >
      <ProjectIntegrationsSection />
    </ContentSection>
  )
}
