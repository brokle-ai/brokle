'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationDangerSection } from '@/features/organizations'

export default function OrganizationDangerPage() {
  return (
    <ContentSection
      title="Danger Zone"
      description="Irreversible and destructive actions for this organization."
    >
      <OrganizationDangerSection />
    </ContentSection>
  )
}
