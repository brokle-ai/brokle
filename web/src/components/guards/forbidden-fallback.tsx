'use client'

import Link from 'next/link'
import { Shield, ArrowLeft, Mail } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { OrganizationRole } from '@/types/auth'

interface ForbiddenFallbackProps {
  title?: string
  message?: string
  requiredRole?: OrganizationRole
  showBackButton?: boolean
  showContactButton?: boolean
}

export function ForbiddenFallback({
  title = 'Insufficient Permissions',
  message,
  requiredRole,
  showBackButton = true,
  showContactButton = true,
}: ForbiddenFallbackProps) {
  const getRoleMessage = (role?: OrganizationRole) => {
    if (!role) return 'You don\'t have permission to access this resource.'
    
    const roleLabels = {
      owner: 'organization owner',
      admin: 'administrator',
      developer: 'developer',
      viewer: 'viewer'
    }
    
    return `This feature requires ${roleLabels[role]} permissions.`
  }

  const finalMessage = message || getRoleMessage(requiredRole)

  return (
    <div className="flex items-center justify-center min-h-[400px] p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-yellow-100">
            <Shield className="h-8 w-8 text-yellow-600" />
          </div>
          <CardTitle className="text-xl">{title}</CardTitle>
          <CardDescription className="text-center">
            {finalMessage}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="text-center text-sm text-muted-foreground">
            If you believe this is an error, please contact your administrator.
          </div>
          
          <div className="flex gap-2">
            {showBackButton && (
              <Button 
                variant="outline" 
                className="flex-1"
                onClick={() => window.history.back()}
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                Go Back
              </Button>
            )}
            {showContactButton && (
              <Button 
                variant="outline" 
                className="flex-1"
                asChild
              >
                <Link href="mailto:support@brokle.com">
                  <Mail className="mr-2 h-4 w-4" />
                  Contact Support
                </Link>
              </Button>
            )}
          </div>
          
          <Link href="/dashboard" className="w-full">
            <Button variant="secondary" className="w-full">
              Return to Dashboard
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  )
}