import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { rawFetch } from '@/lib/api/client'
import { BrokleError } from '@/lib/api/errors'
import { useAuthStore, type SessionUser } from '@/stores/auth-store'

export const Route = createFileRoute('/signup')({
  component: SignUpPage,
})

function SignUpPage() {
  const navigate = useNavigate()
  const setUser = useAuthStore((s) => s.setUser)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setPending(true)
    try {
      const resp = await rawFetch('/v1/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email,
          password,
          first_name: firstName,
          last_name: lastName,
        }),
      })
      const body = (await resp.json()) as { user: SessionUser }
      setUser(body.user)
      await navigate({ to: '/' })
    } catch (err) {
      setError(err instanceof BrokleError ? err.body.error.message : 'Sign-up failed.')
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <form onSubmit={onSubmit} className="w-full max-w-sm space-y-4 rounded-lg border p-6">
        <h1 className="text-lg font-semibold">Create your Brokle account</h1>
        <div className="grid grid-cols-2 gap-3">
          <label className="block">
            <span className="text-sm font-medium">First name</span>
            <input
              name="firstName"
              required
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              className="mt-1 w-full rounded border px-3 py-2"
            />
          </label>
          <label className="block">
            <span className="text-sm font-medium">Last name</span>
            <input
              name="lastName"
              required
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              className="mt-1 w-full rounded border px-3 py-2"
            />
          </label>
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
            minLength={12}
            autoComplete="new-password"
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
          {pending ? 'Creating account…' : 'Sign up'}
        </button>
        <p className="text-sm">
          Already have an account?{' '}
          <Link to="/signin" className="underline">
            Sign in
          </Link>
        </p>
      </form>
    </main>
  )
}
