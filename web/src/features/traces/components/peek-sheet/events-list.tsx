'use client'

import * as React from 'react'
import { format } from 'date-fns'
import { Zap, ChevronDown, ChevronRight, Copy, Check, MessageSquare } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'
import { ChatMLView, extractChatML } from './chatml-view'

// ============================================================================
// Types
// ============================================================================

interface EventsListProps {
  timestamps?: Date[]
  names?: string[]
  attributes?: string[]
  className?: string
}

interface ParsedEvent {
  index: number
  name: string
  timestamp?: Date
  attributes: Record<string, any>
}

// ============================================================================
// Copy Button Component
// ============================================================================

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation()
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant='ghost'
      size='icon'
      className='h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity'
      onClick={handleCopy}
      title='Copy'
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
// Event Card - Expandable card for each event
// ============================================================================

interface EventCardProps {
  event: ParsedEvent
  defaultExpanded?: boolean
}

function EventCard({ event, defaultExpanded = false }: EventCardProps) {
  const [isExpanded, setIsExpanded] = React.useState(defaultExpanded)

  const attrCount = Object.keys(event.attributes).length

  // Check if this is a gen_ai.prompt or gen_ai.completion event (OTEL GenAI convention)
  const isGenAIEvent = event.name === 'gen_ai.prompt' || event.name === 'gen_ai.completion'

  // Try to extract ChatML messages from attributes
  const chatMessages = React.useMemo(() => {
    if (!isGenAIEvent) return null

    // Check for messages in attributes
    const messages = event.attributes['gen_ai.prompt'] || event.attributes['gen_ai.completion']
    if (messages) {
      return extractChatML(messages)
    }

    // Check if the whole attributes object looks like ChatML
    return extractChatML(event.attributes)
  }, [event.attributes, isGenAIEvent])

  const eventJson = React.useMemo(() => {
    return JSON.stringify({
      name: event.name,
      timestamp: event.timestamp?.toISOString(),
      attributes: event.attributes,
    }, null, 2)
  }, [event])

  // Color coding for special events
  const getEventColor = (name: string) => {
    if (name === 'gen_ai.prompt') return 'bg-blue-50 dark:bg-blue-900/30 border-blue-200 dark:border-blue-800'
    if (name === 'gen_ai.completion') return 'bg-purple-50 dark:bg-purple-900/30 border-purple-200 dark:border-purple-800'
    if (name.startsWith('exception')) return 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-800'
    return 'bg-muted/30 border-border'
  }

  return (
    <div className={cn('rounded-lg border', getEventColor(event.name))}>
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <CollapsibleTrigger asChild>
          <div className='group flex items-center gap-2 p-3 cursor-pointer hover:bg-muted/20 transition-colors'>
            {isExpanded ? (
              <ChevronDown className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            ) : (
              <ChevronRight className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            )}

            {isGenAIEvent ? (
              <MessageSquare className='h-4 w-4 text-purple-500 flex-shrink-0' />
            ) : (
              <Zap className='h-4 w-4 text-yellow-500 flex-shrink-0' />
            )}

            <div className='flex-1 min-w-0'>
              <div className='flex items-center gap-2'>
                <span className='text-sm font-medium truncate'>{event.name}</span>
                <Badge variant='outline' className='text-xs font-mono flex-shrink-0'>
                  [{event.index}]
                </Badge>
              </div>
              {event.timestamp && (
                <p className='text-xs text-muted-foreground mt-0.5'>
                  {format(event.timestamp, 'HH:mm:ss.SSS')}
                </p>
              )}
            </div>

            {attrCount > 0 && (
              <Badge variant='secondary' className='text-xs flex-shrink-0'>
                {attrCount} attr{attrCount !== 1 ? 's' : ''}
              </Badge>
            )}

            <CopyButton value={eventJson} />
          </div>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <div className='px-3 pb-3 pt-1 border-t border-border/50'>
            {/* If it's a GenAI event with ChatML messages, render them nicely */}
            {chatMessages && chatMessages.length > 0 ? (
              <div className='mt-2'>
                <ChatMLView messages={chatMessages} collapseAfter={5} />
              </div>
            ) : attrCount > 0 ? (
              <div className='space-y-1.5 mt-2'>
                {Object.entries(event.attributes).map(([key, value]) => (
                  <div key={key} className='flex items-start gap-2'>
                    <span className='text-xs text-muted-foreground font-mono min-w-[120px] flex-shrink-0'>
                      {key}:
                    </span>
                    <span className='text-xs font-mono text-foreground break-all'>
                      {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className='text-xs text-muted-foreground italic mt-2'>
                No attributes
              </p>
            )}
          </div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

// ============================================================================
// EventsList - Main Component
// ============================================================================

/**
 * EventsList - Display OTEL span events
 *
 * Features:
 * - List of events with timestamp and name
 * - Expandable to show event attributes
 * - Special rendering for gen_ai.prompt/gen_ai.completion (ChatML format)
 * - Color-coded by event type
 */
export function EventsList({
  timestamps,
  names,
  attributes,
  className,
}: EventsListProps) {
  // Parse events into structured format
  const events: ParsedEvent[] = React.useMemo(() => {
    if (!names || names.length === 0) return []

    return names.map((name, index) => {
      let parsedAttrs: Record<string, any> = {}
      if (attributes && attributes[index]) {
        try {
          parsedAttrs = JSON.parse(attributes[index])
        } catch {
          // Keep empty if parse fails
        }
      }

      return {
        index,
        name,
        timestamp: timestamps?.[index],
        attributes: parsedAttrs,
      }
    })
  }, [timestamps, names, attributes])

  if (events.length === 0) {
    return (
      <div className={cn('py-6 text-center', className)}>
        <p className='text-sm text-muted-foreground italic'>No events</p>
      </div>
    )
  }

  return (
    <div className={cn('space-y-2', className)}>
      {/* Summary header */}
      <div className='flex items-center gap-2 py-1'>
        <Zap className='h-4 w-4 text-yellow-500' />
        <span className='text-sm font-medium'>{events.length} Event{events.length !== 1 ? 's' : ''}</span>
      </div>

      {/* Event cards */}
      <div className='space-y-2'>
        {events.map((event) => (
          <EventCard
            key={`${event.name}-${event.index}`}
            event={event}
            defaultExpanded={events.length <= 3}
          />
        ))}
      </div>
    </div>
  )
}
