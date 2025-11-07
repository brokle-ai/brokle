'use client'

import { ContentSection } from '@/features/settings'
import { ProjectGeneralSection } from '@/features/projects'

export default function ProjectSettingsPage() {
  return (
    <ContentSection
      title="General"
      description="Manage basic project information and configuration settings."
    >
      <ProjectGeneralSection />
    </ContentSection>
  )
}