import { createFileRoute, Link } from '@tanstack/react-router'
import { useSuspenseQuery } from '@tanstack/react-query'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { InvoicesTable } from '@/features/billing/components'
import { invoiceListQueryOptions } from '@/features/billing/api/queries'

// Same `.catch` pattern as the traces route — a hostile URL falls back
// to sane defaults instead of throwing at validation time.
const searchSchema = z.object({
  page: z.number().int().min(1).catch(1),
  limit: z.number().int().min(1).max(100).catch(20),
})

export const Route = createFileRoute(
  '/_authenticated/o/$orgId/billing/invoices',
)({
  validateSearch: searchSchema,
  loaderDeps: ({ search }) => ({
    page: search.page,
    limit: search.limit,
  }),
  loader: ({ params, context, deps }) =>
    context.queryClient.ensureQueryData(
      invoiceListQueryOptions(params.orgId, {
        page: deps.page,
        limit: deps.limit,
      }),
    ),
  component: InvoicesPage,
})

function InvoicesPage() {
  const { orgId } = Route.useParams()
  const search = Route.useSearch()
  const { data } = useSuspenseQuery(
    invoiceListQueryOptions(orgId, {
      page: search.page,
      limit: search.limit,
    }),
  )

  const { data: rows, pagination } = data

  return (
    <div className="space-y-4">
      <InvoicesTable rows={rows} />

      <nav className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {pagination.total > 0
            ? `Page ${pagination.page} of ${Math.max(1, pagination.total_pages)}`
            : null}
        </p>
        <div className="flex gap-2">
          <Button
            asChild
            variant="outline"
            size="sm"
            disabled={!pagination.has_prev}
          >
            <Link
              to="/o/$orgId/billing/invoices"
              params={{ orgId }}
              search={{
                page: Math.max(1, search.page - 1),
                limit: search.limit,
              }}
            >
              Previous
            </Link>
          </Button>
          <Button
            asChild
            variant="outline"
            size="sm"
            disabled={!pagination.has_next}
          >
            <Link
              to="/o/$orgId/billing/invoices"
              params={{ orgId }}
              search={{
                page: search.page + 1,
                limit: search.limit,
              }}
            >
              Next
            </Link>
          </Button>
        </div>
      </nav>
    </div>
  )
}
