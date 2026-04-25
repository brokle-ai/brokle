import { useMemo, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import {
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  Clock,
  Eye,
  Filter,
  Zap,
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Skeleton } from '@/components/ui/skeleton'
import {
  ExecutionStatusBadge,
  isTerminalStatus,
} from './execution-status-badge'
import { executionListQueryOptions } from '../api/queries'
import type {
  EvaluatorExecution,
  ExecutionListParams,
  TriggerType,
} from '../api/types'

interface EvaluatorExecutionsTableProps {
  orgId: string
  projectId: string
  evaluatorId: string
  limit?: number
}

function formatDuration(ms: number | undefined): string {
  if (ms === undefined || ms === null) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  return `${(ms / 60_000).toFixed(1)}m`
}

function TriggerTypeBadge({ type }: { type: TriggerType }) {
  return (
    <Badge variant={type === 'manual' ? 'secondary' : 'outline'} className="gap-1">
      {type === 'manual' ? (
        <>
          <Zap className="h-3 w-3" />
          Manual
        </>
      ) : (
        <>
          <Clock className="h-3 w-3" />
          Auto
        </>
      )}
    </Badge>
  )
}

interface ExecutionRowProps {
  execution: EvaluatorExecution
  onView: () => void
}

function ExecutionRow({ execution, onView }: ExecutionRowProps) {
  const hasError =
    execution.status === 'failed' && Boolean(execution.error_message)

  return (
    <TableRow>
      <TableCell>
        <ExecutionStatusBadge
          status={execution.status}
          timestamp={execution.created_at}
        />
      </TableCell>
      <TableCell>
        <TriggerTypeBadge type={execution.trigger_type} />
      </TableCell>
      <TableCell className="text-right font-mono text-sm">
        {execution.spans_matched.toLocaleString()}
      </TableCell>
      <TableCell className="text-right font-mono text-sm">
        {execution.spans_scored.toLocaleString()}
      </TableCell>
      <TableCell className="text-right">
        {execution.errors_count > 0 ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <Badge variant="destructive" className="gap-1">
                <AlertCircle className="h-3 w-3" />
                {execution.errors_count}
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              {hasError ? (
                <p className="max-w-xs">{execution.error_message}</p>
              ) : (
                <p>{execution.errors_count} error(s) during execution</p>
              )}
            </TooltipContent>
          </Tooltip>
        ) : (
          <span className="text-muted-foreground">-</span>
        )}
      </TableCell>
      <TableCell className="text-right font-mono text-sm">
        {formatDuration(execution.duration_ms)}
      </TableCell>
      <TableCell className="text-right text-sm text-muted-foreground">
        <Tooltip>
          <TooltipTrigger>
            {formatDistanceToNow(new Date(execution.created_at), {
              addSuffix: true,
            })}
          </TooltipTrigger>
          <TooltipContent>
            {new Date(execution.created_at).toLocaleString()}
          </TooltipContent>
        </Tooltip>
      </TableCell>
      <TableCell className="text-right">
        <Button variant="ghost" size="sm" onClick={onView}>
          <Eye className="h-4 w-4" />
          <span className="sr-only">View details</span>
        </Button>
      </TableCell>
    </TableRow>
  )
}

function TableSkeleton() {
  return (
    <>
      {Array.from({ length: 3 }).map((_, i) => (
        <TableRow key={i}>
          <TableCell>
            <Skeleton className="h-6 w-24" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-6 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-4 w-12" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-4 w-12" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-4 w-8" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-4 w-12" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-4 w-24" />
          </TableCell>
          <TableCell>
            <Skeleton className="ml-auto h-8 w-8" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}

type TriggerTypeFilter = TriggerType | 'all'

export function EvaluatorExecutionsTable({
  orgId,
  projectId,
  evaluatorId,
  limit = 10,
}: EvaluatorExecutionsTableProps) {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [triggerFilter, setTriggerFilter] = useState<TriggerTypeFilter>('all')

  // We use imperative useNavigate rather than `<Link to=…>` here because
  // the parent `evaluators.$evaluatorId` route declares a non-trivial
  // search schema (`page/limit/q?`) that the child execution-detail
  // route opts out of via `z.object({}).catch({})`. The two schemas
  // intersect at the type level into `{[x]: never; page; limit; q?}`,
  // which is impossible to satisfy from a literal Link `search` prop.
  // useNavigate sidesteps that intersection.
  const goToExecution = (executionId: string) => {
    navigate({
      to: '/o/$orgId/p/$projectId/evaluators/$evaluatorId/executions/$executionId',
      params: { orgId, projectId, evaluatorId, executionId },
      search: { page: 1, limit: 20 },
    })
  }

  const params: ExecutionListParams = useMemo(
    () => ({
      page,
      limit,
      trigger_type: triggerFilter === 'all' ? undefined : triggerFilter,
    }),
    [page, limit, triggerFilter],
  )

  // First pass: vanilla query, no polling.
  const initialQuery = useQuery(
    executionListQueryOptions(projectId, evaluatorId, params),
  )

  const hasRunningExecutions = useMemo(
    () =>
      initialQuery.data?.executions?.some(
        (e) => !isTerminalStatus(e.status),
      ) ?? false,
    [initialQuery.data?.executions],
  )

  // Second pass — re-runs only when at least one execution is still
  // making progress. Splitting these means terminal-only pages stay on
  // the cheap, never-refetching path.
  const pollingQuery = useQuery({
    ...executionListQueryOptions(
      projectId,
      evaluatorId,
      params,
      hasRunningExecutions ? 5_000 : false,
    ),
    enabled: hasRunningExecutions,
  })

  const data = pollingQuery.data ?? initialQuery.data
  const isLoading = initialQuery.isLoading
  const isError = initialQuery.isError
  const error = initialQuery.error

  const executions = data?.executions ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / limit))

  const handleFilterChange = (value: TriggerTypeFilter) => {
    setTriggerFilter(value)
    setPage(1)
  }

  if (isError) {
    return (
      <div className="flex items-center justify-center py-8 text-destructive">
        <AlertCircle className="mr-2 h-4 w-4" />
        <span>
          Failed to load executions:{' '}
          {error instanceof Error ? error.message : 'unknown error'}
        </span>
      </div>
    )
  }

  if (!isLoading && executions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Clock className="mb-2 h-8 w-8 opacity-50" />
        <p className="text-sm">No executions yet</p>
        <p className="mt-1 text-xs">
          Trigger a run or wait for automatic evaluation.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <Select
          value={triggerFilter}
          onValueChange={(v) => handleFilterChange(v as TriggerTypeFilter)}
        >
          <SelectTrigger className="w-36">
            <SelectValue placeholder="Filter by trigger" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All triggers</SelectItem>
            <SelectItem value="automatic">Automatic</SelectItem>
            <SelectItem value="manual">Manual</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Trigger</TableHead>
              <TableHead className="text-right">Matched</TableHead>
              <TableHead className="text-right">Scored</TableHead>
              <TableHead className="text-right">Errors</TableHead>
              <TableHead className="text-right">Duration</TableHead>
              <TableHead className="text-right">Time</TableHead>
              <TableHead className="w-[60px] text-right">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableSkeleton />
            ) : (
              executions.map((execution) => (
                <ExecutionRow
                  key={execution.id}
                  execution={execution}
                  onView={() => goToExecution(execution.id)}
                />
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 ? (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {executions.length} of {total} executions
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(Math.max(1, page - 1))}
              disabled={page === 1 || isLoading}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(Math.min(totalPages, page + 1))}
              disabled={page === totalPages || isLoading}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      ) : null}
    </div>
  )
}
