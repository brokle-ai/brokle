import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { useCallback, useMemo } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/layout/page-header'
import {
  PromptEditLayout,
  promptDetailQueryOptions,
  promptVersionsQueryOptions,
  useCreateVersionMutation,
  useVersionQuery,
  usePromptEditState,
} from '@/features/prompts'
import type {
  CreateVersionRequest,
  PromptVersion,
} from '@/features/prompts'

// Decouple from parent `/prompts` search schema. `version` + `restore`
// are the edit-flow deep-link params (wired via usePromptEditState's
// nuqs hook); declaring them here gives callers a typed search signature.
const searchSchema = z
  .object({
    version: z.string().optional().catch(undefined),
    restore: z.boolean().optional().catch(undefined),
  })
  .catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts/$promptId/edit',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    Promise.all([
      context.queryClient.ensureQueryData(
        promptDetailQueryOptions(params.projectId, params.promptId),
      ),
      context.queryClient.ensureQueryData(
        promptVersionsQueryOptions(params.projectId, params.promptId),
      ),
    ]),
  errorComponent: EditErrorBoundary,
  component: EditPromptPage,
})

function EditErrorBoundary({ error }: { error: Error }) {
  if (
    error instanceof BrokleError &&
    error.status >= 400 &&
    error.status < 500
  ) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load prompt
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function EditPromptPage() {
  const { orgId, projectId, promptId } = Route.useParams()
  const navigate = useNavigate()

  const { sourceVersionId, isRestoreFlow, setSourceVersionId } =
    usePromptEditState()

  const { data: prompt } = useSuspenseQuery(
    promptDetailQueryOptions(projectId, promptId),
  )
  const { data: versions } = useSuspenseQuery(
    promptVersionsQueryOptions(projectId, promptId),
  )

  // Fetch source version if specified and different from current
  const shouldFetchVersion =
    !!sourceVersionId && sourceVersionId !== prompt.version_id
  const { data: fetchedVersion } = useVersionQuery(
    projectId,
    promptId,
    sourceVersionId || '',
    { enabled: shouldFetchVersion },
  )

  const sourceVersion = useMemo<PromptVersion | null>(() => {
    if (!sourceVersionId) {
      return versions.length > 0 ? versions[0] : null
    }
    const fromList = versions.find((v) => v.id === sourceVersionId)
    if (fromList) return fromList
    if (fetchedVersion) return fetchedVersion
    return null
  }, [sourceVersionId, versions, fetchedVersion])

  const createVersionMutation = useCreateVersionMutation(projectId, promptId)

  const handleSave = useCallback(
    async (data: CreateVersionRequest) => {
      await createVersionMutation.mutateAsync(data)
      void navigate({
        to: '/o/$orgId/p/$projectId/prompts/$promptId',
        params: { orgId, projectId, promptId },
      })
    },
    [createVersionMutation, navigate, orgId, projectId, promptId],
  )

  const handleCancel = useCallback(() => {
    void navigate({
      to: '/o/$orgId/p/$projectId/prompts/$promptId',
      params: { orgId, projectId, promptId },
    })
  }, [navigate, orgId, projectId, promptId])

  const handleVersionSelect = useCallback(
    (version: PromptVersion) => {
      setSourceVersionId(version.id)
    },
    [setSourceVersionId],
  )

  return (
    <div className="flex h-full flex-col">
      <div className="px-6 pt-4">
        <PageHeader
          title={`Edit: ${prompt.name}`}
          backHref={`/o/${orgId}/p/${projectId}/prompts/${promptId}`}
          badges={
            <Badge variant={prompt.type === 'chat' ? 'default' : 'secondary'}>
              {prompt.type}
            </Badge>
          }
        />
      </div>

      <div className="flex-1 overflow-hidden">
        <PromptEditLayout
          prompt={prompt}
          versions={versions}
          versionsLoading={false}
          sourceVersion={sourceVersion}
          isRestoreFlow={isRestoreFlow}
          onSave={handleSave}
          onCancel={handleCancel}
          onVersionSelect={handleVersionSelect}
          isSaving={createVersionMutation.isPending}
        />
      </div>
    </div>
  )
}

