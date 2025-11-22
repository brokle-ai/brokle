'use client'

import { ContentSection } from '@/features/settings'
import { OrganizationMembersSection } from '@/features/organizations'

export default function OrganizationMembersPage() {
  return (
    <ContentSection
      title="Members"
      description="Manage team members, roles, and invitations."
    >
      <OrganizationMembersSection />
    </ContentSection>
  )
}
