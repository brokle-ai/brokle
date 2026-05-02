import { Avatar, AvatarFallback } from '@/components/ui/avatar'

interface InvitationBannerProps {
  organizationName: string
  inviterName: string
  inviterEmail?: string
}

function getInitials(name: string): string {
  const words = name.trim().split(/\s+/)
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
}

export function InvitationBanner({
  organizationName,
  inviterName,
  inviterEmail,
}: InvitationBannerProps) {
  return (
    <div className="mb-6">
      <div className="flex items-center gap-3">
        <Avatar className="h-12 w-12">
          <AvatarFallback className="bg-primary/10 text-primary text-lg font-medium">
            {getInitials(organizationName)}
          </AvatarFallback>
        </Avatar>
        <div>
          <h2 className="text-lg font-semibold">{organizationName}</h2>
          <p className="text-muted-foreground text-sm">
            Invited by {inviterName}
            {inviterEmail && ` (${inviterEmail})`}
          </p>
        </div>
      </div>
      <div className="border-muted-foreground/30 mt-4 border-b border-dashed" />
    </div>
  )
}
