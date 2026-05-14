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
import { BrokleError } from '@/lib/api/errors'
import {
  experimentDetailQueryOptions,
  experimentItemsListQueryOptions,
  experimentMetricsQueryOptions,
  experimentsKeys,
  rerunExperiment,
} from '../api/queries'
import { ExperimentDeleteButton } from './experiment-delete-button'
import { ExperimentItemsTable } from './experiment-items-table'
import type { ExperimentDetail as ExperimentDetailType } from '../api/types'

interface ExperimentDetailProps {
  orgId: string
  projectId: string
  experimentId: string
  page: number
  limit: number
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: ExperimentDetailType['status'] }) {
  if (status === 'failed') return <Badge variant="destructive">Failed</Badge>
  if (status === 'completed') return <Badge variant="secondary">Completed</Badge>
  if (status === 'running') return <Badge variant="secondary">Running</Badge>
  if (status === 'cancelled') return <Badge variant="outline">Cancelled</Badge>
  return <Badge variant="outline">Draft</Badge>
}

export function ExperimentDetail({
  orgId,
  projectId,
  experimentId,
  page,
  limit,
}: ExperimentDetailProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [rerunError, setRerunError] = useState<string | null>(null)
  const { data: experiment } = useSuspenseQuery(
    experimentDetailQueryOptions(projectId, experimentId),
  )

  const rerunMutation = useMutation({
    mutationFn: () => rerunExperiment(projectId, experimentId),
    onSuccess: async (fresh) => {
      await queryClient.invalidateQueries({
        queryKey: experimentsKeys.lists(),
      })
      await navigate({
        to: '/o/$orgId/p/$projectId/experiments/$experimentId',
        params: { orgId, projectId, experimentId: fresh.id },
        search: { page: 1, limit: 20 },
      })
    },
    onError: (err) => {
      setRerunError(
        err instanceof BrokleError
          ? err.message
          : err instanceof Error
            ? err.message
            : 'Failed to rerun experiment',
      )
    },
  })
  const { data: itemsResp } = useSuspenseQuery(
    experimentItemsListQueryOptions(projectId, experimentId, {
      page,
      limit,
    }),
  )
  const { data: metrics } = useSuspenseQuery(
    experimentMetricsQueryOptions(projectId, experimentId),
  )

  const items = itemsResp.data
  const { total_pages: totalPagesRaw, has_next: hasNext, has_prev: hasPrev } =
    itemsResp.pagination
  const totalPages = Math.max(1, totalPagesRaw)
  const currentPage = page

  // Pick a single "headline" score from the metrics map — the first one
  // present. The metrics endpoint returns a map keyed by score name
  // (arbitrary per-evaluator); callers that need all scores can open
  // the dedicated metrics panel (not shipped yet).
  const scoreEntries = metrics.scores ? Object.entries(metrics.scores) : []
  const headlineScore = scoreEntries[0]

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link
          to="/o/$orgId/p/$projectId/experiments"
          params={{ orgId, projectId }}
          search={{ page: 1, limit: 20, q: undefined }}
          className="hover:text-foreground"
        >
          Experiments
        </Link>
        <span className="mx-2">/</span>
        <span className="text-foreground">{experiment.name}</span>
      </nav>

      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex items-center gap-3">
              <CardTitle>{experiment.name}</CardTitle>
              <StatusBadge status={experiment.status} />
            </div>
            {experiment.description ? (
              <CardDescription>{experiment.description}</CardDescription>
            ) : null}
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={rerunMutation.isPending}
              onClick={() => {
                setRerunError(null)
                rerunMutation.mutate()
              }}
            >
              {rerunMutation.isPending ? 'Rerunning…' : 'Rerun'}
            </Button>
            <ExperimentDeleteButton
              projectId={projectId}
              experimentId={experimentId}
              onDeleted={() =>
                navigate({
                  to: '/o/$orgId/p/$projectId/experiments',
                  params: { orgId, projectId },
                  search: { page: 1, limit: 20, q: undefined },
                })
              }
            />
          </div>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm md:grid-cols-4">
          <div>
            <p className="text-xs text-muted-foreground">Dataset</p>
            <p className="font-mono text-xs">
              {experiment.dataset_id
                ? `${experiment.dataset_id.slice(0, 8)}…`
                : '—'}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Source</p>
            <p className="font-medium">{experiment.source}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Created</p>
            <p className="font-medium">
              {formatTimestamp(experiment.created_at)}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Started</p>
            <p className="font-medium">
              {formatTimestamp(experiment.started_at)}
            </p>
          </div>
        </CardContent>
      </Card>

      {rerunError ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          {rerunError}
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Summary</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm md:grid-cols-5">
          <div>
            <p className="text-xs text-muted-foreground">Items</p>
            <p className="font-medium">
              {metrics.progress.total_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Completed</p>
            <p className="font-medium">
              {metrics.progress.completed_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Failed</p>
            <p
              className={
                metrics.progress.failed_items > 0
                  ? 'font-medium text-destructive'
                  : 'font-medium'
              }
            >
              {metrics.progress.failed_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Progress</p>
            <p className="font-medium">
              {`${metrics.progress.progress_pct.toFixed(1)}%`}
            </p>
          </div>
          {headlineScore ? (
            <div>
              <p className="text-xs text-muted-foreground">
                Avg {headlineScore[0]}
              </p>
              <p className="font-medium">{headlineScore[1].mean.toFixed(3)}</p>
            </div>
          ) : null}
        </CardContent>
      </Card>

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Runs</h2>
        <ExperimentItemsTable rows={items} />

        <nav className="flex items-center justify-between">
          <p className="text-xs text-muted-foreground">
            Page {currentPage} of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
              <Link
                to="/o/$orgId/p/$projectId/experiments/$experimentId"
                params={{ orgId, projectId, experimentId }}
                search={{
                  page: Math.max(1, currentPage - 1),
                  limit,
                }}
              >
                Previous
              </Link>
            </Button>
            <Button asChild variant="outline" size="sm" disabled={!hasNext}>
              <Link
                to="/o/$orgId/p/$projectId/experiments/$experimentId"
                params={{ orgId, projectId, experimentId }}
                search={{ page: currentPage + 1, limit }}
              >
                Next
              </Link>
            </Button>
          </div>
        </nav>
      </section>
    </main>
  )
}
