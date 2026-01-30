'use client'

import { useState, useMemo } from 'react'
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
  ChevronLeft,
  ChevronRight,
  AlertCircle,
  Zap,
  Clock,
  Filter,
  Eye,
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { ExecutionStatusBadge, isTerminalStatus } from './execution-status-badge'
import { ExecutionDetailDialog } from './execution-detail-dialog'
import {
  useEvaluatorExecutionsQuery,
  getRefetchInterval,
} from '../hooks/use-evaluator-executions'
import type { EvaluatorExecution, TriggerType } from '../types'

interface EvaluatorExecutionsTableProps {
  projectId: string
  projectSlug: string
  evaluatorId: string
  limit?: number
}

function formatDuration(ms: number | undefined): string {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${(ms / 60000).toFixed(1)}m`
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
  onViewDetail: (executionId: string) => void
}

function ExecutionRow({ execution, onViewDetail }: ExecutionRowProps) {
  const hasError = execution.status === 'failed' && execution.error_message

  return (
    <TableRow
      className="cursor-pointer hover:bg-muted/50"
      onClick={() => onViewDetail(execution.id)}
    >
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
      <TableCell className="text-right text-muted-foreground text-sm">
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
        <Button
          variant="ghost"
          size="sm"
          onClick={(e) => {
            e.stopPropagation()
            onViewDetail(execution.id)
          }}
        >
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
      {[...Array(3)].map((_, i) => (
        <TableRow key={i}>
          <TableCell>
            <Skeleton className="h-6 w-24" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-6 w-16" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-12 ml-auto" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-12 ml-auto" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-8 ml-auto" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-12 ml-auto" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-4 w-24 ml-auto" />
          </TableCell>
          <TableCell>
            <Skeleton className="h-8 w-8 ml-auto" />
          </TableCell>
        </TableRow>
      ))}
    </>
  )
}

type TriggerTypeFilter = TriggerType | 'all'

export function EvaluatorExecutionsTable({
  projectId,
  projectSlug,
  evaluatorId,
  limit = 10,
}: EvaluatorExecutionsTableProps) {
  const [page, setPage] = useState(1)
  const [triggerFilter, setTriggerFilter] = useState<TriggerTypeFilter>('all')
  const [selectedExecutionId, setSelectedExecutionId] = useState<string | null>(null)
  const [isDetailDialogOpen, setIsDetailDialogOpen] = useState(false)

  const handleViewDetail = (executionId: string) => {
    setSelectedExecutionId(executionId)
    setIsDetailDialogOpen(true)
  }

  // Reset page when filter changes
  const handleFilterChange = (value: TriggerTypeFilter) => {
    setTriggerFilter(value)
    setPage(1)
  }

  // Check if any execution is running to enable auto-refresh
  const { data, isLoading, isError, error } = useEvaluatorExecutionsQuery(
    projectId,
    evaluatorId,
    {
      page,
      limit,
      triggerType: triggerFilter === 'all' ? undefined : triggerFilter,
    }
  )

  const hasRunningExecutions = useMemo(() => {
    return data?.executions.some((e) => !isTerminalStatus(e.status)) ?? false
  }, [data?.executions])

  // Re-query with refetch interval if there are running executions
  const { data: refreshedData } = useEvaluatorExecutionsQuery(projectId, evaluatorId, {
    page,
    limit,
    triggerType: triggerFilter === 'all' ? undefined : triggerFilter,
    refetchInterval: getRefetchInterval(hasRunningExecutions),
    enabled: hasRunningExecutions,
  })

  const executions = refreshedData?.executions ?? data?.executions ?? []
  const total = refreshedData?.total ?? data?.total ?? 0
  const totalPages = Math.ceil(total / limit)

  if (isError) {
    return (
      <div className="flex items-center justify-center py-8 text-destructive">
        <AlertCircle className="h-4 w-4 mr-2" />
        <span>Failed to load executions: {error?.message}</span>
      </div>
    )
  }

  if (!isLoading && executions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Clock className="h-8 w-8 mb-2 opacity-50" />
        <p className="text-sm">No executions yet</p>
        <p className="text-xs mt-1">
          Trigger a run or wait for automatic evaluation.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Filter Bar */}
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-muted-foreground" />
        <Select
          value={triggerFilter}
          onValueChange={(value) => handleFilterChange(value as TriggerTypeFilter)}
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
              <TableHead className="text-right w-[60px]">
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
                  onViewDetail={handleViewDetail}
                />
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
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
      )}

      {/* Execution Detail Dialog */}
      <ExecutionDetailDialog
        projectId={projectId}
        projectSlug={projectSlug}
        evaluatorId={evaluatorId}
        executionId={selectedExecutionId}
        open={isDetailDialogOpen}
        onOpenChange={setIsDetailDialogOpen}
      />
    </div>
  )
}
