import type { ReactNode } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { formatDistanceToNow } from 'date-fns'
import { Copy } from 'lucide-react'
import { toast } from 'sonner'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DataTableColumnHeader } from '@/components/shared/tables/data-table-column-header'
import { TracesRowActions } from './data-table-row-actions'
import { statuses, statusCodeToString } from '../data/constants'
import type { TraceListItem } from '../api/types'
import { formatDuration, formatCost, formatTokens } from '../utils/format-helpers'

/**
 * Column definitions for the traces table.
 *
 * Visual parity with web/ — but column accessors are adapted to the
 * web-vite wire shape (`TraceListItem` / `TraceSummary`) which is the
 * raw Go backend JSON: snake_case, RFC-3339 timestamp strings, decimal
 * cost as string, `total_tokens` / `span_count` / `total_cost` rather
 * than the Next.js app's transformed camelCase `cost` / `tokens` /
 * `spanCount` shape.
 *
 * The table is rendered with a `renderNameLink` prop so the route owns
 * the `<Link>` target — columns stay framework-agnostic. The optional
 * `rowActionsHref` is threaded through the table context in the same
 * way for the "View Detail" entry in the row actions menu.
 */
export interface TracesColumnsContext {
  // Cell-level factories supplied by the table wrapper, so the column
  // defs don't need to know about TanStack Router.
  renderNameLink?: (trace: TraceListItem, children: ReactNode) => ReactNode
  onViewDetail?: (trace: TraceListItem) => void
  onAddToDataset?: (trace: TraceListItem) => void
  onDelete?: (trace: TraceListItem) => void
}

export function buildTracesColumns(
  ctx: TracesColumnsContext = {},
): ColumnDef<TraceListItem>[] {
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && 'indeterminate')
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
          className="translate-y-0.5"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
          className="translate-y-0.5"
          onClick={(e) => e.stopPropagation()}
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'trace_id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Trace ID" />
      ),
      cell: ({ row }) => {
        const traceId = row.original.trace_id
        const shortId = traceId.substring(0, 8)
        const handleCopy = (e: React.MouseEvent) => {
          e.stopPropagation()
          void navigator.clipboard.writeText(traceId)
          toast.success('Trace ID copied to clipboard')
        }
        return (
          <div className="flex items-center gap-1">
            <span className="font-mono text-xs">{shortId}</span>
            <Button
              variant="ghost"
              size="icon"
              className="h-5 w-5"
              onClick={handleCopy}
            >
              <Copy className="h-3 w-3" />
            </Button>
          </div>
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
      cell: ({ row }) => {
        const trace = row.original
        const inner = <span className="font-medium">{trace.name}</span>
        if (ctx.renderNameLink) {
          return (
            <div className="font-medium">{ctx.renderNameLink(trace, trace.name)}</div>
          )
        }
        return inner
      },
      enableSorting: false,
    },
    {
      accessorKey: 'duration',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Duration" />
      ),
      cell: ({ row }) => (
        <div className="font-mono text-sm">
          {formatDuration(row.original.duration)}
        </div>
      ),
      enableSorting: true,
    },
    {
      accessorKey: 'status_code',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => {
        const statusCode = row.original.status_code ?? 0
        const statusStr = statusCodeToString(statusCode)
        const status = statuses.find((s) => s.value === statusStr)
        if (!status) return null
        const StatusIcon = status.icon
        return (
          <div className="flex items-center gap-2">
            <StatusIcon className="h-4 w-4 text-muted-foreground" />
            <span>{status.label}</span>
          </div>
        )
      },
      filterFn: (row, _id, value: string[]) => {
        const statusCode = row.original.status_code ?? 0
        const statusStr = statusCodeToString(statusCode)
        return value.includes(statusStr)
      },
      enableSorting: false,
    },
    {
      accessorKey: 'model_name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Model" />
      ),
      cell: ({ row }) => {
        const model = row.original.model_name
        if (!model) return <span className="text-muted-foreground">-</span>
        return (
          <Badge variant="secondary" className="font-mono text-xs">
            {model}
          </Badge>
        )
      },
      filterFn: (row, _id, value: string[]) => {
        const model = row.original.model_name
        if (!model) return false
        return value.includes(model)
      },
      enableSorting: true,
    },
    {
      accessorKey: 'provider_name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Provider" />
      ),
      cell: ({ row }) => {
        const provider = row.original.provider_name
        if (!provider) return <span className="text-muted-foreground">-</span>
        return <span className="text-sm capitalize">{provider}</span>
      },
      filterFn: (row, _id, value: string[]) => {
        const provider = row.original.provider_name
        if (!provider) return false
        return value.includes(provider)
      },
      enableSorting: false,
    },
    {
      accessorKey: 'total_cost',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Cost" />
      ),
      cell: ({ row }) => (
        <div className="font-mono text-sm">
          {formatCost(row.original.total_cost)}
        </div>
      ),
      enableSorting: true,
    },
    {
      accessorKey: 'total_tokens',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Tokens" />
      ),
      cell: ({ row }) => {
        const tokens = row.original.total_tokens
        if (!tokens) return <span className="text-muted-foreground">-</span>
        return <div className="font-mono text-sm">{formatTokens(tokens)}</div>
      },
      enableSorting: true,
    },
    {
      accessorKey: 'span_count',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Spans" />
      ),
      cell: ({ row }) => (
        <Badge variant="outline" className="font-mono">
          {row.original.span_count}
        </Badge>
      ),
      enableSorting: true,
    },
    {
      accessorKey: 'start_time',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Timestamp" />
      ),
      cell: ({ row }) => {
        const raw = row.original.start_time
        const d = new Date(raw)
        if (Number.isNaN(d.getTime())) {
          return <div className="text-sm text-muted-foreground">-</div>
        }
        return (
          <div className="text-sm text-muted-foreground">
            {formatDistanceToNow(d, { addSuffix: true })}
          </div>
        )
      },
      enableSorting: true,
    },
    {
      id: 'actions',
      cell: ({ row }) => (
        <TracesRowActions
          trace={row.original}
          onViewDetail={ctx.onViewDetail}
          onAddToDataset={ctx.onAddToDataset}
          onDelete={ctx.onDelete}
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
