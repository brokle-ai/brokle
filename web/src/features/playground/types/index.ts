export interface PlaygroundSession {
  id: string
  project_id: string
  name?: string
  description?: string
  variables: Record<string, string>
  config?: ModelConfig
  windows?: WindowState[]
  last_run?: LastRun
  tags: string[]
  created_by?: string
  created_at: string
  updated_at: string
  last_used_at: string
}

export interface PlaygroundSessionSummary {
  id: string
  name?: string
  description?: string
  tags: string[]
  created_at: string
  last_used_at: string
}

export interface LastRun {
  content: string
  metrics?: RunMetrics
  timestamp: string
  error?: string
}

export interface RunMetrics {
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  cost?: number
  latency_ms?: number
  ttft_ms?: number
  model?: string
}

/**
 * Run history entry for in-memory per-window history (Opik pattern).
 * Stores last 10 runs with stale marking when prompt changes.
 */
export interface RunHistoryEntry {
  id: string
  content: string
  metrics: RunMetrics | null
  timestamp: string // ISO string
  error?: string
  isStale: boolean // Opik pattern: true when prompt has changed since this run
  // Snapshot of inputs for restore
  messages: ChatMessage[]
  variables: Record<string, string>
  config: ModelConfig | null
}

export interface WindowState {
  template: ChatTemplate | TextTemplate
  variables?: Record<string, string>
  config?: ModelConfig
  last_run?: LastRun
  // Prompt linking metadata (Opik-style)
  loadedFromPromptId?: string
  loadedFromPromptName?: string
  loadedFromPromptVersionId?: string
  loadedFromPromptVersionNumber?: number
  loadedTemplate?: string // Original template for change detection
}

export interface CreateSessionRequest {
  name: string
  description?: string
  tags?: string[]
  variables?: Record<string, string>
  config?: ModelConfig
  windows: WindowState[]
}

export interface UpdateSessionRequest {
  name?: string
  description?: string
  tags?: string[]
  variables?: Record<string, string>
  config?: ModelConfig
  windows?: WindowState[]
}

export interface ModelConfig {
  model?: string
  provider?: string // Explicit provider (openai, anthropic, azure, gemini, openrouter, custom)
  credential_id?: string // Credential ID for multi-credential scenarios
  temperature?: number
  max_tokens?: number
  top_p?: number
  frequency_penalty?: number
  presence_penalty?: number
  stop?: string[]
  // Enable flags - when false/undefined, parameter uses model default (not sent to API)
  temperature_enabled?: boolean
  max_tokens_enabled?: boolean
  top_p_enabled?: boolean
  frequency_penalty_enabled?: boolean
  presence_penalty_enabled?: boolean
  // Tools/Function Calling
  tools?: Tool[]
  tool_choice?: ToolChoice
  // Structured Output
  response_format?: ResponseFormat
}

/**
 * API-safe model config without UI-only fields.
 * Used when sending config to the backend API.
 * Tools are in API format (without UI-only 'id' field).
 */
export interface APIModelConfig {
  model?: string
  provider?: string
  credential_id?: string
  temperature?: number
  max_tokens?: number
  top_p?: number
  frequency_penalty?: number
  presence_penalty?: number
  stop?: string[]
  tools?: object[] // API format without UI 'id' field
  tool_choice?: ToolChoice
  response_format?: ResponseFormat
}

// Re-export tool types for convenience
export type { Tool, ToolFunction, ToolCall, ToolChoice, SpecificToolChoice, ResponseFormat, JsonSchemaResponseFormat } from './tools'
export { createEmptyTool, createToolFromJSON, validateTool, toolsToAPIFormat, parseToolCalls, formatToolCallArguments } from './tools'

import type { Tool, ToolChoice, ResponseFormat } from './tools'
import { toolsToAPIFormat } from './tools'

export type ParameterKey = 'temperature' | 'max_tokens' | 'top_p' | 'frequency_penalty' | 'presence_penalty'

export interface ParameterDefinition {
  key: ParameterKey
  label: string
  description: string
  type: 'slider' | 'bipolar-slider' | 'number'
  min: number
  max: number
  step: number
  defaultValue: number
  formatValue: (v: number) => string
}

export const PARAMETER_DEFINITIONS: ParameterDefinition[] = [
  {
    key: 'temperature',
    label: 'Temperature',
    description: 'Controls randomness: 0 is focused, higher values are more creative',
    type: 'slider',
    min: 0,
    max: 2,
    step: 0.1,
    defaultValue: 1.0,
    formatValue: (v) => v.toFixed(1),
  },
  {
    key: 'max_tokens',
    label: 'Max Tokens',
    description: 'Maximum length of generated response',
    type: 'number',
    min: 1,
    max: 128000,
    step: 1,
    defaultValue: 4096,
    formatValue: (v) => v.toString(),
  },
  {
    key: 'top_p',
    label: 'Top P',
    description: 'Nucleus sampling threshold (alternative to temperature)',
    type: 'slider',
    min: 0,
    max: 1,
    step: 0.05,
    defaultValue: 1.0,
    formatValue: (v) => v.toFixed(2),
  },
  {
    key: 'frequency_penalty',
    label: 'Frequency Penalty',
    description: 'Reduce repetition based on token frequency (0 to 2)',
    type: 'slider',
    min: 0,
    max: 2,
    step: 0.1,
    defaultValue: 0,
    formatValue: (v) => v.toFixed(1),
  },
  {
    key: 'presence_penalty',
    label: 'Presence Penalty',
    description: 'Reduce repetition based on token presence (0 to 2)',
    type: 'slider',
    min: 0,
    max: 2,
    step: 0.1,
    defaultValue: 0,
    formatValue: (v) => v.toFixed(1),
  },
]

// Provider-specific parameter support
export const PROVIDER_PARAMETER_SUPPORT: Record<string, ParameterKey[]> = {
  openai: ['temperature', 'max_tokens', 'top_p', 'frequency_penalty', 'presence_penalty'],
  anthropic: ['temperature', 'max_tokens', 'top_p'], // No frequency/presence penalty
  azure: ['temperature', 'max_tokens', 'top_p', 'frequency_penalty', 'presence_penalty'],
  gemini: ['temperature', 'max_tokens', 'top_p'], // No frequency/presence penalty
  openrouter: ['temperature', 'max_tokens', 'top_p', 'frequency_penalty', 'presence_penalty'],
  custom: ['temperature', 'max_tokens', 'top_p', 'frequency_penalty', 'presence_penalty'],
}

/**
 * Converts a UI ModelConfig to an API-safe config by:
 * - Filtering to only include enabled parameters
 * - Removing UI-only fields (like tool ids)
 * Returns APIModelConfig suitable for sending to the backend.
 */
export function getEnabledModelConfig(config: ModelConfig | undefined): APIModelConfig | undefined {
  if (!config) return undefined

  const result: APIModelConfig = {
    model: config.model,
    provider: config.provider,
    credential_id: config.credential_id,
    stop: config.stop,
  }

  // Only include parameters that are explicitly enabled
  if (config.temperature_enabled && config.temperature !== undefined) {
    result.temperature = config.temperature
  }
  if (config.max_tokens_enabled && config.max_tokens !== undefined) {
    result.max_tokens = config.max_tokens
  }
  if (config.top_p_enabled && config.top_p !== undefined) {
    result.top_p = config.top_p
  }
  if (config.frequency_penalty_enabled && config.frequency_penalty !== undefined) {
    result.frequency_penalty = config.frequency_penalty
  }
  if (config.presence_penalty_enabled && config.presence_penalty !== undefined) {
    result.presence_penalty = config.presence_penalty
  }

  // Include tools/function calling if present (strip UI-only id field)
  if (config.tools && config.tools.length > 0) {
    // toolsToAPIFormat removes the UI-only 'id' field that would cause provider errors
    result.tools = toolsToAPIFormat(config.tools)
  }
  if (config.tool_choice) {
    result.tool_choice = config.tool_choice
  }

  // Include response format for structured output if present
  if (config.response_format) {
    result.response_format = config.response_format
  }

  return result
}

export interface ChatMessage {
  id: string // Unique ID for drag-and-drop reordering
  role: 'system' | 'user' | 'assistant'
  content: string
}

export interface ChatTemplate {
  messages: ChatMessage[]
}

// Note: TextTemplate kept for API compatibility but playground is chat-only now
export interface TextTemplate {
  content: string
}

// Helper to create a new message with unique ID
export const createMessage = (
  role: ChatMessage['role'] = 'user',
  content: string = ''
): ChatMessage => ({
  id: crypto.randomUUID(),
  role,
  content,
})

export interface StreamChunk {
  type: 'start' | 'content' | 'end' | 'error' | 'metrics'
  content?: string
  error?: string
  finish_reason?: string
  metrics?: StreamMetrics
}

export interface StreamMetrics {
  model?: string
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  cost?: number
  ttft_ms?: number
  total_duration_ms?: number
}

// Execution request (stateless, optionally updates session last_run)
export interface ExecuteRequest {
  template: ChatTemplate | TextTemplate
  prompt_type: 'text' | 'chat'
  variables: Record<string, string>
  config_overrides?: APIModelConfig // API-safe config (tools without UI id field)
  session_id?: string // Optional: updates session's last_run if provided
  project_id: string // Required: for session access validation
}

// Execution response
export interface ExecuteResponse {
  compiled_prompt: any
  response?: {
    content: string
    model: string
    usage?: {
      prompt_tokens: number
      completion_tokens: number
      total_tokens: number
    }
    cost?: number
  }
  latency_ms: number
  error?: string
}
