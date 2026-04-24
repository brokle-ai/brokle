import { Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  datasetDetailQueryOptions,
  datasetItemsListQueryOptions,
} from '../api/queries'
import { DatasetItemsTable } from './dataset-items-table'

interface DatasetDetailProps {
  orgId: string
  projectId: string
  datasetId: string
  page: number
  limit: number
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function DatasetDetail({
  orgId,
  projectId,
  datasetId,
  page,
  limit,
}: DatasetDetailProps) {
  const { data: dataset } = useSuspenseQuery(
    datasetDetailQueryOptions(projectId, datasetId),
  )
  const { data: itemsPage } = useSuspenseQuery(
    datasetItemsListQueryOptions(projectId, datasetId, { page, limit }),
  )

  const items = itemsPage.data
  const totalItems = itemsPage.total
  const totalPages = Math.max(1, Math.ceil(totalItems / Math.max(1, limit)))
  const hasPrev = page > 1
  const hasNext = page < totalPages

  return (
    <main className="mx-auto max-w-7xl p-6 space-y-6">
      <nav className="text-sm text-muted-foreground">
        <Link
          to="/o/$orgId/p/$projectId/datasets"
          params={{ orgId, projectId }}
          search={{ page: 1, limit: 20, q: undefined }}
          className="hover:text-foreground"
        >
          Datasets
        </Link>
        <span className="mx-2">/</span>
        <span className="text-foreground">{dataset.name}</span>
      </nav>

      <Card>
        <CardHeader>
          <CardTitle>{dataset.name}</CardTitle>
          {dataset.description ? (
            <CardDescription>{dataset.description}</CardDescription>
          ) : null}
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm md:grid-cols-4">
          <div>
            <p className="text-xs text-muted-foreground">Items</p>
            <p className="font-medium">{totalItems.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Created</p>
            <p className="font-medium">{formatTimestamp(dataset.created_at)}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">Updated</p>
            <p className="font-medium">{formatTimestamp(dataset.updated_at)}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground">ID</p>
            <p className="font-mono text-xs">{dataset.id}</p>
          </div>
        </CardContent>
      </Card>

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Items</h2>
        <DatasetItemsTable rows={items} />

        <nav className="flex items-center justify-between">
          <p className="text-xs text-muted-foreground">
            Page {page} of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button asChild variant="outline" size="sm" disabled={!hasPrev}>
              <Link
                to="/o/$orgId/p/$projectId/datasets/$datasetId"
                params={{ orgId, projectId, datasetId }}
                search={{ page: Math.max(1, page - 1), limit }}
              >
                Previous
              </Link>
            </Button>
            <Button asChild variant="outline" size="sm" disabled={!hasNext}>
              <Link
                to="/o/$orgId/p/$projectId/datasets/$datasetId"
                params={{ orgId, projectId, datasetId }}
                search={{ page: page + 1, limit }}
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
