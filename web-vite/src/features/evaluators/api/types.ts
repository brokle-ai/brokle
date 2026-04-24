// Wire types for the evaluators list endpoint.
//
// Shape comes from internal/transport/http/handlers/evaluation/evaluator.go
// (listEvaluators) emitting `pageList[*EvaluatorResponse]` →
// `{data, total, page, limit}`. Filter[] / ScorerConfig / VariableMapping
// are intentionally widened to `unknown` here — the list view only needs
// summary fields; detail-view port will narrow them against domain types.

export type EvaluatorStatus = 'active' | 'inactive' | 'paused'
export type EvaluatorTrigger = 'automatic' | 'manual'
export type TargetScope = 'span' | 'trace'
export type ScorerType = 'llm' | 'builtin' | 'regex'

export interface EvaluatorListItem {
  id: string
  project_id: string
  name: string
  description?: string
  status: EvaluatorStatus
  trigger_type: EvaluatorTrigger
  target_scope: TargetScope
  filter: unknown[]
  span_names: string[]
  sampling_rate: number
  scorer_type: ScorerType
  scorer_config: Record<string, unknown>
  variable_mapping: unknown[]
  created_by?: string
  created_at: string
  updated_at: string
}

export interface EvaluationPageList<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

export type EvaluatorListResponse = EvaluationPageList<EvaluatorListItem>

// Detail response — GET /api/v1/projects/{projectId}/evaluators/{evaluatorId}.
// Matches `evaluationDomain.EvaluatorResponse` exactly; the backend list
// endpoint emits the same struct, so detail and list share the shape.
// Scorer config is a provider-agnostic map because it's one of three
// discriminated union variants (LLM / builtin / regex) and the backend
// decodes per `scorer_type` at write time.
export type EvaluatorDetail = EvaluatorListItem

// LLM scorer config shape — only populated when `scorer_type === 'llm'`.
// We keep the cast fully at the read site (not a runtime validator) and
// document the wire fields here so the detail view's prompt-render
// panel has a typed target.
export interface LLMScorerMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
}

export interface LLMScorerOutputField {
  name: string
  type: 'numeric' | 'categorical' | 'boolean'
  description?: string
  min_value?: number
  max_value?: number
  categories?: string[]
}

export interface LLMScorerConfigShape {
  credential_id: string
  model: string
  messages: LLMScorerMessage[]
  temperature: number
  response_format: 'json' | 'text'
  output_schema: LLMScorerOutputField[]
}
