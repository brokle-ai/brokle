import { useSuspenseQuery } from '@tanstack/react-query'
import { useParams } from '@tanstack/react-router'
import { organizationMembershipQueryOptions } from '@/features/organizations/queries'
import { projectMembershipQueryOptions } from '@/features/projects/queries'
import { useAuthStore } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import { rawFetch } from '@/lib/api/client'

export function AppHeader() {
  const { orgId, projectId } = useParams({ strict: false }) as {
    orgId?: string
    projectId?: string
  }
  const user = useAuthStore((s) => s.user)

  return (
    <header className="flex h-14 items-center justify-between border-b px-4">
      <div className="flex items-center gap-2 text-sm">
        {orgId ? <OrgCrumb orgId={orgId} /> : null}
        {orgId && projectId ? (
          <>
            <span className="text-muted-foreground">/</span>
            <ProjectCrumb projectId={projectId} />
          </>
        ) : null}
      </div>
      <div className="flex items-center gap-3">
        <span className="text-sm text-muted-foreground">
          {user?.email ?? ''}
        </span>
        <Button size="sm" variant="outline" onClick={onSignOut}>
          Sign out
        </Button>
      </div>
    </header>
  )
}

function OrgCrumb({ orgId }: { orgId: string }) {
  const { data } = useSuspenseQuery(organizationMembershipQueryOptions(orgId))
  return <span>{data.name}</span>
}

function ProjectCrumb({ projectId }: { projectId: string }) {
  const { data } = useSuspenseQuery(projectMembershipQueryOptions(projectId))
  return <span className="font-medium">{data.name}</span>
}

async function onSignOut() {
  try {
    await rawFetch('/api/v1/auth/logout', { method: 'POST' })
  } finally {
    useAuthStore.getState().expireSession()
    window.location.href = '/signin'
  }
}
