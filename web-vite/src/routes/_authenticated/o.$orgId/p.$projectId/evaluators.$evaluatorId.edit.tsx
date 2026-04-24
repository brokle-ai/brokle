import { createFileRoute, useNavigate } from '@tanstack/react-router'
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import {
  EvaluatorForm,
  parseEvaluatorForm,
} from '@/features/evaluators/components'
import type { EvaluatorFormState } from '@/features/evaluators/components'
import {
  evaluatorDetailQueryOptions,
  evaluatorsKeys,
  updateEvaluator,
} from '@/features/evaluators/api/queries'
import type {
  EvaluatorDetail,
  UpdateEvaluatorRequest,
} from '@/features/evaluators/api/types'

// TanStack Router merges sibling-route search schemas along the URL
// path — mirror the parent `evaluators.tsx` shape so `<Link to="…/edit">`
// callers don't need special search-object shapes.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/evaluators/$evaluatorId/edit',
)({
  validateSearch: searchSchema,
  loader: ({ params, context }) =>
    context.queryClient.ensureQueryData(
      evaluatorDetailQueryOptions(params.projectId, params.evaluatorId),
    ),
  errorComponent: EditErrorBoundary,
  component: EditEvaluatorPage,
})

function EditErrorBoundary({ error }: { error: Error }) {
  if (error instanceof BrokleError && error.status >= 400 && error.status < 500) {
    return (
      <main className="mx-auto max-w-5xl p-6">
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load evaluator
          </p>
          <p className="mt-1 text-sm text-muted-foreground">{error.message}</p>
        </div>
      </main>
    )
  }
  throw error
}

function seedFormState(evaluator: EvaluatorDetail): EvaluatorFormState {
  return {
    name: evaluator.name,
    description: evaluator.description ?? '',
    scorerType: evaluator.scorer_type,
    targetScope: evaluator.target_scope,
    triggerType: evaluator.trigger_type,
    status: evaluator.status,
    samplingRate: String(evaluator.sampling_rate),
    spanNamesCsv: evaluator.span_names.join(', '),
    scorerConfigJson: JSON.stringify(evaluator.scorer_config ?? {}, null, 2),
    filterJson: JSON.stringify(evaluator.filter ?? [], null, 2),
    variableMappingJson: JSON.stringify(
      evaluator.variable_mapping ?? [],
      null,
      2,
    ),
  }
}

function EditEvaluatorPage() {
  const { orgId, projectId, evaluatorId } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)

  const { data: evaluator } = useSuspenseQuery(
    evaluatorDetailQueryOptions(projectId, evaluatorId),
  )

  const initial = seedFormState(evaluator)

  const mutation = useMutation({
    mutationFn: async (state: EvaluatorFormState) => {
      const parsed = parseEvaluatorForm(state)
      if (!parsed.ok) throw new Error(parsed.message)
      const v = parsed.value
      const body: UpdateEvaluatorRequest = {
        name: v.name,
        description: v.description,
        status: v.status,
        trigger_type: v.triggerType,
        target_scope: v.targetScope,
        filter: v.filter,
        span_names: v.spanNames,
        sampling_rate: v.samplingRate,
        scorer_type: v.scorerType,
        scorer_config: v.scorerConfig,
        variable_mapping: v.variableMapping,
      }
      return updateEvaluator(projectId, evaluatorId, body)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: evaluatorsKeys.detail(evaluatorId),
      })
      await queryClient.invalidateQueries({ queryKey: evaluatorsKeys.lists() })
      await navigate({
        to: '/o/$orgId/p/$projectId/evaluators/$evaluatorId',
        params: { orgId, projectId, evaluatorId },
        search: { page: 1, limit: 20, q: undefined },
      })
    },
    onError: (err) => {
      setError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to save evaluator',
      )
    },
  })

  return (
    <main className="mx-auto max-w-5xl p-6 space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">Edit evaluator</h1>
        <p className="text-sm text-muted-foreground">
          Updates apply immediately — in-flight scoring uses the new
          configuration.
        </p>
      </header>
      <EvaluatorForm
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
            to: '/o/$orgId/p/$projectId/evaluators/$evaluatorId',
            params: { orgId, projectId, evaluatorId },
            search: { page: 1, limit: 20, q: undefined },
          })
        }
      />
    </main>
  )
}
