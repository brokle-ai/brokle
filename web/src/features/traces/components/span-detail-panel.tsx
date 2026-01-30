'use client'

import * as React from 'react'
import { X, Clock, DollarSign, Hash, AlertTriangle, Box, Cpu, Server, Radio, Copy, Check, ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import type { Span } from '../data/schema'
import { safeFormat, formatDuration, formatCostDetailed } from '../utils/format-helpers'

// ============================================================================
// Types
// ============================================================================

interface SpanDetailPanelProps {
  span: Span
  onClose: () => void
}

// ============================================================================
// Utility Functions
// ============================================================================

function getSpanKindInfo(kind: number): { icon: React.ElementType; label: string } {
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

// ============================================================================
// CopyButton Component
// ============================================================================

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button variant='ghost' size='icon' className='h-6 w-6' onClick={handleCopy}>
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3' />
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

function CollapsibleSection({ title, defaultOpen = true, children }: CollapsibleSectionProps) {
  const [isOpen, setIsOpen] = React.useState(defaultOpen)

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger className='flex items-center gap-2 w-full py-2 text-sm font-medium hover:text-primary transition-colors'>
        {isOpen ? (
          <ChevronDown className='h-4 w-4' />
        ) : (
          <ChevronRight className='h-4 w-4' />
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
  attributes: Record<string, any> | undefined
  title: string
}

function AttributesView({ attributes, title }: AttributesViewProps) {
  if (!attributes || Object.keys(attributes).length === 0) {
    return null
  }

  return (
    <CollapsibleSection title={title} defaultOpen={false}>
      <div className='space-y-1 pl-6'>
        {Object.entries(attributes).map(([key, value]) => (
          <div key={key} className='flex items-start gap-2 text-xs'>
            <span className='font-mono text-muted-foreground min-w-[100px] truncate'>{key}:</span>
            <span className='font-mono break-all'>
              {typeof value === 'object' ? JSON.stringify(value) : String(value)}
            </span>
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
  const { label: statusLabel, color: statusColor } = getStatusInfo(span.status_code)

  const totalTokens = (span.gen_ai_usage_input_tokens || 0) + (span.gen_ai_usage_output_tokens || 0)

  return (
    <Card className='h-full'>
      <CardHeader className='pb-3'>
        <div className='flex items-start justify-between'>
          <div className='space-y-1 min-w-0 flex-1 pr-2'>
            <CardTitle className='text-base font-semibold truncate'>{span.span_name}</CardTitle>
            <div className='flex items-center gap-2 text-xs text-muted-foreground'>
              <KindIcon className='h-3 w-3' />
              <span>{kindLabel}</span>
              <span>•</span>
              <span className={statusColor}>{statusLabel}</span>
              {span.has_error && <AlertTriangle className='h-3 w-3 text-red-500' />}
            </div>
          </div>
          <Button variant='ghost' size='icon' className='h-8 w-8' onClick={onClose}>
            <X className='h-4 w-4' />
          </Button>
        </div>
      </CardHeader>

      <ScrollArea className='h-[calc(100%-80px)]'>
        <CardContent className='space-y-4 pt-0'>
          {/* IDs Section */}
          <div className='space-y-2'>
            <div className='flex items-center justify-between'>
              <span className='text-xs text-muted-foreground'>Span ID</span>
              <div className='flex items-center gap-1'>
                <code className='text-xs font-mono'>{span.span_id.slice(0, 8)}...</code>
                <CopyButton value={span.span_id} />
              </div>
            </div>
            <div className='flex items-center justify-between'>
              <span className='text-xs text-muted-foreground'>Trace ID</span>
              <div className='flex items-center gap-1'>
                <code className='text-xs font-mono'>{span.trace_id.slice(0, 8)}...</code>
                <CopyButton value={span.trace_id} />
              </div>
            </div>
            {span.parent_span_id && (
              <div className='flex items-center justify-between'>
                <span className='text-xs text-muted-foreground'>Parent ID</span>
                <div className='flex items-center gap-1'>
                  <code className='text-xs font-mono'>{span.parent_span_id.slice(0, 8)}...</code>
                  <CopyButton value={span.parent_span_id} />
                </div>
              </div>
            )}
          </div>

          <Separator />

          {/* Metrics Grid */}
          <div className='grid grid-cols-2 gap-3'>
            <div className='space-y-1'>
              <div className='flex items-center gap-1 text-xs text-muted-foreground'>
                <Clock className='h-3 w-3' />
                Duration
              </div>
              <div className='text-sm font-medium'>{formatDuration(span.duration)}</div>
            </div>

            {span.total_cost != null && (
              <div className='space-y-1'>
                <div className='flex items-center gap-1 text-xs text-muted-foreground'>
                  <DollarSign className='h-3 w-3' />
                  Cost
                </div>
                <div className='text-sm font-medium'>{formatCostDetailed(span.total_cost)}</div>
              </div>
            )}

            {totalTokens > 0 && (
              <div className='space-y-1'>
                <div className='flex items-center gap-1 text-xs text-muted-foreground'>
                  <Hash className='h-3 w-3' />
                  Tokens
                </div>
                <div className='text-sm font-medium'>{totalTokens.toLocaleString()}</div>
              </div>
            )}
          </div>

          {/* AI/Model Info */}
          {(span.model_name || span.provider_name) && (
            <>
              <Separator />
              <div className='space-y-2'>
                {span.provider_name && (
                  <div className='flex items-center justify-between'>
                    <span className='text-xs text-muted-foreground'>Provider</span>
                    <Badge variant='outline' className='text-xs'>{span.provider_name}</Badge>
                  </div>
                )}
                {span.model_name && (
                  <div className='flex items-center justify-between'>
                    <span className='text-xs text-muted-foreground'>Model</span>
                    <Badge variant='secondary' className='text-xs'>{span.model_name}</Badge>
                  </div>
                )}
                {span.gen_ai_usage_input_tokens != null && (
                  <div className='flex items-center justify-between'>
                    <span className='text-xs text-muted-foreground'>Input Tokens</span>
                    <span className='text-xs'>{span.gen_ai_usage_input_tokens.toLocaleString()}</span>
                  </div>
                )}
                {span.gen_ai_usage_output_tokens != null && (
                  <div className='flex items-center justify-between'>
                    <span className='text-xs text-muted-foreground'>Output Tokens</span>
                    <span className='text-xs'>{span.gen_ai_usage_output_tokens.toLocaleString()}</span>
                  </div>
                )}
              </div>
            </>
          )}

          {/* Timing */}
          <Separator />
          <div className='space-y-2'>
            <div className='flex items-center justify-between'>
              <span className='text-xs text-muted-foreground'>Start Time</span>
              <span className='text-xs'>{safeFormat(span.start_time, 'HH:mm:ss.SSS')}</span>
            </div>
            {span.end_time && (
              <div className='flex items-center justify-between'>
                <span className='text-xs text-muted-foreground'>End Time</span>
                <span className='text-xs'>{safeFormat(span.end_time, 'HH:mm:ss.SSS')}</span>
              </div>
            )}
          </div>

          {/* Status Message */}
          {span.status_message && (
            <>
              <Separator />
              <div className='space-y-1'>
                <div className='text-xs text-muted-foreground'>Status Message</div>
                <p className={cn('text-xs', statusColor)}>{span.status_message}</p>
              </div>
            </>
          )}

          {/* Input/Output */}
          {(span.input || span.output) && (
            <>
              <Separator />
              {span.input && (
                <CollapsibleSection title='Input' defaultOpen={true}>
                  <pre className='text-xs bg-muted p-2 rounded-md overflow-x-auto max-h-32 whitespace-pre-wrap'>
                    {span.input}
                  </pre>
                </CollapsibleSection>
              )}
              {span.output && (
                <CollapsibleSection title='Output' defaultOpen={true}>
                  <pre className='text-xs bg-muted p-2 rounded-md overflow-x-auto max-h-32 whitespace-pre-wrap'>
                    {span.output}
                  </pre>
                </CollapsibleSection>
              )}
            </>
          )}

          {/* Attributes */}
          <AttributesView attributes={span.attributes} title='Span Attributes' />
          <AttributesView attributes={span.metadata} title='Metadata' />

          {/* Usage Details */}
          {span.usage_details && Object.keys(span.usage_details).length > 0 && (
            <CollapsibleSection title='Usage Details' defaultOpen={false}>
              <div className='space-y-1 pl-6'>
                {Object.entries(span.usage_details).map(([key, value]) => (
                  <div key={key} className='flex items-center justify-between text-xs'>
                    <span className='text-muted-foreground'>{key}</span>
                    <span className='font-mono'>{value.toLocaleString()}</span>
                  </div>
                ))}
              </div>
            </CollapsibleSection>
          )}

          {/* Cost Details */}
          {span.cost_details && Object.keys(span.cost_details).length > 0 && (
            <CollapsibleSection title='Cost Details' defaultOpen={false}>
              <div className='space-y-1 pl-6'>
                {Object.entries(span.cost_details).map(([key, value]) => (
                  <div key={key} className='flex items-center justify-between text-xs'>
                    <span className='text-muted-foreground'>{key}</span>
                    <span className='font-mono'>${value}</span>
                  </div>
                ))}
              </div>
            </CollapsibleSection>
          )}
        </CardContent>
      </ScrollArea>
    </Card>
  )
}
