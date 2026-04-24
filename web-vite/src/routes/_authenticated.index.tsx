import { createFileRoute, Link } from '@tanstack/react-router'
import { useCurrentUser } from '@/features/auth/hooks'

// Pilot route. The full dashboard home lives under the tenancy path
// `/o/$orgId/p/$projectId/` (Phase 1.3c); this index lands a
// signed-in user here if they haven't picked an org/project yet.
export const Route = createFileRoute('/_authenticated/')({
  component: AuthenticatedHome,
})

function AuthenticatedHome() {
  const { data: user, isPending } = useCurrentUser()
  if (isPending) return <p className="p-6">Loading…</p>

  return (
    <main className="p-6 space-y-4">
      <h1 className="text-xl font-semibold">
        Welcome {user?.first_name ?? 'back'}
      </h1>
      <p className="text-sm text-muted-foreground">
        Phase 1.3a scaffold — tenancy routes arrive in 1.3c.
      </p>
      <Link
        to="/"
        className="underline text-sm"
      >
        back to public landing
      </Link>
    </main>
  )
}
