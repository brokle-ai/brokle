'use client'

import { ContentSection } from '@/features/settings'
import { ProjectSecuritySection } from '@/features/projects'

export default function ProjectSecurityPage() {
  return (
    <ContentSection
      title="Security"
      description="Configure advanced security features and access control for your project."
    >
      <ProjectSecuritySection />
    </ContentSection>
  )
}
