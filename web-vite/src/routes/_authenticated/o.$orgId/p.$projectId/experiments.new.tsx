import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import {
  EMPTY_EXPERIMENT_FORM,
  ExperimentForm,
} from '@/features/experiments/components'
import type { ExperimentFormState } from '@/features/experiments/components'
import {
  createExperiment,
  experimentsKeys,
} from '@/features/experiments/api/queries'
import type { CreateExperimentRequest } from '@/features/experiments/api/types'

// TanStack Router merges sibling-route search schemas along the URL
// path — mirror the parent `experiments.tsx` list shape so `<Link>`
// callers pointing here don't need special search-object shapes.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/experiments/new',
)({
  validateSearch: searchSchema,
  component: NewExperimentPage,
})

function NewExperimentPage() {
  const { orgId, projectId } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: async (state: ExperimentFormState) => {
      const trimmedDesc = state.description.trim()
      const body: CreateExperimentRequest = {
        name: state.name,
        description: trimmedDesc.length > 0 ? trimmedDesc : undefined,
        dataset_id:
          state.datasetId.length > 0 ? state.datasetId : undefined,
      }
      return createExperiment(projectId, body)
    },
    onSuccess: async (created) => {
      await queryClient.invalidateQueries({
        queryKey: experimentsKeys.lists(),
      })
      await navigate({
        to: '/o/$orgId/p/$projectId/experiments/$experimentId',
        params: { orgId, projectId, experimentId: created.id },
        search: { page: 1, limit: 20 },
      })
    },
    onError: (err) => {
      setError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to create experiment',
      )
    },
  })

  return (
    <main className="mx-auto max-w-5xl p-6 space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">New experiment</h1>
        <p className="text-sm text-muted-foreground">
          Create an experiment record. Items can be attached by
          importing from a dataset or streaming from production
          traces via the SDK.
        </p>
      </header>
      <ExperimentForm
        projectId={projectId}
        initial={EMPTY_EXPERIMENT_FORM}
        onSubmit={(state) => {
          setError(null)
          mutation.mutate(state)
        }}
        isSubmitting={mutation.isPending}
        error={error}
        onCancel={() =>
          navigate({
            to: '/o/$orgId/p/$projectId/experiments',
            params: { orgId, projectId },
            search: { page: 1, limit: 20, q: undefined },
          })
        }
      />
    </main>
  )
}
