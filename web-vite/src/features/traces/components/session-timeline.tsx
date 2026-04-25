import { memo } from 'react'
import { ChevronLeft, ChevronRight, MessageSquare } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { SessionGroup } from '../hooks/use-session-grouping'

interface SessionTimelineProps {
  /** The session group containing multiple traces / turns. */
  session: SessionGroup
  /** Currently selected trace ID. */
  currentTraceId?: string
  /** Callback when a trace is selected from the timeline. */
  onTraceSelect: (traceId: string) => void
  className?: string
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

/**
 * SessionTimeline — multi-turn conversation navigator. Renders a
 * compact strip of numbered buttons (one per trace in the session)
 * with prev/next arrows so a user can hop between turns of the same
 * agent conversation without leaving the detail view.
 *
 * The component renders nothing for single-turn sessions (no
 * navigation surface needed).
 */
function SessionTimelineComponent({
  session,
  currentTraceId,
  onTraceSelect,
  className,
}: SessionTimelineProps) {
  const currentIndex = session.traces.findIndex(
    (t) => t.trace_id === currentTraceId,
  )
  const hasPrev = currentIndex > 0
  const hasNext = currentIndex < session.traces.length - 1

  const goToPrev = () => {
    if (hasPrev) onTraceSelect(session.traces[currentIndex - 1].trace_id)
  }
  const goToNext = () => {
    if (hasNext) onTraceSelect(session.traces[currentIndex + 1].trace_id)
  }

  if (session.turns <= 1) return null

  return (
    <TooltipProvider>
      <div
        className={cn(
          'flex items-center gap-2 border-b bg-muted/30 px-3 py-2',
          className,
        )}
      >
        <div className="flex items-center gap-1.5 text-muted-foreground">
          <MessageSquare className="h-3.5 w-3.5" />
          <span className="text-xs font-medium">Session</span>
        </div>

        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6"
          onClick={goToPrev}
          disabled={!hasPrev}
          aria-label="Previous turn"
          title="Previous turn"
        >
          <ChevronLeft className="h-3.5 w-3.5" />
        </Button>

        <div className="flex items-center gap-1">
          {session.traces.map((trace, index) => {
            const isSelected = trace.trace_id === currentTraceId
            const turnNumber = index + 1
            return (
              <Tooltip key={trace.trace_id}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    onClick={() => onTraceSelect(trace.trace_id)}
                    className={cn(
                      'flex h-6 w-6 items-center justify-center rounded-full text-xs font-medium transition-colors',
                      isSelected
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted text-muted-foreground hover:bg-muted-foreground/20',
                    )}
                  >
                    {turnNumber}
                  </button>
                </TooltipTrigger>
                <TooltipContent side="bottom" className="text-xs">
                  <div className="space-y-0.5">
                    <p className="font-medium">Turn {turnNumber}</p>
                    <p className="text-muted-foreground">
                      {formatTime(trace.start_time)}
                    </p>
                  </div>
                </TooltipContent>
              </Tooltip>
            )
          })}
        </div>

        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6"
          onClick={goToNext}
          disabled={!hasNext}
          aria-label="Next turn"
          title="Next turn"
        >
          <ChevronRight className="h-3.5 w-3.5" />
        </Button>

        <span className="ml-1 text-xs text-muted-foreground">
          {currentIndex + 1} of {session.turns} turn
          {session.turns > 1 ? 's' : ''}
        </span>
      </div>
    </TooltipProvider>
  )
}

export const SessionTimeline = memo(SessionTimelineComponent)
