import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { rawFetch } from '@/lib/api/client'

export const Route = createFileRoute('/forgot-password')({
  component: ForgotPasswordPage,
})

function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [pending, setPending] = useState(false)

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setPending(true)
    try {
      await rawFetch('/api/v1/auth/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })
    } finally {
      // Anti-enumeration: always show success, even if the email didn't exist.
      setSubmitted(true)
      setPending(false)
    }
  }

  if (submitted) {
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <div className="max-w-sm space-y-3 rounded-lg border p-6 text-sm">
          <h1 className="text-lg font-semibold">Check your inbox</h1>
          <p>
            If an account exists for <strong>{email}</strong>, you&apos;ll receive a reset
            link shortly.
          </p>
          <Link to="/signin" className="underline">
            Back to sign in
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <form onSubmit={onSubmit} className="w-full max-w-sm space-y-4 rounded-lg border p-6">
        <h1 className="text-lg font-semibold">Reset your password</h1>
        <p className="text-sm text-muted-foreground">
          Enter your email and we&apos;ll send you a reset link.
        </p>
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
        <button
          type="submit"
          disabled={pending}
          className="w-full rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
        >
          {pending ? 'Sending…' : 'Send reset link'}
        </button>
        <p className="text-sm">
          <Link to="/signin" className="underline">
            Back to sign in
          </Link>
        </p>
      </form>
    </main>
  )
}
