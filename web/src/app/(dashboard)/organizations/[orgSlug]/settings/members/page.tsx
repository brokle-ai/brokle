'use client'

import { ContentSection } from '@/features/settings'
import { MemberManagement } from '@/features/organizations'

export default function OrganizationMembersPage() {
  return (
    <ContentSection
      title="Members"
      description="Manage team members, roles, and invitations."
    >
      <MemberManagement />
    </ContentSection>
  )
}
