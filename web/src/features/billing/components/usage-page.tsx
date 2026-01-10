'use client'

import * as React from 'react'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/layout/page-header'
import { TimeRangePicker, type TimeRange } from '@/components/shared/time-range-picker'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { useSearchParams, useRouter, usePathname } from 'next/navigation'

import {
  useUsageOverviewQuery,
  useUsageByProjectQuery,
} from '../hooks'
import { UsageSummaryCards } from './usage-summary-cards'
import { UsageTrendChart } from './usage-trend-chart'
import { formatBytes, formatNumber } from '../types'

interface UsagePageProps {
  organizationId: string
  className?: string
}

// Default time range
const DEFAULT_TIME_RANGE: TimeRange = { relative: '30d' }

export function UsagePage({ organizationId, className }: UsagePageProps) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  // Parse time range from URL or use default
  const timeRange = React.useMemo<TimeRange>(() => {
    const relativeParam = searchParams.get('range')
    if (relativeParam) {
      return { relative: relativeParam as TimeRange['relative'] }
    }
    const from = searchParams.get('from')
    const to = searchParams.get('to')
    if (from && to) {
      return { from, to, relative: 'custom' }
    }
    return DEFAULT_TIME_RANGE
  }, [searchParams])

  // Update URL when time range changes
  const setTimeRange = React.useCallback(
    (newRange: TimeRange) => {
      const params = new URLSearchParams(searchParams.toString())

      // Clear existing time params
      params.delete('range')
      params.delete('from')
      params.delete('to')

      if (newRange.relative && newRange.relative !== 'custom') {
        params.set('range', newRange.relative)
      } else if (newRange.from && newRange.to) {
        params.set('from', newRange.from)
        params.set('to', newRange.to)
      }

      router.push(`${pathname}?${params.toString()}`)
    },
    [searchParams, pathname, router]
  )

  const {
    data: overview,
    isLoading: isOverviewLoading,
    isFetching: isOverviewFetching,
    error: overviewError,
    refetch: refetchOverview,
  } = useUsageOverviewQuery(organizationId)

  const {
    data: projectUsage,
    isLoading: isProjectUsageLoading,
    error: projectUsageError,
  } = useUsageByProjectQuery(organizationId, timeRange)

  const handleRefresh = () => {
    refetchOverview()
  }

  const isLoading = isOverviewLoading
  const isRefetching = isOverviewFetching && !isOverviewLoading

  const errorMessage = overviewError
    ? typeof overviewError === 'object' && 'message' in overviewError
      ? overviewError.message as string
      : String(overviewError)
    : null

  const projectErrorMessage = projectUsageError
    ? typeof projectUsageError === 'object' && 'message' in projectUsageError
      ? projectUsageError.message as string
      : String(projectUsageError)
    : null

  return (
    <div className={cn('space-y-6', className)}>
      <PageHeader title="Usage">
        <div className="flex items-center gap-2">
          <TimeRangePicker value={timeRange} onChange={setTimeRange} />
          <Button
            variant="outline"
            size="icon"
            onClick={handleRefresh}
            disabled={isLoading || isRefetching}
          >
            <RefreshCw
              className={cn('h-4 w-4', (isLoading || isRefetching) && 'animate-spin')}
            />
          </Button>
        </div>
      </PageHeader>

      {/* Usage Summary Cards (3 dimensions + estimated cost) */}
      <UsageSummaryCards
        usage={overview ?? null}
        isLoading={isLoading}
        error={errorMessage}
      />

      {/* Usage Trend Chart */}
      <UsageTrendChart
        organizationId={organizationId}
        timeRange={timeRange}
      />

      {/* Usage by Project Table */}
      <Card>
        <CardHeader>
          <CardTitle>Usage by Project</CardTitle>
          <CardDescription>
            Breakdown of usage across your projects for the selected time range
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isProjectUsageLoading ? (
            <div className="space-y-2">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : projectErrorMessage ? (
            <p className="text-sm text-destructive">{projectErrorMessage}</p>
          ) : !projectUsage || projectUsage.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No usage data for the selected time range
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Project</TableHead>
                  <TableHead className="text-right">Spans</TableHead>
                  <TableHead className="text-right">Data Processed</TableHead>
                  <TableHead className="text-right">Scores</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {projectUsage.map((project) => (
                  <TableRow key={project.project_id ?? 'unknown'}>
                    <TableCell className="font-medium">
                      {project.project_name ?? project.project_id ?? 'Unknown'}
                    </TableCell>
                    <TableCell className="text-right">
                      {formatNumber(project.total_spans)}
                    </TableCell>
                    <TableCell className="text-right">
                      {formatBytes(project.total_bytes)}
                    </TableCell>
                    <TableCell className="text-right">
                      {formatNumber(project.total_scores)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
