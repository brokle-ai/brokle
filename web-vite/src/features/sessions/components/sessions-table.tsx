import { useState, type ReactNode } from 'react'
import {
  type ColumnDef,
  type ExpandedState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  type Row,
  useReactTable,
} from '@tanstack/react-table'
import { useQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  ExternalLink,
} from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { SessionListItem } from '../api/types'
import { traceListQueryOptions } from '@/features/traces/api/queries'

interface SessionsTableProps {
  rows: SessionListItem[]
  /**
   * Project context — required for the inline expansion to fetch
   * traces scoped to a session via the traces list endpoint
   * (`/api/v1/traces?session_id=...`).
   */
  projectId: string
  /**
   * Wrap the session_id cell in a router-typed `<Link>`. Parent owns
   * routing, table stays framework-agnostic. Pattern mirrors
   * `TracesTable.renderNameLink`. Not called for the "no-session"
   * synthetic row (session_id === 'no-session') since those traces
   * have no identifying attribute to link to.
   */
  renderSessionLink?: (
    session: SessionListItem,
    children: ReactNode,
  ) => ReactNode
  /**
   * Render the "open detail" trailing icon as a router-typed Link.
   * Kept as a render prop so the table doesn't depend on the router.
   */
  renderDetailLink?: (
    session: SessionListItem,
    children: ReactNode,
  ) => ReactNode
  /**
   * Render a trace row's name as a Link to the trace-detail route.
   * Used inside the expanded preview panel.
   */
  renderTraceLink?: (
    traceId: string,
    children: ReactNode,
  ) => ReactNode
}

// Top-N traces shown inline on row expansion. Backend caps page size
// at 100; 5 is enough for an at-a-glance preview without making the
// row balloon.
const PREVIEW_LIMIT = 5

// Duration arrives in nanoseconds (summed OTLP span durations).
function formatDurationNs(ns: number | undefined): string {
  if (ns === undefined || ns === null || ns === 0) return '—'
  const ms = ns / 1_000_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  const s = ms / 1_000
  if (s < 60) return `${s.toFixed(2)}s`
  const m = s / 60
  if (m < 60) return `${m.toFixed(1)}m`
  const h = m / 60
  return `${h.toFixed(1)}h`
}

function formatCost(cost: number | undefined): string {
  if (cost === undefined || cost === null) return '—'
  if (cost === 0) return '$0.00'
  if (cost < 0.01) return `$${cost.toFixed(6)}`
  return `$${cost.toFixed(4)}`
}

// Trace-list cost arrives as a decimal string from shopspring/decimal —
// parse at the render boundary.
function formatCostString(cost: string | undefined): string {
  if (!cost) return '—'
  const n = Number(cost)
  if (!Number.isFinite(n)) return cost
  return formatCost(n)
}

function formatTokens(tokens: number | undefined): string {
  if (tokens === undefined || tokens === null || tokens === 0) return '—'
  return tokens.toLocaleString()
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function SessionsTable({
  rows,
  projectId,
  renderSessionLink,
  renderDetailLink,
  renderTraceLink,
}: SessionsTableProps) {
  const [expanded, setExpanded] = useState<ExpandedState>({})

  const columns: ColumnDef<SessionListItem>[] = [
    {
      id: 'expander',
      header: () => null,
      cell: ({ row }) => (
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6"
          onClick={(e) => {
            e.stopPropagation()
            row.toggleExpanded()
          }}
          aria-label={row.getIsExpanded() ? 'Collapse row' : 'Expand row'}
        >
          {row.getIsExpanded() ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </Button>
      ),
    },
    {
      id: 'session_id',
      header: 'Session ID',
      cell: ({ row }) => {
        const s = row.original
        const isNoSession = s.session_id === 'no-session'
        const label = (
          <span className="font-mono text-xs">{s.session_id.slice(0, 24)}</span>
        )
        if (isNoSession) {
          return (
            <span className="italic text-muted-foreground">No session</span>
          )
        }
        return renderSessionLink ? renderSessionLink(s, label) : label
      },
    },
    {
      id: 'trace_count',
      header: 'Traces',
      cell: ({ row }) => (
        <Badge variant="outline" className="font-mono">
          {row.original.trace_count.toLocaleString()}
        </Badge>
      ),
    },
    {
      id: 'error_count',
      header: 'Errors',
      cell: ({ row }) =>
        row.original.error_count > 0 ? (
          <Badge variant="destructive">{row.original.error_count}</Badge>
        ) : (
          <span className="text-muted-foreground">0</span>
        ),
    },
    {
      id: 'users',
      header: 'Users',
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.user_ids.length}
        </span>
      ),
    },
    {
      id: 'duration',
      header: 'Duration',
      cell: ({ row }) => formatDurationNs(row.original.total_duration),
    },
    {
      id: 'tokens',
      header: 'Tokens',
      cell: ({ row }) => formatTokens(row.original.total_tokens),
    },
    {
      id: 'cost',
      header: 'Cost',
      cell: ({ row }) => formatCost(row.original.total_cost),
    },
    {
      id: 'first_trace',
      header: 'First seen',
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {formatTimestamp(row.original.first_trace)}
        </span>
      ),
    },
    {
      id: 'last_trace',
      header: 'Last seen',
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {formatTimestamp(row.original.last_trace)}
        </span>
      ),
    },
    {
      id: 'actions',
      header: () => null,
      cell: ({ row }) => {
        const s = row.original
        if (s.session_id === 'no-session') return null
        const icon = (
          <ExternalLink className="h-3.5 w-3.5" />
        )
        if (renderDetailLink) {
          return (
            <span
              className="inline-flex items-center text-muted-foreground hover:text-foreground"
              onClick={(e) => e.stopPropagation()}
              title="Open session detail"
            >
              {renderDetailLink(s, icon)}
            </span>
          )
        }
        return null
      },
    },
  ]

  const table = useReactTable({
    data: rows,
    columns,
    state: { expanded },
    getRowId: (row) => row.session_id,
    onExpandedChange: setExpanded,
    getRowCanExpand: (row) => row.original.session_id !== 'no-session',
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
  })

  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-12 text-center">
        <p className="text-sm text-muted-foreground">No sessions yet</p>
        <p className="mt-1 text-xs text-muted-foreground">
          Sessions appear here once traces carry a <code>session_id</code>{' '}
          attribute.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((hg) => (
            <TableRow key={hg.id}>
              {hg.headers.map((header) => (
                <TableHead key={header.id}>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.map((row) => (
            <RowWithExpansion
              key={row.id}
              row={row}
              colSpan={columns.length}
              projectId={projectId}
              renderTraceLink={renderTraceLink}
              renderDetailLink={renderDetailLink}
            />
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

function RowWithExpansion({
  row,
  colSpan,
  projectId,
  renderTraceLink,
  renderDetailLink,
}: {
  row: Row<SessionListItem>
  colSpan: number
  projectId: string
  renderTraceLink?: (traceId: string, children: ReactNode) => ReactNode
  renderDetailLink?: (
    session: SessionListItem,
    children: ReactNode,
  ) => ReactNode
}) {
  const expandable = row.getCanExpand()
  return (
    <>
      <TableRow
        className={expandable ? 'cursor-pointer hover:bg-muted/40' : ''}
        onClick={() => {
          if (expandable) row.toggleExpanded()
        }}
      >
        {row.getVisibleCells().map((cell) => (
          <TableCell key={cell.id}>
            {flexRender(cell.column.columnDef.cell, cell.getContext())}
          </TableCell>
        ))}
      </TableRow>
      {row.getIsExpanded() && (
        <TableRow>
          <TableCell colSpan={colSpan} className="bg-muted/30 p-0">
            <ExpandedSessionPanel
              session={row.original}
              projectId={projectId}
              renderTraceLink={renderTraceLink}
              renderDetailLink={renderDetailLink}
            />
          </TableCell>
        </TableRow>
      )}
    </>
  )
}

function ExpandedSessionPanel({
  session,
  projectId,
  renderTraceLink,
  renderDetailLink,
}: {
  session: SessionListItem
  projectId: string
  renderTraceLink?: (traceId: string, children: ReactNode) => ReactNode
  renderDetailLink?: (
    session: SessionListItem,
    children: ReactNode,
  ) => ReactNode
}) {
  const { data, isLoading, isError, error } = useQuery({
    ...traceListQueryOptions(projectId, {
      page: 1,
      limit: PREVIEW_LIMIT,
      range: 'all',
      sessionId: session.session_id,
    }),
  })

  const traces = data?.data ?? []

  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center justify-between">
        <div className="text-sm font-medium">Session summary</div>
        {renderDetailLink && (
          <span className="inline-block">
            {renderDetailLink(
              session,
              <Button variant="outline" size="sm" className="gap-1.5">
                <ExternalLink className="h-3.5 w-3.5" />
                Open session
              </Button>,
            )}
          </span>
        )}
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <Stat label="Traces" value={session.trace_count.toLocaleString()} />
        <Stat label="Duration" value={formatDurationNs(session.total_duration)} />
        <Stat label="Tokens" value={formatTokens(session.total_tokens)} />
        <Stat label="Cost" value={formatCost(session.total_cost)} />
      </div>

      {(session.error_count > 0 || session.user_ids.length > 0) && (
        <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
          {session.error_count > 0 && (
            <span className="inline-flex items-center gap-1.5 text-destructive">
              <AlertTriangle className="h-3.5 w-3.5" />
              {session.error_count} error
              {session.error_count === 1 ? '' : 's'}
            </span>
          )}
          {session.user_ids.length > 0 && (
            <span>
              Users: {session.user_ids.slice(0, 5).join(', ')}
              {session.user_ids.length > 5 &&
                ` +${session.user_ids.length - 5} more`}
            </span>
          )}
        </div>
      )}

      <div className="space-y-2">
        <div className="flex items-baseline justify-between">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Recent traces
          </p>
          {session.trace_count > PREVIEW_LIMIT && (
            <p className="text-xs text-muted-foreground">
              Showing {Math.min(traces.length, PREVIEW_LIMIT)} of{' '}
              {session.trace_count.toLocaleString()}
            </p>
          )}
        </div>
        {isLoading ? (
          <p className="text-sm text-muted-foreground">Loading traces…</p>
        ) : isError ? (
          <p className="text-sm text-destructive">
            {error instanceof Error
              ? error.message
              : 'Failed to load traces'}
          </p>
        ) : traces.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            No traces in this session.
          </p>
        ) : (
          <div className="overflow-hidden rounded-md border bg-background">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Trace</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Tokens</TableHead>
                  <TableHead className="text-right">Cost</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {traces.map((trace) => {
                  const nameNode = (
                    <span className="font-medium">
                      {trace.name || trace.trace_id.slice(0, 16)}
                    </span>
                  )
                  return (
                    <TableRow key={trace.trace_id}>
                      <TableCell>
                        {renderTraceLink
                          ? renderTraceLink(trace.trace_id, nameNode)
                          : nameNode}
                      </TableCell>
                      <TableCell>
                        {trace.has_error ? (
                          <Badge variant="destructive">error</Badge>
                        ) : (
                          <Badge variant="outline">ok</Badge>
                        )}
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {formatDurationNs(trace.duration)}
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {formatTokens(trace.total_tokens)}
                      </TableCell>
                      <TableCell className="text-right font-mono text-xs">
                        {formatCostString(trace.total_cost)}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="rounded-md border bg-background p-3">
      <p className="text-xs uppercase tracking-wide text-muted-foreground">
        {label}
      </p>
      <p className="mt-1 font-mono text-base font-semibold">{value}</p>
    </div>
  )
}
