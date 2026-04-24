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

// Create/update request shapes. Mirrors
// `evaluationDomain.CreateEvaluatorRequest` / `UpdateEvaluatorRequest`
// at internal/core/domain/evaluation/rule.go. `filter` and
// `variable_mapping` are typed `unknown[]` here to match the list DTO
// — the v1 form serialises them as raw JSON strings and relies on the
// backend for schema validation. Filter typing can narrow later when
// the form grows a structured editor.
export interface CreateEvaluatorRequest {
  name: string
  description?: string
  status?: EvaluatorStatus
  trigger_type?: EvaluatorTrigger
  target_scope?: TargetScope
  filter?: unknown[]
  span_names?: string[]
  sampling_rate?: number
  scorer_type: ScorerType
  scorer_config: Record<string, unknown>
  variable_mapping?: unknown[]
}

export interface UpdateEvaluatorRequest {
  name?: string
  description?: string
  status?: EvaluatorStatus
  trigger_type?: EvaluatorTrigger
  target_scope?: TargetScope
  filter?: unknown[]
  span_names?: string[]
  sampling_rate?: number
  scorer_type?: ScorerType
  scorer_config?: Record<string, unknown>
  variable_mapping?: unknown[]
}

// Test-evaluator endpoint shapes — used by the "Test" button on the
// evaluator detail page. Mirrors the backend's
// `/evaluators/{id}/test` POST: the request picks the sample source
// (trace/span IDs, or a time-range + limit for backfill-style sampling)
// and the response returns per-span resolutions plus an aggregate
// summary. Fields typed `unknown` where the variant depends on the
// scorer's output schema (numeric vs categorical vs boolean).

export interface TestEvaluatorRequest {
  trace_id?: string
  span_id?: string
  span_ids?: string[]
  limit?: number
  time_range?: '1h' | '24h' | '7d'
}

export interface TestScoreResult {
  score_name: string
  value: number | string | boolean
  reasoning?: string
  confidence?: number
}

export interface TestResolvedVariable {
  variable_name: string
  source: string
  json_path?: string
  resolved_value: unknown
}

export interface TestExecution {
  span_id: string
  trace_id: string
  span_name: string
  matched_filter: boolean
  status: 'success' | 'failed' | 'skipped' | 'filtered'
  score_results: TestScoreResult[]
  llm_response?: string
  variables_resolved: TestResolvedVariable[]
  error_message?: string
  latency_ms?: number
}

export interface TestSummary {
  total_spans: number
  matched_spans: number
  evaluated_spans: number
  success_count: number
  failure_count: number
  skipped_count: number
  average_score?: number
  average_latency_ms?: number
}

export interface TestEvaluatorPreview {
  name: string
  scorer_type: ScorerType
  filter_description: string
  variable_names: string[]
  prompt_preview?: string
  matching_count?: number
}

export interface TestEvaluatorResponse {
  summary: TestSummary
  executions: TestExecution[]
  evaluator_preview: TestEvaluatorPreview
}
