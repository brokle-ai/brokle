import { createFileRoute, Link, useSearch } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { ArrowLeft, Loader2 } from 'lucide-react'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { AuthLayout } from '@/components/layout/auth-layout'
import {
  InvitationBanner,
  TwoStepSignUpForm,
  validateInvitation,
} from '@/features/authentication'
import type { InvitationDetails } from '@/features/authentication'

const searchSchema = z.object({
  token: z.string().optional().catch(undefined),
  session: z.string().optional().catch(undefined),
  redirect: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/signup')({
  validateSearch: searchSchema,
  component: SignUpPage,
})

type SignupStep = 'auth' | 'personalization'

function SignUpPage() {
  const { token: invitationToken, session: oauthSessionId, redirect } = useSearch({
    from: '/signup',
  })

  const [currentStep, setCurrentStep] = useState<SignupStep>(
    oauthSessionId ? 'personalization' : 'auth',
  )

  const [invitationDetails, setInvitationDetails] =
    useState<InvitationDetails | null>(null)
  const [invitationLoading, setInvitationLoading] = useState<boolean>(
    Boolean(invitationToken),
  )
  const [invitationError, setInvitationError] = useState<string | null>(null)

  useEffect(() => {
    if (!invitationToken) return

    let cancelled = false
    validateInvitation(invitationToken)
      .then((data) => {
        if (cancelled) return
        if (data.is_expired) {
          setInvitationError(
            'This invitation has expired. Please ask for a new invitation.',
          )
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
        if (!cancelled) {
          setInvitationError('Invalid or expired invitation link.')
          setInvitationDetails(null)
        }
      })
      .finally(() => {
        if (!cancelled) setInvitationLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [invitationToken])

  if (invitationLoading) {
    return (
      <AuthLayout>
        <div className="flex min-h-[400px] items-center justify-center">
          <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
        </div>
      </AuthLayout>
    )
  }

  if (invitationToken && invitationError) {
    return (
      <AuthLayout>
        <Card className="gap-4">
          <CardHeader>
            <CardTitle className="text-destructive text-lg tracking-tight">
              Invalid Invitation
            </CardTitle>
            <CardDescription>{invitationError}</CardDescription>
          </CardHeader>
          <CardFooter>
            <Button asChild variant="outline">
              <Link to="/signin">Go to Sign In</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <div className="w-full max-w-md">
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
                to="/signin"
                className="hover:text-primary underline underline-offset-4"
              >
                Sign In
              </Link>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <TwoStepSignUpForm
              invitationToken={invitationToken}
              invitationDetails={invitationDetails}
              oauthSessionId={oauthSessionId}
              redirectTo={redirect}
              onStepChange={setCurrentStep}
            />
          </CardContent>
          <CardFooter className="flex flex-col items-center gap-4">
            <p className="text-muted-foreground px-8 text-center text-sm">
              By creating an account, you agree to our{' '}
              <a
                href="/terms"
                className="hover:text-primary underline underline-offset-4"
              >
                Terms of Service
              </a>{' '}
              and{' '}
              <a
                href="/privacy"
                className="hover:text-primary underline underline-offset-4"
              >
                Privacy Policy
              </a>
              .
            </p>
            {currentStep === 'personalization' && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
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
