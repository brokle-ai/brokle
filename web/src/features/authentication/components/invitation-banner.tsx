'use client'

import { Avatar, AvatarFallback } from '@/components/ui/avatar'

interface InvitationBannerProps {
  organizationName: string
  inviterName: string
  inviterEmail?: string
}

export function InvitationBanner({
  organizationName,
  inviterName,
  inviterEmail,
}: InvitationBannerProps) {
  // Get initials from organization name (first 2 characters, or first letter of each word)
  const getInitials = (name: string) => {
    const words = name.trim().split(/\s+/)
    if (words.length >= 2) {
      return (words[0][0] + words[1][0]).toUpperCase()
    }
    return name.substring(0, 2).toUpperCase()
  }

  return (
    <div className="mb-6">
      <div className="flex items-center gap-3">
        <Avatar className="h-12 w-12">
          <AvatarFallback className="bg-primary/10 text-primary text-lg font-medium">
            {getInitials(organizationName)}
          </AvatarFallback>
        </Avatar>
        <div>
          <h2 className="font-semibold text-lg">{organizationName}</h2>
          <p className="text-sm text-muted-foreground">
            Invited by {inviterName}
            {inviterEmail && ` (${inviterEmail})`}
          </p>
        </div>
      </div>
      <div className="border-b border-dashed border-muted-foreground/30 mt-4" />
    </div>
  )
}
