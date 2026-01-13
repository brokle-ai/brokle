'use client'

import { Suspense, useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ArrowLeft, Loader2 } from 'lucide-react'
import {
  AuthLayout,
  TwoStepSignUpForm,
  AuthFormWrapper
} from '@/features/authentication'
import { InvitationBanner } from '@/features/authentication/components/invitation-banner'
import { validateInvitation } from '@/features/authentication/api/auth-api'
import type { InvitationDetails } from '@/features/authentication/types'
import Link from 'next/link'
import { ROUTES } from '@/lib/routes'

type SignupStep = 'auth' | 'personalization'

function SignUpContent() {
  const searchParams = useSearchParams()
  const invitationToken = searchParams.get('token') || undefined
  const oauthSessionId = searchParams.get('session') || undefined
  const [currentStep, setCurrentStep] = useState<SignupStep>(oauthSessionId ? 'personalization' : 'auth')

  // Invitation details fetched at page level
  const [invitationDetails, setInvitationDetails] = useState<InvitationDetails | null>(null)
  const [invitationLoading, setInvitationLoading] = useState(!!invitationToken)
  const [invitationError, setInvitationError] = useState<string | null>(null)

  // Fetch invitation details when token is present
  useEffect(() => {
    if (!invitationToken) {
      return
    }

    validateInvitation(invitationToken)
      .then((data) => {
        if (data.is_expired) {
          setInvitationError('This invitation has expired. Please ask for a new invitation.')
          setInvitationDetails(null)
        } else {
          setInvitationDetails({
            organizationName: data.organization_name,
            organizationId: data.organization_id,
            inviterName: data.inviter_name,
            role: data.role,
            email: data.email,
            expiresAt: data.expires_at,
            isExpired: data.is_expired,
          })
          setInvitationError(null)
        }
      })
      .catch(() => {
        setInvitationError('Invalid or expired invitation link.')
        setInvitationDetails(null)
      })
      .finally(() => {
        setInvitationLoading(false)
      })
  }, [invitationToken])

  // Show loading while fetching invitation
  if (invitationLoading) {
    return (
      <AuthLayout>
        <div className="flex min-h-[400px] items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </AuthLayout>
    )
  }

  // Show error if invitation is invalid
  if (invitationToken && invitationError) {
    return (
      <AuthLayout>
        <Card className="gap-4">
          <CardHeader>
            <CardTitle className="text-lg tracking-tight text-destructive">
              Invalid Invitation
            </CardTitle>
            <CardDescription>{invitationError}</CardDescription>
          </CardHeader>
          <CardFooter>
            <Button asChild variant="outline">
              <Link href={ROUTES.SIGNIN}>Go to Sign In</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <div className="w-full max-w-md">
        {/* Show invitation banner when invitation exists */}
        {invitationDetails && (
          <InvitationBanner
            organizationName={invitationDetails.organizationName}
            inviterName={invitationDetails.inviterName}
          />
        )}

        <Card className="gap-4">
          <CardHeader>
            <CardTitle className="text-lg tracking-tight">
              {invitationDetails ? 'Join Organization' : 'Create your account'}
            </CardTitle>
            <CardDescription>
              {invitationDetails ? (
                <>
                  A new account will be created for{' '}
                  <strong>{invitationDetails.email}</strong>
                </>
              ) : (
                'Get started with Brokle in seconds'
              )}
              <br />
              Already have an account?{' '}
              <Link
                href={ROUTES.SIGNIN}
                className="hover:text-primary underline underline-offset-4"
              >
                Sign In
              </Link>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AuthFormWrapper>
              <TwoStepSignUpForm
                invitationToken={invitationToken}
                invitationDetails={invitationDetails}
                oauthSessionId={oauthSessionId}
                onStepChange={setCurrentStep}
              />
            </AuthFormWrapper>
          </CardContent>
          <CardFooter className="flex flex-col items-center gap-4">
            <p className="text-muted-foreground px-8 text-center text-sm">
              By creating an account, you agree to our{' '}
              <Link
                href="/terms"
                className="hover:text-primary underline underline-offset-4"
              >
                Terms of Service
              </Link>{' '}
              and{' '}
              <Link
                href="/privacy"
                className="hover:text-primary underline underline-offset-4"
              >
                Privacy Policy
              </Link>
              .
            </p>
            {currentStep === 'personalization' && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  // Trigger step change in the form component
                  window.dispatchEvent(new CustomEvent('signup-go-back'))
                }}
                className="text-muted-foreground"
              >
                <ArrowLeft className="mr-2 h-4 w-4" /> Back
              </Button>
            )}
          </CardFooter>
        </Card>
      </div>
    </AuthLayout>
  )
}

export default function SignUpPage() {
  return (
    <Suspense
      fallback={
        <AuthLayout>
          <div className="flex min-h-[400px] items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        </AuthLayout>
      }
    >
      <SignUpContent />
    </Suspense>
  )
}
