'use client'

import { ContentSection } from '@/features/settings'
import { ProjectDangerSection } from '@/features/projects'

export default function ProjectDangerPage() {
  return (
    <ContentSection
      title="Danger Zone"
      description="Irreversible and destructive actions for this project."
    >
      <ProjectDangerSection />
    </ContentSection>
  )
}
