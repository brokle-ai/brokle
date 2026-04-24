import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { rawFetch } from '@/lib/api/client'
import { BrokleError, ValidationError } from '@/lib/api/errors'
import { useAuthStore, type SessionUser } from '@/stores/auth-store'

export const Route = createFileRoute('/signup')({
  component: SignUpPage,
})

const ROLES = [
  { value: 'engineer', label: 'Engineer' },
  { value: 'product', label: 'Product' },
  { value: 'designer', label: 'Designer' },
  { value: 'executive', label: 'Executive' },
  { value: 'other', label: 'Other' },
] as const

function SignUpPage() {
  const navigate = useNavigate()
  const setUser = useAuthStore((s) => s.setUser)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<typeof ROLES[number]['value']>('engineer')
  const [organizationName, setOrganizationName] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [fieldIssues, setFieldIssues] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setFieldIssues({})
    setPending(true)
    try {
      const resp = await rawFetch('/api/v1/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email,
          password,
          first_name: firstName,
          last_name: lastName,
          role,
          organization_name: organizationName,
        }),
      })
      const body = (await resp.json()) as { user: SessionUser }
      setUser(body.user)
      await navigate({ to: '/' })
    } catch (err) {
      if (err instanceof ValidationError && err.fieldIssues) {
        const byField: Record<string, string> = {}
        for (const issue of err.fieldIssues) {
          if (issue.location) byField[issue.location] = issue.message
        }
        setFieldIssues(byField)
        setError(err.body.error.message)
      } else if (err instanceof BrokleError) {
        setError(err.body.error.message)
      } else {
        setError('Sign-up failed.')
      }
    } finally {
      setPending(false)
    }
  }

  const issue = (field: string) => fieldIssues[field]

  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <form onSubmit={onSubmit} className="w-full max-w-sm space-y-4 rounded-lg border p-6">
        <h1 className="text-lg font-semibold">Create your Brokle account</h1>
        <div className="grid grid-cols-2 gap-3">
          <Field label="First name" error={issue('body.first_name')}>
            <input
              name="firstName"
              required
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              className="mt-1 w-full rounded border px-3 py-2"
            />
          </Field>
          <Field label="Last name" error={issue('body.last_name')}>
            <input
              name="lastName"
              required
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              className="mt-1 w-full rounded border px-3 py-2"
            />
          </Field>
        </div>
        <Field label="Email" error={issue('body.email')}>
          <input
            name="email"
            type="email"
            required
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="mt-1 w-full rounded border px-3 py-2"
          />
        </Field>
        <Field label="Password" error={issue('body.password')}>
          <input
            name="password"
            type="password"
            required
            minLength={8}
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="mt-1 w-full rounded border px-3 py-2"
          />
        </Field>
        <Field label="Your role" error={issue('body.role')}>
          <select
            name="role"
            required
            value={role}
            onChange={(e) => setRole(e.target.value as typeof role)}
            className="mt-1 w-full rounded border px-3 py-2"
          >
            {ROLES.map((r) => (
              <option key={r.value} value={r.value}>
                {r.label}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Organization name" error={issue('body.organization_name')}>
          <input
            name="organizationName"
            required
            value={organizationName}
            onChange={(e) => setOrganizationName(e.target.value)}
            className="mt-1 w-full rounded border px-3 py-2"
            placeholder="Acme Inc."
          />
        </Field>
        {error && Object.keys(fieldIssues).length === 0 ? (
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

function Field({
  label,
  error,
  children,
}: {
  label: string
  error: string | undefined
  children: React.ReactNode
}) {
  return (
    <label className="block">
      <span className="text-sm font-medium">{label}</span>
      {children}
      {error ? <p className="mt-1 text-xs text-red-600">{error}</p> : null}
    </label>
  )
}
