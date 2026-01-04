'use client'

import { ArrowDown, ArrowUp, Minus } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Widget } from '../../types'

interface StatWidgetProps {
  widget: Widget
  data: StatData | null
  isLoading: boolean
  error?: string
}

interface StatData {
  value: number | string
  previousValue?: number
  unit?: string
  label?: string
}

// Normalize backend data to expected format
// Backend returns: [{ measure_key: 123 }] (array with first row containing measure)
// Frontend expects: { value: number, unit?: string }
function normalizeStatData(data: unknown): StatData | null {
  if (!data) return null

  // If already in expected format (object with 'value' key)
  if (typeof data === 'object' && !Array.isArray(data) && 'value' in (data as object)) {
    return data as StatData
  }

  // Backend returns array - extract first row and first numeric value
  if (Array.isArray(data) && data.length > 0) {
    const firstRow = data[0] as Record<string, unknown>
    const keys = Object.keys(firstRow)

    // Find the first numeric value (the measure result)
    const measureKey = keys.find((k) => typeof firstRow[k] === 'number')
    if (measureKey) {
      return {
        value: Number(firstRow[measureKey]),
      }
    }

    // Handle string values (e.g., formatted results)
    const stringKey = keys.find((k) => typeof firstRow[k] === 'string')
    if (stringKey) {
      return {
        value: String(firstRow[stringKey]),
      }
    }
  }

  return null
}

function formatValue(value: number | string, unit?: string): string {
  if (typeof value === 'string') return value

  // Format large numbers
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M${unit ? ` ${unit}` : ''}`
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K${unit ? ` ${unit}` : ''}`
  }

  // Format decimals
  if (Number.isInteger(value)) {
    return `${value}${unit ? ` ${unit}` : ''}`
  }
  return `${value.toFixed(2)}${unit ? ` ${unit}` : ''}`
}

function calculateChange(current: number, previous: number): { percent: number; direction: 'up' | 'down' | 'neutral' } {
  if (previous === 0) return { percent: 0, direction: 'neutral' }
  const percent = ((current - previous) / previous) * 100
  return {
    percent: Math.abs(percent),
    direction: percent > 0.5 ? 'up' : percent < -0.5 ? 'down' : 'neutral'
  }
}

export function StatWidget({ widget, data: rawData, isLoading, error }: StatWidgetProps) {
  // Normalize data from backend format
  const data = normalizeStatData(rawData)

  if (isLoading) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">
            <Skeleton className="h-4 w-24" />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-20 mb-2" />
          <Skeleton className="h-4 w-16" />
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
        <CardContent>
          <p className="text-sm text-destructive">{error}</p>
        </CardContent>
      </Card>
    )
  }

  if (!data) {
    return (
      <Card className="h-full">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{widget.title}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No data available</p>
        </CardContent>
      </Card>
    )
  }

  const currentValue = typeof data.value === 'number' ? data.value : parseFloat(data.value) || 0
  const change = data.previousValue !== undefined
    ? calculateChange(currentValue, data.previousValue)
    : null

  return (
    <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {widget.title}
        </CardTitle>
        {widget.description && (
          <CardDescription className="text-xs">{widget.description}</CardDescription>
        )}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">
          {formatValue(data.value, data.unit)}
        </div>
        {change && (
          <div className={cn(
            "flex items-center gap-1 text-xs mt-1",
            change.direction === 'up' && "text-green-600",
            change.direction === 'down' && "text-red-600",
            change.direction === 'neutral' && "text-muted-foreground"
          )}>
            {change.direction === 'up' && <ArrowUp className="h-3 w-3" />}
            {change.direction === 'down' && <ArrowDown className="h-3 w-3" />}
            {change.direction === 'neutral' && <Minus className="h-3 w-3" />}
            <span>{change.percent.toFixed(1)}% from previous period</span>
          </div>
        )}
        {data.label && (
          <p className="text-xs text-muted-foreground mt-1">{data.label}</p>
        )}
      </CardContent>
    </Card>
  )
}

export type { StatData }
