'use client'

import { Activity, DollarSign, Timer, AlertTriangle, BarChart3 } from 'lucide-react'
import { useTraceMetrics, TIME_RANGES, type TimeRange } from '../hooks/use-trace-metrics'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { AreaChart } from '@/components/shared/charts/area-chart'
import { BarChart } from '@/components/shared/charts/bar-chart'
import { DataTableEmptyState } from '@/components/data-table'
import { formatCost } from '../utils/format-helpers'

/**
 * Format large numbers with K/M suffixes
 */
function formatNumber(num: number): string {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`
  return num.toLocaleString()
}

/**
 * Format latency in ms
 */
function formatLatency(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${ms.toFixed(0)}ms`
}

/**
 * Time range selector component
 */
function TimeRangeSelector({
  value,
  onChange,
}: {
  value: TimeRange
  onChange: (range: TimeRange) => void
}) {
  const labels: Record<TimeRange, string> = {
    '24h': '24h',
    '7d': '7d',
    '30d': '30d',
    all: 'All',
  }

  return (
    <div className='flex gap-1 rounded-lg bg-muted p-1'>
      {TIME_RANGES.map((range) => (
        <Button
          key={range}
          variant={value === range ? 'default' : 'ghost'}
          size='sm'
          className='h-7 px-3'
          onClick={() => onChange(range)}
        >
          {labels[range]}
        </Button>
      ))}
    </div>
  )
}

/**
 * Stat card component
 */
function StatCard({
  title,
  value,
  description,
  icon: Icon,
  loading = false,
  trend,
}: {
  title: string
  value: string
  description?: string
  icon: React.ComponentType<{ className?: string }>
  loading?: boolean
  trend?: { value: number; label: string }
}) {
  if (loading) {
    return (
      <Card>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
          <Skeleton className='h-4 w-24' />
          <Skeleton className='h-4 w-4' />
        </CardHeader>
        <CardContent>
          <Skeleton className='h-8 w-20 mb-1' />
          <Skeleton className='h-3 w-32' />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
        <CardTitle className='text-sm font-medium'>{title}</CardTitle>
        <Icon className='h-4 w-4 text-muted-foreground' />
      </CardHeader>
      <CardContent>
        <div className='text-2xl font-bold'>{value}</div>
        {description && <p className='text-xs text-muted-foreground'>{description}</p>}
        {trend && (
          <p
            className={`text-xs ${trend.value >= 0 ? 'text-green-600' : 'text-red-600'}`}
          >
            {trend.value >= 0 ? '+' : ''}
            {trend.value.toFixed(1)}% {trend.label}
          </p>
        )}
      </CardContent>
    </Card>
  )
}

/**
 * Model/Provider breakdown card
 */
function BreakdownCard({
  title,
  data,
  loading = false,
  valueKey,
  labelKey,
}: {
  title: string
  data: Array<{ [key: string]: string | number }>
  loading?: boolean
  valueKey: string
  labelKey: string
}) {
  if (loading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className='h-5 w-32' />
        </CardHeader>
        <CardContent className='space-y-3'>
          {[1, 2, 3].map((i) => (
            <div key={i} className='flex items-center justify-between'>
              <Skeleton className='h-4 w-24' />
              <Skeleton className='h-4 w-12' />
            </div>
          ))}
        </CardContent>
      </Card>
    )
  }

  const total = data.reduce((sum, item) => sum + (item[valueKey] as number), 0)

  return (
    <Card>
      <CardHeader className='pb-3'>
        <CardTitle className='text-base'>{title}</CardTitle>
      </CardHeader>
      <CardContent className='space-y-3'>
        {data.length === 0 ? (
          <p className='text-sm text-muted-foreground'>No data</p>
        ) : (
          data.slice(0, 5).map((item, index) => {
            const value = item[valueKey] as number
            const label = item[labelKey] as string
            const percentage = total > 0 ? (value / total) * 100 : 0

            return (
              <div key={index} className='space-y-1'>
                <div className='flex items-center justify-between text-sm'>
                  <span className='truncate max-w-[150px] font-medium'>{label}</span>
                  <span className='text-muted-foreground'>
                    {formatNumber(value)} ({percentage.toFixed(1)}%)
                  </span>
                </div>
                <div className='h-1.5 rounded-full bg-muted overflow-hidden'>
                  <div
                    className='h-full rounded-full bg-primary transition-all'
                    style={{ width: `${percentage}%` }}
                  />
                </div>
              </div>
            )
          })
        )}
      </CardContent>
    </Card>
  )
}

export function MetricsView() {
  const { metrics, timeRange, setTimeRange, isLoading, isFetching, error, hasProject, refetch } =
    useTraceMetrics()

  // Error state
  if (error) {
    return (
      <div className='flex flex-col items-center justify-center py-12 space-y-4'>
        <div className='rounded-lg bg-destructive/10 p-6 text-center max-w-md'>
          <h3 className='font-semibold text-destructive mb-2'>Failed to load metrics</h3>
          <p className='text-sm text-muted-foreground mb-4'>{error}</p>
          <button
            onClick={() => refetch()}
            className='inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors'
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  // Empty state
  if (!isLoading && metrics && metrics.totalTraces === 0) {
    return (
      <DataTableEmptyState
        icon={<BarChart3 className='h-full w-full' />}
        title='No metrics data'
        description='Start sending traces to see usage analytics and cost metrics.'
      />
    )
  }

  return (
    <div className='space-y-6'>
      {/* Header with time range selector */}
      <div className='flex items-center justify-between'>
        <div>
          <h2 className='text-lg font-semibold'>Usage Metrics</h2>
          <p className='text-sm text-muted-foreground'>
            Track your AI usage, costs, and performance
          </p>
        </div>
        <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
      </div>

      {/* Summary stats cards */}
      <div className='grid gap-4 md:grid-cols-2 lg:grid-cols-4'>
        <StatCard
          title='Total Traces'
          value={isLoading ? '...' : formatNumber(metrics?.totalTraces || 0)}
          description='Requests processed'
          icon={Activity}
          loading={isLoading}
        />
        <StatCard
          title='Total Tokens'
          value={isLoading ? '...' : formatNumber(metrics?.totalTokens || 0)}
          description='Input + output tokens'
          icon={Activity}
          loading={isLoading}
        />
        <StatCard
          title='Total Cost'
          value={isLoading ? '...' : formatCost(metrics?.totalCost || 0)}
          description='Estimated cost'
          icon={DollarSign}
          loading={isLoading}
        />
        <StatCard
          title='Avg Latency'
          value={isLoading ? '...' : formatLatency(metrics?.averageLatency || 0)}
          description='Response time'
          icon={Timer}
          loading={isLoading}
        />
      </div>

      {/* Error rate card */}
      {!isLoading && metrics && metrics.errorRate > 0 && (
        <Card className='border-orange-200 bg-orange-50/50 dark:border-orange-900 dark:bg-orange-950/20'>
          <CardHeader className='flex flex-row items-center gap-2 pb-2'>
            <AlertTriangle className='h-4 w-4 text-orange-600' />
            <CardTitle className='text-sm font-medium text-orange-600'>
              Error Rate: {metrics.errorRate.toFixed(1)}%
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className='text-sm text-muted-foreground'>
              {Math.round((metrics.totalTraces * metrics.errorRate) / 100)} of{' '}
              {metrics.totalTraces} traces encountered errors
            </p>
          </CardContent>
        </Card>
      )}

      {/* Charts row */}
      <div className='grid gap-4 md:grid-cols-2'>
        {/* Traces over time */}
        <AreaChart
          title='Traces Over Time'
          description='Daily trace volume'
          data={metrics?.timeSeries || []}
          xKey='date'
          yKey='traces'
          loading={isLoading}
          height={250}
          colors={['#3b82f6']}
        />

        {/* Cost over time */}
        <AreaChart
          title='Cost Over Time'
          description='Daily cost breakdown'
          data={metrics?.timeSeries || []}
          xKey='date'
          yKey='cost'
          loading={isLoading}
          height={250}
          colors={['#10b981']}
          formatYAxis={(value) => `$${value.toFixed(2)}`}
        />
      </div>

      {/* Breakdowns row */}
      <div className='grid gap-4 md:grid-cols-2'>
        <BreakdownCard
          title='Usage by Model'
          data={metrics?.byModel || []}
          loading={isLoading}
          valueKey='tokens'
          labelKey='model'
        />
        <BreakdownCard
          title='Cost by Provider'
          data={metrics?.byProvider || []}
          loading={isLoading}
          valueKey='cost'
          labelKey='provider'
        />
      </div>

      {/* Token usage chart */}
      <AreaChart
        title='Token Usage Over Time'
        description='Daily token consumption'
        data={metrics?.timeSeries || []}
        xKey='date'
        yKey='tokens'
        loading={isLoading}
        height={250}
        colors={['#8b5cf6']}
        formatYAxis={(value) => formatNumber(value)}
      />
    </div>
  )
}
