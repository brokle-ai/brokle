import { createFileRoute, Link, useNavigate, useSearch } from '@tanstack/react-router'
import { useState } from 'react'
import { z } from 'zod'
import { rawFetch } from '@/lib/api/client'
import { BrokleError } from '@/lib/api/errors'
import { useAuthStore, type SessionUser } from '@/stores/auth-store'

const searchSchema = z.object({
  redirect: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/signin')({
  validateSearch: searchSchema,
  component: SignInPage,
})

function SignInPage() {
  const navigate = useNavigate()
  const { redirect } = useSearch({ from: '/signin' })
  const setUser = useAuthStore((s) => s.setUser)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setPending(true)
    try {
      const resp = await rawFetch('/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      const body = (await resp.json()) as { user: SessionUser }
      setUser(body.user)
      await navigate({ to: redirect ?? '/' })
    } catch (err) {
      if (err instanceof BrokleError) {
        setError(err.body.error.message)
      } else {
        setError('Sign-in failed. Please try again.')
      }
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <form
        onSubmit={onSubmit}
        className="w-full max-w-sm space-y-4 rounded-lg border p-6"
        noValidate
      >
        <div className="space-y-1">
          <h1 className="text-lg font-semibold">Sign in to Brokle</h1>
          <p className="text-sm text-muted-foreground">
            Use your organization account to continue.
          </p>
        </div>
        <label className="block">
          <span className="text-sm font-medium">Email</span>
          <input
            name="email"
            type="email"
            required
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 w-full rounded border px-3 py-2"
          />
        </label>
        <label className="block">
          <span className="text-sm font-medium">Password</span>
          <input
            name="password"
            type="password"
            required
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 w-full rounded border px-3 py-2"
          />
        </label>
        {error ? (
          <p role="alert" className="text-sm text-red-600">
            {error}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={pending}
          className="w-full rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
        >
          {pending ? 'Signing in…' : 'Sign in'}
        </button>
        <div className="flex items-center justify-between text-sm">
          <Link to="/forgot-password" className="underline">
            Forgot password?
          </Link>
          <Link to="/signup" className="underline">
            Create account
          </Link>
        </div>
      </form>
    </main>
  )
}
