'use client'

import { useMemo } from 'react'
import type { Span } from '../data/schema'
import { detectSpanCategory } from '../utils/span-type-detector'

/**
 * Result of agent detection analysis
 */
export interface AgentDetectionResult {
  /** Whether the trace has agent workflow patterns worth visualizing as a graph */
  hasAgentWorkflow: boolean
  /** Number of spans detected as agent spans */
  agentSpanCount: number
  /** Number of spans detected as tool/API spans */
  toolSpanCount: number
  /** Number of spans detected as LLM spans */
  llmSpanCount: number
  /** Total number of spans analyzed */
  totalSpanCount: number
}

/**
 * Flatten nested spans for analysis
 */
function flattenSpans(spans: Span[]): Span[] {
  const result: Span[] = []

  function traverse(span: Span) {
    result.push(span)
    if (span.child_spans) {
      span.child_spans.forEach(traverse)
    }
  }

  spans.forEach(traverse)
  return result
}

/**
 * Hook to detect if a trace has agent workflow patterns
 *
 * Detection Logic:
 * - Count spans with category 'agent'
 * - Count spans with category 'api' or name containing 'tool'
 * - Return hasAgentWorkflow: true if:
 *   - agentSpanCount > 0 OR
 *   - toolSpanCount >= 2 OR
 *   - llmSpanCount >= 2 (multi-step LLM workflow)
 *
 * This determines whether to show the Graph tab in the navigation panel.
 *
 * @param spans - Array of spans to analyze
 * @returns Detection result with counts and workflow flag
 */
export function useAgentDetection(spans: Span[]): AgentDetectionResult {
  return useMemo(() => {
    if (!spans || spans.length === 0) {
      return {
        hasAgentWorkflow: false,
        agentSpanCount: 0,
        toolSpanCount: 0,
        llmSpanCount: 0,
        totalSpanCount: 0,
      }
    }

    // Flatten nested spans
    const flatSpans = flattenSpans(spans)

    let agentSpanCount = 0
    let toolSpanCount = 0
    let llmSpanCount = 0

    for (const span of flatSpans) {
      const category = detectSpanCategory(span.span_name, span.attributes)
      const nameLower = span.span_name.toLowerCase()

      switch (category) {
        case 'agent':
          agentSpanCount++
          break
        case 'api':
          toolSpanCount++
          break
        case 'llm':
          llmSpanCount++
          break
        default:
          // Check for tool patterns in span name
          if (
            nameLower.includes('tool') ||
            nameLower.includes('function_call') ||
            nameLower.includes('tool_use')
          ) {
            toolSpanCount++
          }
      }
    }

    // Determine if this is an agent workflow worth visualizing as a graph
    // Criteria:
    // 1. Has explicit agent spans
    // 2. Has multiple tool calls (agent orchestrating tools)
    // 3. Has multiple LLM calls (multi-step reasoning)
    // 4. Has enough spans to make a graph useful (>= 3)
    const hasAgentWorkflow =
      agentSpanCount > 0 ||
      toolSpanCount >= 2 ||
      llmSpanCount >= 2 ||
      flatSpans.length >= 3

    return {
      hasAgentWorkflow,
      agentSpanCount,
      toolSpanCount,
      llmSpanCount,
      totalSpanCount: flatSpans.length,
    }
  }, [spans])
}
