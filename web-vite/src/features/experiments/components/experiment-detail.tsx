import { Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  experimentDetailQueryOptions,
  experimentItemsListQueryOptions,
  experimentMetricsQueryOptions,
} from '../api/queries'
import { ExperimentItemsTable } from './experiment-items-table'
import type { ExperimentDetail as ExperimentDetailType } from '../api/types'

interface ExperimentDetailProps {
  orgId: string
  projectId: string
  experimentId: string
  limit: number
  offset: number
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
  limit,
  offset,
}: ExperimentDetailProps) {
  const { data: experiment } = useSuspenseQuery(
    experimentDetailQueryOptions(projectId, experimentId),
  )
  const { data: itemsResp } = useSuspenseQuery(
    experimentItemsListQueryOptions(projectId, experimentId, {
      limit,
      offset,
    }),
  )
  const { data: metrics } = useSuspenseQuery(
    experimentMetricsQueryOptions(projectId, experimentId),
  )

  const items = itemsResp.items
  const total = itemsResp.total
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, limit)))
  const currentPage = Math.floor(offset / Math.max(1, limit)) + 1
  const hasPrev = offset > 0
  const hasNext = offset + limit < total

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
        <CardHeader>
          <div className="flex items-center gap-3">
            <CardTitle>{experiment.name}</CardTitle>
            <StatusBadge status={experiment.status} />
          </div>
          {experiment.description ? (
            <CardDescription>{experiment.description}</CardDescription>
          ) : null}
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
                  page: currentPage,
                  limit,
                  offset: Math.max(0, offset - limit),
                }}
              >
                Previous
              </Link>
            </Button>
            <Button asChild variant="outline" size="sm" disabled={!hasNext}>
              <Link
                to="/o/$orgId/p/$projectId/experiments/$experimentId"
                params={{ orgId, projectId, experimentId }}
                search={{ page: currentPage, limit, offset: offset + limit }}
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
