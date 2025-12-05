'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  TreeDeciduous,
  GanttChart,
  Search,
  PanelLeftClose,
  ChevronsUpDown,
  ChevronsDownUp,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Span } from '../../data/schema'
import { SpanTree, getParentSpanIds } from '../span-tree'
import { SpanTimeline } from '../span-timeline'
import {
  TraceSettingsDropdown,
  useTraceDisplaySettings,
  type TraceDisplaySettings,
} from './trace-settings-dropdown'

export type ViewMode = 'tree' | 'timeline'

interface SpanNavigationPanelProps {
  spans: Span[]
  selectedSpanId?: string
  onSpanSelect: (span: Span) => void
  viewMode: ViewMode
  onViewModeChange: (mode: ViewMode) => void
  onCollapse?: () => void
  isCollapsed?: boolean
  className?: string
}

/**
 * SpanNavigationPanel - Left panel for span tree/timeline navigation
 *
 * Features:
 * - View mode toggle (Tree/Timeline)
 * - Search/filter spans
 * - Settings dropdown (show/hide metrics, level filter, heatmap)
 * - Expand/Collapse all buttons
 * - Span count display
 */
export function SpanNavigationPanel({
  spans,
  selectedSpanId,
  onSpanSelect,
  viewMode,
  onViewModeChange,
  onCollapse,
  isCollapsed,
  className,
}: SpanNavigationPanelProps) {
  const [searchQuery, setSearchQuery] = React.useState('')
  const [displaySettings, setDisplaySettings] = useTraceDisplaySettings()
  const [collapsedNodes, setCollapsedNodes] = React.useState<Set<string>>(new Set())

  // Get parent span IDs for expand/collapse all
  const parentSpanIds = React.useMemo(() => getParentSpanIds(spans), [spans])

  // Filter spans by search query
  const filteredSpans = React.useMemo(() => {
    if (!searchQuery.trim()) return spans

    const query = searchQuery.toLowerCase()
    return spans.filter(
      (span) =>
        span.span_name.toLowerCase().includes(query) ||
        span.span_id.toLowerCase().includes(query) ||
        span.model_name?.toLowerCase().includes(query) ||
        span.provider_name?.toLowerCase().includes(query)
    )
  }, [spans, searchQuery])

  // Handle expand all
  const handleExpandAll = React.useCallback(() => {
    setCollapsedNodes(new Set())
  }, [])

  // Handle collapse all
  const handleCollapseAll = React.useCallback(() => {
    setCollapsedNodes(new Set(parentSpanIds))
  }, [parentSpanIds])

  // Handle individual node toggle
  const handleToggleNode = React.useCallback((spanId: string) => {
    setCollapsedNodes((prev) => {
      const next = new Set(prev)
      if (next.has(spanId)) {
        next.delete(spanId)
      } else {
        next.add(spanId)
      }
      return next
    })
  }, [])

  // Check if all expanded or all collapsed
  const allExpanded = collapsedNodes.size === 0
  const allCollapsed = collapsedNodes.size === parentSpanIds.length && parentSpanIds.length > 0

  if (isCollapsed) return null

  return (
    <div className={cn('flex flex-col h-full bg-muted/30', className)}>
      {/* Header with view toggle, settings, and search */}
      <div className='p-3 space-y-3 border-b'>
        {/* View Mode Toggle + Actions */}
        <div className='flex items-center justify-between'>
          <Tabs value={viewMode} onValueChange={(v) => onViewModeChange(v as ViewMode)}>
            <TabsList className='h-8'>
              <TabsTrigger value='tree' className='h-7 px-3 text-xs gap-1.5'>
                <TreeDeciduous className='h-3.5 w-3.5' />
                Tree
              </TabsTrigger>
              <TabsTrigger value='timeline' className='h-7 px-3 text-xs gap-1.5'>
                <GanttChart className='h-3.5 w-3.5' />
                Timeline
              </TabsTrigger>
            </TabsList>
          </Tabs>

          <div className='flex items-center gap-1'>
            {/* Expand/Collapse All (only in tree view) */}
            {viewMode === 'tree' && parentSpanIds.length > 0 && (
              <>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-7 w-7'
                  onClick={handleExpandAll}
                  disabled={allExpanded}
                  title='Expand all'
                >
                  <ChevronsUpDown className='h-4 w-4' />
                </Button>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-7 w-7'
                  onClick={handleCollapseAll}
                  disabled={allCollapsed}
                  title='Collapse all'
                >
                  <ChevronsDownUp className='h-4 w-4' />
                </Button>
              </>
            )}

            {/* Settings Dropdown */}
            <TraceSettingsDropdown
              settings={displaySettings}
              onSettingsChange={setDisplaySettings}
              className='h-7 w-7'
            />

            {/* Collapse Panel Button */}
            {onCollapse && (
              <Button
                variant='ghost'
                size='icon'
                className='h-7 w-7'
                onClick={onCollapse}
                title='Collapse panel'
              >
                <PanelLeftClose className='h-4 w-4' />
              </Button>
            )}
          </div>
        </div>

        {/* Search Input */}
        <div className='relative'>
          <Search className='absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground' />
          <Input
            type='text'
            placeholder='Search spans...'
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className='h-8 pl-8 text-xs'
          />
        </div>

        {/* Span count */}
        <div className='text-xs text-muted-foreground'>
          {filteredSpans.length === spans.length ? (
            <>{spans.length} span{spans.length !== 1 ? 's' : ''}</>
          ) : (
            <>
              {filteredSpans.length} of {spans.length} span{spans.length !== 1 ? 's' : ''}
            </>
          )}
          {displaySettings.minLevel !== 'all' && (
            <span className='ml-1 text-yellow-600 dark:text-yellow-400'>
              (filtered by level)
            </span>
          )}
        </div>
      </div>

      {/* Span Visualization */}
      <div className='flex-1 min-h-0 overflow-hidden'>
        {filteredSpans.length === 0 ? (
          <div className='flex items-center justify-center h-full text-sm text-muted-foreground'>
            {spans.length === 0 ? 'No spans found' : 'No matching spans'}
          </div>
        ) : viewMode === 'tree' ? (
          <SpanTree
            spans={filteredSpans}
            onSpanSelect={onSpanSelect}
            selectedSpanId={selectedSpanId}
            displaySettings={displaySettings}
            collapsedNodes={collapsedNodes}
            onToggleNode={handleToggleNode}
            className='h-full'
          />
        ) : (
          <SpanTimeline
            spans={filteredSpans}
            onSpanSelect={onSpanSelect}
            selectedSpanId={selectedSpanId}
          />
        )}
      </div>
    </div>
  )
}
