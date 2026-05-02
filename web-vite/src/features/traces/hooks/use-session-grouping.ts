import { useMemo } from 'react'
import type { TraceSummary } from '../api/types'

/**
 * A group of traces belonging to the same session — i.e. multi-turn
 * conversation. Web-vite consumes traces in the trimmed `TraceSummary`
 * wire shape, which only carries the top-level `session_id` field
 * (the backend extracts and materialises it from
 * `langfuse.session_id` / `session.id` attributes during ingestion).
 * That means we don't need to crawl `resource_attributes` here the way
 * web/'s implementation did.
 */
export interface SessionGroup {
  sessionId: string
  traces: TraceSummary[]
  turns: number
  startTime: string
  endTime: string
}

export interface UseSessionGroupingResult {
  /** All session groups found in the supplied trace list. */
  sessions: SessionGroup[]
  /** The session containing `currentTraceId`, if any. */
  currentSession: SessionGroup | null
  /** True when more than one session is present. */
  hasMultipleSessions: boolean
}

/**
 * Resolve a stable session-id for a trace. Falls back to a synthetic
 * `trace-<id>` key when the trace is genuinely standalone, so each
 * trace still gets its own group instead of collapsing into a "no
 * session" bucket.
 */
function extractSessionId(trace: TraceSummary): string {
  if (trace.session_id && trace.session_id.length > 0) {
    return trace.session_id
  }
  return `trace-${trace.trace_id}`
}

/**
 * Group traces by session for multi-turn conversation visualisation.
 * Used by the trace detail's session timeline to surface the parent
 * conversation when a single trace is part of a larger flow.
 */
export function useSessionGrouping(
  traces: TraceSummary[],
  currentTraceId?: string,
): UseSessionGroupingResult {
  const sessions = useMemo(() => {
    if (!traces || traces.length === 0) return []

    const sessionMap = new Map<string, TraceSummary[]>()
    for (const trace of traces) {
      const sessionId = extractSessionId(trace)
      const bucket = sessionMap.get(sessionId)
      if (bucket) {
        bucket.push(trace)
      } else {
        sessionMap.set(sessionId, [trace])
      }
    }

    return Array.from(sessionMap.entries())
      .map(([sessionId, sessionTraces]) => {
        const sorted = [...sessionTraces].sort(
          (a, b) =>
            new Date(a.start_time).getTime() -
            new Date(b.start_time).getTime(),
        )
        const last = sorted[sorted.length - 1]
        return {
          sessionId,
          traces: sorted,
          turns: sorted.length,
          startTime: sorted[0].start_time,
          endTime: last.end_time ?? last.start_time,
        }
      })
      .sort(
        (a, b) =>
          new Date(a.startTime).getTime() - new Date(b.startTime).getTime(),
      )
  }, [traces])

  const currentSession = useMemo(() => {
    if (!currentTraceId) return null
    return (
      sessions.find((s) =>
        s.traces.some((t) => t.trace_id === currentTraceId),
      ) ?? null
    )
  }, [sessions, currentTraceId])

  return {
    sessions,
    currentSession,
    hasMultipleSessions: sessions.length > 1,
  }
}
