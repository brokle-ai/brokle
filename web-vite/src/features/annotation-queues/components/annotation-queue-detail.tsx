import { Link, useNavigate } from '@tanstack/react-router'
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
  queueDetailQueryOptions,
  queueItemsListQueryOptions,
} from '../api/queries'
import { QueueItemsTable } from './queue-items-table'
import type { QueueItemStatus, QueueStatus } from '../api/types'

interface AnnotationQueueDetailProps {
  orgId: string
  projectId: string
  queueId: string
  page: number
  limit: number
  status?: QueueItemStatus
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function StatusBadge({ status }: { status: QueueStatus }) {
  switch (status) {
    case 'active':
      return <Badge variant="secondary">Active</Badge>
    case 'paused':
      return <Badge variant="outline">Paused</Badge>
    case 'archived':
      return <Badge variant="destructive">Archived</Badge>
  }
}

export function AnnotationQueueDetail({
  orgId,
  projectId,
  queueId,
  page,
  limit,
  status,
}: AnnotationQueueDetailProps) {
  const navigate = useNavigate()
  const { data: detail } = useSuspenseQuery(
    queueDetailQueryOptions(projectId, queueId),
  )
  const { data: itemsPage } = useSuspenseQuery(
    queueItemsListQueryOptions(projectId, queueId, { page, limit, status }),
  )

  const { queue, stats } = detail
  const rows = itemsPage.data
  const { total } = itemsPage
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, limit)))
  const hasPrev = page > 1
  const hasNext = page < totalPages

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link
          to="/o/$orgId/p/$projectId/annotation-queues"
          params={{ orgId, projectId }}
          search={{
            page: 1,
            limit: 20,
            status: undefined,
            q: undefined,
          }}
          className="hover:text-foreground"
        >
          Annotation queues
        </Link>
        <span className="mx-2">/</span>
        <span className="text-foreground">{queue.name}</span>
      </nav>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <CardTitle>{queue.name}</CardTitle>
              <StatusBadge status={queue.status} />
            </div>
            <Button
              size="sm"
              onClick={() => {
                // TanStack Router cascades parent search schemas into
                // the leaf's inferred type, so even though the review
                // route's own `validateSearch` is an open record, the
                // target search still requires the parent's
                // page/limit/status/q fields. We supply defaults
                // here; the route strips them at runtime.
                void navigate({
                  to: '/o/$orgId/p/$projectId/annotation-queues/$queueId/review',
                  params: { orgId, projectId, queueId },
                  search: {
                    page: 1,
                    limit: 20,
                    itemStatus: undefined,
                    status: undefined,
                    q: undefined,
                  },
                })
              }}
            >
              Start review
            </Button>
          </div>
          {queue.description ? (
            <CardDescription>{queue.description}</CardDescription>
          ) : null}
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm md:grid-cols-5">
          <div>
            <p className="text-xs text-muted-foreground">Total</p>
            <p className="font-medium">{stats.total_items.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Pending</p>
            <p className="font-medium">
              {stats.pending_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">In review</p>
            <p className="font-medium">
              {stats.in_progress_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Completed</p>
            <p className="font-medium">
              {stats.completed_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Skipped</p>
            <p className="font-medium">
              {stats.skipped_items.toLocaleString()}
            </p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Updated</p>
            <p className="font-medium">{formatTimestamp(queue.updated_at)}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">ID</p>
            <p className="font-mono text-xs">{queue.id}</p>
          </div>
        </CardContent>
      </Card>

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Items</h2>
        <QueueItemsTable rows={rows} />

        <nav className="flex items-center justify-between">
          <p className="text-xs text-muted-foreground">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
              <Link
                to="/o/$orgId/p/$projectId/annotation-queues/$queueId"
                params={{ orgId, projectId, queueId }}
                search={{
                  page: Math.max(1, page - 1),
                  limit,
                  itemStatus: status,
                }}
              >
                Previous
              </Link>
            </Button>
            <Button asChild variant="outline" size="sm" disabled={!hasNext}>
              <Link
                to="/o/$orgId/p/$projectId/annotation-queues/$queueId"
                params={{ orgId, projectId, queueId }}
                search={{ page: page + 1, limit, itemStatus: status }}
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
