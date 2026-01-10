'use client'

import * as React from 'react'
import {
  Area,
  AreaChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  CartesianGrid,
} from 'recharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { useUsageTimeSeriesQuery } from '../hooks'
import type { BillableUsage } from '../types'
import type { TimeRange } from '@/components/shared/time-range-picker'
import { formatBytes, formatNumber } from '../types'

interface UsageTrendChartProps {
  organizationId: string
  timeRange: TimeRange
  className?: string
}

type MetricType = 'spans' | 'bytes' | 'scores' | 'all'

interface ChartDataPoint {
  date: string
  displayDate: string
  spans: number
  bytes: number
  scores: number
}

const COLORS = {
  spans: 'hsl(var(--chart-1))',
  bytes: 'hsl(var(--chart-2))',
  scores: 'hsl(var(--chart-3))',
}

function detectHourlyGranularity(usage: BillableUsage[]): boolean {
  if (usage.length === 0) return false
  // Check if any bucket_time has a non-midnight time component
  // Daily data: 2024-01-15T00:00:00Z, Hourly data: 2024-01-15T14:00:00Z
  for (const item of usage) {
    const timePart = item.bucket_time.split('T')[1]
    if (timePart && !timePart.startsWith('00:00:00')) {
      return true
    }
  }
  return false
}

function formatDisplayDate(bucketTime: string, isHourly: boolean): string {
  if (isHourly) {
    // Full ISO timestamp - format with time in local timezone
    const date = new Date(bucketTime)
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    })
  } else {
    // Date-only: parse as local date to avoid timezone shift
    const [year, month, day] = bucketTime.split('T')[0].split('-').map(Number)
    const localDate = new Date(year, month - 1, day)
    return localDate.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    })
  }
}

function transformData(usage: BillableUsage[] | undefined): ChartDataPoint[] {
  if (!usage || usage.length === 0) return []

  const isHourly = detectHourlyGranularity(usage)

  const grouped = usage.reduce<Record<string, ChartDataPoint>>((acc, item) => {
    const key = isHourly ? item.bucket_time : item.bucket_time.split('T')[0]

    if (!acc[key]) {
      acc[key] = {
        date: key,
        displayDate: formatDisplayDate(item.bucket_time, isHourly),
        spans: 0,
        bytes: 0,
        scores: 0,
      }
    }
    acc[key].spans += item.span_count
    acc[key].bytes += item.bytes_processed
    acc[key].scores += item.score_count
    return acc
  }, {})

  return Object.values(grouped).sort(
    (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
  )
}

function CustomTooltip({
  active,
  payload,
  label,
}: {
  active?: boolean
  payload?: Array<{ dataKey: string; value: number; color: string }>
  label?: string
}) {
  if (!active || !payload || payload.length === 0) return null

  return (
    <div className="rounded-lg border bg-background p-3 shadow-sm">
      <p className="text-sm font-medium mb-2">{label}</p>
      <div className="space-y-1">
        {payload.map((entry) => (
          <div key={entry.dataKey} className="flex items-center gap-2 text-sm">
            <div
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-muted-foreground capitalize">{entry.dataKey}:</span>
            <span className="font-medium">
              {entry.dataKey === 'bytes'
                ? formatBytes(entry.value)
                : formatNumber(entry.value)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

export function UsageTrendChart({
  organizationId,
  timeRange,
  className,
}: UsageTrendChartProps) {
  const [metric, setMetric] = React.useState<MetricType>('all')

  const {
    data: usage,
    isLoading,
    error,
  } = useUsageTimeSeriesQuery(organizationId, timeRange)

  const chartData = React.useMemo(() => transformData(usage), [usage])

  const errorMessage = error
    ? typeof error === 'object' && 'message' in error
      ? (error.message as string)
      : String(error)
    : null

  const showSpans = metric === 'all' || metric === 'spans'
  const showBytes = metric === 'all' || metric === 'bytes'
  const showScores = metric === 'all' || metric === 'scores'

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle className="text-base">Usage Trend</CardTitle>
          <CardDescription>
            Usage across all dimensions for the selected time range
          </CardDescription>
        </div>
        <Select value={metric} onValueChange={(v) => setMetric(v as MetricType)}>
          <SelectTrigger className="w-[130px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Metrics</SelectItem>
            <SelectItem value="spans">Spans</SelectItem>
            <SelectItem value="bytes">Data</SelectItem>
            <SelectItem value="scores">Scores</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-[300px] w-full" />
        ) : errorMessage ? (
          <div className="flex items-center justify-center h-[300px]">
            <p className="text-sm text-destructive">{errorMessage}</p>
          </div>
        ) : chartData.length === 0 ? (
          <div className="flex items-center justify-center h-[300px]">
            <p className="text-sm text-muted-foreground">
              No usage data for the selected time range
            </p>
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart
              data={chartData}
              margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
            >
              <defs>
                <linearGradient id="colorSpans" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={COLORS.spans} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={COLORS.spans} stopOpacity={0} />
                </linearGradient>
                <linearGradient id="colorBytes" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={COLORS.bytes} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={COLORS.bytes} stopOpacity={0} />
                </linearGradient>
                <linearGradient id="colorScores" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor={COLORS.scores} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={COLORS.scores} stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid
                strokeDasharray="3 3"
                className="stroke-muted"
                vertical={false}
              />
              <XAxis
                dataKey="displayDate"
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                className="text-muted-foreground"
              />
              <YAxis
                tick={{ fontSize: 12 }}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => formatNumber(value)}
                className="text-muted-foreground"
                width={60}
              />
              <Tooltip content={<CustomTooltip />} />
              <Legend
                wrapperStyle={{ fontSize: 12 }}
                formatter={(value) => (
                  <span className="text-muted-foreground capitalize">{value}</span>
                )}
              />
              {showSpans && (
                <Area
                  type="monotone"
                  dataKey="spans"
                  stroke={COLORS.spans}
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorSpans)"
                />
              )}
              {showBytes && (
                <Area
                  type="monotone"
                  dataKey="bytes"
                  stroke={COLORS.bytes}
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorBytes)"
                />
              )}
              {showScores && (
                <Area
                  type="monotone"
                  dataKey="scores"
                  stroke={COLORS.scores}
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorScores)"
                />
              )}
            </AreaChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
