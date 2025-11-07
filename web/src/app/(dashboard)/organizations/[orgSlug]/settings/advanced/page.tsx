'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationAdvancedSection } from '@/features/organizations'

export default function OrganizationAdvancedPage() {
  return (
    <ContentSection
      title="Advanced"
      description="Manage advanced settings, webhooks, and custom configurations."
    >
      <OrganizationAdvancedSection />
    </ContentSection>
  )
}
