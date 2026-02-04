'use client'

import { UserPlus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ContentSection } from '@/features/settings'
import {
  OrganizationMembersSection,
  PendingInvitations,
  InviteMemberModal,
} from '@/features/organizations'
import { useHasAccess } from '@/hooks/rbac/use-has-access'

export default function OrganizationMembersPage() {
  const canInviteMembers = useHasAccess({ scope: 'members:invite' })

  return (
    <ContentSection
      title="Members"
      description="Manage team members, roles, and invitations."
      action={
        canInviteMembers ? (
          <InviteMemberModal
            trigger={
              <Button>
                <UserPlus className="mr-2 h-4 w-4" />
                Invite Member
              </Button>
            }
          />
        ) : undefined
      }
    >
      <div className="space-y-8">
        <OrganizationMembersSection />
        <PendingInvitations />
      </div>
    </ContentSection>
  )
}
