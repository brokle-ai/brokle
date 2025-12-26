'use client'

import * as React from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import {
  Copy,
  Check,
  AlertTriangle,
  ListTree,
  ChevronDown,
  ChevronRight,
  ArrowLeftRight,
  Download,
  TreeDeciduous,
  Code,
  Star,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Trace, Span } from '../../data/schema'
import { InputOutputSection } from './input-output-section'
import { MetadataBadgeRow } from './metadata-badge-row'
import { PathValueTree } from './path-value-tree'
import { CollapsibleSection } from './collapsible-section'
import { AttributesTable } from './attributes-table'
import { EventsList } from './events-list'
import { LinksList } from './links-list'
import { ScoresTabContent } from './scores-tab-content'
import { useLocalStorage } from '@/hooks/use-local-storage'
import { format } from 'date-fns'

// View mode type for formatted vs JSON display
export type ViewMode = 'formatted' | 'json'

interface DetailPanelProps {
  trace: Trace
  selectedSpan?: Span | null
  spans?: Span[]
  projectId?: string
  className?: string
}

// ============================================================================
// Copy Button Component
// ============================================================================

function CopyButton({ value, className }: { value: string; className?: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant='ghost'
      size='icon'
      className={cn('h-5 w-5 hover:bg-muted', className)}
      onClick={handleCopy}
    >
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3 text-muted-foreground' />
      )}
    </Button>
  )
}

// ============================================================================
// Detail Row Component
// ============================================================================

interface DetailRowProps {
  label: string
  value?: string | React.ReactNode
  copyValue?: string
  className?: string
}

function DetailRow({ label, value, copyValue, className }: DetailRowProps) {
  if (!value && value !== 0) return null

  return (
    <div className={cn('flex items-start justify-between py-2.5', className)}>
      <span className='text-sm text-muted-foreground min-w-[120px] flex-shrink-0'>
        {label}
      </span>
      <div className='flex items-center gap-1.5 text-right'>
        {typeof value === 'string' ? (
          <span className='text-sm font-medium text-foreground font-mono'>
            {value}
          </span>
        ) : (
          value
        )}
        {copyValue && <CopyButton value={copyValue} />}
      </div>
    </div>
  )
}


// ============================================================================
// Span Preview Content - Adaptive metrics based on span type
// ============================================================================

function SpanPreviewContent({
  span,
  trace,
  viewMode,
}: {
  span: Span
  trace: Trace
  viewMode: ViewMode
}) {
  // Calculate counts for badges
  const spanAttrCount = span.attributes ? Object.keys(span.attributes).length : 0
  const resourceAttrCount = trace.resource_attributes ? Object.keys(trace.resource_attributes).length : 0
  const eventsCount = span.events_name?.length ?? 0
  const linksCount = span.links_trace_id?.length ?? 0

  return (
    <div className='space-y-1'>
      {/* Status Message (Error) */}
      {span.status_message && (
        <div className='mb-4'>
          <div className='space-y-2'>
            <span className='text-sm text-muted-foreground'>Status Message</span>
            <div className='p-3 rounded-md bg-destructive/10 border border-destructive/20'>
              <p className='text-sm text-destructive font-mono whitespace-pre-wrap break-all'>
                {span.status_message}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Input Section - Collapsible, default expanded */}
      <InputOutputSection title='Input' content={span.input} viewMode={viewMode} defaultExpanded={true} />

      {/* Output Section - Collapsible, default expanded */}
      <InputOutputSection title='Output' content={span.output} viewMode={viewMode} defaultExpanded={true} />

      {/* Span Attributes - Collapsible, default collapsed */}
      <CollapsibleSection
        title='Span Attributes'
        count={spanAttrCount}
        defaultExpanded={false}
        emptyMessage='No span attributes'
      >
        {spanAttrCount > 0 && (
          <AttributesTable data={span.attributes} emptyMessage='No span attributes' />
        )}
      </CollapsibleSection>

      {/* Resource Attributes - Collapsible, default collapsed */}
      <CollapsibleSection
        title='Resource Attributes'
        count={resourceAttrCount}
        defaultExpanded={false}
        emptyMessage='No resource attributes'
      >
        {resourceAttrCount > 0 && (
          <AttributesTable data={trace.resource_attributes} emptyMessage='No resource attributes' />
        )}
      </CollapsibleSection>

      {/* Events - Collapsible, default collapsed */}
      <CollapsibleSection
        title='Events'
        count={eventsCount}
        defaultExpanded={false}
        emptyMessage='No events'
      >
        {eventsCount > 0 && (
          <EventsList
            timestamps={span.events_timestamp}
            names={span.events_name}
            attributes={span.events_attributes}
          />
        )}
      </CollapsibleSection>

      {/* Links - Collapsible, default collapsed */}
      <CollapsibleSection
        title='Links'
        count={linksCount}
        defaultExpanded={false}
        emptyMessage='No links'
      >
        {linksCount > 0 && (
          <LinksList
            traceIds={span.links_trace_id}
            spanIds={span.links_span_id}
            attributes={span.links_attributes}
          />
        )}
      </CollapsibleSection>
    </div>
  )
}

// ============================================================================
// Span Observation Item - Collapsible span entry for Log View
// ============================================================================

interface SpanObservationItemProps {
  span: Span
  defaultOpen?: boolean
}

function SpanObservationItem({ span, defaultOpen = false }: SpanObservationItemProps) {
  const [isOpen, setIsOpen] = React.useState(defaultOpen)

  // Build observation data object for this span
  const observationData = React.useMemo(() => {
    const data: Record<string, unknown> = {
      id: span.span_id,
      type: span.span_type || 'SPAN',
      name: span.span_name,
      startTime: span.start_time?.toISOString(),
      endTime: span.end_time?.toISOString(),
      latency: span.duration ? Math.round(span.duration / 1_000_000) : undefined, // ms
      level: span.level || 'DEFAULT',
      statusCode: span.status_code,
    }

    // Add input/output if present
    if (span.input) {
      try {
        data.input = typeof span.input === 'string' ? JSON.parse(span.input) : span.input
      } catch {
        data.input = span.input
      }
    }
    if (span.output) {
      try {
        data.output = typeof span.output === 'string' ? JSON.parse(span.output) : span.output
      } catch {
        data.output = span.output
      }
    }

    // Add metadata/attributes if present
    if (span.attributes && Object.keys(span.attributes).length > 0) {
      data.metadata = span.attributes
    }

    // Add usage if present
    if (span.gen_ai_usage_input_tokens || span.gen_ai_usage_output_tokens) {
      data.usage = {
        inputTokens: span.gen_ai_usage_input_tokens,
        outputTokens: span.gen_ai_usage_output_tokens,
        totalTokens: (span.gen_ai_usage_input_tokens || 0) + (span.gen_ai_usage_output_tokens || 0),
      }
    }

    // Add cost if present
    if (span.total_cost) {
      data.cost = span.total_cost
    }

    // Add model info if present
    if (span.gen_ai_request_model) {
      data.model = span.gen_ai_request_model
    }

    return data
  }, [span])

  const itemCount = Object.keys(observationData).filter(k => observationData[k] !== undefined).length

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger asChild>
        <div className='flex items-center gap-2 py-1.5 cursor-pointer hover:bg-muted/50 rounded-md px-2 -mx-2'>
          {isOpen ? (
            <ChevronDown className='h-4 w-4 text-muted-foreground flex-shrink-0' />
          ) : (
            <ChevronRight className='h-4 w-4 text-muted-foreground flex-shrink-0' />
          )}
          <ArrowLeftRight className='h-3.5 w-3.5 text-muted-foreground' />
          <span className='text-sm font-medium truncate'>{span.span_name}</span>
          <span className='text-xs text-muted-foreground font-mono'>
            ({span.span_id.substring(0, 8)})
          </span>
          <Badge variant='secondary' className='text-xs ml-auto'>
            {itemCount} items
          </Badge>
        </div>
      </CollapsibleTrigger>

      <CollapsibleContent>
        <div className='ml-4 py-2 border-l border-border/50 pl-3'>
          <PathValueTree
            data={observationData}
            maxInitialDepth={1}
            showCopyButtons={true}
          />
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

// ============================================================================
// Performance Warning Thresholds
// ============================================================================

const LOG_VIEW_WARNING_THRESHOLD = 150 // Show warning at this count
const LOG_VIEW_DISABLE_THRESHOLD = 350 // Disable Log View at this count

// ============================================================================
// Performance Warning Banner
// ============================================================================

function PerformanceWarning({
  spanCount,
  onDismiss,
}: {
  spanCount: number
  onDismiss: () => void
}) {
  return (
    <div className='bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3 mb-4'>
      <div className='flex items-start gap-3'>
        <AlertTriangle className='h-5 w-5 text-yellow-600 dark:text-yellow-400 flex-shrink-0 mt-0.5' />
        <div className='flex-1 space-y-1'>
          <p className='text-sm font-medium text-yellow-800 dark:text-yellow-200'>
            Performance Warning
          </p>
          <p className='text-xs text-yellow-700 dark:text-yellow-300'>
            This trace has {spanCount.toLocaleString()} spans. Loading the full Log View may
            cause performance issues. Consider using the Preview tab or filtering spans in the
            tree view.
          </p>
        </div>
        <Button
          variant='ghost'
          size='sm'
          className='h-6 px-2 text-xs text-yellow-700 hover:text-yellow-900 dark:text-yellow-300'
          onClick={onDismiss}
        >
          Dismiss
        </Button>
      </div>
    </div>
  )
}

// ============================================================================
// Log View Content - Concatenated Observation Log
// ============================================================================

function LogViewContent({
  trace,
  spans,
  viewMode,
}: {
  trace: Trace
  spans: Span[]
  viewMode: ViewMode
}) {
  const [copied, setCopied] = React.useState(false)
  const [warningDismissed, setWarningDismissed] = React.useState(false)

  // Use passed spans (fetched separately) or fallback to trace.spans
  const allSpans = spans.length > 0 ? spans : (trace.spans || [])

  // Show performance warning for large traces
  const showWarning = allSpans.length >= LOG_VIEW_WARNING_THRESHOLD && !warningDismissed

  // Build full JSON for copy/download
  const fullJson = React.useMemo(() => {
    try {
      const observations = allSpans.map(span => ({
        id: span.span_id,
        type: span.span_type || 'SPAN',
        name: span.span_name,
        startTime: span.start_time?.toISOString(),
        endTime: span.end_time?.toISOString(),
        latency: span.duration ? Math.round(span.duration / 1_000_000) : undefined,
        level: span.level || 'DEFAULT',
        input: span.input,
        output: span.output,
        metadata: span.attributes,
      }))
      return JSON.stringify(observations, null, 2)
    } catch {
      return '[]'
    }
  }, [allSpans])

  const handleCopy = async () => {
    await navigator.clipboard.writeText(fullJson)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDownload = () => {
    const blob = new Blob([fullJson], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `trace-${trace.trace_id.substring(0, 8)}-observations.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  // JSON View - raw JSON display
  if (viewMode === 'json') {
    return (
      <div className='space-y-3'>
        {/* Performance Warning */}
        {showWarning && (
          <PerformanceWarning
            spanCount={allSpans.length}
            onDismiss={() => setWarningDismissed(true)}
          />
        )}

        {/* Header */}
        <div className='flex items-center justify-between'>
          <span className='text-sm font-medium'>Concatenated Observation Log</span>
          <div className='flex items-center gap-1'>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7'
              onClick={handleDownload}
              title='Download JSON'
            >
              <Download className='h-3.5 w-3.5' />
            </Button>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7'
              onClick={handleCopy}
              title='Copy all'
            >
              {copied ? (
                <Check className='h-3.5 w-3.5 text-green-500' />
              ) : (
                <Copy className='h-3.5 w-3.5' />
              )}
            </Button>
          </div>
        </div>

        {/* Raw JSON View */}
        <pre className='bg-muted/50 rounded-md p-3 text-xs font-mono overflow-x-auto max-h-[600px] whitespace-pre-wrap break-words'>
          {fullJson}
        </pre>
      </div>
    )
  }

  // Formatted View - tree structure
  return (
    <div className='space-y-3'>
      {/* Performance Warning */}
      {showWarning && (
        <PerformanceWarning
          spanCount={allSpans.length}
          onDismiss={() => setWarningDismissed(true)}
        />
      )}

      {/* Header */}
      <div className='flex items-center justify-between'>
        <span className='text-sm font-medium'>Concatenated Observation Log</span>
        <div className='flex items-center gap-1'>
          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7'
            onClick={handleDownload}
            title='Download JSON'
          >
            <Download className='h-3.5 w-3.5' />
          </Button>
          <Button
            variant='ghost'
            size='icon'
            className='h-7 w-7'
            onClick={handleCopy}
            title='Copy all'
          >
            {copied ? (
              <Check className='h-3.5 w-3.5 text-green-500' />
            ) : (
              <Copy className='h-3.5 w-3.5' />
            )}
          </Button>
        </div>
      </div>

      {/* Path | Value header */}
      <div className='grid grid-cols-2 gap-2 text-xs text-muted-foreground border-b pb-1.5 px-2'>
        <span>Path</span>
        <span>Value</span>
      </div>

      {/* Span observations */}
      {allSpans.length === 0 ? (
        <div className='text-sm text-muted-foreground italic py-4 text-center'>
          No observations in this trace
        </div>
      ) : (
        <div className='space-y-1'>
          {allSpans.map((span, index) => (
            <SpanObservationItem
              key={span.span_id}
              span={span}
              defaultOpen={index === 0}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// ============================================================================
// Detail Panel Header - Shows name, ID, timestamp, badges above tabs
// ============================================================================

function DetailPanelHeader({ trace, selectedSpan }: { trace: Trace; selectedSpan?: Span | null }) {
  const item = selectedSpan || trace
  const name = selectedSpan ? selectedSpan.span_name : trace.name
  const id = selectedSpan ? selectedSpan.span_id : trace.trace_id
  const startTime = selectedSpan ? selectedSpan.start_time : trace.start_time

  const formatDateTime = (date: Date | undefined) => {
    if (!date) return '-'
    return format(date, 'yyyy-MM-dd HH:mm:ss.SSS')
  }

  return (
    <div className='px-4 pt-4 pb-2 space-y-2'>
      {/* Name with icon and copy ID button */}
      <div className='flex items-center gap-2'>
        <ListTree className='h-4 w-4 text-muted-foreground flex-shrink-0' />
        <span className='text-sm font-medium truncate' title={name}>
          {name}
        </span>
        <CopyButton value={id} />
        <span className='text-xs text-muted-foreground'>ID</span>
        {selectedSpan?.span_type && (
          <Badge variant='outline' className='text-xs font-normal'>
            {selectedSpan.span_type}
          </Badge>
        )}
        {(selectedSpan?.has_error || (!selectedSpan && trace.has_error)) && (
          <Badge variant='destructive' className='text-xs gap-1'>
            <AlertTriangle className='h-3 w-3' />
            Error
          </Badge>
        )}
      </div>

      {/* Timestamp */}
      <p className='text-xs text-muted-foreground'>
        {formatDateTime(startTime)}
      </p>

      {/* Badges row */}
      <MetadataBadgeRow
        environment={trace.environment}
        duration={selectedSpan ? selectedSpan.duration : trace.duration}
        cost={selectedSpan ? selectedSpan.total_cost : trace.cost}
        inputTokens={selectedSpan?.gen_ai_usage_input_tokens}
        outputTokens={selectedSpan?.gen_ai_usage_output_tokens}
        totalTokens={!selectedSpan ? trace.tokens : undefined}
        version={selectedSpan?.version || trace.service_version}
        modelName={selectedSpan?.gen_ai_request_model || trace.model_name}
        providerName={selectedSpan?.gen_ai_provider_name || trace.provider_name}
        level={selectedSpan?.level}
      />
    </div>
  )
}

// ============================================================================
// View Mode Toggle Component
// ============================================================================

interface ViewModeToggleProps {
  value: ViewMode
  onValueChange: (value: ViewMode) => void
}

function ViewModeToggle({ value, onValueChange }: ViewModeToggleProps) {
  return (
    <ToggleGroup
      type='single'
      value={value}
      onValueChange={(v: string) => v && onValueChange(v as ViewMode)}
      className='h-7'
    >
      <ToggleGroupItem
        value='formatted'
        className='h-7 px-2 text-xs gap-1'
        title='Formatted tree view'
      >
        <TreeDeciduous className='h-3 w-3' />
        <span className='hidden sm:inline'>Formatted</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value='json'
        className='h-7 px-2 text-xs gap-1'
        title='Raw JSON view'
      >
        <Code className='h-3 w-3' />
        <span className='hidden sm:inline'>JSON</span>
      </ToggleGroupItem>
    </ToggleGroup>
  )
}

// ============================================================================
// Main Detail Panel Export
// ============================================================================

/**
 * DetailPanel - Right panel showing trace/span details
 * Structured layout with header above tabs
 *
 * Tab behavior:
 * - At trace level (no span selected): Show Preview + Log View tabs
 * - At span level (span selected): Show only Preview tab
 * - Log View disabled at 350+ spans for performance
 */
export function DetailPanel({
  trace,
  selectedSpan,
  spans = [],
  projectId,
  className,
}: DetailPanelProps) {
  // Persist view mode preference across sessions
  const [viewMode, setViewMode] = useLocalStorage<ViewMode>('jsonViewPreference', 'formatted')

  // Calculate span count for Log View tab logic
  const spanCount = spans.length > 0 ? spans.length : (trace.spans?.length || 0)

  // Determine if Log View tab should be shown (when trace has spans)
  const showLogViewTab = spanCount > 0

  // Disable Log View for very large traces
  const logViewDisabled = spanCount >= LOG_VIEW_DISABLE_THRESHOLD

  return (
    <div className={cn('flex flex-col h-full overflow-hidden', className)}>
      {/* Header with name, ID, timestamp, badges */}
      <DetailPanelHeader trace={trace} selectedSpan={selectedSpan} />

      {/* Tabs below header */}
      <Tabs defaultValue='preview' className='flex flex-col flex-1 min-h-0 overflow-hidden'>
        {/* Tab Headers with View Mode Toggle - fixed height */}
        <div className='border-b px-4 flex-shrink-0'>
          <div className='flex items-center justify-between'>
            <TabsList className='h-8'>
              <TabsTrigger value='preview' className='h-7 px-3 text-xs'>
                Preview
              </TabsTrigger>
              {showLogViewTab && (
                <TabsTrigger
                  value='log'
                  className='h-7 px-3 text-xs'
                  disabled={logViewDisabled}
                  title={logViewDisabled ? `Disabled for traces with ${LOG_VIEW_DISABLE_THRESHOLD}+ spans` : undefined}
                >
                  Log View
                  {logViewDisabled && (
                    <AlertTriangle className='h-3 w-3 ml-1 text-yellow-500' />
                  )}
                </TabsTrigger>
              )}
              <TabsTrigger value='scores' className='h-7 px-3 text-xs'>
                <Star className='h-3 w-3 mr-1' />
                Scores
              </TabsTrigger>
            </TabsList>

            {/* Formatted/JSON Toggle */}
            <ViewModeToggle value={viewMode} onValueChange={setViewMode} />
          </div>
        </div>

        {/* Tab Content - scrollable individually */}
        <TabsContent value='preview' className='flex-1 min-h-0 overflow-y-auto p-4 mt-0'>
          {selectedSpan ? (
            <SpanPreviewContent span={selectedSpan} trace={trace} viewMode={viewMode} />
          ) : (
            // No span selected (should rarely happen with auto-select) - show empty state
            <div className='flex flex-col items-center justify-center py-12 text-center'>
              <ListTree className='h-12 w-12 text-muted-foreground/50 mb-4' />
              <p className='text-sm text-muted-foreground'>
                No spans in this trace
              </p>
              <p className='text-xs text-muted-foreground/70 mt-1'>
                Select a span from the tree to view details
              </p>
            </div>
          )}
        </TabsContent>

        {showLogViewTab && !logViewDisabled && (
          <TabsContent value='log' className='flex-1 min-h-0 overflow-y-auto p-4 mt-0'>
            <LogViewContent trace={trace} spans={spans} viewMode={viewMode} />
          </TabsContent>
        )}

        <TabsContent value='scores' className='flex-1 min-h-0 overflow-y-auto p-4 mt-0'>
          {projectId ? (
            <ScoresTabContent
              projectId={projectId}
              traceId={trace.trace_id}
              spans={spans}
            />
          ) : (
            <div className='text-sm text-muted-foreground text-center py-8'>
              Project context required to load scores
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
