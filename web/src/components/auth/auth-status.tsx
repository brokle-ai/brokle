'use client'

import { useAuth } from '@/hooks/auth/use-auth'
import { useTokenRefresh } from '@/hooks/auth/use-token-refresh'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { CheckCircle, AlertCircle, Clock, User, Building } from 'lucide-react'
import { cn } from '@/lib/utils'

interface AuthStatusProps {
  variant?: 'compact' | 'detailed'
  showTokenInfo?: boolean
  className?: string
}

export function AuthStatus({ 
  variant = 'detailed',
  showTokenInfo = false,
  className 
}: AuthStatusProps) {
  const { user, organization, isAuthenticated, isLoading } = useAuth()
  const { tokenTimeLeft, isTokenValid } = useTokenRefresh()

  if (isLoading) {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <div className="h-4 w-4 animate-pulse rounded-full bg-gray-300" />
        <span className="text-sm text-muted-foreground">Loading...</span>
      </div>
    )
  }

  if (!isAuthenticated || !user) {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <AlertCircle className="h-4 w-4 text-red-500" />
        <span className="text-sm text-red-600">Not authenticated</span>
      </div>
    )
  }

  // Compact variant
  if (variant === 'compact') {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <CheckCircle className="h-4 w-4 text-green-500" />
        <span className="text-sm text-green-600">Authenticated</span>
        {showTokenInfo && (
          <>
            <Separator orientation="vertical" className="h-4" />
            <TokenStatusBadge timeLeft={tokenTimeLeft} isValid={isTokenValid} />
          </>
        )}
      </div>
    )
  }

  // Detailed variant
  const userMembership = organization?.members.find(m => m.userId === user.id)

  return (
    <Card className={cn('', className)}>
      <CardHeader className="pb-3">
        <CardTitle className="text-base flex items-center gap-2">
          <CheckCircle className="h-4 w-4 text-green-500" />
          Authentication Status
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* User Info */}
        <div className="flex items-center gap-3">
          <Avatar className="h-10 w-10">
            <AvatarImage src={user.avatar} />
            <AvatarFallback>
              <User className="h-4 w-4" />
            </AvatarFallback>
          </Avatar>
          <div className="flex-1">
            <p className="text-sm font-medium">
              {user.firstName && user.lastName 
                ? `${user.firstName} ${user.lastName}`
                : user.name || 'Unnamed User'}
            </p>
            <p className="text-xs text-muted-foreground">{user.email}</p>
          </div>
          <Badge variant={user.isEmailVerified ? 'secondary' : 'destructive'}>
            {user.isEmailVerified ? 'Verified' : 'Unverified'}
          </Badge>
        </div>

        {/* Organization Info */}
        {organization && (
          <>
            <Separator />
            <div className="flex items-center gap-3">
              <Building className="h-4 w-4 text-muted-foreground" />
              <div className="flex-1">
                <p className="text-sm font-medium">{organization.name}</p>
                <p className="text-xs text-muted-foreground">
                  {organization.members.length} member{organization.members.length !== 1 ? 's' : ''}
                </p>
              </div>
              {userMembership && (
                <Badge variant="outline">
                  {userMembership.role}
                </Badge>
              )}
            </div>
          </>
        )}

        {/* Token Info */}
        {showTokenInfo && (
          <>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Token Status</span>
              <TokenStatusBadge timeLeft={tokenTimeLeft} isValid={isTokenValid} />
            </div>
          </>
        )}

        {/* Session Info */}
        <div className="grid grid-cols-2 gap-4 text-xs text-muted-foreground">
          <div>
            <span className="font-medium">User ID:</span>
            <br />
            <code className="text-[10px]">{user.id.split('-')[0]}...</code>
          </div>
          <div>
            <span className="font-medium">Last Login:</span>
            <br />
            {user.lastLoginAt 
              ? new Date(user.lastLoginAt).toLocaleString()
              : 'N/A'
            }
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function TokenStatusBadge({ timeLeft, isValid }: { timeLeft: number; isValid: boolean }) {
  if (!isValid) {
    return (
      <Badge variant="destructive" className="text-xs">
        <AlertCircle className="mr-1 h-3 w-3" />
        Invalid
      </Badge>
    )
  }

  const minutes = Math.floor(timeLeft / (1000 * 60))
  const hours = Math.floor(minutes / 60)

  let timeDisplay: string
  let variant: 'default' | 'secondary' | 'destructive' = 'default'

  if (hours > 0) {
    timeDisplay = `${hours}h ${minutes % 60}m`
    variant = 'default'
  } else if (minutes > 5) {
    timeDisplay = `${minutes}m`
    variant = 'default'
  } else if (minutes > 0) {
    timeDisplay = `${minutes}m`
    variant = 'secondary'
  } else {
    timeDisplay = 'Expiring soon'
    variant = 'destructive'
  }

  return (
    <Badge variant={variant} className="text-xs">
      <Clock className="mr-1 h-3 w-3" />
      {timeDisplay}
    </Badge>
  )
}