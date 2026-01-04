'use client'

import { useMemo, useCallback, useState, useEffect, useRef } from 'react'
import { GridLayout, type Layout, type LayoutItem as RGLLayoutItem } from 'react-grid-layout'
import { Edit, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import type { Widget, LayoutItem, WidgetType, TimeRange } from '../types'
import { WidgetRenderer } from './widgets'
import { WidgetErrorBoundary } from './widget-error-boundary'
import { WidgetSkeletonRenderer } from './widgets/widget-skeletons'

import 'react-grid-layout/css/styles.css'
import './dashboard-grid.css'

// Widget type-specific constraints for min dimensions
interface WidgetConstraints {
  minW: number
  minH: number
  maxW?: number
  maxH?: number
}

const WIDGET_CONSTRAINTS: Record<WidgetType, WidgetConstraints> = {
  stat: { minW: 1, minH: 1, maxW: 4, maxH: 2 },
  time_series: { minW: 2, minH: 2 },
  bar: { minW: 2, minH: 2 },
  pie: { minW: 2, minH: 2 },
  table: { minW: 2, minH: 2 },
  heatmap: { minW: 2, minH: 2 },
  histogram: { minW: 2, minH: 2 },
  trace_list: { minW: 2, minH: 2 },
  text: { minW: 1, minH: 1 },
}

const DEFAULT_CONSTRAINT: WidgetConstraints = { minW: 1, minH: 1 }

const COLS = 12
const ROW_HEIGHT = 100
const MARGIN: [number, number] = [16, 16]

interface DashboardGridProps {
  widgets: Widget[]
  layout: LayoutItem[]
  queryResults?: Record<string, { data: unknown; error?: string }>
  isLoading?: boolean
  isEditMode?: boolean
  isLocked?: boolean
  projectSlug?: string
  timeRange?: TimeRange
  onLayoutChange?: (layout: LayoutItem[]) => void
  onEditWidget?: (widget: Widget) => void
  onDeleteWidget?: (widgetId: string) => void
  className?: string
}

/**
 * Widget edit overlay with edit and delete buttons
 */
function WidgetEditOverlay({
  onEdit,
  onDelete,
}: {
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <div className="absolute top-2 right-2 z-10 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
      <Button
        variant="secondary"
        size="icon"
        className="h-7 w-7"
        onClick={(e) => {
          e.stopPropagation()
          onEdit()
        }}
      >
        <Edit className="h-3.5 w-3.5" />
      </Button>
      <Button
        variant="secondary"
        size="icon"
        className="h-7 w-7 hover:bg-destructive hover:text-destructive-foreground"
        onClick={(e) => {
          e.stopPropagation()
          onDelete()
        }}
      >
        <Trash2 className="h-3.5 w-3.5" />
      </Button>
    </div>
  )
}

/**
 * Converts LayoutItem[] to react-grid-layout Layout
 */
function toGridLayout(
  layout: LayoutItem[],
  widgets: Widget[],
  isStatic: boolean
): RGLLayoutItem[] {
  return layout.map((item) => {
    const widget = widgets.find((w) => w.id === item.widget_id)
    const widgetType = widget?.type as WidgetType | undefined
    const constraints = widgetType ? WIDGET_CONSTRAINTS[widgetType] : DEFAULT_CONSTRAINT

    return {
      i: item.widget_id,
      x: item.x,
      y: item.y,
      w: item.w,
      h: item.h,
      minW: constraints.minW,
      minH: constraints.minH,
      maxW: constraints.maxW,
      maxH: constraints.maxH,
      static: isStatic,
    }
  })
}

/**
 * Converts react-grid-layout Layout back to LayoutItem[]
 */
function fromGridLayout(gridLayout: Layout): LayoutItem[] {
  return gridLayout.map((item) => ({
    widget_id: item.i,
    x: item.x,
    y: item.y,
    w: item.w,
    h: item.h,
  }))
}

/**
 * Creates a default layout for widgets that don't have one
 */
function createDefaultLayout(widgets: Widget[], existingLayout: LayoutItem[]): LayoutItem[] {
  const existingIds = new Set(existingLayout.map((l) => l.widget_id))
  const newItems: LayoutItem[] = []

  widgets.forEach((widget) => {
    if (!existingIds.has(widget.id)) {
      const widgetType = widget.type as WidgetType
      const constraints = WIDGET_CONSTRAINTS[widgetType] || DEFAULT_CONSTRAINT

      // Stack new widgets vertically
      const y = existingLayout.length + newItems.length
      newItems.push({
        widget_id: widget.id,
        x: 0,
        y: y * 2,
        w: Math.max(constraints.minW, 3),
        h: Math.max(constraints.minH, 2),
      })
    }
  })

  return [...existingLayout, ...newItems]
}

export function DashboardGrid({
  widgets,
  layout,
  queryResults,
  isLoading = false,
  isEditMode = false,
  isLocked = false,
  projectSlug,
  timeRange,
  onLayoutChange,
  onEditWidget,
  onDeleteWidget,
  className,
}: DashboardGridProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [containerWidth, setContainerWidth] = useState(1200)

  // Ensure all widgets have layout items
  const normalizedLayout = useMemo(
    () => createDefaultLayout(widgets, layout),
    [widgets, layout]
  )

  // Convert to grid layout format
  const gridLayout = useMemo(
    () => toGridLayout(normalizedLayout, widgets, isLocked || !isEditMode),
    [normalizedLayout, widgets, isLocked, isEditMode]
  )

  const handleLayoutChange = useCallback(
    (newLayout: Layout) => {
      if (!onLayoutChange) return
      const updatedLayout = fromGridLayout(newLayout)
      onLayoutChange(updatedLayout)
    },
    [onLayoutChange]
  )

  // Measure container width
  useEffect(() => {
    const updateWidth = () => {
      if (containerRef.current) {
        setContainerWidth(containerRef.current.clientWidth)
      }
    }

    updateWidth()

    // Use ResizeObserver for more accurate updates
    const resizeObserver = new ResizeObserver(updateWidth)
    if (containerRef.current) {
      resizeObserver.observe(containerRef.current)
    }

    return () => resizeObserver.disconnect()
  }, [])

  if (widgets.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center border rounded-lg border-dashed">
        <p className="text-muted-foreground mb-2">No widgets configured</p>
        <p className="text-xs text-muted-foreground">
          Add widgets to visualize your observability data
        </p>
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      className={cn(
        'dashboard-grid',
        isEditMode && !isLocked && 'dashboard-grid--edit-mode',
        className
      )}
    >
      <GridLayout
        className="layout"
        layout={gridLayout}
        width={containerWidth}
        gridConfig={{
          cols: COLS,
          rowHeight: ROW_HEIGHT,
          margin: MARGIN,
          containerPadding: [0, 0],
        }}
        dragConfig={{
          enabled: isEditMode && !isLocked,
          handle: '.widget-drag-handle',
        }}
        resizeConfig={{
          enabled: isEditMode && !isLocked,
        }}
        onLayoutChange={handleLayoutChange}
        autoSize
      >
        {widgets.map((widget) => {
          const widgetResult = queryResults?.[widget.id]
          const widgetData = widgetResult?.data ?? null
          const widgetError = widgetResult?.error
          const widgetLoading = isLoading && !queryResults

          return (
            <div
              key={widget.id}
              className={cn(
                'dashboard-grid__item',
                isEditMode && !isLocked && 'group relative'
              )}
            >
              {isEditMode && !isLocked && (
                <>
                  <div className="widget-drag-handle" />
                  {onEditWidget && onDeleteWidget && (
                    <WidgetEditOverlay
                      onEdit={() => onEditWidget(widget)}
                      onDelete={() => onDeleteWidget(widget.id)}
                    />
                  )}
                </>
              )}
              <WidgetErrorBoundary
                widgetId={widget.id}
                widgetTitle={widget.title}
                className="h-full"
              >
                {widgetLoading ? (
                  <div className="h-full rounded-lg border bg-card">
                    <WidgetSkeletonRenderer
                      type={widget.type as WidgetType}
                      className="h-full"
                    />
                  </div>
                ) : (
                  <WidgetRenderer
                    widget={widget}
                    data={widgetData}
                    isLoading={widgetLoading}
                    error={widgetError}
                    projectSlug={projectSlug}
                    timeRange={timeRange}
                  />
                )}
              </WidgetErrorBoundary>
            </div>
          )
        })}
      </GridLayout>
    </div>
  )
}
