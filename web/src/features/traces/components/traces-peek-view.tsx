'use client'

import * as React from 'react'
import { useEffect, useRef } from 'react'
import { useSearchParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { Sheet, SheetContent, SheetTitle } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@/components/ui/resizable'
import { Skeleton } from '@/components/ui/skeleton'
import { PanelLeftOpen, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

import { usePeekNavigation } from '../hooks/use-peek-navigation'
import { useDetailNavigation } from '../hooks/use-detail-navigation'
import { usePeekData } from '../hooks/use-peek-data'
import { getSpansForTrace } from '../api/traces-api'
import type { Span } from '../data/schema'

import { PeekSheetHeader } from './peek-sheet/peek-sheet-header'
import { SpanNavigationPanel, type ViewMode } from './peek-sheet/span-navigation-panel'
import { DetailPanel } from './peek-sheet/detail-panel'

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
// Main Peek View Component
// ============================================================================

/**
 * Peek sheet component for viewing trace details
 * Two-panel layout with tree/timeline on left, details on right
 */
export function TracesPeekView() {
  const searchParams = useSearchParams()
  const peekId = searchParams.get('peek')
  const { closePeek, expandPeek } = usePeekNavigation()
  const { handlePrev, handleNext, canGoPrev, canGoNext } = useDetailNavigation()
  const { trace, isLoading, error, projectId } = usePeekData()

  // Local state for span selection and view mode
  // Store user's manual selection (null = use auto-selected root span)
  const [manuallySelectedSpanId, setManuallySelectedSpanId] = React.useState<string | null>(null)
  const [viewMode, setViewMode] = React.useState<ViewMode>('tree')
  const [isLeftPanelCollapsed, setIsLeftPanelCollapsed] = React.useState(false)

  // Fetch spans for this trace
  const {
    data: spans = [],
    isLoading: spansLoading,
  } = useQuery({
    queryKey: ['traceSpans', projectId, peekId],
    queryFn: () => getSpansForTrace(projectId!, peekId!),
    enabled: !!projectId && !!peekId && !!trace,
    staleTime: 30_000,
  })

  // Compute root span synchronously (no useEffect needed)
  const rootSpan = React.useMemo(() => {
    if (spans.length === 0) return null
    // Find root span (no parent_span_id) or fallback to first span
    return spans.find((s) => !s.parent_span_id) || spans[0]
  }, [spans])

  // Compute effective selected span ID: manual selection takes precedence, otherwise use root
  const selectedSpanId = manuallySelectedSpanId || rootSpan?.span_id || null

  // Find the selected span object
  const selectedSpan = React.useMemo(() => {
    if (!selectedSpanId) return null
    return spans.find((s) => s.span_id === selectedSpanId) || null
  }, [spans, selectedSpanId])

  // Reset manual selection when trace changes (reverts to auto-selected root)
  React.useEffect(() => {
    setManuallySelectedSpanId(null)
  }, [peekId])

  // Ref pattern for stable keyboard handlers
  const handlersRef = useRef({ handlePrev, handleNext, closePeek })

  useEffect(() => {
    handlersRef.current = { handlePrev, handleNext, closePeek }
  }, [handlePrev, handleNext, closePeek])

  // Keyboard shortcuts
  useEffect(() => {
    if (!peekId) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Arrow keys for navigation
      if (e.key === 'ArrowLeft' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlersRef.current.handlePrev()
      }
      if (e.key === 'ArrowRight' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault()
        handlersRef.current.handleNext()
      }
      // Escape to close
      if (e.key === 'Escape') {
        e.preventDefault()
        handlersRef.current.closePeek()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [peekId])

  // Handle span selection (manual user selection)
  const handleSpanSelect = React.useCallback((span: Span) => {
    setManuallySelectedSpanId(span.span_id)
  }, [])

  if (!peekId) return null

  return (
    <Sheet open={!!peekId} onOpenChange={(open) => !open && closePeek()} modal={false}>
      <SheetContent
        side='right'
        className='flex max-h-full min-h-0 min-w-[70vw] flex-col gap-0 overflow-hidden rounded-l-xl p-0'
        onPointerDownOutside={(e) => {
          // Prevent sheet closure when clicking outside with modal={false}
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
          <div className='flex flex-col h-full'>
            {/* Header with badges and navigation */}
            <PeekSheetHeader
              trace={trace}
              onPrevious={handlePrev}
              onNext={handleNext}
              onExpand={() => expandPeek(false)}
              onExpandNewTab={() => expandPeek(true)}
              onClose={closePeek}
              hasPrevious={canGoPrev}
              hasNext={canGoNext}
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
                          onViewModeChange={setViewMode}
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
        )}
      </SheetContent>
    </Sheet>
  )
}
