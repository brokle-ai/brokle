'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { LineChart } from '@/components/shared/charts/line-chart'
import { ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { TimeSeriesPoint, OverviewTimeRange } from '../types'

interface TraceVolumeChartProps {
  data: TimeSeriesPoint[]
  timeRange: OverviewTimeRange
  isLoading?: boolean
  error?: string | null
  className?: string
  projectSlug?: string
}

function formatTime(timestamp: string, timeRange: OverviewTimeRange): string {
  const date = new Date(timestamp)

  switch (timeRange) {
    case '24h':
      return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
    case '7d':
      return date.toLocaleDateString('en-US', { weekday: 'short', hour: '2-digit' })
    case '30d':
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    default:
      return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
  }
}

export function TraceVolumeChart({
  data,
  timeRange,
  isLoading,
  error,
  className,
  projectSlug,
}: TraceVolumeChartProps) {
  const router = useRouter()

  const chartData = React.useMemo(() => {
    return data.map((point) => ({
      time: formatTime(point.timestamp, timeRange),
      traces: point.value,
    }))
  }, [data, timeRange])

  const handleViewAll = () => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/traces`)
    }
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base font-medium">Trace Volume</CardTitle>
        <Button
          variant="ghost"
          size="sm"
          className="gap-1 text-xs"
          onClick={handleViewAll}
        >
          View All
          <ArrowRight className="h-3 w-3" />
        </Button>
      </CardHeader>
      <CardContent>
        {data.length === 0 && !isLoading ? (
          <div className="h-[200px] flex items-center justify-center text-muted-foreground">
            No trace data yet. Start sending traces to see activity.
          </div>
        ) : (
          <LineChart
            data={chartData}
            xKey="time"
            yKey="traces"
            height={200}
            loading={isLoading}
            error={error ?? undefined}
            colors={['#3b82f6']}
            formatYAxis={(value) => value.toLocaleString()}
          />
        )}
      </CardContent>
    </Card>
  )
}
