import * as React from 'react'
import {
  X,
  Clock,
  DollarSign,
  Hash,
  AlertTriangle,
  Box,
  Cpu,
  Server,
  Radio,
  Copy,
  Check,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import type { Span } from '../api/types'
import {
  safeFormat,
  formatDuration,
  formatCostDetailed,
} from '../utils/format-helpers'
import {
  detectSpanCategory,
  SPAN_CATEGORY_COLORS,
  SPAN_CATEGORY_LABELS,
} from '../utils/span-type-detector'

// ============================================================================
// Types
// ============================================================================

interface SpanDetailPanelProps {
  span: Span
  onClose?: () => void
}

// ============================================================================
// Utility Functions
// ============================================================================

function getSpanKindInfo(kind: number): {
  icon: React.ElementType
  label: string
} {
  switch (kind) {
    case 0:
      return { icon: Box, label: 'Unspecified' }
    case 1:
      return { icon: Cpu, label: 'Internal' }
    case 2:
      return { icon: Server, label: 'Server' }
    case 3:
      return { icon: Radio, label: 'Client' }
    case 4:
      return { icon: Radio, label: 'Producer' }
    case 5:
      return { icon: Radio, label: 'Consumer' }
    default:
      return { icon: Box, label: 'Unknown' }
  }
}

function getStatusInfo(statusCode: number): { label: string; color: string } {
  switch (statusCode) {
    case 0:
      return { label: 'Unset', color: 'text-muted-foreground' }
    case 1:
      return { label: 'OK', color: 'text-green-600 dark:text-green-400' }
    case 2:
      return { label: 'Error', color: 'text-red-600 dark:text-red-400' }
    default:
      return { label: 'Unknown', color: 'text-muted-foreground' }
  }
}

function sumTokens(usage: Record<string, number> | undefined): number {
  if (!usage) return 0
  let total = 0
  for (const v of Object.values(usage)) {
    if (typeof v === 'number' && Number.isFinite(v)) total += v
  }
  return total
}

// ============================================================================
// CopyButton Component
// ============================================================================

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      /* clipboard permission denied — silent */
    }
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-6 w-6"
      onClick={handleCopy}
      aria-label="Copy"
    >
      {copied ? (
        <Check className="h-3 w-3 text-green-500" />
      ) : (
        <Copy className="h-3 w-3" />
      )}
    </Button>
  )
}

// ============================================================================
// CollapsibleSection Component
// ============================================================================

interface CollapsibleSectionProps {
  title: string
  defaultOpen?: boolean
  children: React.ReactNode
}

function CollapsibleSection({
  title,
  defaultOpen = true,
  children,
}: CollapsibleSectionProps) {
  const [isOpen, setIsOpen] = React.useState(defaultOpen)

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger className="flex w-full items-center gap-2 py-2 text-sm font-medium transition-colors hover:text-primary">
        {isOpen ? (
          <ChevronDown className="h-4 w-4" />
        ) : (
          <ChevronRight className="h-4 w-4" />
        )}
        {title}
      </CollapsibleTrigger>
      <CollapsibleContent>{children}</CollapsibleContent>
    </Collapsible>
  )
}

// ============================================================================
// AttributesView Component
// ============================================================================

interface AttributesViewProps {
  attributes: Record<string, string> | undefined
  title: string
  defaultOpen?: boolean
}

function AttributesView({
  attributes,
  title,
  defaultOpen = false,
}: AttributesViewProps) {
  if (!attributes || Object.keys(attributes).length === 0) {
    return null
  }

  return (
    <CollapsibleSection title={title} defaultOpen={defaultOpen}>
      <div className="space-y-1 pl-6">
        {Object.entries(attributes).map(([key, value]) => (
          <div key={key} className="flex items-start gap-2 text-xs">
            <span className="min-w-[100px] truncate font-mono text-muted-foreground">
              {key}:
            </span>
            <span className="break-all font-mono">{String(value)}</span>
          </div>
        ))}
      </div>
    </CollapsibleSection>
  )
}

// ============================================================================
// Main SpanDetailPanel Component
// ============================================================================

export function SpanDetailPanel({ span, onClose }: SpanDetailPanelProps) {
  const { icon: KindIcon, label: kindLabel } = getSpanKindInfo(span.span_kind)
  const { label: statusLabel, color: statusColor } = getStatusInfo(
    span.status_code,
  )

  const totalTokens = sumTokens(span.usage_details)

  // Unified attribute bag for span-type detection.
  const detectorAttrs = React.useMemo<Record<string, unknown>>(() => {
    const bag: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(span.span_attributes ?? {})) bag[k] = v
    for (const [k, v] of Object.entries(span.resource_attributes ?? {})) bag[k] = v
    return bag
  }, [span.span_attributes, span.resource_attributes])

  const category = detectSpanCategory(span.span_name, detectorAttrs)
  const categoryColor = SPAN_CATEGORY_COLORS[category]

  return (
    <Card className="h-full">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="min-w-0 flex-1 space-y-1 pr-2">
            <div className="flex items-center gap-2">
              <Badge
                variant="outline"
                className={cn(
                  'text-[10px] tracking-wider',
                  categoryColor.bg,
                  categoryColor.text,
                  categoryColor.border,
                )}
              >
                {SPAN_CATEGORY_LABELS[category]}
              </Badge>
            </div>
            <CardTitle className="truncate text-base font-semibold">
              {span.span_name}
            </CardTitle>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <KindIcon className="h-3 w-3" />
              <span>{kindLabel}</span>
              <span>•</span>
              <span className={statusColor}>{statusLabel}</span>
              {span.has_error && (
                <AlertTriangle className="h-3 w-3 text-red-500" />
              )}
            </div>
          </div>
          {onClose ? (
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={onClose}
              aria-label="Close"
            >
              <X className="h-4 w-4" />
            </Button>
          ) : null}
        </div>
      </CardHeader>

      <ScrollArea className="h-[calc(100%-80px)]">
        <CardContent className="space-y-4 pt-0">
          {/* IDs Section */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">Span ID</span>
              <div className="flex items-center gap-1">
                <code className="font-mono text-xs">
                  {span.span_id.slice(0, 8)}...
                </code>
                <CopyButton value={span.span_id} />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">Trace ID</span>
              <div className="flex items-center gap-1">
                <code className="font-mono text-xs">
                  {span.trace_id.slice(0, 8)}...
                </code>
                <CopyButton value={span.trace_id} />
              </div>
            </div>
            {span.parent_span_id ? (
              <div className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">Parent ID</span>
                <div className="flex items-center gap-1">
                  <code className="font-mono text-xs">
                    {span.parent_span_id.slice(0, 8)}...
                  </code>
                  <CopyButton value={span.parent_span_id} />
                </div>
              </div>
            ) : null}
          </div>

          <Separator />

          {/* Metrics Grid */}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <div className="flex items-center gap-1 text-xs text-muted-foreground">
                <Clock className="h-3 w-3" />
                Duration
              </div>
              <div className="text-sm font-medium">
                {formatDuration(span.duration)}
              </div>
            </div>

            {span.total_cost != null ? (
              <div className="space-y-1">
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <DollarSign className="h-3 w-3" />
                  Cost
                </div>
                <div className="text-sm font-medium">
                  {formatCostDetailed(span.total_cost)}
                </div>
              </div>
            ) : null}

            {totalTokens > 0 ? (
              <div className="space-y-1">
                <div className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Hash className="h-3 w-3" />
                  Tokens
                </div>
                <div className="text-sm font-medium">
                  {totalTokens.toLocaleString()}
                </div>
              </div>
            ) : null}
          </div>

          {/* AI/Model Info */}
          {span.model_name || span.provider_name ? (
            <>
              <Separator />
              <div className="space-y-2">
                {span.provider_name ? (
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">
                      Provider
                    </span>
                    <Badge variant="outline" className="text-xs">
                      {span.provider_name}
                    </Badge>
                  </div>
                ) : null}
                {span.model_name ? (
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">Model</span>
                    <Badge variant="secondary" className="text-xs">
                      {span.model_name}
                    </Badge>
                  </div>
                ) : null}
              </div>
            </>
          ) : null}

          {/* Timing */}
          <Separator />
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">Start Time</span>
              <span className="text-xs">
                {safeFormat(span.start_time, 'HH:mm:ss.SSS')}
              </span>
            </div>
            {span.end_time ? (
              <div className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">End Time</span>
                <span className="text-xs">
                  {safeFormat(span.end_time, 'HH:mm:ss.SSS')}
                </span>
              </div>
            ) : null}
          </div>

          {/* Status Message */}
          {span.status_message ? (
            <>
              <Separator />
              <div className="space-y-1">
                <div className="text-xs text-muted-foreground">
                  Status Message
                </div>
                <p className={cn('text-xs', statusColor)}>
                  {span.status_message}
                </p>
              </div>
            </>
          ) : null}

          {/* Input/Output */}
          {span.input || span.output ? (
            <>
              <Separator />
              {span.input ? (
                <CollapsibleSection title="Input" defaultOpen={true}>
                  <pre className="max-h-48 overflow-x-auto whitespace-pre-wrap rounded-md bg-muted p-2 text-xs">
                    {span.input}
                  </pre>
                </CollapsibleSection>
              ) : null}
              {span.output ? (
                <CollapsibleSection title="Output" defaultOpen={true}>
                  <pre className="max-h-48 overflow-x-auto whitespace-pre-wrap rounded-md bg-muted p-2 text-xs">
                    {span.output}
                  </pre>
                </CollapsibleSection>
              ) : null}
            </>
          ) : null}

          {/* Attribute bags */}
          <AttributesView
            attributes={span.span_attributes}
            title="Span Attributes"
          />
          <AttributesView
            attributes={span.resource_attributes}
            title="Resource Attributes"
          />
          <AttributesView
            attributes={span.scope_attributes}
            title="Scope Attributes"
          />

          {/* Usage Details */}
          {span.usage_details && Object.keys(span.usage_details).length > 0 ? (
            <CollapsibleSection title="Usage Details" defaultOpen={false}>
              <div className="space-y-1 pl-6">
                {Object.entries(span.usage_details).map(([key, value]) => (
                  <div
                    key={key}
                    className="flex items-center justify-between text-xs"
                  >
                    <span className="text-muted-foreground">{key}</span>
                    <span className="font-mono">
                      {value.toLocaleString()}
                    </span>
                  </div>
                ))}
              </div>
            </CollapsibleSection>
          ) : null}

          {/* Cost Details */}
          {span.cost_details && Object.keys(span.cost_details).length > 0 ? (
            <CollapsibleSection title="Cost Details" defaultOpen={false}>
              <div className="space-y-1 pl-6">
                {Object.entries(span.cost_details).map(([key, value]) => (
                  <div
                    key={key}
                    className="flex items-center justify-between text-xs"
                  >
                    <span className="text-muted-foreground">{key}</span>
                    <span className="font-mono">${value}</span>
                  </div>
                ))}
              </div>
            </CollapsibleSection>
          ) : null}
        </CardContent>
      </ScrollArea>
    </Card>
  )
}
