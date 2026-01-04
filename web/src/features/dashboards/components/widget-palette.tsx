'use client'

import { useState } from 'react'
import {
  ActivityIcon,
  BarChart3Icon,
  HashIcon,
  LayoutGridIcon,
  ListIcon,
  PieChartIcon,
  TableIcon,
  TextIcon,
  TrendingUpIcon,
  Plus,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import type { WidgetType } from '../types'

interface WidgetTypeDefinition {
  type: WidgetType
  label: string
  description: string
  icon: React.ElementType
  defaultSize: { w: number; h: number }
  minSize: { w: number; h: number }
}

const WIDGET_TYPES: WidgetTypeDefinition[] = [
  {
    type: 'stat',
    label: 'Stat',
    description: 'Display a single metric value with optional trend indicator',
    icon: HashIcon,
    defaultSize: { w: 2, h: 1 },
    minSize: { w: 1, h: 1 },
  },
  {
    type: 'time_series',
    label: 'Time Series',
    description: 'Line or area chart showing data over time',
    icon: TrendingUpIcon,
    defaultSize: { w: 4, h: 2 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'bar',
    label: 'Bar Chart',
    description: 'Horizontal or vertical bar chart for comparisons',
    icon: BarChart3Icon,
    defaultSize: { w: 4, h: 2 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'pie',
    label: 'Pie Chart',
    description: 'Circular chart showing proportional data',
    icon: PieChartIcon,
    defaultSize: { w: 3, h: 2 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'table',
    label: 'Table',
    description: 'Display data in rows and columns with sorting',
    icon: TableIcon,
    defaultSize: { w: 4, h: 3 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'heatmap',
    label: 'Heatmap',
    description: 'Grid visualization with color-coded values',
    icon: LayoutGridIcon,
    defaultSize: { w: 4, h: 2 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'histogram',
    label: 'Histogram',
    description: 'Distribution of data across buckets',
    icon: ActivityIcon,
    defaultSize: { w: 4, h: 2 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'trace_list',
    label: 'Trace List',
    description: 'List of recent traces with quick navigation',
    icon: ListIcon,
    defaultSize: { w: 4, h: 3 },
    minSize: { w: 2, h: 2 },
  },
  {
    type: 'text',
    label: 'Text',
    description: 'Static text content with markdown support',
    icon: TextIcon,
    defaultSize: { w: 3, h: 1 },
    minSize: { w: 1, h: 1 },
  },
]

interface WidgetPaletteProps {
  onSelectWidget: (type: WidgetType, defaultSize: { w: number; h: number }) => void
  disabled?: boolean
}

export function WidgetPalette({ onSelectWidget, disabled }: WidgetPaletteProps) {
  const [isOpen, setIsOpen] = useState(false)

  const handleSelectWidget = (widgetDef: WidgetTypeDefinition) => {
    onSelectWidget(widgetDef.type, widgetDef.defaultSize)
    setIsOpen(false)
  }

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        <Button disabled={disabled} size="sm" className="gap-1.5">
          <Plus className="h-4 w-4" />
          Add Widget
        </Button>
      </SheetTrigger>
      <SheetContent side="right" className="w-[400px] sm:w-[540px]">
        <SheetHeader>
          <SheetTitle>Add Widget</SheetTitle>
          <SheetDescription>
            Select a widget type to add to your dashboard. You can configure it after adding.
          </SheetDescription>
        </SheetHeader>
        <ScrollArea className="h-[calc(100vh-140px)] mt-6 pr-4">
          <div className="grid gap-3">
            {WIDGET_TYPES.map((widgetDef) => (
              <WidgetTypeCard
                key={widgetDef.type}
                widgetDef={widgetDef}
                onClick={() => handleSelectWidget(widgetDef)}
              />
            ))}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}

interface WidgetTypeCardProps {
  widgetDef: WidgetTypeDefinition
  onClick: () => void
}

function WidgetTypeCard({ widgetDef, onClick }: WidgetTypeCardProps) {
  const Icon = widgetDef.icon

  return (
    <Card
      className="cursor-pointer transition-colors hover:bg-accent hover:border-primary/50"
      onClick={onClick}
    >
      <CardHeader className="flex flex-row items-start gap-4 p-4">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
          <Icon className="h-5 w-5 text-primary" />
        </div>
        <div className="flex-1 space-y-1">
          <CardTitle className="text-base">{widgetDef.label}</CardTitle>
          <CardDescription className="text-sm">
            {widgetDef.description}
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="px-4 pb-4 pt-0">
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span>
            Default size: {widgetDef.defaultSize.w}×{widgetDef.defaultSize.h}
          </span>
          <span>
            Min: {widgetDef.minSize.w}×{widgetDef.minSize.h}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}

// Export widget type definitions for use elsewhere
export { WIDGET_TYPES, type WidgetTypeDefinition }
