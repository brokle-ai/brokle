import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { useCallback } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/layout/page-header'
import {
  PromptDetailLayout,
  promptDetailQueryOptions,
  promptVersionsQueryOptions,
  usePromptDetailState,
  useProtectedLabelsQuery,
  useSetLabelsMutation,
} from '@/features/prompts'
import type { PromptVersion } from '@/features/prompts'

// Explicit empty validateSearch decouples this leaf from the parent
// `/prompts` search schema. nuqs keys for detail state (`version`,
// `tab`) are still accepted via passthrough.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts/$promptId/',
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
  errorComponent: PromptDetailErrorBoundary,
  component: PromptDetailPage,
})

function PromptDetailErrorBoundary({ error }: { error: Error }) {
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

function PromptDetailPage() {
  const { orgId, projectId, promptId } = Route.useParams()
  const navigate = useNavigate()

  const { data: prompt } = useSuspenseQuery(
    promptDetailQueryOptions(projectId, promptId),
  )
  const { data: versions } = useSuspenseQuery(
    promptVersionsQueryOptions(projectId, promptId),
  )
  const { data: protectedLabels = [] } = useProtectedLabelsQuery(projectId)

  const setLabelsMutation = useSetLabelsMutation(projectId, promptId)
  const { selectedVersionId, setSelectedVersionId } = usePromptDetailState()

  const availableLabels = Array.from(
    new Set(versions.flatMap((v) => v.labels)),
  )

  const handleLabelsChange = useCallback(
    async (labels: string[]) => {
      const versionId = selectedVersionId || prompt.version_id
      await setLabelsMutation.mutateAsync({ versionId, labels })
    },
    [prompt.version_id, selectedVersionId, setLabelsMutation],
  )

  const handleRestore = useCallback(
    (version: PromptVersion) => {
      void navigate({
        to: '/o/$orgId/p/$projectId/prompts/$promptId/edit',
        params: { orgId, projectId, promptId },
        // Pass version + restore flag via raw search; typed-search
        // would require these keys in the route schema and we keep
        // the edit route's schema open.
        search: { version: version.id, restore: true },
      })
    },
    [navigate, orgId, projectId, promptId],
  )

  const handleCreateVersion = useCallback(() => {
    void navigate({
      to: '/o/$orgId/p/$projectId/prompts/$promptId/edit',
      params: { orgId, projectId, promptId },
    })
  }, [navigate, orgId, projectId, promptId])

  return (
    <div className="flex h-full flex-col">
      <div className="px-6 pt-4">
        <PageHeader
          title={prompt.name}
          backHref={`/o/${orgId}/p/${projectId}/prompts`}
          description={prompt.description}
          badges={
            <Badge variant={prompt.type === 'chat' ? 'default' : 'secondary'}>
              {prompt.type}
            </Badge>
          }
        />
      </div>

      <div className="flex-1 overflow-hidden">
        <PromptDetailLayout
          prompt={prompt}
          versions={versions}
          versionsLoading={false}
          protectedLabels={protectedLabels}
          availableLabels={availableLabels}
          projectId={projectId}
          projectSlug={projectId}
          selectedVersionId={selectedVersionId}
          onVersionChange={setSelectedVersionId}
          onLabelsChange={handleLabelsChange}
          onRestore={handleRestore}
          onCreateVersion={handleCreateVersion}
          isLabelsLoading={setLabelsMutation.isPending}
        />
      </div>
    </div>
  )
}
