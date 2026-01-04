'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Widget } from '../../types'

interface HeatmapWidgetProps {
  widget: Widget
  data: HeatmapData[] | null
  isLoading: boolean
  error?: string
}

interface HeatmapData {
  x: string
  y: string
  value: number
}

// Color scale from light to dark (based on value intensity)
const getHeatmapColor = (value: number, min: number, max: number): string => {
  if (max === min) return 'hsl(var(--primary) / 0.5)'
  const intensity = (value - min) / (max - min)

  // Blue color scale: lighter to darker
  if (intensity < 0.2) return 'hsl(210 100% 95%)'
  if (intensity < 0.4) return 'hsl(210 100% 80%)'
  if (intensity < 0.6) return 'hsl(210 100% 65%)'
  if (intensity < 0.8) return 'hsl(210 100% 50%)'
  return 'hsl(210 100% 35%)'
}

function formatValue(value: number): string {
  if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(1)
}

export function HeatmapWidget({ widget, data, isLoading, error }: HeatmapWidgetProps) {
  const showValues = widget.config?.showValues !== false
  const showLabels = widget.config?.showLabels !== false

  // Extract unique x and y values and compute min/max
  const { xLabels, yLabels, dataMap, minValue, maxValue } = useMemo(() => {
    if (!data || data.length === 0) {
      return { xLabels: [], yLabels: [], dataMap: new Map(), minValue: 0, maxValue: 0 }
    }

    const xSet = new Set<string>()
    const ySet = new Set<string>()
    const map = new Map<string, number>()
    let min = Infinity
    let max = -Infinity

    data.forEach((item) => {
      xSet.add(item.x)
      ySet.add(item.y)
      const key = `${item.x}|${item.y}`
      map.set(key, item.value)
      min = Math.min(min, item.value)
      max = Math.max(max, item.value)
    })

    return {
      xLabels: Array.from(xSet),
      yLabels: Array.from(ySet),
      dataMap: map,
      minValue: min === Infinity ? 0 : min,
      maxValue: max === -Infinity ? 0 : max,
    }
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

  if (!data || data.length === 0 || xLabels.length === 0 || yLabels.length === 0) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-[200px]">
          <p className="text-sm text-muted-foreground">No heatmap data available</p>
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
      <CardContent className="overflow-auto">
        <div className="min-w-fit">
          {/* Header row with X labels */}
          {showLabels && (
            <div className="flex">
              <div className="w-16 shrink-0" /> {/* Empty corner */}
              {xLabels.map((x) => (
                <div
                  key={x}
                  className="flex-1 min-w-[40px] text-center text-[10px] text-muted-foreground truncate px-0.5"
                  title={x}
                >
                  {x}
                </div>
              ))}
            </div>
          )}

          {/* Data rows */}
          {yLabels.map((y) => (
            <div key={y} className="flex">
              {/* Y label */}
              {showLabels && (
                <div
                  className="w-16 shrink-0 text-[10px] text-muted-foreground truncate pr-2 flex items-center"
                  title={y}
                >
                  {y}
                </div>
              )}

              {/* Cells */}
              {xLabels.map((x) => {
                const key = `${x}|${y}`
                const value = dataMap.get(key) ?? 0
                const color = getHeatmapColor(value, minValue, maxValue)

                return (
                  <div
                    key={key}
                    className={cn(
                      'flex-1 min-w-[40px] min-h-[28px] flex items-center justify-center',
                      'border border-background/50 transition-colors hover:ring-1 hover:ring-primary/50'
                    )}
                    style={{ backgroundColor: color }}
                    title={`${x}, ${y}: ${formatValue(value)}`}
                  >
                    {showValues && (
                      <span className="text-[9px] font-medium text-foreground/80">
                        {formatValue(value)}
                      </span>
                    )}
                  </div>
                )
              })}
            </div>
          ))}

          {/* Legend */}
          <div className="flex items-center justify-end gap-2 mt-3">
            <span className="text-[10px] text-muted-foreground">{formatValue(minValue)}</span>
            <div className="flex h-2 w-24 rounded-sm overflow-hidden">
              <div className="flex-1" style={{ backgroundColor: 'hsl(210 100% 95%)' }} />
              <div className="flex-1" style={{ backgroundColor: 'hsl(210 100% 80%)' }} />
              <div className="flex-1" style={{ backgroundColor: 'hsl(210 100% 65%)' }} />
              <div className="flex-1" style={{ backgroundColor: 'hsl(210 100% 50%)' }} />
              <div className="flex-1" style={{ backgroundColor: 'hsl(210 100% 35%)' }} />
            </div>
            <span className="text-[10px] text-muted-foreground">{formatValue(maxValue)}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export type { HeatmapData }
