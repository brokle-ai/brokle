import { createFileRoute, redirect } from '@tanstack/react-router'
import { organizationListQueryOptions } from '@/features/organizations/queries'
import { projectListQueryOptions } from '@/features/projects/queries'
import { useAuthStore } from '@/stores/auth-store'

// Authenticated root. Every dashboard URL lives under /o/$orgId/p/$projectId/
// (Linear/Vercel precedent — see plan §Re-architecture). This index
// resolves the right default for the current user and forwards them;
// callers that deliberately want a picker screen can navigate to /o
// directly.
export const Route = createFileRoute('/_authenticated/')({
  loader: async ({ context }) => {
    const defaultOrgId = useAuthStore.getState().user?.default_organization_id
    const targetOrgId = defaultOrgId ?? (await resolveFirstOrg(context.queryClient))
    if (!targetOrgId) {
      throw redirect({ to: '/o' })
    }

    const targetProjectId = await resolveFirstProject(context.queryClient, targetOrgId)
    if (!targetProjectId) {
      throw redirect({
        to: '/o/$orgId',
        params: { orgId: targetOrgId },
      })
    }

    throw redirect({
      to: '/o/$orgId/p/$projectId',
      params: { orgId: targetOrgId, projectId: targetProjectId },
    })
  },
  component: () => null,
})

async function resolveFirstOrg(
  qc: import('@tanstack/react-query').QueryClient,
): Promise<string | null> {
  const list = await qc.ensureQueryData(organizationListQueryOptions())
  return list.data[0]?.id ?? null
}

async function resolveFirstProject(
  qc: import('@tanstack/react-query').QueryClient,
  orgId: string,
): Promise<string | null> {
  const list = await qc.ensureQueryData(projectListQueryOptions(orgId))
  return list.data[0]?.id ?? null
}
