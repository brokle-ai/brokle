import type { Span } from '@/features/traces/data/schema'
import type { ChatMessage, ModelConfig } from '../types'
import { createMessage } from '../types'

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
export function isLLMSpan(span: Span): boolean {
  // Check for model indicators
  if (span.model_name || span.gen_ai_request_model) {
    return true
  }

  // Check for provider indicators
  if (span.provider_name || span.gen_ai_provider_name) {
    return true
  }

  // Check span_type for LLM-related keywords
  const spanType = span.span_type?.toLowerCase() || ''
  if (spanType.includes('generation') ||
      spanType.includes('llm') ||
      spanType.includes('chat') ||
      spanType.includes('completion')) {
    return true
  }

  // Check gen_ai_operation_name
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
function extractContent(content: string | Array<{ type?: string; text?: string }> | undefined): string {
  if (!content) return ''

  if (typeof content === 'string') {
    return content
  }

  // Handle Anthropic-style content arrays
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
    .filter((m): m is RawMessage => m && typeof m === 'object')
    .map((m) => createMessage(
      normalizeRole(m.role),
      extractContent(m.content)
    ))
    .filter((m) => m.content.trim() !== '') // Filter out empty messages
}

/**
 * Parse span.input to ChatMessage array
 *
 * Handles multiple formats:
 * 1. OpenAI format: { messages: [{role: "user", content: "..."}] }
 * 2. Direct array: [{role: "user", content: "..."}]
 * 3. Anthropic format: { messages: [{role: "user", content: [...]}] } (content as array)
 * 4. Plain text: Convert to single user message
 *
 * @param input - Raw input string from span
 * @param output - Optional output string to include as assistant message
 * @param includeOutput - Whether to include output as assistant message
 */
export function parseSpanToMessages(
  input?: string,
  output?: string,
  includeOutput = false
): ChatMessage[] {
  if (!input) return []

  let messages: ChatMessage[] = []

  try {
    const parsed = JSON.parse(input)

    // Format 1: OpenAI/Anthropic format with messages property
    if (parsed && typeof parsed === 'object' && 'messages' in parsed) {
      messages = parseMessagesArray(parsed.messages)
    }
    // Format 2: Direct array of messages
    else if (Array.isArray(parsed)) {
      messages = parseMessagesArray(parsed)
    }
    // Format 3: Single message object
    else if (parsed && typeof parsed === 'object' && 'role' in parsed) {
      const msg = createMessage(
        normalizeRole(parsed.role),
        extractContent(parsed.content)
      )
      if (msg.content.trim()) {
        messages = [msg]
      }
    }
    // Format 4: Object with prompt/content property
    else if (parsed && typeof parsed === 'object') {
      const text = parsed.prompt || parsed.content || parsed.text || parsed.input
      if (typeof text === 'string' && text.trim()) {
        messages = [createMessage('user', text)]
      }
    }
  } catch {
    // Not valid JSON - treat as plain text
    if (input.trim()) {
      messages = [createMessage('user', input.trim())]
    }
  }

  // Optionally include output as assistant message
  if (includeOutput && output) {
    try {
      const parsedOutput = JSON.parse(output)

      // Handle structured output
      if (parsedOutput && typeof parsedOutput === 'object') {
        // Check for content property (common in responses)
        const content = parsedOutput.content ||
                       parsedOutput.text ||
                       parsedOutput.message ||
                       parsedOutput.output

        if (typeof content === 'string' && content.trim()) {
          messages.push(createMessage('assistant', content))
        } else if (Array.isArray(parsedOutput.choices)) {
          // OpenAI response format
          const choice = parsedOutput.choices[0]
          const messageContent = choice?.message?.content || choice?.text
          if (messageContent && typeof messageContent === 'string') {
            messages.push(createMessage('assistant', messageContent))
          }
        } else {
          // Unknown structure - stringify it
          messages.push(createMessage('assistant', JSON.stringify(parsedOutput, null, 2)))
        }
      } else if (typeof parsedOutput === 'string' && parsedOutput.trim()) {
        messages.push(createMessage('assistant', parsedOutput))
      }
    } catch {
      // Not valid JSON - treat as plain text
      if (output.trim()) {
        messages.push(createMessage('assistant', output.trim()))
      }
    }
  }

  return messages
}

/**
 * Extract model configuration from span attributes
 *
 * Maps span fields to ModelConfig:
 * - gen_ai_request_model or model_name → model
 * - gen_ai_provider_name or provider_name → provider
 * - gen_ai_request_temperature → temperature
 * - gen_ai_request_max_tokens → max_tokens
 * - gen_ai_request_top_p → top_p
 */
export function extractModelConfig(span: Span): ModelConfig | null {
  const model = span.gen_ai_request_model || span.model_name
  const provider = span.gen_ai_provider_name || span.provider_name

  // If no model info, return null
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

  // Extract parameters (only set if present in span)
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
export function getDisabledReason(span: Span): string | null {
  if (!isLLMSpan(span)) {
    return 'Only LLM spans can be opened in the playground'
  }

  if (!span.input) {
    return 'Span has no input data'
  }

  // Try to parse messages
  const messages = parseSpanToMessages(span.input)
  if (messages.length === 0) {
    return 'Could not parse messages from span input'
  }

  return null
}
