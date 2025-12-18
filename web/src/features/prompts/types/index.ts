export type PromptType = 'text' | 'chat'

export interface ChatMessage {
  type: string // 'message' | 'placeholder'
  role?: string // 'system' | 'user' | 'assistant'
  content?: string
  name?: string // For placeholders
}

export interface ModelConfig {
  model?: string
  temperature?: number
  max_tokens?: number
  top_p?: number
  frequency_penalty?: number
  presence_penalty?: number
  stop?: string[]
}

export interface TextTemplate {
  content: string
}

export interface ChatTemplate {
  messages: ChatMessage[]
}

export interface Prompt {
  id: string
  name: string
  type: PromptType
  description?: string
  tags: string[]
  version: number
  labels: string[]
  template: TextTemplate | ChatTemplate
  config?: ModelConfig
  variables: string[]
  commit_message?: string
  created_at: string
  created_by?: string
  is_fallback?: boolean
}

export interface PromptListItem {
  id: string
  name: string
  type: PromptType
  description?: string
  tags: string[]
  latest_version: number
  labels: PromptLabelInfo[]
  created_at: string
  updated_at: string
}

export interface PromptLabelInfo {
  name: string
  version: number
}

export interface PromptVersion {
  id: string
  version: number
  template: TextTemplate | ChatTemplate
  config?: ModelConfig
  variables: string[]
  commit_message?: string
  labels: string[]
  created_at: string
  created_by?: string
}

export interface VersionDiff {
  from_version: number
  to_version: number
  template_from: TextTemplate | ChatTemplate
  template_to: TextTemplate | ChatTemplate
  config_from?: ModelConfig
  config_to?: ModelConfig
  variables_added: string[]
  variables_removed: string[]
}

export interface ExecutePromptResponse {
  compiled_prompt: any
  response?: LLMResponse
  latency_ms: number
  error?: string
}

export interface LLMResponse {
  content: string
  model: string
  usage?: LLMUsage
  cost?: number
}

export interface LLMUsage {
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export interface UpsertResponse {
  id: string
  name: string
  type: PromptType
  version: number
  is_new_prompt: boolean
  labels: string[]
  created_at: string
}

export interface CreatePromptRequest {
  name: string
  type?: PromptType
  description?: string
  tags?: string[]
  template: TextTemplate | ChatTemplate
  config?: ModelConfig
  labels?: string[]
  commit_message?: string
}

export interface UpdatePromptRequest {
  name?: string
  description?: string
  tags?: string[]
}

export interface CreateVersionRequest {
  template: TextTemplate | ChatTemplate
  config?: ModelConfig
  labels?: string[]
  commit_message?: string
}

export interface SetLabelsRequest {
  labels: string[]
}

export interface ExecutePromptRequest {
  variables: Record<string, string>
  config_overrides?: ModelConfig
}

export interface GetPromptsParams {
  projectId: string
  type?: PromptType
  tags?: string[]
  search?: string
  page?: number
  limit?: number
}

export interface GetPromptParams {
  projectId: string
  promptId?: string
  name?: string
  label?: string
  version?: number
}
