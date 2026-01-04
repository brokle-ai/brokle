'use client'

import { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Legend,
  Tooltip,
  type PieLabelRenderProps,
} from 'recharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Widget, TimeRange } from '../../types'
import { buildFiltersFromDataPoint, buildDrilldownUrl } from '../../utils'

interface PieWidgetProps {
  widget: Widget
  data: PieData[] | null
  isLoading: boolean
  error?: string
  projectSlug?: string
  timeRange?: TimeRange
}

interface PieData {
  name: string
  value: number
  [key: string]: string | number
}

// Normalize backend data to expected format
// Backend returns: { dimension_key: "value", measure_key: 123 }
// Frontend expects: { name: "value", value: 123 }
function normalizePieData(data: unknown[]): PieData[] {
  if (!data || data.length === 0) return []

  const firstRow = data[0] as Record<string, unknown>
  const keys = Object.keys(firstRow)

  // If already has 'name' and 'value', return as-is
  if ('name' in firstRow && 'value' in firstRow) {
    return data as PieData[]
  }

  // Find the dimension key (string value) and measure key (number value)
  const dimensionKey = keys.find((k) => typeof firstRow[k] === 'string')
  const measureKey = keys.find((k) => typeof firstRow[k] === 'number')

  if (!dimensionKey || !measureKey) {
    return data as PieData[]
  }

  return data.map((row) => {
    const r = row as Record<string, unknown>
    return {
      name: String(r[dimensionKey] ?? ''),
      value: Number(r[measureKey] ?? 0),
    }
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

export function PieWidget({
  widget,
  data: rawData,
  isLoading,
  error,
  projectSlug,
  timeRange,
}: PieWidgetProps) {
  const router = useRouter()
  const donut = widget.config?.donut === true
  const showLegend = widget.config?.showLegend !== false
  const showLabels = widget.config?.showLabels !== false
  const enableDrilldown = projectSlug && widget.query.dimensions && widget.query.dimensions.length > 0

  // Normalize data from backend format
  const data = useMemo(() => {
    if (!rawData) return null
    return normalizePieData(rawData as unknown[])
  }, [rawData])

  const total = useMemo(() => {
    if (!data) return 0
    return data.reduce((sum, item) => sum + item.value, 0)
  }, [data])

  // Drilldown click handler
  const handlePieClick = useCallback(
    (pieData: PieData) => {
      if (!enableDrilldown || !projectSlug) return

      const filters = buildFiltersFromDataPoint(pieData as Record<string, unknown>, widget.query)
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
          <Skeleton className="h-[200px] w-full rounded-full mx-auto max-w-[200px]" />
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

  const renderLabel = (props: PieLabelRenderProps) => {
    const { cx, cy, midAngle, innerRadius, outerRadius, percent } = props as {
      cx: number
      cy: number
      midAngle: number
      innerRadius: number
      outerRadius: number
      percent: number
    }
    if (!percent || percent < 0.05) return null // Don't show labels for very small slices
    const RADIAN = Math.PI / 180
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5
    const x = cx + radius * Math.cos(-midAngle * RADIAN)
    const y = cy + radius * Math.sin(-midAngle * RADIAN)

    return (
      <text
        x={x}
        y={y}
        fill="white"
        textAnchor="middle"
        dominantBaseline="central"
        fontSize={11}
        fontWeight={500}
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
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
        <div className="h-[200px] relative">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart className={cn(enableDrilldown && 'cursor-pointer')}>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={showLabels ? renderLabel : false}
                innerRadius={donut ? 50 : 0}
                outerRadius={80}
                paddingAngle={2}
                dataKey="value"
                onClick={(e) => {
                  if (e && enableDrilldown) {
                    handlePieClick(e.payload)
                  }
                }}
              >
                {data.map((_, index) => (
                  <Cell
                    key={`cell-${index}`}
                    fill={CHART_COLORS[index % CHART_COLORS.length]}
                    stroke="hsl(var(--background))"
                    strokeWidth={2}
                    className={cn(enableDrilldown && 'cursor-pointer hover:opacity-80 transition-opacity')}
                  />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  background: 'hsl(var(--popover))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                  fontSize: '12px',
                }}
                formatter={(value) => [
                  `${formatValue(value as number)} (${(((value as number) / total) * 100).toFixed(1)}%)`,
                  '',
                ]}
              />
              {showLegend && (
                <Legend
                  layout="horizontal"
                  verticalAlign="bottom"
                  wrapperStyle={{ fontSize: '11px', paddingTop: '8px' }}
                />
              )}
            </PieChart>
          </ResponsiveContainer>
          {donut && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <div className="text-center">
                <div className="text-lg font-bold">{formatValue(total)}</div>
                <div className="text-xs text-muted-foreground">Total</div>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export type { PieData }
