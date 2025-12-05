/**
 * Span Type Detection Utility
 *
 * Detects the category of an OTEL span based on span_name patterns and attributes.
 * Used by the peek sheet to show adaptive metrics based on span type.
 *
 * @see https://opentelemetry.io/docs/specs/semconv/
 */

/**
 * Categories of spans that determine what metrics to display
 */
export type SpanCategory =
  | 'llm' // LLM spans with gen_ai.* attributes (chat completions, embeddings)
  | 'conversation' // Multi-turn conversation spans
  | 'agent' // AI agent spans (iterations, tool calls, reasoning)
  | 'pipeline' // RAG/pipeline spans (retrieval, vector search)
  | 'batch' // Batch processing spans
  | 'api' // HTTP API spans
  | 'worker' // Worker/orchestration spans
  | 'generic' // Fallback for unknown spans

/**
 * Human-readable labels for each span category
 */
export const SPAN_CATEGORY_LABELS: Record<SpanCategory, string> = {
  llm: 'LLM',
  conversation: 'CONVERSATION',
  agent: 'AGENT',
  pipeline: 'PIPELINE',
  batch: 'BATCH',
  api: 'API',
  worker: 'WORKER',
  generic: 'SPAN',
}

/**
 * Badge colors for each span category (Tailwind CSS classes)
 */
export const SPAN_CATEGORY_COLORS: Record<
  SpanCategory,
  { bg: string; text: string; border: string }
> = {
  llm: {
    bg: 'bg-purple-100 dark:bg-purple-900/30',
    text: 'text-purple-700 dark:text-purple-300',
    border: 'border-purple-200 dark:border-purple-800',
  },
  conversation: {
    bg: 'bg-blue-100 dark:bg-blue-900/30',
    text: 'text-blue-700 dark:text-blue-300',
    border: 'border-blue-200 dark:border-blue-800',
  },
  agent: {
    bg: 'bg-orange-100 dark:bg-orange-900/30',
    text: 'text-orange-700 dark:text-orange-300',
    border: 'border-orange-200 dark:border-orange-800',
  },
  pipeline: {
    bg: 'bg-green-100 dark:bg-green-900/30',
    text: 'text-green-700 dark:text-green-300',
    border: 'border-green-200 dark:border-green-800',
  },
  batch: {
    bg: 'bg-cyan-100 dark:bg-cyan-900/30',
    text: 'text-cyan-700 dark:text-cyan-300',
    border: 'border-cyan-200 dark:border-cyan-800',
  },
  api: {
    bg: 'bg-amber-100 dark:bg-amber-900/30',
    text: 'text-amber-700 dark:text-amber-300',
    border: 'border-amber-200 dark:border-amber-800',
  },
  worker: {
    bg: 'bg-teal-100 dark:bg-teal-900/30',
    text: 'text-teal-700 dark:text-teal-300',
    border: 'border-teal-200 dark:border-teal-800',
  },
  generic: {
    bg: 'bg-muted',
    text: 'text-muted-foreground',
    border: 'border-border',
  },
}

/**
 * Known pipeline/RAG operation names
 */
const PIPELINE_OPERATIONS = [
  'retrieval',
  'vector_search',
  'embed_query',
  'augmentation',
  'generation',
  'prompt_construction',
  'reranking',
  'chunking',
]

/**
 * Detect the category of a span based on its name and attributes
 *
 * Detection order (most specific first):
 * 1. LLM: llm.* prefix OR gen_ai.request.model attribute
 * 2. Agent: agent.* prefix
 * 3. Batch: batch.* prefix
 * 4. Conversation: conversation prefix OR .turn_ in name
 * 5. Pipeline: known RAG operations
 * 6. Worker: worker in name OR parallel_orchestrator
 * 7. API: api_request OR http.method attribute
 * 8. Generic: fallback
 *
 * @param spanName - The span_name field from OTEL
 * @param attributes - The span_attributes map (optional)
 * @returns The detected span category
 */
export function detectSpanCategory(
  spanName: string,
  attributes?: Record<string, unknown> | null
): SpanCategory {
  const attrs = attributes ?? {}
  const name = spanName.toLowerCase()

  // 1. LLM spans - most common, check first
  if (
    name.startsWith('llm.') ||
    attrs['gen_ai.request.model'] ||
    attrs['gen_ai.system']
  ) {
    return 'llm'
  }

  // 2. Agent spans
  if (name.startsWith('agent.')) {
    return 'agent'
  }

  // 3. Batch spans
  if (name.startsWith('batch.')) {
    return 'batch'
  }

  // 4. Conversation spans (but not llm.chat.completions.turn_N which is handled by LLM)
  if (name === 'conversation' || (name.includes('.turn_') && !name.startsWith('llm.'))) {
    return 'conversation'
  }

  // 5. Pipeline/RAG spans
  if (name === 'llm.pipeline' || PIPELINE_OPERATIONS.includes(name)) {
    return 'pipeline'
  }

  // 6. Worker/Orchestration spans
  if (name.includes('worker') || name === 'parallel_orchestrator') {
    return 'worker'
  }

  // 7. API spans
  if (name === 'api_request' || attrs['http.method'] || attrs['http.url']) {
    return 'api'
  }

  // 8. Generic fallback
  return 'generic'
}

/**
 * Get relevant attribute keys for a span category
 * Used to highlight important attributes in the UI
 *
 * @param category - The span category
 * @returns Array of attribute key patterns to highlight
 */
export function getRelevantAttributeKeys(category: SpanCategory): string[] {
  switch (category) {
    case 'llm':
      return [
        'gen_ai.request.model',
        'gen_ai.system',
        'gen_ai.response.finish_reason',
        'gen_ai.usage.input_tokens',
        'gen_ai.usage.output_tokens',
        'gen_ai.usage.total_tokens',
      ]
    case 'agent':
      return [
        'agent.name',
        'agent.max_iterations',
        'iteration',
        'tool.name',
        'tool.status',
        'answer.tokens',
      ]
    case 'batch':
      return ['batch.id', 'batch.size', 'parallelism', 'item.index', 'item.status']
    case 'conversation':
      return ['conversation.id', 'conversation.turns', 'conversation.turn']
    case 'pipeline':
      return ['pipeline.name', 'pipeline.version', 'retrieval.documents', 'embedding.model']
    case 'worker':
      return ['worker_id', 'task', 'child_count', 'operation']
    case 'api':
      return ['http.method', 'http.url', 'http.status_code', 'http.response_content_length']
    case 'generic':
    default:
      return []
  }
}

/**
 * Check if a span has LLM-specific data (tokens, cost)
 *
 * @param span - The span object
 * @returns True if the span has LLM-specific data
 */
export function hasLLMData(span: {
  usage_details?: Record<string, number>
  total_cost?: number
  gen_ai_usage_input_tokens?: number
  gen_ai_usage_output_tokens?: number
}): boolean {
  return !!(
    span.usage_details?.input ||
    span.usage_details?.output ||
    span.total_cost ||
    span.gen_ai_usage_input_tokens ||
    span.gen_ai_usage_output_tokens
  )
}

/**
 * Get the display icon for a span category (lucide-react icon name)
 */
export function getSpanCategoryIcon(category: SpanCategory): string {
  switch (category) {
    case 'llm':
      return 'Brain'
    case 'agent':
      return 'Bot'
    case 'batch':
      return 'Layers'
    case 'conversation':
      return 'MessageSquare'
    case 'pipeline':
      return 'Workflow'
    case 'worker':
      return 'Cpu'
    case 'api':
      return 'Globe'
    case 'generic':
    default:
      return 'Circle'
  }
}
