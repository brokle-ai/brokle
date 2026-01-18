'use client'

import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Sheet, SheetContent, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@/components/ui/resizable'
import { PanelLeftOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Skeleton } from '@/components/ui/skeleton'
import { AlertCircle } from 'lucide-react'

import type { Trace, Span } from '../../data/schema'
import { traceQueryKeys } from '../../hooks/trace-query-keys'
import { getSpansForTrace } from '../../api/traces-api'
import { usePeekNavigation } from '../../hooks/use-peek-navigation'
import { useDetailNavigation } from '../../hooks/use-detail-navigation'
import { usePeekData } from '../../hooks/use-peek-data'
import { usePeekSheetState } from '../../hooks/use-peek-sheet-state'

import { PeekSheetHeader } from './peek-sheet-header'
import { SpanNavigationPanel, type ViewMode } from './span-navigation-panel'
import { DetailPanel } from './detail-panel'

// Re-export components for external use
export { MetadataBadgeRow } from './metadata-badge-row'
export { InputOutputSection } from './input-output-section'
export { MetadataSection } from './metadata-section'
export { PeekSheetHeader } from './peek-sheet-header'
export { SpanNavigationPanel } from './span-navigation-panel'
export { DetailPanel } from './detail-panel'

// ============================================================================
// Loading State
// ============================================================================

function PeekSheetSkeleton() {
  return (
    <div className='flex flex-col h-full'>
      {/* Header skeleton */}
      <div className='border-b p-4 space-y-3'>
        <Skeleton className='h-6 w-48' />
        <Skeleton className='h-4 w-32' />
        <div className='flex gap-2'>
          <Skeleton className='h-5 w-16' />
          <Skeleton className='h-5 w-20' />
          <Skeleton className='h-5 w-24' />
        </div>
      </div>

      {/* Content skeleton */}
      <div className='flex-1 flex'>
        <div className='w-1/3 border-r p-4 space-y-3'>
          <Skeleton className='h-8 w-full' />
          <Skeleton className='h-8 w-full' />
          <Skeleton className='h-32 w-full' />
        </div>
        <div className='flex-1 p-4 space-y-4'>
          <Skeleton className='h-8 w-full' />
          <Skeleton className='h-24 w-full' />
          <Skeleton className='h-24 w-full' />
        </div>
      </div>
    </div>
  )
}

// ============================================================================
// Error State
// ============================================================================

function PeekSheetError({ message }: { message: string }) {
  return (
    <div className='flex flex-col items-center justify-center h-full py-12 text-destructive'>
      <AlertCircle className='h-12 w-12 mb-4' />
      <h3 className='text-lg font-semibold mb-2'>Failed to load trace</h3>
      <p className='text-sm text-muted-foreground'>{message}</p>
    </div>
  )
}

// ============================================================================
// Main Content
// ============================================================================

interface PeekSheetContentProps {
  trace: Trace
  projectId: string
  onClose: () => void
  onExpand: () => void
  onExpandNewTab: () => void
  hasPrevious: boolean
  hasNext: boolean
  onPrevious: () => void
  onNext: () => void
}

function PeekSheetContent({
  trace,
  projectId,
  onClose,
  onExpand,
  onExpandNewTab,
  hasPrevious,
  hasNext,
  onPrevious,
  onNext,
}: PeekSheetContentProps) {
  const { selectedSpanId, viewMode, setSelectedSpan, setViewMode } = usePeekSheetState()
  const [isLeftPanelCollapsed, setIsLeftPanelCollapsed] = React.useState(false)

  // Fetch spans for this trace
  const {
    data: spans = [],
    isLoading: spansLoading,
  } = useQuery({
    queryKey: traceQueryKeys.spans(projectId, trace.trace_id),
    queryFn: () => getSpansForTrace(projectId, trace.trace_id),
    enabled: !!projectId && !!trace.trace_id,
    staleTime: 30_000,
  })

  // Find the selected span object
  const selectedSpan = React.useMemo(() => {
    if (!selectedSpanId) return null
    return spans.find((s) => s.span_id === selectedSpanId) || null
  }, [spans, selectedSpanId])

  // Handle span selection
  const handleSpanSelect = React.useCallback(
    (span: Span) => {
      setSelectedSpan(span.span_id)
    },
    [setSelectedSpan]
  )

  // Handle view mode change
  const handleViewModeChange = React.useCallback(
    (mode: ViewMode) => {
      setViewMode(mode)
    },
    [setViewMode]
  )

  return (
    <div className='flex flex-col h-full'>
      {/* Header */}
      <PeekSheetHeader
        trace={trace}
        projectId={projectId}
        onPrevious={onPrevious}
        onNext={onNext}
        onExpand={onExpand}
        onExpandNewTab={onExpandNewTab}
        onClose={onClose}
        hasPrevious={hasPrevious}
        hasNext={hasNext}
      />

      {/* Two-Panel Layout */}
      <div className='flex-1 min-h-0'>
        {spansLoading ? (
          <div className='flex items-center justify-center h-full'>
            <div className='flex flex-col items-center space-y-2'>
              <div className='h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent' />
              <p className='text-sm text-muted-foreground'>Loading spans...</p>
            </div>
          </div>
        ) : spans.length === 0 ? (
          // No spans - just show detail panel
          <DetailPanel trace={trace} selectedSpan={null} spans={[]} projectId={projectId} />
        ) : (
          // Has spans - show resizable two-panel layout
          <ResizablePanelGroup direction='horizontal' className='h-full'>
            {/* Left Panel - Span Navigation */}
            {!isLeftPanelCollapsed && (
              <>
                <ResizablePanel
                  defaultSize={35}
                  minSize={20}
                  maxSize={50}
                  className='min-w-0'
                >
                  <SpanNavigationPanel
                    spans={spans}
                    selectedSpanId={selectedSpanId || undefined}
                    onSpanSelect={handleSpanSelect}
                    viewMode={viewMode}
                    onViewModeChange={handleViewModeChange}
                    onCollapse={() => setIsLeftPanelCollapsed(true)}
                  />
                </ResizablePanel>
                <ResizableHandle withHandle />
              </>
            )}

            {/* Right Panel - Detail View */}
            <ResizablePanel defaultSize={isLeftPanelCollapsed ? 100 : 65} className='min-w-0'>
              <div className='relative h-full'>
                {/* Expand button when collapsed */}
                {isLeftPanelCollapsed && (
                  <Button
                    variant='ghost'
                    size='icon'
                    className='absolute left-2 top-2 z-10 h-8 w-8'
                    onClick={() => setIsLeftPanelCollapsed(false)}
                    title='Expand span panel'
                  >
                    <PanelLeftOpen className='h-4 w-4' />
                  </Button>
                )}
                <DetailPanel
                  trace={trace}
                  selectedSpan={selectedSpan}
                  spans={spans}
                  projectId={projectId}
                  className={cn(isLeftPanelCollapsed && 'pl-12')}
                />
              </div>
            </ResizablePanel>
          </ResizablePanelGroup>
        )}
      </div>
    </div>
  )
}

// ============================================================================
// Main Export: TracePeekSheet
// ============================================================================

/**
 * TracePeekSheet - Peek sheet component for trace visualization
 *
 * Features:
 * - Two-panel resizable layout
 * - Tree/Timeline span visualization toggle
 * - Collapsible left panel
 * - Rich header with metadata badges
 * - Input/Output sections with copy
 * - Hierarchical metadata display
 * - URL-based state management
 * - Keyboard navigation
 */
export function TracePeekSheet() {
  const { trace, isLoading, error, peekId, projectId } = usePeekData()
  const { closePeek, expandPeek } = usePeekNavigation()
  const { handlePrev, handleNext, canGoPrev, canGoNext } = useDetailNavigation()

  // Keyboard shortcuts
  React.useEffect(() => {
    if (!peekId) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Arrow keys for navigation
      if (e.key === 'ArrowLeft' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlePrev()
      }
      if (e.key === 'ArrowRight' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handleNext()
      }
      // Escape to close
      if (e.key === 'Escape') {
        e.preventDefault()
        closePeek()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [peekId, handlePrev, handleNext, closePeek])

  if (!peekId) return null

  return (
    <Sheet open={!!peekId} onOpenChange={(open) => !open && closePeek()} modal={false}>
      <SheetContent
        side='right'
        className='flex max-h-full min-h-0 min-w-[70vw] flex-col gap-0 overflow-hidden rounded-l-xl p-0'
        onPointerDownOutside={(e) => {
          // Prevent sheet closure when clicking outside
          e.preventDefault()
        }}
        tabIndex={-1}
        hideCloseButton
      >
        {/* Visually hidden title for accessibility */}
        <SheetTitle className='sr-only'>Trace Details</SheetTitle>

        {isLoading ? (
          <PeekSheetSkeleton />
        ) : error || !trace ? (
          <PeekSheetError message={error?.message || 'Trace not found'} />
        ) : (
          <PeekSheetContent
            trace={trace}
            projectId={projectId!}
            onClose={closePeek}
            onExpand={() => expandPeek(false)}
            onExpandNewTab={() => expandPeek(true)}
            hasPrevious={canGoPrev}
            hasNext={canGoNext}
            onPrevious={handlePrev}
            onNext={handleNext}
          />
        )}
      </SheetContent>
    </Sheet>
  )
}
