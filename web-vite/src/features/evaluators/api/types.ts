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
