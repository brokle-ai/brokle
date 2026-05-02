/**
 * Span Type Detection Utility.
 *
 * Detects the category of an OTEL span based on span_name patterns and
 * attributes. Used by the span detail panel to show adaptive metrics.
 */

export type SpanCategory =
  | 'llm'
  | 'conversation'
  | 'agent'
  | 'pipeline'
  | 'batch'
  | 'api'
  | 'worker'
  | 'generic'

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
 * Detect the category of a span based on its name and attributes.
 */
export function detectSpanCategory(
  spanName: string,
  attributes?: Record<string, unknown> | null,
): SpanCategory {
  const attrs = attributes ?? {}
  const name = spanName.toLowerCase()

  if (
    name.startsWith('llm.') ||
    attrs['gen_ai.request.model'] ||
    attrs['gen_ai.system']
  ) {
    return 'llm'
  }

  if (name.startsWith('agent.')) return 'agent'
  if (name.startsWith('batch.')) return 'batch'

  if (
    name === 'conversation' ||
    (name.includes('.turn_') && !name.startsWith('llm.'))
  ) {
    return 'conversation'
  }

  if (name === 'llm.pipeline' || PIPELINE_OPERATIONS.includes(name)) {
    return 'pipeline'
  }

  if (name.includes('worker') || name === 'parallel_orchestrator') {
    return 'worker'
  }

  if (name === 'api_request' || attrs['http.method'] || attrs['http.url']) {
    return 'api'
  }

  return 'generic'
}

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
      return [
        'pipeline.name',
        'pipeline.version',
        'retrieval.documents',
        'embedding.model',
      ]
    case 'worker':
      return ['worker_id', 'task', 'child_count', 'operation']
    case 'api':
      return [
        'http.method',
        'http.url',
        'http.status_code',
        'http.response_content_length',
      ]
    case 'generic':
    default:
      return []
  }
}
