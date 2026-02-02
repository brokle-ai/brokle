'use client'

import { ContentSection } from '@/features/settings'
import {
  OrganizationMembersSection,
  PendingInvitations,
} from '@/features/organizations'

export default function OrganizationMembersPage() {
  return (
    <ContentSection
      title="Members"
      description="Manage team members, roles, and invitations."
    >
      <div className="space-y-8">
        <OrganizationMembersSection />
        <PendingInvitations />
      </div>
    </ContentSection>
  )
}
