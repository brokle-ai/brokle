'use client'

import { useEffect, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { Loader2, CheckCircle, XCircle, AlertTriangle, Building2, UserCircle, Shield } from 'lucide-react'
import { AuthLayout, useAuth } from '@/features/authentication'
import { validateInvitationToken, acceptInvitation } from '@/features/organizations'
import { ROUTES, signinWithRedirect } from '@/lib/routes'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'

type InvitationState =
  | { status: 'loading' }
  | { status: 'invalid'; message: string }
  | { status: 'expired' }
  | { status: 'valid'; details: InvitationDetails }
  | { status: 'accepting'; details: InvitationDetails }
  | { status: 'accepted'; orgName: string; orgId: string }
  | { status: 'error'; message: string }

interface InvitationDetails {
  organizationName: string
  organizationId: string
  role: string
  email: string
  inviterName: string
  expiresAt: Date
}

export default function AcceptInvitePage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const { user, isLoading: authLoading } = useAuth()
  const token = searchParams.get('token')

  const [state, setState] = useState<InvitationState>({ status: 'loading' })

  // Validate token on mount
  useEffect(() => {
    if (!token) {
      setState({ status: 'invalid', message: 'No invitation token provided' })
      return
    }

    const validate = async () => {
      try {
        const response = await validateInvitationToken(token)

        if (!response.valid) {
          setState({ status: 'invalid', message: 'This invitation link is no longer valid' })
          return
        }

        setState({
          status: 'valid',
          details: {
            organizationName: response.organization_name,
            organizationId: response.organization_id,
            role: response.role_name,
            email: response.email,
            inviterName: response.invited_by_name,
            expiresAt: new Date(response.expires_at),
          },
        })
      } catch (error: unknown) {
        // Check if it's an expired invitation (410 status)
        if (error && typeof error === 'object' && 'status' in error && error.status === 410) {
          setState({ status: 'expired' })
          return
        }

        // Check if it's a not found error (404)
        if (error && typeof error === 'object' && 'status' in error && error.status === 404) {
          setState({ status: 'invalid', message: 'This invitation link is invalid or has been revoked' })
          return
        }

        console.error('Failed to validate invitation:', error)
        setState({ status: 'invalid', message: 'Failed to validate invitation' })
      }
    }

    validate()
  }, [token])

  // Handle accept action
  const handleAccept = async () => {
    if (!token || state.status !== 'valid') return

    const { details } = state
    setState({ status: 'accepting', details })

    try {
      const response = await acceptInvitation(token)
      setState({
        status: 'accepted',
        orgName: response.organization_name || details.organizationName,
        orgId: response.organization_id || details.organizationId,
      })
      toast.success('Successfully joined organization!')
    } catch (error: unknown) {
      console.error('Failed to accept invitation:', error)

      // Handle specific error cases
      let message = 'Failed to accept invitation. Please try again.'

      if (error && typeof error === 'object' && 'message' in error) {
        const errorMessage = String(error.message).toLowerCase()
        if (errorMessage.includes('already a member')) {
          message = 'You are already a member of this organization'
        } else if (errorMessage.includes('email') && errorMessage.includes('mismatch')) {
          message = 'This invitation was sent to a different email address'
        } else if (errorMessage.includes('expired')) {
          setState({ status: 'expired' })
          return
        }
      }

      setState({ status: 'error', message })
    }
  }

  // Handle redirect to sign in
  const handleSignIn = () => {
    if (!token) return
    // Redirect to sign in with return URL to this page
    const returnUrl = `/accept-invite?token=${encodeURIComponent(token)}`
    router.push(signinWithRedirect(returnUrl))
  }

  // Show loading while checking auth or validating
  if (authLoading || state.status === 'loading') {
    return (
      <AuthLayout>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground mb-4" />
            <p className="text-muted-foreground">Validating invitation...</p>
          </CardContent>
        </Card>
      </AuthLayout>
    )
  }

  // Invalid token
  if (state.status === 'invalid') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
              <XCircle className="h-6 w-6 text-destructive" />
            </div>
            <CardTitle>Invalid Invitation</CardTitle>
            <CardDescription>{state.message}</CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button asChild>
              <Link href={ROUTES.HOME}>Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  // Expired invitation
  if (state.status === 'expired') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-yellow-100 dark:bg-yellow-900/20">
              <AlertTriangle className="h-6 w-6 text-yellow-600 dark:text-yellow-500" />
            </div>
            <CardTitle>Invitation Expired</CardTitle>
            <CardDescription>
              This invitation has expired. Please ask the organization admin to send a new invitation.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button asChild>
              <Link href={ROUTES.HOME}>Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  // Error state
  if (state.status === 'error') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
              <XCircle className="h-6 w-6 text-destructive" />
            </div>
            <CardTitle>Something Went Wrong</CardTitle>
            <CardDescription>{state.message}</CardDescription>
          </CardHeader>
          <CardFooter className="flex flex-col gap-2">
            <Button onClick={() => setState({ status: 'loading' })} className="w-full">
              Try Again
            </Button>
            <Button variant="outline" asChild className="w-full">
              <Link href={ROUTES.HOME}>Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  // Successfully accepted
  if (state.status === 'accepted') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/20">
              <CheckCircle className="h-6 w-6 text-green-600 dark:text-green-500" />
            </div>
            <CardTitle>Welcome to {state.orgName}!</CardTitle>
            <CardDescription>
              You have successfully joined the organization.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button asChild>
              <Link href={ROUTES.HOME}>Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  // At this point, state can only be 'valid' or 'accepting' (both have details)
  if (state.status !== 'valid' && state.status !== 'accepting') {
    return null // Should not reach here, but satisfy TypeScript
  }

  // Valid invitation - show details and accept button
  const { details } = state
  const isAccepting = state.status === 'accepting'

  // Not logged in - prompt to sign in
  if (!user) {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
              <Building2 className="h-6 w-6 text-primary" />
            </div>
            <CardTitle>You're Invited!</CardTitle>
            <CardDescription>
              You've been invited to join <strong>{details.organizationName}</strong>
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-lg border p-4 space-y-3">
              <div className="flex items-center gap-3">
                <Building2 className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm text-muted-foreground">Organization</p>
                  <p className="font-medium">{details.organizationName}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Shield className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm text-muted-foreground">Your Role</p>
                  <p className="font-medium capitalize">{details.role}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <UserCircle className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm text-muted-foreground">Invited by</p>
                  <p className="font-medium">{details.inviterName}</p>
                </div>
              </div>
            </div>

            <div className="text-center text-sm text-muted-foreground">
              <p>Sign in to accept this invitation</p>
              <p className="text-xs mt-1">
                Invitation for: <strong>{details.email}</strong>
              </p>
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-2">
            <Button onClick={handleSignIn} className="w-full">
              Sign In to Accept
            </Button>
            <div className="text-center text-sm">
              <span className="text-muted-foreground">Don't have an account? </span>
              <Link
                href={`${ROUTES.SIGNUP}?email=${encodeURIComponent(details.email)}&redirect=${encodeURIComponent(`/accept-invite?token=${token}`)}`}
                className="font-medium underline underline-offset-4 hover:text-primary"
              >
                Sign up
              </Link>
            </div>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  // User is logged in - show accept button
  return (
    <AuthLayout>
      <Card>
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            <Building2 className="h-6 w-6 text-primary" />
          </div>
          <CardTitle>You're Invited!</CardTitle>
          <CardDescription>
            You've been invited to join <strong>{details.organizationName}</strong>
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="rounded-lg border p-4 space-y-3">
            <div className="flex items-center gap-3">
              <Building2 className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-sm text-muted-foreground">Organization</p>
                <p className="font-medium">{details.organizationName}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-sm text-muted-foreground">Your Role</p>
                <p className="font-medium capitalize">{details.role}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <UserCircle className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-sm text-muted-foreground">Invited by</p>
                <p className="font-medium">{details.inviterName}</p>
              </div>
            </div>
          </div>

          {/* Email mismatch warning */}
          {user.email !== details.email && (
            <div className="rounded-lg border border-yellow-200 bg-yellow-50 dark:border-yellow-900/50 dark:bg-yellow-900/10 p-3 text-sm">
              <p className="text-yellow-800 dark:text-yellow-200">
                <strong>Note:</strong> This invitation was sent to <strong>{details.email}</strong>,
                but you're signed in as <strong>{user.email}</strong>.
              </p>
            </div>
          )}
        </CardContent>
        <CardFooter className="flex flex-col gap-2">
          <Button
            onClick={handleAccept}
            className="w-full"
            disabled={isAccepting}
          >
            {isAccepting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Accepting...
              </>
            ) : (
              'Accept Invitation'
            )}
          </Button>
          <Button variant="outline" asChild className="w-full">
            <Link href={ROUTES.HOME}>Decline</Link>
          </Button>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
