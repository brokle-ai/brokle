import type { Span, TraceDetail as TraceDetailType } from '../api/types'
import { TraceDetailLayout } from './trace-detail-layout'

interface TraceDetailProps {
  trace: TraceDetailType
  spans: Span[]
  orgId: string
  projectId: string
}

/**
 * TraceDetail - Thin wrapper around TraceDetailLayout preserving the
 * original public surface. The layout owns the header (with the
 * annotations + comments drawer triggers, bookmark, back link, tags
 * editor) so the route no longer passes a separate `headerActions`
 * slot — the route component now just hands trace + spans + scope.
 */
export function TraceDetail({
  trace,
  spans,
  orgId,
  projectId,
}: TraceDetailProps) {
  return (
    <TraceDetailLayout
      trace={trace}
      spans={spans}
      orgId={orgId}
      projectId={projectId}
      className="h-[calc(100vh-6rem)]"
    />
  )
}
