'use client'

import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import {
  Fan,
  ArrowRight,
  Circle,
  Bot,
  Wrench,
  Search,
  Layers,
  Link,
  Shield,
  Wand2,
  Zap,
} from 'lucide-react'
import { cn } from '@/lib/utils'

/**
 * Span/Observation types supported by Brokle
 * Maps to OTEL semantic conventions and LLM observability patterns
 */
export type SpanType =
  | 'generation'
  | 'completion'
  | 'chat'
  | 'span'
  | 'agent'
  | 'tool'
  | 'retriever'
  | 'embedding'
  | 'chain'
  | 'guardrail'
  | 'evaluator'
  | 'event'
  | 'default'

/**
 * Icon mapping for span types
 */
const typeIcons: Record<SpanType, React.ComponentType<{ className?: string }>> = {
  generation: Fan,
  completion: Fan,
  chat: Fan,
  span: ArrowRight,
  agent: Bot,
  tool: Wrench,
  retriever: Search,
  embedding: Layers,
  chain: Link,
  guardrail: Shield,
  evaluator: Wand2,
  event: Zap,
  default: Circle,
}

/**
 * Badge variants using CVA for consistent styling
 */
const itemBadgeVariants = cva(
  'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset',
  {
    variants: {
      type: {
        generation: 'bg-fuchsia-50 text-fuchsia-700 ring-fuchsia-600/20 dark:bg-fuchsia-900/30 dark:text-fuchsia-300 dark:ring-fuchsia-400/30',
        completion: 'bg-fuchsia-50 text-fuchsia-700 ring-fuchsia-600/20 dark:bg-fuchsia-900/30 dark:text-fuchsia-300 dark:ring-fuchsia-400/30',
        chat: 'bg-fuchsia-50 text-fuchsia-700 ring-fuchsia-600/20 dark:bg-fuchsia-900/30 dark:text-fuchsia-300 dark:ring-fuchsia-400/30',
        span: 'bg-blue-50 text-blue-700 ring-blue-600/20 dark:bg-blue-900/30 dark:text-blue-300 dark:ring-blue-400/30',
        agent: 'bg-purple-50 text-purple-700 ring-purple-600/20 dark:bg-purple-900/30 dark:text-purple-300 dark:ring-purple-400/30',
        tool: 'bg-orange-50 text-orange-700 ring-orange-600/20 dark:bg-orange-900/30 dark:text-orange-300 dark:ring-orange-400/30',
        retriever: 'bg-teal-50 text-teal-700 ring-teal-600/20 dark:bg-teal-900/30 dark:text-teal-300 dark:ring-teal-400/30',
        embedding: 'bg-amber-50 text-amber-700 ring-amber-600/20 dark:bg-amber-900/30 dark:text-amber-300 dark:ring-amber-400/30',
        chain: 'bg-pink-50 text-pink-700 ring-pink-600/20 dark:bg-pink-900/30 dark:text-pink-300 dark:ring-pink-400/30',
        guardrail: 'bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-900/30 dark:text-red-300 dark:ring-red-400/30',
        evaluator: 'bg-indigo-50 text-indigo-700 ring-indigo-600/20 dark:bg-indigo-900/30 dark:text-indigo-300 dark:ring-indigo-400/30',
        event: 'bg-green-50 text-green-700 ring-green-600/20 dark:bg-green-900/30 dark:text-green-300 dark:ring-green-400/30',
        default: 'bg-gray-50 text-gray-700 ring-gray-600/20 dark:bg-gray-900/30 dark:text-gray-300 dark:ring-gray-400/30',
      },
      size: {
        default: 'text-xs',
        sm: 'text-[10px] px-1 py-0.5',
      },
    },
    defaultVariants: {
      type: 'default',
      size: 'default',
    },
  }
)

export interface ItemBadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof itemBadgeVariants> {
  /** The span/observation type */
  spanType?: string | null
  /** Whether to show the label text */
  showLabel?: boolean
  /** Whether to use small size */
  isSmall?: boolean
}

/**
 * Normalize span_type string to a known SpanType
 * Handles various formats: "gen_ai.completion", "generation", "SPAN", etc.
 */
function normalizeSpanType(spanType: string | null | undefined): SpanType {
  if (!spanType) return 'default'

  const normalized = spanType.toLowerCase().trim()

  // Handle gen_ai.* conventions
  if (normalized.includes('completion') || normalized.includes('generation')) {
    return 'generation'
  }
  if (normalized.includes('chat')) {
    return 'chat'
  }
  if (normalized.includes('embedding')) {
    return 'embedding'
  }
  if (normalized.includes('retriev')) {
    return 'retriever'
  }

  // Direct matches
  if (normalized === 'span' || normalized === 'internal') {
    return 'span'
  }
  if (normalized === 'agent') {
    return 'agent'
  }
  if (normalized === 'tool' || normalized === 'function') {
    return 'tool'
  }
  if (normalized === 'chain') {
    return 'chain'
  }
  if (normalized === 'guardrail' || normalized === 'guard') {
    return 'guardrail'
  }
  if (normalized === 'evaluator' || normalized === 'eval') {
    return 'evaluator'
  }
  if (normalized === 'event') {
    return 'event'
  }

  return 'default'
}

/**
 * Get display label for span type
 */
function getTypeLabel(type: SpanType): string {
  const labels: Record<SpanType, string> = {
    generation: 'Generation',
    completion: 'Completion',
    chat: 'Chat',
    span: 'Span',
    agent: 'Agent',
    tool: 'Tool',
    retriever: 'Retriever',
    embedding: 'Embedding',
    chain: 'Chain',
    guardrail: 'Guardrail',
    evaluator: 'Evaluator',
    event: 'Event',
    default: 'Span',
  }
  return labels[type]
}

/**
 * ItemBadge - Type-colored badge for span/observation types
 *
 * Displays an icon and optional label with colors based on span type.
 */
export function ItemBadge({
  spanType,
  showLabel = true,
  isSmall = false,
  className,
  ...props
}: ItemBadgeProps) {
  const normalizedType = normalizeSpanType(spanType)
  const Icon = typeIcons[normalizedType]
  const label = getTypeLabel(normalizedType)

  return (
    <span
      className={cn(
        itemBadgeVariants({ type: normalizedType, size: isSmall ? 'sm' : 'default' }),
        className
      )}
      title={label}
      {...props}
    >
      <Icon className={cn('shrink-0', isSmall ? 'h-2.5 w-2.5' : 'h-3 w-3')} />
      {showLabel && <span>{label}</span>}
    </span>
  )
}

/**
 * Export utilities for external use
 */
export { normalizeSpanType, getTypeLabel, typeIcons }
