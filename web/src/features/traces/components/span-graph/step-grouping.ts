/**
 * Step Grouping Algorithm
 *
 * Groups spans into execution steps based on temporal overlap (Langfuse-style).
 * Spans that start before any span in the current step finishes are considered
 * part of the same step (parallel execution).
 */

import type { Span } from '../../data/schema'

/**
 * A group of spans that execute in parallel within the same step
 */
export interface StepGroup {
  step: number
  spans: Span[]
  startTime: number
  endTime: number
}

/**
 * Groups spans into steps based on temporal overlap.
 * Spans that start before any span in the current step finishes
 * are considered part of the same step (parallel execution).
 *
 * @param spans - Array of spans to group
 * @returns Array of step groups sorted by step number
 */
export function buildStepGroups(spans: Span[]): StepGroup[] {
  if (spans.length === 0) return []

  // Sort spans by start_time
  const sorted = [...spans].sort(
    (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
  )

  const groups: StepGroup[] = []
  let currentGroup: Span[] = []
  let currentEndTime = 0
  let step = 0

  for (const span of sorted) {
    const spanStart = new Date(span.start_time).getTime()
    const spanEnd = span.end_time ? new Date(span.end_time).getTime() : spanStart

    if (currentGroup.length === 0 || spanStart < currentEndTime) {
      // Add to current group (parallel execution)
      currentGroup.push(span)
      currentEndTime = Math.max(currentEndTime, spanEnd)
    } else {
      // Start new group (sequential execution)
      groups.push({
        step: step++,
        spans: currentGroup,
        startTime: Math.min(
          ...currentGroup.map((s) => new Date(s.start_time).getTime())
        ),
        endTime: currentEndTime,
      })
      currentGroup = [span]
      currentEndTime = spanEnd
    }
  }

  // Push final group
  if (currentGroup.length > 0) {
    groups.push({
      step: step,
      spans: currentGroup,
      startTime: Math.min(
        ...currentGroup.map((s) => new Date(s.start_time).getTime())
      ),
      endTime: currentEndTime,
    })
  }

  return groups
}

/**
 * Build edges connecting spans between consecutive steps.
 * Each span in step N connects to each span in step N+1.
 *
 * @param groups - Array of step groups
 * @returns Array of edge definitions for React Flow
 */
export function buildStepEdges(
  groups: StepGroup[]
): Array<{ id: string; source: string; target: string }> {
  const edges: Array<{ id: string; source: string; target: string }> = []

  for (let i = 0; i < groups.length - 1; i++) {
    const currentStep = groups[i]
    const nextStep = groups[i + 1]

    // Connect each span in current step to each span in next step
    for (const fromSpan of currentStep.spans) {
      for (const toSpan of nextStep.spans) {
        edges.push({
          id: `step-${fromSpan.span_id}-${toSpan.span_id}`,
          source: fromSpan.span_id,
          target: toSpan.span_id,
        })
      }
    }
  }

  return edges
}

/**
 * Get the step number for a given span ID
 *
 * @param groups - Array of step groups
 * @param spanId - The span ID to find
 * @returns Step number or -1 if not found
 */
export function getStepForSpan(groups: StepGroup[], spanId: string): number {
  for (const group of groups) {
    if (group.spans.some((s) => s.span_id === spanId)) {
      return group.step
    }
  }
  return -1
}
