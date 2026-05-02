import type { Span as TraceSpan } from '@/features/traces/api/types'
import type { ChatMessage, ModelConfig } from '../types'
import { createMessage } from '../types'

// The traces wire type models the OTEL-core Span fields (timing,
// IDs, kind, attributes). Gen-AI specifics live in the attribute bags
// (`span_attributes`) on the ClickHouse side but are also projected to
// top-level columns on the internal query path; the web-vite wire DTO
// currently surfaces only the core fields. We extend Span locally with
// the gen-ai projection names (as optional) so this util can accept
// either shape — today's thin DTO or a future richer one — without
// forcing a traces/types edit.
export interface LLMSpan extends TraceSpan {
  gen_ai_request_model?: string
  gen_ai_provider_name?: string
  gen_ai_operation_name?: string
  gen_ai_request_temperature?: number
  gen_ai_request_max_tokens?: number
  gen_ai_request_top_p?: number
}

/**
 * Message format from various providers
 */
interface RawMessage {
  role?: string
  content?: string | Array<{ type?: string; text?: string }>
}

/**
 * Check if a span is an LLM span (can be opened in playground)
 *
 * A span is considered an LLM span if it has any of:
 * - model_name
 * - gen_ai_request_model
 * - provider_name
 * - gen_ai_provider_name
 * - span_type containing "generation", "llm", or "chat"
 */
export function isLLMSpan(span: LLMSpan): boolean {
  if (span.model_name || span.gen_ai_request_model) {
    return true
  }

  if (span.provider_name || span.gen_ai_provider_name) {
    return true
  }

  const spanType = span.span_type?.toLowerCase() || ''
  if (
    spanType.includes('generation') ||
    spanType.includes('llm') ||
    spanType.includes('chat') ||
    spanType.includes('completion')
  ) {
    return true
  }

  if (span.gen_ai_operation_name) {
    return true
  }

  return false
}

/**
 * Normalize role to valid ChatMessage role
 */
function normalizeRole(role: string | undefined): ChatMessage['role'] {
  const r = role?.toLowerCase()
  switch (r) {
    case 'system':
      return 'system'
    case 'assistant':
    case 'model': // Gemini uses 'model'
    case 'ai':
      return 'assistant'
    case 'user':
    case 'human':
    default:
      return 'user'
  }
}

/**
 * Extract text content from message content (handles Anthropic arrays)
 */
function extractContent(
  content: string | Array<{ type?: string; text?: string }> | undefined,
): string {
  if (!content) return ''

  if (typeof content === 'string') {
    return content
  }

  if (Array.isArray(content)) {
    return content
      .filter((block) => block.type === 'text' || !block.type)
      .map((block) => block.text || '')
      .join('\n')
  }

  return ''
}

/**
 * Parse messages from raw input data
 */
function parseMessagesArray(messages: unknown): ChatMessage[] {
  if (!Array.isArray(messages)) return []

  return messages
    .filter((m): m is RawMessage => !!m && typeof m === 'object')
    .map((m) => createMessage(normalizeRole(m.role), extractContent(m.content)))
    .filter((m) => m.content.trim() !== '')
}

/**
 * Parse span.input to ChatMessage array
 *
 * Handles multiple formats:
 * 1. OpenAI format: { messages: [{role: "user", content: "..."}] }
 * 2. Direct array: [{role: "user", content: "..."}]
 * 3. Anthropic format: { messages: [{role: "user", content: [...]}] } (content as array)
 * 4. Plain text: Convert to single user message
 */
export function parseSpanToMessages(
  input?: string,
  output?: string,
  includeOutput = false,
): ChatMessage[] {
  if (!input) return []

  let messages: ChatMessage[] = []

  try {
    const parsed = JSON.parse(input)

    if (parsed && typeof parsed === 'object' && 'messages' in parsed) {
      messages = parseMessagesArray((parsed as { messages: unknown }).messages)
    } else if (Array.isArray(parsed)) {
      messages = parseMessagesArray(parsed)
    } else if (parsed && typeof parsed === 'object' && 'role' in parsed) {
      const obj = parsed as RawMessage
      const msg = createMessage(normalizeRole(obj.role), extractContent(obj.content))
      if (msg.content.trim()) {
        messages = [msg]
      }
    } else if (parsed && typeof parsed === 'object') {
      const obj = parsed as Record<string, unknown>
      const text = obj.prompt ?? obj.content ?? obj.text ?? obj.input
      if (typeof text === 'string' && text.trim()) {
        messages = [createMessage('user', text)]
      }
    }
  } catch {
    if (input.trim()) {
      messages = [createMessage('user', input.trim())]
    }
  }

  if (includeOutput && output) {
    try {
      const parsedOutput = JSON.parse(output) as unknown

      if (parsedOutput && typeof parsedOutput === 'object') {
        const obj = parsedOutput as Record<string, unknown>
        const content = obj.content ?? obj.text ?? obj.message ?? obj.output

        if (typeof content === 'string' && content.trim()) {
          messages.push(createMessage('assistant', content))
        } else if (Array.isArray(obj.choices)) {
          // OpenAI response format
          const choice = obj.choices[0] as
            | { message?: { content?: string }; text?: string }
            | undefined
          const messageContent = choice?.message?.content || choice?.text
          if (messageContent && typeof messageContent === 'string') {
            messages.push(createMessage('assistant', messageContent))
          }
        } else {
          messages.push(createMessage('assistant', JSON.stringify(parsedOutput, null, 2)))
        }
      } else if (typeof parsedOutput === 'string' && parsedOutput.trim()) {
        messages.push(createMessage('assistant', parsedOutput))
      }
    } catch {
      if (output.trim()) {
        messages.push(createMessage('assistant', output.trim()))
      }
    }
  }

  return messages
}

/**
 * Extract model configuration from span attributes
 */
export function extractModelConfig(span: LLMSpan): ModelConfig | null {
  const model = span.gen_ai_request_model || span.model_name
  const provider = span.gen_ai_provider_name || span.provider_name

  if (!model && !provider) {
    return null
  }

  const config: ModelConfig = {}

  if (model) {
    config.model = model
  }

  if (provider) {
    config.provider = provider.toLowerCase()
  }

  if (span.gen_ai_request_temperature !== undefined && span.gen_ai_request_temperature !== null) {
    config.temperature = span.gen_ai_request_temperature
    config.temperature_enabled = true
  }

  if (span.gen_ai_request_max_tokens !== undefined && span.gen_ai_request_max_tokens !== null) {
    config.max_tokens = span.gen_ai_request_max_tokens
    config.max_tokens_enabled = true
  }

  if (span.gen_ai_request_top_p !== undefined && span.gen_ai_request_top_p !== null) {
    config.top_p = span.gen_ai_request_top_p
    config.top_p_enabled = true
  }

  return config
}

/**
 * Get a descriptive reason why a span cannot be opened in playground
 */
export function getDisabledReason(span: LLMSpan): string | null {
  if (!isLLMSpan(span)) {
    return 'Only LLM spans can be opened in the playground'
  }

  if (!span.input) {
    return 'Span has no input data'
  }

  const messages = parseSpanToMessages(span.input)
  if (messages.length === 0) {
    return 'Could not parse messages from span input'
  }

  return null
}
