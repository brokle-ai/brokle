'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationGeneralSection } from '@/features/organizations'

export default function OrganizationSettingsPage() {
  return (
    <ContentSection
      title="Organization"
      description="Manage your organization details and configuration."
    >
      <OrganizationGeneralSection />
    </ContentSection>
  )
}
