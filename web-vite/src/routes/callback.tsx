import {
  createFileRoute,
  useNavigate,
  useRouter,
  useSearch,
} from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { Loader2 } from 'lucide-react'
import { z } from 'zod'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { exchangeLoginSession } from '@/features/authentication'
import { resetSession } from '@/lib/auth/session'

const searchSchema = z.object({
  session: z.string().optional().catch(undefined),
  type: z.string().optional().catch(undefined),
  error: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/callback')({
  validateSearch: searchSchema,
  component: OAuthCallbackPage,
})

function OAuthCallbackPage() {
  const navigate = useNavigate()
  const router = useRouter()
  const { session, error: queryError } = useSearch({ from: '/callback' })
  const [error, setError] = useState<string | null>(queryError ?? null)

  useEffect(() => {
    if (queryError) {
      setError(queryError)
      const t = setTimeout(() => navigate({ to: '/signin' }), 3000)
      return () => clearTimeout(t)
    }

    if (!session) {
      setError('No session found. Please try again.')
      const t = setTimeout(() => navigate({ to: '/signin' }), 2000)
      return () => clearTimeout(t)
    }

    let cancelled = false
    exchangeLoginSession(session)
      .then((response) => {
        if (cancelled) return
        if (!response || !response.user) {
          setError('Invalid session data. Please try again.')
          setTimeout(() => navigate({ to: '/signin' }), 2000)
          return
        }
        // Trigger the navigation; `_authenticated.beforeLoad` owns the
        // /me probe and any failure routing from here:
        //   - happy path: dashboard renders.
        //   - cookie dropped: beforeLoad redirects to
        //     /signin?session=expired (toast handled by
        //     SignInToastHandler).
        //   - /me 5xx: RootErrorFallback renders.
        // No try/catch around navigate — TanStack Router resolves the
        // promise on redirects/errors, so caller-side handling here
        // would be unreachable for those cases.
        resetSession(router.options.context.queryClient)
        void navigate({ to: '/', replace: true })
      })
      .catch(() => {
        if (cancelled) return
        setError('Failed to complete authentication. Please try again.')
        setTimeout(() => navigate({ to: '/signin' }), 3000)
      })

    return () => {
      cancelled = true
    }
  }, [session, queryError, navigate, router])

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Alert variant="destructive" className="max-w-md">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-4">
      <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
      <p className="text-muted-foreground mt-4 text-sm">
        Completing authentication...
      </p>
    </div>
  )
}
