import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation, useQueryClient, useSuspenseQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import { PromptForm } from '@/features/prompts/components'
import {
  createPromptVersion,
  promptDetailQueryOptions,
  promptsKeys,
} from '@/features/prompts/api/queries'
import type {
  CreateVersionRequest,
  PromptFormState,
} from '@/features/prompts/api/types'
import { templateToBody } from '@/features/prompts/api/types'

// Decouples from parent `/prompts` search schema — see CRITICAL note.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts/$promptId/edit',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      promptDetailQueryOptions(params.projectId, params.promptId),
    ),
  errorComponent: EditErrorBoundary,
  component: EditPromptPage,
})

function EditErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
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
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)

  const { data: prompt } = useSuspenseQuery(
    promptDetailQueryOptions(projectId, promptId),
  )

  const initial: PromptFormState = {
    name: prompt.name,
    type: prompt.type,
    tagsCsv: prompt.tags.join(', '),
    body: templateToBody(prompt.template),
    variablesCsv: prompt.variables.join(', '),
    commitMessage: '',
  }

  const mutation = useMutation({
    mutationFn: async (state: PromptFormState) => {
      const body: CreateVersionRequest = {
        template: { content: state.body },
        commit_message: state.commitMessage || undefined,
      }
      return createPromptVersion(projectId, promptId, body)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: promptsKeys.detail(projectId, promptId),
      })
      await queryClient.invalidateQueries({
        queryKey: promptsKeys.versions(projectId, promptId),
      })
      await queryClient.invalidateQueries({ queryKey: promptsKeys.lists() })
      await navigate({
        to: '/o/$orgId/p/$projectId/prompts/$promptId',
        params: { orgId, projectId, promptId },
      })
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to save version')
    },
  })

  return (
    <main className="mx-auto max-w-5xl p-6 space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Edit prompt</h1>
        <p className="text-sm text-muted-foreground">
          Saving creates a new version; the current version is preserved.
        </p>
      </header>
      <PromptForm
        mode="edit"
        initial={initial}
        onSubmit={(state) => {
          setError(null)
          mutation.mutate(state)
        }}
        isSubmitting={mutation.isPending}
        error={error}
        onCancel={() =>
          navigate({
            to: '/o/$orgId/p/$projectId/prompts/$promptId',
            params: { orgId, projectId, promptId },
          })
        }
      />
    </main>
  )
}

