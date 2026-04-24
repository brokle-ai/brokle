import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { PromptForm } from '@/features/prompts/components'
import {
  createPrompt,
  promptsKeys,
} from '@/features/prompts/api/queries'
import type {
  CreatePromptRequest,
  PromptFormState,
} from '@/features/prompts/api/types'

// Decouples from parent `/prompts` search schema — see CRITICAL note.
const searchSchema = z.object({}).catch({})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/prompts/new',
)({
  validateSearch: searchSchema,
  component: NewPromptPage,
})

function parseCsv(csv: string): string[] {
  return csv
    .split(',')
    .map((x) => x.trim())
    .filter((x) => x.length > 0)
}

const INITIAL: PromptFormState = {
  name: '',
  type: 'text',
  tagsCsv: '',
  body: '',
  variablesCsv: '',
  commitMessage: '',
}

function NewPromptPage() {
  const { orgId, projectId } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: async (state: PromptFormState) => {
      const body: CreatePromptRequest = {
        name: state.name,
        type: state.type,
        tags: parseCsv(state.tagsCsv),
        template: { content: state.body },
        commit_message: state.commitMessage || undefined,
      }
      return createPrompt(projectId, body)
    },
    onSuccess: async (created) => {
      await queryClient.invalidateQueries({ queryKey: promptsKeys.lists() })
      await navigate({
        to: '/o/$orgId/p/$projectId/prompts/$promptId',
        params: { orgId, projectId, promptId: created.id },
      })
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to create prompt')
    },
  })

  return (
    <main className="mx-auto max-w-5xl p-6 space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">New prompt</h1>
        <p className="text-sm text-muted-foreground">
          Create the first version of a new prompt template.
        </p>
      </header>
      <PromptForm
        mode="create"
        initial={INITIAL}
        onSubmit={(state) => {
          setError(null)
          mutation.mutate(state)
        }}
        isSubmitting={mutation.isPending}
        error={error}
        onCancel={() =>
          navigate({
            to: '/o/$orgId/p/$projectId/prompts',
            params: { orgId, projectId },
            search: { page: 1, limit: 20, q: undefined },
          })
        }
      />
    </main>
  )
}
