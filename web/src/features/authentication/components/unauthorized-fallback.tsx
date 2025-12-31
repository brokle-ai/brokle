'use client'

import Link from 'next/link'
import { ShieldAlert, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ROUTES } from '@/lib/routes'

interface UnauthorizedFallbackProps {
  title?: string
  message?: string
  showBackButton?: boolean
  showLoginButton?: boolean
}

export function UnauthorizedFallback({
  title = 'Access Denied',
  message = 'You need to be logged in to access this page.',
  showBackButton = true,
  showLoginButton = true,
}: UnauthorizedFallbackProps) {
  return (
    <div className="flex items-center justify-center min-h-[400px] p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
            <ShieldAlert className="h-8 w-8 text-red-600" />
          </div>
          <CardTitle className="text-xl">{title}</CardTitle>
          <CardDescription className="text-center">
            {message}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {showLoginButton && (
            <Link href={ROUTES.SIGNIN} className="w-full">
              <Button className="w-full">
                Sign In to Continue
              </Button>
            </Link>
          )}
          {showBackButton && (
            <Button 
              variant="outline" 
              className="w-full"
              onClick={() => window.history.back()}
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Go Back
            </Button>
          )}
        </CardContent>
      </Card>
    </div>
  )
}