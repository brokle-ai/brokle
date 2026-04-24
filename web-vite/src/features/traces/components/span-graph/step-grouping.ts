/**
 * Step Grouping Algorithm
 *
 * Groups spans into execution steps based on temporal overlap (Langfuse-style).
 * Spans that start before any span in the current step finishes are considered
 * part of the same step (parallel execution).
 */

import type { Span } from '../../api/types'

/**
 * A group of spans that execute in parallel within the same step.
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
 */
export function buildStepGroups(spans: Span[]): StepGroup[] {
  if (spans.length === 0) return []

  const sorted = [...spans].sort(
    (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime(),
  )

  const groups: StepGroup[] = []
  let currentGroup: Span[] = []
  let currentEndTime = 0
  let step = 0

  for (const span of sorted) {
    const spanStart = new Date(span.start_time).getTime()
    const spanEnd = span.end_time ? new Date(span.end_time).getTime() : spanStart

    if (currentGroup.length === 0 || spanStart < currentEndTime) {
      currentGroup.push(span)
      currentEndTime = Math.max(currentEndTime, spanEnd)
    } else {
      groups.push({
        step: step++,
        spans: currentGroup,
        startTime: Math.min(
          ...currentGroup.map((s) => new Date(s.start_time).getTime()),
        ),
        endTime: currentEndTime,
      })
      currentGroup = [span]
      currentEndTime = spanEnd
    }
  }

  if (currentGroup.length > 0) {
    groups.push({
      step,
      spans: currentGroup,
      startTime: Math.min(
        ...currentGroup.map((s) => new Date(s.start_time).getTime()),
      ),
      endTime: currentEndTime,
    })
  }

  return groups
}

/**
 * Build edges connecting spans between consecutive steps.
 * Each span in step N connects to each span in step N+1.
 */
export function buildStepEdges(
  groups: StepGroup[],
): Array<{ id: string; source: string; target: string }> {
  const edges: Array<{ id: string; source: string; target: string }> = []

  for (let i = 0; i < groups.length - 1; i++) {
    const currentStep = groups[i]
    const nextStep = groups[i + 1]

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
 * Get the step number for a given span ID.
 */
export function getStepForSpan(groups: StepGroup[], spanId: string): number {
  for (const group of groups) {
    if (group.spans.some((s) => s.span_id === spanId)) {
      return group.step
    }
  }
  return -1
}
