// Playground feature types - All sessions are saved to database

// ----------------------------
// Session Types
// ----------------------------

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

// ----------------------------
// Session API Request/Response Types
// ----------------------------

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

// ----------------------------
// Config Types
// ----------------------------

export interface ModelConfig {
  model?: string
  temperature?: number
  max_tokens?: number
  top_p?: number
  frequency_penalty?: number
  presence_penalty?: number
  stop?: string[]
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
  config_overrides?: ModelConfig
  session_id?: string // Optional: updates session's last_run if provided
  project_id: string // Required: for project-scoped credential resolution
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
