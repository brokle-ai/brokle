'use client'

import { useMemo } from 'react'
import type { Trace } from '../data/schema'

/**
 * A group of traces belonging to the same session (multi-turn conversation)
 */
export interface SessionGroup {
  sessionId: string
  traces: Trace[]
  turns: number
  startTime: string
  endTime: string
}

/**
 * Result of session grouping analysis
 */
export interface UseSessionGroupingResult {
  /** All session groups found */
  sessions: SessionGroup[]
  /** The session containing the current trace */
  currentSession: SessionGroup | null
  /** Whether there are multiple sessions */
  hasMultipleSessions: boolean
}

/**
 * Extract session ID from trace attributes
 * Looks for common session/conversation/thread ID patterns
 */
function extractSessionId(trace: Trace): string {
  // Check resource_attributes first (most common location)
  const attrs = trace.resource_attributes || {}

  // Common session ID attribute names
  const sessionIdKeys = [
    'session_id',
    'session.id',
    'conversation_id',
    'conversation.id',
    'thread_id',
    'thread.id',
    'chat_id',
    'chat.id',
    'langchain.session_id',
    'langfuse.session_id',
  ]

  for (const key of sessionIdKeys) {
    const value = attrs[key]
    if (value && typeof value === 'string') {
      return value
    }
  }

  // Check trace-level session_id
  if (trace.session_id) {
    return trace.session_id
  }

  // Default: each trace is its own session
  return `trace-${trace.trace_id}`
}

/**
 * Hook to group traces by session for multi-turn visualization
 *
 * Sessions represent complete multi-turn conversations with an agent.
 * This is useful for visualizing:
 * - Chat conversations across multiple turns
 * - Agent workflows that span multiple traces
 * - Thread-based interactions
 *
 * @param traces - Array of traces to group
 * @param currentTraceId - Currently selected trace ID
 * @returns Session groups and current session info
 */
export function useSessionGrouping(
  traces: Trace[],
  currentTraceId?: string
): UseSessionGroupingResult {
  const sessions = useMemo(() => {
    if (!traces || traces.length === 0) {
      return []
    }

    // Group traces by session_id from attributes
    const sessionMap = new Map<string, Trace[]>()

    for (const trace of traces) {
      const sessionId = extractSessionId(trace)

      if (!sessionMap.has(sessionId)) {
        sessionMap.set(sessionId, [])
      }
      sessionMap.get(sessionId)!.push(trace)
    }

    // Convert to session groups
    return Array.from(sessionMap.entries())
      .map(([sessionId, sessionTraces]) => {
        // Sort traces by start_time
        const sorted = sessionTraces.sort(
          (a, b) =>
            new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
        )

        return {
          sessionId,
          traces: sorted,
          turns: sorted.length,
          startTime: sorted[0].start_time.toString(),
          endTime: sorted[sorted.length - 1].end_time?.toString() ||
            sorted[sorted.length - 1].start_time.toString(),
        }
      })
      // Sort sessions by start time
      .sort(
        (a, b) =>
          new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
      )
  }, [traces])

  const currentSession = useMemo(() => {
    if (!currentTraceId) return null
    return (
      sessions.find((s) =>
        s.traces.some((t) => t.trace_id === currentTraceId)
      ) || null
    )
  }, [sessions, currentTraceId])

  return {
    sessions,
    currentSession,
    hasMultipleSessions: sessions.length > 1,
  }
}
