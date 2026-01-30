'use client'

import { memo } from 'react'
import { MessageSquare, ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { SessionGroup } from '../../hooks/use-session-grouping'

/**
 * Props for the SessionTimeline component
 */
interface SessionTimelineProps {
  /** The session group containing multiple traces/turns */
  session: SessionGroup
  /** Currently selected trace ID */
  currentTraceId?: string
  /** Callback when a trace is selected */
  onTraceSelect: (traceId: string) => void
  /** Optional className for styling */
  className?: string
}

/**
 * Format timestamp for tooltip display
 */
function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

/**
 * SessionTimeline - Displays multi-turn conversation navigation
 *
 * Shows numbered buttons for each turn in a session, allowing users to
 * quickly navigate between traces in the same conversation/session.
 *
 * Features:
 * - Numbered turn indicators with selection state
 * - Tooltips showing turn time
 * - Previous/Next navigation buttons
 * - Session metadata display
 */
function SessionTimelineComponent({
  session,
  currentTraceId,
  onTraceSelect,
  className,
}: SessionTimelineProps) {
  // Find current turn index
  const currentIndex = session.traces.findIndex(
    (t) => t.trace_id === currentTraceId
  )
  const hasPrev = currentIndex > 0
  const hasNext = currentIndex < session.traces.length - 1

  // Navigation handlers
  const goToPrev = () => {
    if (hasPrev) {
      onTraceSelect(session.traces[currentIndex - 1].trace_id)
    }
  }

  const goToNext = () => {
    if (hasNext) {
      onTraceSelect(session.traces[currentIndex + 1].trace_id)
    }
  }

  // Don't render if only one turn (no multi-turn navigation needed)
  if (session.turns <= 1) {
    return null
  }

  return (
    <div
      className={cn(
        'flex items-center gap-2 px-3 py-2 border-b bg-muted/30',
        className
      )}
    >
      {/* Session icon and label */}
      <div className="flex items-center gap-1.5 text-muted-foreground">
        <MessageSquare className="h-3.5 w-3.5" />
        <span className="text-xs font-medium">Session</span>
      </div>

      {/* Previous button */}
      <Button
        variant="ghost"
        size="icon"
        className="h-6 w-6"
        onClick={goToPrev}
        disabled={!hasPrev}
        title="Previous turn"
      >
        <ChevronLeft className="h-3.5 w-3.5" />
      </Button>

      {/* Turn buttons */}
      <div className="flex items-center gap-1">
        {session.traces.map((trace, index) => {
          const isSelected = trace.trace_id === currentTraceId
          const turnNumber = index + 1

          return (
            <Tooltip key={trace.trace_id}>
              <TooltipTrigger asChild>
                <button
                  onClick={() => onTraceSelect(trace.trace_id)}
                  className={cn(
                    'w-6 h-6 rounded-full text-xs font-medium',
                    'flex items-center justify-center',
                    'transition-colors duration-150',
                    isSelected
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted hover:bg-muted-foreground/20 text-muted-foreground'
                  )}
                >
                  {turnNumber}
                </button>
              </TooltipTrigger>
              <TooltipContent side="bottom" className="text-xs">
                <div className="space-y-0.5">
                  <p className="font-medium">Turn {turnNumber}</p>
                  <p className="text-muted-foreground">
                    {formatTime(trace.start_time.toString())}
                  </p>
                </div>
              </TooltipContent>
            </Tooltip>
          )
        })}
      </div>

      {/* Next button */}
      <Button
        variant="ghost"
        size="icon"
        className="h-6 w-6"
        onClick={goToNext}
        disabled={!hasNext}
        title="Next turn"
      >
        <ChevronRight className="h-3.5 w-3.5" />
      </Button>

      {/* Turn count */}
      <span className="text-xs text-muted-foreground ml-1">
        {currentIndex + 1} of {session.turns} turn{session.turns > 1 ? 's' : ''}
      </span>
    </div>
  )
}

// Memoize for performance
export const SessionTimeline = memo(SessionTimelineComponent)
