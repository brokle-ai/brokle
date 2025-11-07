'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationGeneralSection } from '@/features/organizations'

export default function OrganizationSettingsPage() {
  return (
    <ContentSection
      title="General"
      description="Manage your organization details and configuration."
    >
      <OrganizationGeneralSection />
    </ContentSection>
  )
}
