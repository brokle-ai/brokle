'use client'

import { useMemo } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import type { Widget } from '../../types'

interface HistogramWidgetProps {
  widget: Widget
  data: HistogramData[] | null
  isLoading: boolean
  error?: string
}

interface HistogramData {
  bucket: string | number
  lower_bound?: number
  upper_bound?: number
  count: number
}

interface HistogramStats {
  mean: number
  median: number
  p50: number
  p95: number
  p99: number
}

function formatValue(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(1)
}

// Format bucket label from lower/upper bounds or use existing bucket string
function getBucketLabel(d: HistogramData): string {
  if (d.lower_bound !== undefined && d.upper_bound !== undefined) {
    return `${formatValue(d.lower_bound)}-${formatValue(d.upper_bound)}`
  }
  return String(d.bucket)
}

// Get bucket midpoint value for statistics calculation
function getBucketMidpoint(d: HistogramData): number {
  if (d.lower_bound !== undefined && d.upper_bound !== undefined) {
    return (d.lower_bound + d.upper_bound) / 2
  }
  // Parse from bucket string like "0-100"
  const match = String(d.bucket).match(/(\d+(?:\.\d+)?)/g)
  if (match && match.length >= 2) {
    const low = parseFloat(match[0])
    const high = parseFloat(match[1])
    return (low + high) / 2
  }
  // Try to parse as a single value
  const value = parseFloat(String(d.bucket).replace(/[^\d.-]/g, ''))
  return isNaN(value) ? 0 : value
}

export function HistogramWidget({ widget, data, isLoading, error }: HistogramWidgetProps) {
  const showGrid = widget.config?.showGrid !== false
  const showStats = widget.config?.showStats === true
  const color = (widget.config?.color as string) || '#3b82f6'

  // Transform data to include formatted bucket labels
  const chartData = useMemo(() => {
    if (!data) return []
    return data.map((d) => ({
      ...d,
      bucketLabel: getBucketLabel(d),
      midpoint: getBucketMidpoint(d),
    }))
  }, [data])

  // Calculate statistics from the histogram data
  const stats = useMemo((): HistogramStats | null => {
    if (!chartData || chartData.length === 0 || !showStats) return null

    const totalCount = chartData.reduce((sum, d) => sum + d.count, 0)
    if (totalCount === 0) return null

    // Use midpoint values calculated from getBucketMidpoint
    const valuesWithCounts = chartData.map((d) => ({
      value: d.midpoint,
      count: d.count,
    }))

    // Calculate mean
    const weightedSum = valuesWithCounts.reduce((sum, v) => sum + v.value * v.count, 0)
    const mean = weightedSum / totalCount

    // Calculate percentiles
    const sortedValues = valuesWithCounts
      .flatMap((v) => Array(v.count).fill(v.value))
      .sort((a, b) => a - b)

    const percentile = (p: number) => {
      const index = Math.ceil((p / 100) * sortedValues.length) - 1
      return sortedValues[Math.max(0, index)] || 0
    }

    return {
      mean,
      median: percentile(50),
      p50: percentile(50),
      p95: percentile(95),
      p99: percentile(99),
    }
  }, [chartData, showStats])

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">
            <Skeleton className="h-4 w-32" />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-[200px] w-full" />
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-destructive">{error}</p>
        </CardContent>
      </Card>
    )
  }

  if (!data || data.length === 0) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-muted-foreground">No histogram data available</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        {widget.description && (
          <CardDescription className="text-xs">{widget.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent>
        <div className="h-[200px]">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
              {showGrid && (
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" vertical={false} />
              )}
              <XAxis
                dataKey="bucketLabel"
                tick={{ fontSize: 9 }}
                tickLine={false}
                axisLine={false}
                className="fill-muted-foreground"
                interval={0}
                angle={-45}
                textAnchor="end"
                height={50}
              />
              <YAxis
                tick={{ fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                className="fill-muted-foreground"
                tickFormatter={formatValue}
                width={40}
                label={{
                  value: 'Count',
                  angle: -90,
                  position: 'insideLeft',
                  style: { fontSize: 10, fill: 'hsl(var(--muted-foreground))' },
                }}
              />
              <Tooltip
                contentStyle={{
                  background: 'hsl(var(--popover))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  fontSize: '12px',
                }}
                formatter={(value) => [formatValue(value as number), 'Count']}
                labelFormatter={(label) => `Range: ${label}`}
              />
              {stats && (
                <>
                  <ReferenceLine
                    x={chartData.find((d) => {
                      // Find the bucket that contains the p95 value
                      if (d.lower_bound !== undefined && d.upper_bound !== undefined) {
                        return stats.p95 >= d.lower_bound && stats.p95 < d.upper_bound
                      }
                      // Fallback to midpoint comparison for legacy bucket format
                      const nextBucket = chartData[chartData.indexOf(d) + 1]
                      if (nextBucket) {
                        return stats.p95 >= d.midpoint && stats.p95 < nextBucket.midpoint
                      }
                      return false
                    })?.bucketLabel}
                    stroke="#ef4444"
                    strokeDasharray="3 3"
                    label={{ value: 'p95', fontSize: 9, fill: '#ef4444' }}
                  />
                </>
              )}
              <Bar dataKey="count" fill={color} radius={[2, 2, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {stats && (
          <div className="flex gap-4 mt-2 text-xs text-muted-foreground justify-center">
            <span>
              Mean: <span className="font-medium text-foreground">{formatValue(stats.mean)}</span>
            </span>
            <span>
              p50: <span className="font-medium text-foreground">{formatValue(stats.p50)}</span>
            </span>
            <span>
              p95: <span className="font-medium text-foreground">{formatValue(stats.p95)}</span>
            </span>
            <span>
              p99: <span className="font-medium text-foreground">{formatValue(stats.p99)}</span>
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export type { HistogramData, HistogramStats }
