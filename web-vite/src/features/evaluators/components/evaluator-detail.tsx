import { Link, useNavigate } from '@tanstack/react-router'
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query'
import { useState } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ReadonlyCode } from '@/editors/readonly-code'
import { BrokleError } from '@/lib/api/errors'
import {
  activateEvaluator,
  deactivateEvaluator,
  evaluatorDetailQueryOptions,
  evaluatorsKeys,
} from '../api/queries'
import type { EvaluatorDetail as EvaluatorDetailType } from '../api/types'
import { EvaluatorDeleteButton } from './evaluator-delete-button'
import { TestEvaluatorDialog } from './test-evaluator-dialog'
import { ScorerConfigDisplay } from './scorer-config-display'
import { EvaluatorExecutionsTable } from './evaluator-executions-table'
import { EvaluatorAnalyticsTab } from './evaluator-analytics-tab'

interface EvaluatorDetailProps {
  orgId: string
  projectId: string
  evaluatorId: string
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: EvaluatorDetailType['status'] }) {
  if (status === 'active') return <Badge variant="secondary">Active</Badge>
  if (status === 'paused') return <Badge variant="outline">Paused</Badge>
  return <Badge variant="outline">Inactive</Badge>
}

function prettyJSON(value: unknown): string {
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

export function EvaluatorDetail({
  orgId,
  projectId,
  evaluatorId,
}: EvaluatorDetailProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [toggleError, setToggleError] = useState<string | null>(null)
  const [testOpen, setTestOpen] = useState(false)
  const { data: evaluator } = useSuspenseQuery(
    evaluatorDetailQueryOptions(projectId, evaluatorId),
  )

  const toggleMutation = useMutation({
    mutationFn: () => {
      if (evaluator.status === 'active') {
        return deactivateEvaluator(projectId, evaluatorId)
      }
      return activateEvaluator(projectId, evaluatorId)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: evaluatorsKeys.detail(evaluatorId),
      })
      await queryClient.invalidateQueries({ queryKey: evaluatorsKeys.lists() })
    },
    onError: (err) => {
      setToggleError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to update status',
      )
    },
  })

  const llmModel =
    evaluator.scorer_type === 'llm'
      ? (evaluator.scorer_config as { model?: string }).model
      : undefined

  return (
    <main className="mx-auto max-w-7xl space-y-6 p-6">
      <nav className="text-sm text-muted-foreground">
        <Link
          to="/o/$orgId/p/$projectId/evaluators"
          params={{ orgId, projectId }}
          search={{ page: 1, limit: 20, q: undefined }}
          className="hover:text-foreground"
        >
          Evaluators
        </Link>
        <span className="mx-2">/</span>
        <span className="text-foreground">{evaluator.name}</span>
      </nav>

      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex items-center gap-3">
              <CardTitle>{evaluator.name}</CardTitle>
              <StatusBadge status={evaluator.status} />
            </div>
            {evaluator.description ? (
              <CardDescription>{evaluator.description}</CardDescription>
            ) : null}
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setTestOpen(true)}
            >
              Test
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={toggleMutation.isPending}
              onClick={() => {
                setToggleError(null)
                toggleMutation.mutate()
              }}
            >
              {toggleMutation.isPending
                ? 'Saving…'
                : evaluator.status === 'active'
                  ? 'Deactivate'
                  : 'Activate'}
            </Button>
            <Button asChild variant="outline" size="sm">
              <Link
                to="/o/$orgId/p/$projectId/evaluators/$evaluatorId/edit"
                params={{ orgId, projectId, evaluatorId }}
                search={{ page: 1, limit: 20, q: undefined }}
              >
                Edit
              </Link>
            </Button>
            <EvaluatorDeleteButton
              projectId={projectId}
              evaluatorId={evaluatorId}
              onDeleted={() =>
                navigate({
                  to: '/o/$orgId/p/$projectId/evaluators',
                  params: { orgId, projectId },
                  search: { page: 1, limit: 20, q: undefined },
                })
              }
            />
          </div>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm md:grid-cols-4">
          <div>
            <p className="text-xs text-muted-foreground">Kind</p>
            <p className="font-medium">{evaluator.scorer_type}</p>
          </div>
          {llmModel ? (
            <div>
              <p className="text-xs text-muted-foreground">Model</p>
              <p className="font-medium">{llmModel}</p>
            </div>
          ) : null}
          <div>
            <p className="text-xs text-muted-foreground">Trigger</p>
            <p className="font-medium">{evaluator.trigger_type}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Scope</p>
            <p className="font-medium">{evaluator.target_scope}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Sampling</p>
            <p className="font-medium">
              {`${(evaluator.sampling_rate * 100).toFixed(0)}%`}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Updated</p>
            <p className="font-medium">
              {formatTimestamp(evaluator.updated_at)}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">ID</p>
            <p className="font-mono text-xs">{evaluator.id}</p>
          </div>
        </CardContent>
      </Card>

      {toggleError ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {toggleError}
        </div>
      ) : null}

      <Tabs defaultValue="config" className="w-full">
        <TabsList>
          <TabsTrigger value="config">Configuration</TabsTrigger>
          <TabsTrigger value="runs">Runs</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
        </TabsList>

        <TabsContent value="config" className="space-y-6 mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Scorer config</CardTitle>
              <CardDescription>
                Visualised configuration for the {evaluator.scorer_type}{' '}
                scorer.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ScorerConfigDisplay
                scorerType={evaluator.scorer_type}
                config={evaluator.scorer_config}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Filter rules</CardTitle>
              <CardDescription>
                Incoming traces/spans must match every clause to be scored.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {evaluator.filter && evaluator.filter.length > 0 ? (
                <ReadonlyCode
                  code={prettyJSON(evaluator.filter)}
                  language="json"
                />
              ) : (
                <p className="text-sm text-muted-foreground">
                  No filters — every span/trace in scope is evaluated.
                </p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Variable mapping</CardTitle>
              <CardDescription>
                Pulls template variables from span input/output/metadata.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {evaluator.variable_mapping &&
              evaluator.variable_mapping.length > 0 ? (
                <ReadonlyCode
                  code={prettyJSON(evaluator.variable_mapping)}
                  language="json"
                />
              ) : (
                <p className="text-sm text-muted-foreground">
                  No variable mappings configured.
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="runs" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Execution history</CardTitle>
              <CardDescription>
                Recent automatic and manual evaluator runs. Click a row to
                inspect span-level scores and prompt/response data.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <EvaluatorExecutionsTable
                orgId={orgId}
                projectId={projectId}
                evaluatorId={evaluatorId}
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="analytics" className="mt-6">
          <EvaluatorAnalyticsTab
            projectId={projectId}
            evaluatorId={evaluatorId}
          />
        </TabsContent>
      </Tabs>

      <TestEvaluatorDialog
        projectId={projectId}
        evaluatorId={evaluatorId}
        evaluatorName={evaluator.name}
        open={testOpen}
        onOpenChange={setTestOpen}
      />
    </main>
  )
}
