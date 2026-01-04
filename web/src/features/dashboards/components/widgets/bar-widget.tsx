'use client'

import { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  Cell,
} from 'recharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Widget, TimeRange } from '../../types'
import { buildFiltersFromDataPoint, buildDrilldownUrl } from '../../utils'

interface BarWidgetProps {
  widget: Widget
  data: BarData[] | null
  isLoading: boolean
  error?: string
  projectSlug?: string
  timeRange?: TimeRange
}

interface BarData {
  name: string
  value: number
  [key: string]: string | number
}

// Normalize backend data to expected format
// Backend returns: { dimension_key: "value", measure_key: 123 }
// Frontend expects: { name: "value", value: 123, ... }
function normalizeBarData(data: unknown[]): BarData[] {
  if (!data || data.length === 0) return []

  const firstRow = data[0] as Record<string, unknown>
  const keys = Object.keys(firstRow)

  // If already has 'name' and 'value', return as-is
  if ('name' in firstRow && 'value' in firstRow) {
    return data as BarData[]
  }

  // Find the dimension key (string value) and measure key (number value)
  const dimensionKey = keys.find((k) => typeof firstRow[k] === 'string')
  const measureKeys = keys.filter((k) => typeof firstRow[k] === 'number')

  if (!dimensionKey || measureKeys.length === 0) {
    return data as BarData[]
  }

  return data.map((row) => {
    const r = row as Record<string, unknown>
    const result: BarData = {
      name: String(r[dimensionKey] ?? ''),
      value: Number(r[measureKeys[0]] ?? 0),
    }
    // Include additional measure keys for multi-series
    measureKeys.slice(1).forEach((key) => {
      result[key] = Number(r[key] ?? 0)
    })
    return result
  })
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

function formatValue(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(2)
}

export function BarWidget({
  widget,
  data: rawData,
  isLoading,
  error,
  projectSlug,
  timeRange,
}: BarWidgetProps) {
  const router = useRouter()
  const horizontal = widget.config?.horizontal === true
  const stacked = widget.config?.stacked === true
  const showLegend = widget.config?.showLegend !== false
  const showGrid = widget.config?.showGrid !== false
  const colorByValue = widget.config?.colorByValue === true
  const enableDrilldown = projectSlug && widget.query.dimensions && widget.query.dimensions.length > 0

  // Normalize data from backend format
  const data = useMemo(() => {
    if (!rawData) return null
    return normalizeBarData(rawData as unknown[])
  }, [rawData])

  // Drilldown click handler
  const handleBarClick = useCallback(
    (barData: BarData) => {
      if (!enableDrilldown || !projectSlug) return

      const filters = buildFiltersFromDataPoint(barData as Record<string, unknown>, widget.query)
      const url = buildDrilldownUrl(projectSlug, filters, timeRange)
      router.push(url)
    },
    [enableDrilldown, projectSlug, widget.query, timeRange, router]
  )

  // Determine which keys to plot (excluding name)
  const dataKeys = useMemo(() => {
    if (!data || data.length === 0) return ['value']
    const firstRow = data[0]
    return Object.keys(firstRow).filter(
      (key) => key !== 'name' && typeof firstRow[key] === 'number'
    )
  }, [data])

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
          <p className="text-sm text-muted-foreground">No data available</p>
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
            <BarChart
              data={data}
              layout={horizontal ? 'vertical' : 'horizontal'}
              margin={{ top: 5, right: 10, left: 0, bottom: 5 }}
              onClick={(e: unknown) => {
                const event = e as { activePayload?: { payload: BarData }[] }
                if (event?.activePayload?.[0]?.payload && enableDrilldown) {
                  handleBarClick(event.activePayload[0].payload)
                }
              }}
              className={cn(enableDrilldown && 'cursor-pointer')}
            >
              {showGrid && (
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              )}
              {horizontal ? (
                <>
                  <XAxis
                    type="number"
                    tick={{ fontSize: 10 }}
                    tickLine={false}
                    axisLine={false}
                    className="fill-muted-foreground"
                    tickFormatter={formatValue}
                  />
                  <YAxis
                    type="category"
                    dataKey="name"
                    tick={{ fontSize: 10 }}
                    tickLine={false}
                    axisLine={false}
                    className="fill-muted-foreground"
                    width={80}
                  />
                </>
              ) : (
                <>
                  <XAxis
                    dataKey="name"
                    tick={{ fontSize: 10 }}
                    tickLine={false}
                    axisLine={false}
                    className="fill-muted-foreground"
                    interval={0}
                    angle={-45}
                    textAnchor="end"
                    height={60}
                  />
                  <YAxis
                    tick={{ fontSize: 10 }}
                    tickLine={false}
                    axisLine={false}
                    className="fill-muted-foreground"
                    tickFormatter={formatValue}
                    width={45}
                  />
                </>
              )}
              <Tooltip
                contentStyle={{
                  background: 'hsl(var(--popover))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  fontSize: '12px',
                }}
                formatter={(value) => [formatValue(value as number), '']}
              />
              {showLegend && dataKeys.length > 1 && (
                <Legend wrapperStyle={{ fontSize: '12px' }} />
              )}
              {dataKeys.map((key, keyIndex) => (
                <Bar
                  key={key}
                  dataKey={key}
                  name={key === 'value' ? widget.title : key}
                  stackId={stacked ? '1' : undefined}
                  fill={CHART_COLORS[keyIndex % CHART_COLORS.length]}
                  radius={[4, 4, 0, 0]}
                >
                  {colorByValue &&
                    data.map((_, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={CHART_COLORS[index % CHART_COLORS.length]}
                      />
                    ))}
                </Bar>
              ))}
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}

export type { BarData }
