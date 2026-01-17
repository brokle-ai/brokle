'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@/components/ui/resizable'
import { PanelLeftOpen } from 'lucide-react'
import { cn } from '@/lib/utils'

import type { Trace, Span } from '../data/schema'
import { useTraceDetailState } from '../hooks/use-trace-detail-state'
import { useTraces } from '../context/traces-context'

import { TraceDetailHeader } from './trace-detail-header'
import { SpanNavigationPanel } from './peek-sheet/span-navigation-panel'
import { DetailPanel } from './peek-sheet/detail-panel'

interface TraceDetailLayoutProps {
  trace: Trace
  spans: Span[]
  spansLoading: boolean
  projectId: string
  context: 'peek' | 'page'
}

/**
 * TraceDetailLayout - Core layout component for trace details
 *
 * Used in both peek mode (Sheet) and full-page mode.
 * Provides:
 * - Header with trace ID, navigation, and mode controls
 * - Resizable two-panel layout (span tree + detail panel)
 * - Collapsible left panel
 *
 * Behavior varies by context:
 * - peek: Close button, prev/next navigation, expand to full page
 * - page: Back button, open in new tab
 */
export function TraceDetailLayout({
  trace,
  spans,
  spansLoading,
  projectId: _projectId,
  context,
}: TraceDetailLayoutProps) {
  const router = useRouter()
  const { projectSlug } = useTraces()

  // Only use hook state in peek mode (when trace context is from URL params)
  const hookState = useTraceDetailState()

  // For page mode, we need different state management
  const isPeek = context === 'peek'

  // Get state values - either from hook (peek) or from props/route params (page)
  const manuallySelectedSpanId = hookState.selectedSpanId
  const viewMode = hookState.viewMode

  const [isLeftPanelCollapsed, setIsLeftPanelCollapsed] = React.useState(false)

  // Compute root span synchronously for auto-selection
  const rootSpan = React.useMemo(() => {
    if (spans.length === 0) return null
    // Find root span (no parent_span_id) or fallback to first span
    return spans.find((s) => !s.parent_span_id) || spans[0]
  }, [spans])

  // Effective selected span ID: manual selection takes precedence, otherwise auto-select root
  const selectedSpanId = manuallySelectedSpanId || rootSpan?.span_id || null

  // Find the selected span object
  const selectedSpan = React.useMemo(() => {
    if (!selectedSpanId) return null
    return spans.find((s) => s.span_id === selectedSpanId) || null
  }, [spans, selectedSpanId])

  // Handle span selection
  const handleSpanSelect = React.useCallback(
    (span: Span) => {
      hookState.selectSpan(span.span_id)
    },
    [hookState]
  )

  // Handle back navigation (page mode)
  const handleBack = React.useCallback(() => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/traces`)
    } else {
      router.back()
    }
  }, [router, projectSlug])

  // Handle open in new tab
  const handleOpenInNewTab = React.useCallback(() => {
    if (isPeek) {
      // In peek mode, use the hook's expandToFullPage with newTab=true
      hookState.expandToFullPage(true)
    } else {
      // In page mode, just open the current URL in a new tab
      window.open(window.location.href, '_blank')
    }
  }, [isPeek, hookState])

  // Handle expand to full page (peek mode only)
  const handleExpand = React.useCallback(() => {
    hookState.expandToFullPage(false)
  }, [hookState])

  // Keyboard shortcuts for navigation (only in peek mode)
  React.useEffect(() => {
    if (!isPeek) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Arrow keys for prev/next navigation
      if (e.key === 'ArrowLeft' && !e.metaKey && !e.ctrlKey) {
        const target = e.target as HTMLElement
        if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA') {
          e.preventDefault()
          hookState.goToPrev()
        }
      }
      if (e.key === 'ArrowRight' && !e.metaKey && !e.ctrlKey) {
        const target = e.target as HTMLElement
        if (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA') {
          e.preventDefault()
          hookState.goToNext()
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isPeek, hookState])

  return (
    <div className='flex flex-col h-full'>
      {/* Header */}
      <TraceDetailHeader
        trace={trace}
        spans={spans}
        projectId={_projectId}
        selectedSpanId={selectedSpanId || undefined}
        context={context}
        onPrevious={isPeek ? hookState.goToPrev : undefined}
        onNext={isPeek ? hookState.goToNext : undefined}
        onExpand={handleExpand}
        onOpenInNewTab={handleOpenInNewTab}
        onClose={isPeek ? hookState.closeTrace : undefined}
        onBack={!isPeek ? handleBack : undefined}
        hasPrevious={isPeek ? hookState.canGoPrev : false}
        hasNext={isPeek ? hookState.canGoNext : false}
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
          <DetailPanel trace={trace} selectedSpan={null} spans={[]} projectId={_projectId} />
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
                    onViewModeChange={hookState.setViewMode}
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
                  projectId={_projectId}
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
