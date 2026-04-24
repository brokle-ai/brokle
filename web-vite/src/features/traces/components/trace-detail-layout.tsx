import * as React from 'react'
import { Button } from '@/components/ui/button'
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from '@/components/ui/resizable'
import { Network, FileText } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Span, TraceDetail } from '../api/types'
import { SpanTree } from './span-tree'
import { SpanDetailPanel } from './span-detail-panel'
import { IoPreview } from './io-preview'
import { SpanGraph } from './span-graph'
import { TraceDetailHeader } from './trace-detail-header'

interface TraceDetailLayoutProps {
  trace: TraceDetail
  spans: Span[]
  orgId: string
  projectId: string
  className?: string
}

type MiddleView = 'graph' | 'io'

/**
 * TraceDetailLayout - Three-panel resizable layout for the trace page.
 *
 *   ┌──────────────┬──────────────────────────┬─────────────┐
 *   │  span tree   │   span-graph │ io-preview │ span detail │
 *   │     30%      │           45%             │    25%      │
 *   └──────────────┴──────────────────────────┴─────────────┘
 *
 * The middle panel toggles between the reactflow span graph and the
 * IO preview of the currently-selected span. Selection is a single
 * state shared across all three panels (tree click ⇔ graph node click
 * ⇔ detail panel contents), matching the Langfuse-style interaction
 * where every view is a lens over the same focused span.
 */
export function TraceDetailLayout({
  trace,
  spans,
  orgId,
  projectId,
  className,
}: TraceDetailLayoutProps) {
  const [selectedSpanId, setSelectedSpanId] = React.useState<
    string | undefined
  >(() => spans.find((s) => !s.parent_span_id)?.span_id ?? spans[0]?.span_id)
  const [middleView, setMiddleView] = React.useState<MiddleView>('graph')

  // Keep selection in sync when the spans array changes (e.g., refetch).
  React.useEffect(() => {
    if (!selectedSpanId) return
    if (!spans.some((s) => s.span_id === selectedSpanId)) {
      setSelectedSpanId(
        spans.find((s) => !s.parent_span_id)?.span_id ?? spans[0]?.span_id,
      )
    }
  }, [spans, selectedSpanId])

  const selectedSpan = React.useMemo(
    () => spans.find((s) => s.span_id === selectedSpanId),
    [spans, selectedSpanId],
  )

  const handleSpanSelect = React.useCallback((span: Span) => {
    setSelectedSpanId(span.span_id)
  }, [])

  return (
    <div className={cn('flex h-full flex-col', className)}>
      <TraceDetailHeader
        trace={trace}
        spans={spans}
        orgId={orgId}
        projectId={projectId}
      />

      <div className="min-h-0 flex-1">
        {spans.length === 0 ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            This trace has no spans.
          </div>
        ) : (
          <ResizablePanelGroup direction="horizontal" className="h-full">
            {/* Left — span tree */}
            <ResizablePanel
              defaultSize={30}
              minSize={20}
              maxSize={50}
              className="min-w-0"
            >
              <div className="flex h-full flex-col">
                <div className="flex items-center justify-between border-b px-3 py-2">
                  <span className="text-xs font-medium">
                    Spans ({spans.length})
                  </span>
                </div>
                <div className="min-h-0 flex-1 overflow-auto p-3">
                  <SpanTree
                    spans={spans}
                    selectedSpanId={selectedSpanId}
                    onSpanSelect={handleSpanSelect}
                  />
                </div>
              </div>
            </ResizablePanel>

            <ResizableHandle withHandle />

            {/* Middle — graph or I/O preview */}
            <ResizablePanel defaultSize={45} minSize={25} className="min-w-0">
              <div className="flex h-full flex-col">
                <div className="flex items-center justify-between border-b px-3 py-1.5">
                  <div className="flex items-center gap-1">
                    <Button
                      variant={middleView === 'graph' ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-7 gap-1.5 px-2 text-xs"
                      onClick={() => setMiddleView('graph')}
                    >
                      <Network className="h-3.5 w-3.5" />
                      Graph
                    </Button>
                    <Button
                      variant={middleView === 'io' ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-7 gap-1.5 px-2 text-xs"
                      onClick={() => setMiddleView('io')}
                    >
                      <FileText className="h-3.5 w-3.5" />
                      I/O
                    </Button>
                  </div>
                </div>
                <div className="min-h-0 flex-1">
                  {middleView === 'graph' ? (
                    <SpanGraph
                      spans={spans}
                      selectedSpanId={selectedSpanId}
                      onSpanSelect={handleSpanSelect}
                    />
                  ) : (
                    <div className="h-full space-y-4 overflow-auto p-4">
                      <IoPreview
                        value={selectedSpan?.input}
                        label="Input"
                      />
                      <IoPreview
                        value={selectedSpan?.output}
                        label="Output"
                      />
                    </div>
                  )}
                </div>
              </div>
            </ResizablePanel>

            <ResizableHandle withHandle />

            {/* Right — span detail panel */}
            <ResizablePanel defaultSize={25} minSize={20} className="min-w-0">
              <div className="h-full overflow-auto p-3">
                {selectedSpan ? (
                  <SpanDetailPanel span={selectedSpan} />
                ) : (
                  <p className="text-sm text-muted-foreground">
                    Select a span to see details.
                  </p>
                )}
              </div>
            </ResizablePanel>
          </ResizablePanelGroup>
        )}
      </div>
    </div>
  )
}
