'use client'

import * as React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { IOPreview } from '@/components/traces/IOPreview'
import type { Trace, Span } from '../data/schema'
import { statuses, statusCodeToString } from '../data/constants'
import { safeFormat, formatDuration, formatCost } from '../utils/format-helpers'
import { getSpansForTrace, getTraceWithScores } from '../api/traces-api'
import { SpanTree } from './span-tree'
import { SpanTimeline } from './span-timeline'
import { SpanDetailPanel } from './span-detail-panel'
import { Clock, DollarSign, Layers, Server, Tag, TreeDeciduous, GanttChart, MessageSquare, Info, FileInput } from 'lucide-react'

// ============================================================================
// Types
// ============================================================================

interface TraceDetailViewProps {
  trace: Trace
  projectId: string
}

// ============================================================================
// Metrics Grid Component
// ============================================================================

function MetricsGrid({ trace }: { trace: Trace }) {
  return (
    <div className='grid grid-cols-2 md:grid-cols-4 gap-4'>
      <Card>
        <CardHeader className='pb-2'>
          <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
            <Clock className='h-4 w-4' />
            Duration
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold'>{formatDuration(trace.duration)}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className='pb-2'>
          <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
            <DollarSign className='h-4 w-4' />
            Cost
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold'>{formatCost(trace.cost)}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className='pb-2'>
          <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
            <Layers className='h-4 w-4' />
            Spans
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold'>{trace.spanCount}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className='pb-2'>
          <CardTitle className='text-sm font-medium text-muted-foreground flex items-center gap-2'>
            <Server className='h-4 w-4' />
            Tokens
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className='text-2xl font-bold'>{trace.tokens?.toLocaleString() || '-'}</div>
        </CardContent>
      </Card>
    </div>
  )
}

// ============================================================================
// Trace Metadata Card Component
// ============================================================================

function TraceMetadataCard({ trace }: { trace: Trace }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <Info className='h-4 w-4' />
          Trace Metadata
        </CardTitle>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid grid-cols-2 gap-4'>
          <div>
            <div className='text-sm font-medium text-muted-foreground'>Start Time</div>
            <div className='text-sm'>{safeFormat(trace.start_time, 'PPpp')}</div>
          </div>
          {trace.end_time && (
            <div>
              <div className='text-sm font-medium text-muted-foreground'>End Time</div>
              <div className='text-sm'>{safeFormat(trace.end_time, 'PPpp')}</div>
            </div>
          )}
          {trace.environment && (
            <div>
              <div className='text-sm font-medium text-muted-foreground'>Environment</div>
              <Badge variant='outline'>{trace.environment}</Badge>
            </div>
          )}
          {trace.service_name && (
            <div>
              <div className='text-sm font-medium text-muted-foreground'>Service</div>
              <div className='text-sm'>{trace.service_name}</div>
            </div>
          )}
          {trace.service_version && (
            <div>
              <div className='text-sm font-medium text-muted-foreground'>Version</div>
              <div className='text-sm font-mono'>{trace.service_version}</div>
            </div>
          )}
          {trace.user_id && (
            <div>
              <div className='text-sm font-medium text-muted-foreground'>User ID</div>
              <div className='text-sm font-mono truncate'>{trace.user_id}</div>
            </div>
          )}
          {trace.session_id && (
            <div>
              <div className='text-sm font-medium text-muted-foreground'>Session ID</div>
              <div className='text-sm font-mono truncate'>{trace.session_id}</div>
            </div>
          )}
        </div>

        {trace.tags && trace.tags.length > 0 && (
          <>
            <Separator />
            <div>
              <div className='text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2'>
                <Tag className='h-4 w-4' />
                Tags
              </div>
              <div className='flex flex-wrap gap-2'>
                {trace.tags.map((tag) => (
                  <Badge key={tag} variant='secondary'>
                    {tag}
                  </Badge>
                ))}
              </div>
            </div>
          </>
        )}

        {/* I/O Data */}
        {(trace.input || trace.output) && (
          <>
            <Separator />
            <div className='space-y-3'>
              {trace.input && (
                <div>
                  <div className='text-sm font-medium text-muted-foreground mb-1'>Input</div>
                  <pre className='text-xs bg-muted p-2 rounded-md overflow-x-auto max-h-32'>
                    {trace.input}
                  </pre>
                </div>
              )}
              {trace.output && (
                <div>
                  <div className='text-sm font-medium text-muted-foreground mb-1'>Output</div>
                  <pre className='text-xs bg-muted p-2 rounded-md overflow-x-auto max-h-32'>
                    {trace.output}
                  </pre>
                </div>
              )}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}

// ============================================================================
// Spans Tab Content Component
// ============================================================================

interface SpansTabContentProps {
  projectId: string
  traceId: string
  spanCount: number
}

function SpansTabContent({ projectId, traceId, spanCount }: SpansTabContentProps) {
  const [selectedSpan, setSelectedSpan] = React.useState<Span | null>(null)
  const [viewMode, setViewMode] = React.useState<'tree' | 'timeline'>('tree')

  // Fetch spans for this trace
  const {
    data: spans,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['traceSpans', projectId, traceId],
    queryFn: () => getSpansForTrace(projectId, traceId),
    enabled: !!projectId && !!traceId,
    staleTime: 30_000,
  })

  if (isLoading) {
    return (
      <div className='flex items-center justify-center py-8'>
        <div className='flex flex-col items-center space-y-2'>
          <div className='h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent' />
          <p className='text-sm text-muted-foreground'>Loading spans...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className='flex items-center justify-center py-8'>
        <p className='text-sm text-destructive'>Failed to load spans</p>
      </div>
    )
  }

  if (!spans || spans.length === 0) {
    return (
      <div className='flex items-center justify-center py-8'>
        <p className='text-sm text-muted-foreground'>No spans found for this trace</p>
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      {/* View mode toggle */}
      <div className='flex items-center justify-between'>
        <p className='text-sm text-muted-foreground'>
          {spans.length} span{spans.length !== 1 ? 's' : ''} in this trace
        </p>
        <TabsList className='h-8'>
          <TabsTrigger
            value='tree'
            className='h-7 px-3 text-xs'
            onClick={() => setViewMode('tree')}
            data-state={viewMode === 'tree' ? 'active' : 'inactive'}
          >
            <TreeDeciduous className='h-3.5 w-3.5 mr-1.5' />
            Tree
          </TabsTrigger>
          <TabsTrigger
            value='timeline'
            className='h-7 px-3 text-xs'
            onClick={() => setViewMode('timeline')}
            data-state={viewMode === 'timeline' ? 'active' : 'inactive'}
          >
            <GanttChart className='h-3.5 w-3.5 mr-1.5' />
            Timeline
          </TabsTrigger>
        </TabsList>
      </div>

      {/* Span visualization */}
      <div className='flex gap-4'>
        {/* Main span view */}
        <div className='flex-1 min-w-0'>
          {viewMode === 'tree' ? (
            <Card>
              <CardContent className='p-4'>
                <ScrollArea className='h-[500px]'>
                  <SpanTree
                    spans={spans}
                    onSpanSelect={setSelectedSpan}
                    selectedSpanId={selectedSpan?.span_id}
                  />
                </ScrollArea>
              </CardContent>
            </Card>
          ) : (
            <SpanTimeline
              spans={spans}
              onSpanSelect={setSelectedSpan}
              selectedSpanId={selectedSpan?.span_id}
            />
          )}
        </div>

        {/* Span detail panel (shows when a span is selected) */}
        {selectedSpan && (
          <div className='w-[350px] flex-shrink-0'>
            <SpanDetailPanel span={selectedSpan} onClose={() => setSelectedSpan(null)} />
          </div>
        )}
      </div>
    </div>
  )
}

// ============================================================================
// Scores Tab Content Component
// ============================================================================

interface ScoresTabContentProps {
  projectId: string
  traceId: string
}

function ScoresTabContent({ projectId, traceId }: ScoresTabContentProps) {
  const {
    data: scores,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['traceScores', projectId, traceId],
    queryFn: async () => {
      const response = await getTraceWithScores(projectId, traceId)
      return response.scores || []
    },
    enabled: !!projectId && !!traceId,
    staleTime: 30_000,
  })

  if (isLoading) {
    return (
      <div className='flex items-center justify-center py-8'>
        <div className='flex flex-col items-center space-y-2'>
          <div className='h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent' />
          <p className='text-sm text-muted-foreground'>Loading scores...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className='flex items-center justify-center py-8'>
        <p className='text-sm text-destructive'>Failed to load scores</p>
      </div>
    )
  }

  if (!scores || scores.length === 0) {
    return (
      <div className='flex items-center justify-center py-8 text-center'>
        <div>
          <p className='text-sm text-muted-foreground mb-2'>No scores found for this trace</p>
          <p className='text-xs text-muted-foreground'>
            Quality scores can be added via the SDK or Annotation UI
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className='space-y-3'>
      {scores.map((score) => (
        <Card key={score.id}>
          <CardContent className='p-4'>
            <div className='flex items-center justify-between'>
              <div>
                <div className='font-medium'>{score.name}</div>
                <div className='text-xs text-muted-foreground'>
                  {score.source} • {score.data_type}
                </div>
              </div>
              <div className='text-right'>
                {score.value !== undefined && (
                  <div className='text-lg font-bold'>{score.value}</div>
                )}
                {score.string_value && (
                  <Badge variant='secondary'>{score.string_value}</Badge>
                )}
              </div>
            </div>
            {score.comment && (
              <div className='mt-2 text-sm text-muted-foreground'>{score.comment}</div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

// ============================================================================
// I/O Tab Content Component
// ============================================================================

function IOTabContent({ trace }: { trace: Trace }) {
  if (!trace.input && !trace.output) {
    return (
      <Card>
        <CardContent className='py-8'>
          <div className='text-center text-muted-foreground'>
            <FileInput className='h-12 w-12 mx-auto mb-4 opacity-50' />
            <p>No input/output data available for this trace</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <FileInput className='h-5 w-5' />
          Input / Output
        </CardTitle>
      </CardHeader>
      <CardContent className='space-y-6'>
        {trace.input && (
          <IOPreview
            value={trace.input}
            mimeType='application/json'
            label='Input'
          />
        )}
        {trace.input && trace.output && <Separator />}
        {trace.output && (
          <IOPreview
            value={trace.output}
            mimeType='application/json'
            label='Output'
          />
        )}
      </CardContent>
    </Card>
  )
}

// ============================================================================
// Main TraceDetailView Component
// ============================================================================

export function TraceDetailView({ trace, projectId }: TraceDetailViewProps) {
  const statusStr = statusCodeToString(trace.status_code)
  const status = statuses.find((s) => s.value === statusStr)
  const StatusIcon = status?.icon

  return (
    <div className='space-y-6'>
      {/* Header */}
      <div>
        <div className='flex items-center gap-3 mb-2'>
          <h2 className='text-2xl font-bold'>{trace.name}</h2>
          {StatusIcon && (
            <div className='flex items-center gap-2'>
              <StatusIcon className='h-4 w-4' />
              <span className='text-sm'>{status.label}</span>
            </div>
          )}
        </div>
        <p className='text-sm text-muted-foreground font-mono'>{trace.trace_id}</p>
      </div>

      {/* Metrics Grid */}
      <MetricsGrid trace={trace} />

      {/* Tabbed Content */}
      <Tabs defaultValue='spans' className='space-y-4'>
        <TabsList>
          <TabsTrigger value='spans' className='gap-2'>
            <Layers className='h-4 w-4' />
            Spans ({trace.spanCount})
          </TabsTrigger>
          <TabsTrigger value='scores' className='gap-2'>
            <MessageSquare className='h-4 w-4' />
            Scores
          </TabsTrigger>
          <TabsTrigger value='io' className='gap-2'>
            <FileInput className='h-4 w-4' />
            I/O
          </TabsTrigger>
          <TabsTrigger value='metadata' className='gap-2'>
            <Info className='h-4 w-4' />
            Metadata
          </TabsTrigger>
        </TabsList>

        <TabsContent value='spans'>
          <SpansTabContent
            projectId={projectId}
            traceId={trace.trace_id}
            spanCount={trace.spanCount}
          />
        </TabsContent>

        <TabsContent value='scores'>
          <ScoresTabContent projectId={projectId} traceId={trace.trace_id} />
        </TabsContent>

        <TabsContent value='io'>
          <IOTabContent trace={trace} />
        </TabsContent>

        <TabsContent value='metadata'>
          <TraceMetadataCard trace={trace} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
