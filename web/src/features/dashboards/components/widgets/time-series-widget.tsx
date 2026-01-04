'use client'

import { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import { format, parseISO } from 'date-fns'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Widget, TimeRange } from '../../types'
import { buildFiltersFromDataPoint, buildDrilldownUrl } from '../../utils'

interface TimeSeriesWidgetProps {
  widget: Widget
  data: TimeSeriesData[] | null
  isLoading: boolean
  error?: string
  projectSlug?: string
  timeRange?: TimeRange
}

interface TimeSeriesData {
  timestamp?: string
  time?: string
  [key: string]: string | number | undefined
}

const CHART_COLORS = [
  '#3b82f6', // blue
  '#10b981', // green
  '#f59e0b', // amber
  '#ef4444', // red
  '#8b5cf6', // violet
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#84cc16', // lime
]

type ChartVariant = 'line' | 'area'

function formatTimestamp(timestamp: string): string {
  try {
    const date = parseISO(timestamp)
    return format(date, 'MMM d, HH:mm')
  } catch {
    return timestamp
  }
}

function formatYAxisValue(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(2)
}

export function TimeSeriesWidget({
  widget,
  data,
  isLoading,
  error,
  projectSlug,
  timeRange,
}: TimeSeriesWidgetProps) {
  const router = useRouter()
  const variant: ChartVariant = (widget.config?.variant as ChartVariant) || 'line'
  const stacked = widget.config?.stacked === true
  const showLegend = widget.config?.showLegend !== false
  const showGrid = widget.config?.showGrid !== false
  const enableDrilldown = projectSlug && widget.query.dimensions && widget.query.dimensions.length > 0

  // Get the time key (backend may return 'time' or 'timestamp')
  const getTimeValue = (point: TimeSeriesData): string => {
    return point.timestamp || point.time || ''
  }

  // Determine which keys to plot (excluding timestamp/time)
  const dataKeys = useMemo(() => {
    if (!data || data.length === 0) return []
    const firstRow = data[0]
    return Object.keys(firstRow).filter(
      key => key !== 'timestamp' && key !== 'time' && typeof firstRow[key] === 'number'
    )
  }, [data])

  // Format data for chart
  const chartData = useMemo(() => {
    if (!data) return []
    return data.map(point => ({
      ...point,
      formattedTime: formatTimestamp(getTimeValue(point))
    }))
  }, [data])

  // Drilldown click handler
  const handleDataPointClick = useCallback(
    (dataPoint: TimeSeriesData) => {
      if (!enableDrilldown || !projectSlug) return

      const filters = buildFiltersFromDataPoint(dataPoint as Record<string, unknown>, widget.query)
      const url = buildDrilldownUrl(projectSlug, filters, timeRange)
      router.push(url)
    },
    [enableDrilldown, projectSlug, widget.query, timeRange, router]
  )

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

  if (!data || data.length === 0 || dataKeys.length === 0) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-muted-foreground">No time series data available</p>
        </CardContent>
      </Card>
    )
  }

  const ChartComponent = variant === 'area' ? AreaChart : LineChart
  const DataComponent = variant === 'area' ? Area : Line

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
            <ChartComponent
              data={chartData}
              margin={{ top: 5, right: 10, left: 0, bottom: 5 }}
              onClick={(e: unknown) => {
                const event = e as { activePayload?: { payload: TimeSeriesData }[] }
                if (event?.activePayload?.[0]?.payload && enableDrilldown) {
                  handleDataPointClick(event.activePayload[0].payload)
                }
              }}
              className={cn(enableDrilldown && 'cursor-pointer')}
            >
              {showGrid && (
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              )}
              <XAxis
                dataKey="formattedTime"
                tick={{ fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                className="fill-muted-foreground"
                interval="preserveStartEnd"
              />
              <YAxis
                tick={{ fontSize: 10 }}
                tickLine={false}
                axisLine={false}
                className="fill-muted-foreground"
                tickFormatter={formatYAxisValue}
                width={45}
              />
              <Tooltip
                contentStyle={{
                  background: 'hsl(var(--popover))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  fontSize: '12px',
                }}
                labelStyle={{ fontWeight: 600, marginBottom: '4px' }}
                formatter={(value) => [formatYAxisValue(value as number), '']}
              />
              {showLegend && dataKeys.length > 1 && (
                <Legend wrapperStyle={{ fontSize: '12px' }} />
              )}
              {dataKeys.map((key, index) => (
                variant === 'area' ? (
                  <Area
                    key={key}
                    type="monotone"
                    dataKey={key}
                    name={key}
                    stackId={stacked ? '1' : undefined}
                    stroke={CHART_COLORS[index % CHART_COLORS.length]}
                    fill={CHART_COLORS[index % CHART_COLORS.length]}
                    fillOpacity={0.4}
                    strokeWidth={2}
                  />
                ) : (
                  <Line
                    key={key}
                    type="monotone"
                    dataKey={key}
                    name={key}
                    stroke={CHART_COLORS[index % CHART_COLORS.length]}
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4 }}
                  />
                )
              ))}
            </ChartComponent>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}

export type { TimeSeriesData }
