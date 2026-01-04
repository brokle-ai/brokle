'use client'

import { AlertCircle } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { StatWidget, type StatData } from './stat-widget'
import { TimeSeriesWidget, type TimeSeriesData } from './time-series-widget'
import { TableWidget, type TableData } from './table-widget'
import { BarWidget, type BarData } from './bar-widget'
import { PieWidget, type PieData } from './pie-widget'
import { HeatmapWidget, type HeatmapData } from './heatmap-widget'
import { HistogramWidget, type HistogramData } from './histogram-widget'
import { TraceListWidget, type TraceListData } from './trace-list-widget'
import { TextWidget, type TextData } from './text-widget'
import type { Widget, TimeRange } from '../../types'

interface WidgetRendererProps {
  widget: Widget
  data: unknown
  isLoading: boolean
  error?: string
  projectSlug?: string
  timeRange?: TimeRange
}

export function WidgetRenderer({ widget, data, isLoading, error, projectSlug, timeRange }: WidgetRendererProps) {
  switch (widget.type) {
    case 'stat':
      return (
        <StatWidget
          widget={widget}
          data={data as StatData | null}
          isLoading={isLoading}
          error={error}
        />
      )
    case 'time_series':
      return (
        <TimeSeriesWidget
          widget={widget}
          data={data as TimeSeriesData[] | null}
          isLoading={isLoading}
          error={error}
          projectSlug={projectSlug}
          timeRange={timeRange}
        />
      )
    case 'table':
      return (
        <TableWidget
          widget={widget}
          data={data as TableData | null}
          isLoading={isLoading}
          error={error}
        />
      )
    case 'bar':
      return (
        <BarWidget
          widget={widget}
          data={data as BarData[] | null}
          isLoading={isLoading}
          error={error}
          projectSlug={projectSlug}
          timeRange={timeRange}
        />
      )
    case 'pie':
      return (
        <PieWidget
          widget={widget}
          data={data as PieData[] | null}
          isLoading={isLoading}
          error={error}
          projectSlug={projectSlug}
          timeRange={timeRange}
        />
      )
    case 'heatmap':
      return (
        <HeatmapWidget
          widget={widget}
          data={data as HeatmapData[] | null}
          isLoading={isLoading}
          error={error}
        />
      )
    case 'histogram':
      return (
        <HistogramWidget
          widget={widget}
          data={data as HistogramData[] | null}
          isLoading={isLoading}
          error={error}
        />
      )
    case 'trace_list':
      return (
        <TraceListWidget
          widget={widget}
          data={data as TraceListData | null}
          isLoading={isLoading}
          error={error}
          projectSlug={projectSlug}
        />
      )
    case 'text':
      return (
        <TextWidget
          widget={widget}
          data={data as TextData | null}
          isLoading={isLoading}
          error={error}
        />
      )
    default:
      return (
        <Card className="h-full">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <AlertCircle className="h-4 w-4 text-muted-foreground" />
              {widget.title || 'Unknown Widget'}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex items-center justify-center h-[200px]">
            <p className="text-sm text-muted-foreground">
              Unknown widget type: {widget.type}
            </p>
          </CardContent>
        </Card>
      )
  }
}
