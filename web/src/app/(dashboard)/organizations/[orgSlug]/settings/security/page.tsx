'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationSecuritySection } from '@/features/organizations'

export default function OrganizationSecurityPage() {
  return (
    <ContentSection
      title="Security"
      description="Configure security features and access control for your organization."
    >
      <OrganizationSecuritySection />
    </ContentSection>
  )
}
