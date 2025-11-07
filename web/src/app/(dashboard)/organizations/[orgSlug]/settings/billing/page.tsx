'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationBillingSection } from '@/features/organizations'

export default function OrganizationBillingPage() {
  return (
    <ContentSection
      title="Billing & Usage"
      description="Manage your subscription, usage, and payment information."
    >
      <OrganizationBillingSection />
    </ContentSection>
  )
}
