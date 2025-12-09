'use client'

import * as React from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { AttributesTable } from './attributes-table'
import { EventsList } from './events-list'
import { LinksList } from './links-list'
import type { Trace, Span } from '../../data/schema'

// ============================================================================
// Types
// ============================================================================

interface OtelDataTabsProps {
  /** Span attributes (from span.attributes) */
  spanAttributes?: Record<string, any>
  /** Resource attributes (from trace.resource_attributes) */
  resourceAttributes?: Record<string, any>
  /** Events arrays */
  eventsTimestamp?: Date[]
  eventsName?: string[]
  eventsAttributes?: string[]
  /** Links arrays */
  linksTraceId?: string[]
  linksSpanId?: string[]
  linksAttributes?: string[]
  /** Handler for link navigation */
  onLinkClick?: (traceId: string, spanId?: string) => void
  className?: string
}

// ============================================================================
// Helper: Count badge for tab triggers
// ============================================================================

function CountBadge({ count }: { count: number }) {
  if (count === 0) return null
  return (
    <Badge variant='secondary' className='ml-1.5 h-4 min-w-[1rem] px-1 text-[10px] font-normal'>
      {count}
    </Badge>
  )
}

// ============================================================================
// OtelDataTabs - Main Component
// ============================================================================

/**
 * OtelDataTabs - Container component for displaying all OTEL data
 *
 * This component provides a tabbed interface showing:
 * - SpanAttributes: Key-value table from span_attributes
 * - ResourceAttributes: Key-value table from resource_attributes
 * - Events: OTEL events with timestamps (gen_ai.prompt/completion here)
 * - Links: Span links with trace/span IDs
 *
 * Inspired by OpenLIT's OTEL-native data display pattern.
 */
export function OtelDataTabs({
  spanAttributes,
  resourceAttributes,
  eventsTimestamp,
  eventsName,
  eventsAttributes,
  linksTraceId,
  linksSpanId,
  linksAttributes,
  onLinkClick,
  className,
}: OtelDataTabsProps) {
  // Calculate counts for tab badges
  const spanAttrCount = spanAttributes ? Object.keys(spanAttributes).length : 0
  const resourceAttrCount = resourceAttributes ? Object.keys(resourceAttributes).length : 0
  const eventsCount = eventsName?.length ?? 0
  const linksCount = linksTraceId?.length ?? 0

  // Check if there's any data to display
  const hasAnyData = spanAttrCount > 0 || resourceAttrCount > 0 || eventsCount > 0 || linksCount > 0

  if (!hasAnyData) {
    return (
      <div className={cn('py-6 text-center', className)}>
        <p className='text-sm text-muted-foreground italic'>No OTEL data available</p>
      </div>
    )
  }

  // Determine default tab (first with data)
  const defaultTab = spanAttrCount > 0
    ? 'span-attributes'
    : resourceAttrCount > 0
      ? 'resource-attributes'
      : eventsCount > 0
        ? 'events'
        : 'links'

  return (
    <div className={cn('space-y-2', className)}>
      <Tabs defaultValue={defaultTab} className='w-full'>
        <TabsList className='h-8 w-full grid grid-cols-4 bg-muted/50'>
          <TabsTrigger
            value='span-attributes'
            className='h-7 text-xs px-2 data-[state=active]:bg-background'
            disabled={spanAttrCount === 0}
          >
            <span className='truncate'>Span Attrs</span>
            <CountBadge count={spanAttrCount} />
          </TabsTrigger>
          <TabsTrigger
            value='resource-attributes'
            className='h-7 text-xs px-2 data-[state=active]:bg-background'
            disabled={resourceAttrCount === 0}
          >
            <span className='truncate'>Resource</span>
            <CountBadge count={resourceAttrCount} />
          </TabsTrigger>
          <TabsTrigger
            value='events'
            className='h-7 text-xs px-2 data-[state=active]:bg-background'
            disabled={eventsCount === 0}
          >
            Events
            <CountBadge count={eventsCount} />
          </TabsTrigger>
          <TabsTrigger
            value='links'
            className='h-7 text-xs px-2 data-[state=active]:bg-background'
            disabled={linksCount === 0}
          >
            Links
            <CountBadge count={linksCount} />
          </TabsTrigger>
        </TabsList>

        <TabsContent value='span-attributes' className='mt-3'>
          <AttributesTable
            data={spanAttributes}
            emptyMessage='No span attributes'
          />
        </TabsContent>

        <TabsContent value='resource-attributes' className='mt-3'>
          <AttributesTable
            data={resourceAttributes}
            emptyMessage='No resource attributes'
          />
        </TabsContent>

        <TabsContent value='events' className='mt-3'>
          <EventsList
            timestamps={eventsTimestamp}
            names={eventsName}
            attributes={eventsAttributes}
          />
        </TabsContent>

        <TabsContent value='links' className='mt-3'>
          <LinksList
            traceIds={linksTraceId}
            spanIds={linksSpanId}
            attributes={linksAttributes}
            onLinkClick={onLinkClick}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ============================================================================
// Convenience wrapper for Span data
// ============================================================================

interface OtelDataTabsFromSpanProps {
  span: Span
  trace: Trace
  onLinkClick?: (traceId: string, spanId?: string) => void
  className?: string
}

/**
 * OtelDataTabsFromSpan - Convenience wrapper that extracts data from Span/Trace objects
 */
export function OtelDataTabsFromSpan({
  span,
  trace,
  onLinkClick,
  className,
}: OtelDataTabsFromSpanProps) {
  return (
    <OtelDataTabs
      spanAttributes={span.attributes}
      resourceAttributes={trace.resource_attributes}
      eventsTimestamp={span.events_timestamp}
      eventsName={span.events_name}
      eventsAttributes={span.events_attributes}
      linksTraceId={span.links_trace_id}
      linksSpanId={span.links_span_id}
      linksAttributes={span.links_attributes}
      onLinkClick={onLinkClick}
      className={className}
    />
  )
}

