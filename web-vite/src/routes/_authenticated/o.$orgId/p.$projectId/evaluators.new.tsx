import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { z } from 'zod'
import { BrokleError } from '@/lib/api/errors'
import {
  EMPTY_EVALUATOR_FORM,
  EvaluatorForm,
  parseEvaluatorForm,
} from '@/features/evaluators/components'
import type { EvaluatorFormState } from '@/features/evaluators/components'
import {
  createEvaluator,
  evaluatorsKeys,
} from '@/features/evaluators/api/queries'
import type {
  CreateEvaluatorRequest,
  VariableMap,
} from '@/features/evaluators/api/types'

// TanStack Router merges sibling-route search schemas along the URL
// path — `/evaluators/new` inherits `evaluators.tsx`'s page/limit/q
// in the type layer even though this route doesn't read them. Mirror
// the parent shape so Link callers pointing here don't need special
// search-object shapes. `.catch(...)` on every field keeps hostile
// URLs harmless.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
  q: z.string().optional().catch(undefined),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/p/$projectId/evaluators/new',
)({
  validateSearch: searchSchema,
  component: NewEvaluatorPage,
})

function NewEvaluatorPage() {
  const { orgId, projectId } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [error, setError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: async (state: EvaluatorFormState) => {
      const parsed = parseEvaluatorForm(state)
      if (!parsed.ok) throw new Error(parsed.message)
      const v = parsed.value
      const body: CreateEvaluatorRequest = {
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
        // The form ships `unknown[]` from the JSON textarea; the backend
        // owns shape validation, so we narrow to the wire type here.
        variable_mapping: v.variableMapping as VariableMap[],
      }
      return createEvaluator(projectId, body)
    },
    onSuccess: async (created) => {
      await queryClient.invalidateQueries({ queryKey: evaluatorsKeys.lists() })
      await navigate({
        to: '/o/$orgId/p/$projectId/evaluators/$evaluatorId',
        params: { orgId, projectId, evaluatorId: created.id },
        search: { page: 1, limit: 20, q: undefined },
      })
    },
    onError: (err) => {
      setError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to create evaluator',
      )
    },
  })

  return (
    <main className="mx-auto max-w-5xl p-6 space-y-4">
      <header>
        <h1 className="text-2xl font-semibold">New evaluator</h1>
        <p className="text-sm text-muted-foreground">
          Configure scoring rules for spans or traces in this project.
        </p>
      </header>
      <EvaluatorForm
        mode="create"
        initial={EMPTY_EVALUATOR_FORM}
        onSubmit={(state) => {
          setError(null)
          mutation.mutate(state)
        }}
        isSubmitting={mutation.isPending}
        error={error}
        onCancel={() =>
          navigate({
            to: '/o/$orgId/p/$projectId/evaluators',
            params: { orgId, projectId },
            search: { page: 1, limit: 20, q: undefined },
          })
        }
      />
    </main>
  )
}
