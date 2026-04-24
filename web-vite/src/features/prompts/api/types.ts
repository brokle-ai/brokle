// Wire types for the prompts endpoints. Shapes match
// `internal/transport/http/handlers/prompt/types.go` on the backend.
// Timestamps arrive as RFC 3339 strings; components format on render.

export type PromptType = 'text' | 'chat'

export type TemplateDialect = 'simple' | 'mustache' | 'jinja2' | 'auto'

export interface PromptLabelInfo {
  name: string
  version: number
}

export interface ChatMessage {
  type: string // 'message' | 'placeholder'
  role?: string // 'system' | 'user' | 'assistant'
  content?: string
  name?: string
}

export interface ModelConfig {
  model?: string
  provider?: string
  credential_id?: string
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

export type PromptTemplate = TextTemplate | ChatTemplate

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

// Detail response (single prompt at its current/latest version).
export interface Prompt {
  id: string
  name: string
  type: PromptType
  description?: string
  tags: string[]
  version: number
  version_id: string
  labels: string[]
  template: PromptTemplate
  config?: ModelConfig | null
  variables: string[]
  commit_message?: string
  created_at: string
  created_by?: string
  is_fallback?: boolean
}

export interface PromptVersion {
  id: string
  version: number
  template: PromptTemplate
  config?: ModelConfig | null
  variables: string[]
  commit_message?: string
  labels: string[]
  created_at: string
  created_by?: string
}

// The dashboard list route returns a flat `{data, total, page, limit}`
// envelope rather than the `{data, pagination}` shape used by the
// observability plane. Kept 1:1 with the backend so conversions stay at
// the render layer.
export interface PromptListResponse {
  data: PromptListItem[]
  total: number
  page: number
  limit: number
}

export interface CreatePromptRequest {
  name: string
  type?: PromptType
  description?: string
  tags?: string[]
  template: PromptTemplate
  config?: ModelConfig
  labels?: string[]
  commit_message?: string
}

export interface CreateVersionRequest {
  template: PromptTemplate
  config?: ModelConfig
  labels?: string[]
  commit_message?: string
}

// Form-local state for the shared prompt form. `body` is the raw
// template string (text prompts only — chat editing is deferred).
export interface PromptFormState {
  name: string
  type: PromptType
  tagsCsv: string
  body: string
  variablesCsv: string
  commitMessage: string
}

// Utility: whether a template is a plain text template.
export function isTextTemplate(t: PromptTemplate): t is TextTemplate {
  return typeof (t as TextTemplate).content === 'string'
}

// Extract the raw template body for display/edit. Chat templates are
// rendered as JSON for now since the form UI handles only text.
export function templateToBody(t: PromptTemplate): string {
  if (isTextTemplate(t)) return t.content
  return JSON.stringify(t, null, 2)
}
