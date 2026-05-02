import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import {
  AlertTriangle,
  Ban,
  Building2,
  CheckCircle,
  Loader2,
  Shield,
  UserCircle,
  XCircle,
} from 'lucide-react'
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
  acceptInvitation,
  declineInvitation,
  validateInvitation,
} from '@/features/authentication'
import { useCurrentUser } from '@/features/auth'
import { BrokleError } from '@/lib/api/errors'

const searchSchema = z.object({
  token: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/accept-invite')({
  validateSearch: searchSchema,
  component: AcceptInvitePage,
})

interface InvitationDetailsView {
  organizationName: string
  organizationId: string
  role: string
  email: string
  inviterName: string
  expiresAt: Date
}

type InvitationState =
  | { status: 'loading' }
  | { status: 'invalid'; message: string }
  | { status: 'expired' }
  | { status: 'valid'; details: InvitationDetailsView }
  | { status: 'accepting'; details: InvitationDetailsView }
  | { status: 'declining'; details: InvitationDetailsView }
  | { status: 'accepted'; orgName: string; orgId: string }
  | { status: 'declined'; orgName: string }
  | { status: 'error'; message: string }

function AcceptInvitePage() {
  const navigate = useNavigate()
  const { token } = Route.useSearch()
  const { data: user, isLoading: authLoading } = useCurrentUser()

  const [state, setState] = useState<InvitationState>(() =>
    token
      ? { status: 'loading' }
      : { status: 'invalid', message: 'No invitation token provided' },
  )

  useEffect(() => {
    if (!token) return
    let cancelled = false

    validateInvitation(token)
      .then((response) => {
        if (cancelled) return
        if (response.is_expired) {
          setState({ status: 'expired' })
          return
        }
        setState({
          status: 'valid',
          details: {
            organizationName: response.organization_name,
            organizationId: response.organization_id,
            role: response.role,
            email: response.email,
            inviterName: response.inviter_name,
            expiresAt: new Date(response.expires_at),
          },
        })
      })
      .catch((error: unknown) => {
        if (cancelled) return
        if (error instanceof BrokleError && error.status === 410) {
          setState({ status: 'expired' })
          return
        }
        if (error instanceof BrokleError && error.status === 404) {
          setState({
            status: 'invalid',
            message: 'This invitation link is invalid or has been revoked',
          })
          return
        }
        setState({ status: 'invalid', message: 'Failed to validate invitation' })
      })

    return () => {
      cancelled = true
    }
  }, [token])

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
      let message = 'Failed to accept invitation. Please try again.'
      if (error instanceof BrokleError) {
        const msg = error.body.error.message.toLowerCase()
        if (msg.includes('already a member')) {
          message = 'You are already a member of this organization'
        } else if (msg.includes('email') && msg.includes('mismatch')) {
          message = 'This invitation was sent to a different email address'
        } else if (msg.includes('expired')) {
          setState({ status: 'expired' })
          return
        } else {
          message = error.body.error.message
        }
      }
      setState({ status: 'error', message })
    }
  }

  const handleSignIn = () => {
    if (!token) return
    navigate({
      to: '/signin',
      search: { redirect: `/accept-invite?token=${encodeURIComponent(token)}` },
    })
  }

  const handleDecline = async () => {
    if (!token || (state.status !== 'valid' && state.status !== 'accepting')) return
    const details =
      state.status === 'valid' || state.status === 'accepting' ? state.details : null
    if (!details) return
    setState({ status: 'declining', details })
    try {
      await declineInvitation(token)
      setState({ status: 'declined', orgName: details.organizationName })
      toast.success('Invitation declined')
    } catch {
      toast.error('Failed to decline invitation')
      setState({ status: 'valid', details })
    }
  }

  const [countdown, setCountdown] = useState(5)

  useEffect(() => {
    if (state.status !== 'accepted') return
    if (countdown <= 0) {
      window.location.href = '/'
      return
    }
    const timer = setTimeout(() => setCountdown((v) => v - 1), 1000)
    return () => clearTimeout(timer)
  }, [state.status, countdown])

  if (authLoading || state.status === 'loading') {
    return (
      <AuthLayout>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Loader2 className="text-muted-foreground mb-4 h-8 w-8 animate-spin" />
            <p className="text-muted-foreground">Validating invitation...</p>
          </CardContent>
        </Card>
      </AuthLayout>
    )
  }

  if (state.status === 'invalid') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="bg-destructive/10 mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full">
              <XCircle className="text-destructive h-6 w-6" />
            </div>
            <CardTitle>Invalid Invitation</CardTitle>
            <CardDescription>{state.message}</CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button asChild>
              <Link to="/">Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

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
              This invitation has expired. Please ask the organization admin to
              send a new invitation.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button asChild>
              <Link to="/">Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  if (state.status === 'error') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="bg-destructive/10 mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full">
              <XCircle className="text-destructive h-6 w-6" />
            </div>
            <CardTitle>Something Went Wrong</CardTitle>
            <CardDescription>{state.message}</CardDescription>
          </CardHeader>
          <CardFooter className="flex flex-col gap-2">
            <Button
              onClick={() => setState({ status: 'loading' })}
              className="w-full"
            >
              Try Again
            </Button>
            <Button variant="outline" asChild className="w-full">
              <Link to="/">Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  if (state.status === 'accepted') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/20">
              <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-500" />
            </div>
            <CardTitle className="text-xl">
              Welcome to {state.orgName}!
            </CardTitle>
            <CardDescription>
              You have successfully joined the organization.
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <p className="text-muted-foreground text-sm">
              Redirecting to dashboard in{' '}
              <span className="text-foreground font-semibold">{countdown}</span>{' '}
              seconds...
            </p>
          </CardContent>
          <CardFooter className="flex justify-center">
            <Button
              onClick={() => {
                window.location.href = '/'
              }}
            >
              Go to Dashboard Now
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  if (state.status === 'declined') {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="bg-muted mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full">
              <Ban className="text-muted-foreground h-8 w-8" />
            </div>
            <CardTitle className="text-xl">Invitation Declined</CardTitle>
            <CardDescription>
              You&apos;ve declined the invitation to join {state.orgName}.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button variant="outline" asChild>
              <Link to="/">Go to Dashboard</Link>
            </Button>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  if (
    state.status !== 'valid' &&
    state.status !== 'accepting' &&
    state.status !== 'declining'
  ) {
    return null
  }

  const { details } = state
  const isAccepting = state.status === 'accepting'
  const isDeclining = state.status === 'declining'

  if (!user) {
    return (
      <AuthLayout>
        <Card>
          <CardHeader className="text-center">
            <div className="bg-primary/10 mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full">
              <Building2 className="text-primary h-6 w-6" />
            </div>
            <CardTitle>You&apos;re Invited!</CardTitle>
            <CardDescription>
              You&apos;ve been invited to join{' '}
              <strong>{details.organizationName}</strong>
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3 rounded-lg border p-4">
              <div className="flex items-center gap-3">
                <Building2 className="text-muted-foreground h-4 w-4" />
                <div>
                  <p className="text-muted-foreground text-sm">Organization</p>
                  <p className="font-medium">{details.organizationName}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Shield className="text-muted-foreground h-4 w-4" />
                <div>
                  <p className="text-muted-foreground text-sm">Your Role</p>
                  <p className="font-medium capitalize">{details.role}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <UserCircle className="text-muted-foreground h-4 w-4" />
                <div>
                  <p className="text-muted-foreground text-sm">Invited by</p>
                  <p className="font-medium">{details.inviterName}</p>
                </div>
              </div>
            </div>
            <div className="text-muted-foreground text-center text-sm">
              <p>Sign in to accept this invitation</p>
              <p className="mt-1 text-xs">
                Invitation for: <strong>{details.email}</strong>
              </p>
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-3">
            <Button
              onClick={handleSignIn}
              className="w-full"
              disabled={isDeclining}
            >
              Sign In to Accept
            </Button>
            <Button
              variant="outline"
              className="w-full"
              onClick={handleDecline}
              disabled={isDeclining}
            >
              {isDeclining ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Declining...
                </>
              ) : (
                'Decline Invitation'
              )}
            </Button>
            <div className="text-center text-sm">
              <span className="text-muted-foreground">
                Don&apos;t have an account?{' '}
              </span>
              <Link
                to="/signup"
                search={{ token }}
                className="hover:text-primary font-medium underline underline-offset-4"
              >
                Sign up
              </Link>
            </div>
          </CardFooter>
        </Card>
      </AuthLayout>
    )
  }

  const emailMismatch =
    user.email.toLowerCase() !== details.email.toLowerCase()

  return (
    <AuthLayout>
      <Card>
        <CardHeader className="text-center">
          <div className="bg-primary/10 mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full">
            <Building2 className="text-primary h-6 w-6" />
          </div>
          <CardTitle>You&apos;re Invited!</CardTitle>
          <CardDescription>
            You&apos;ve been invited to join{' '}
            <strong>{details.organizationName}</strong>
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-3 rounded-lg border p-4">
            <div className="flex items-center gap-3">
              <Building2 className="text-muted-foreground h-4 w-4" />
              <div>
                <p className="text-muted-foreground text-sm">Organization</p>
                <p className="font-medium">{details.organizationName}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Shield className="text-muted-foreground h-4 w-4" />
              <div>
                <p className="text-muted-foreground text-sm">Your Role</p>
                <p className="font-medium capitalize">{details.role}</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <UserCircle className="text-muted-foreground h-4 w-4" />
              <div>
                <p className="text-muted-foreground text-sm">Invited by</p>
                <p className="font-medium">{details.inviterName}</p>
              </div>
            </div>
          </div>

          {emailMismatch && (
            <div className="border-destructive/50 bg-destructive/10 rounded-lg border p-3 text-sm">
              <p className="text-destructive">
                <strong>Cannot accept:</strong> This invitation was sent to{' '}
                <strong>{details.email}</strong>, but you&apos;re signed in as{' '}
                <strong>{user.email}</strong>. Please sign in with the correct
                account.
              </p>
            </div>
          )}
        </CardContent>
        <CardFooter className="flex flex-col gap-2">
          <Button
            onClick={handleAccept}
            className="w-full"
            disabled={isAccepting || isDeclining || emailMismatch}
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
          <Button
            variant="outline"
            className="w-full"
            onClick={handleDecline}
            disabled={isAccepting || isDeclining}
          >
            {isDeclining ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Declining...
              </>
            ) : (
              'Decline'
            )}
          </Button>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
