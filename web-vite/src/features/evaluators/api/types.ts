// Wire types for the evaluators dashboard API.
//
// Shape comes from internal/transport/http/handlers/evaluation/* and
// internal/core/domain/evaluation/*. The list endpoint emits the
// canonical `pageList[*EvaluatorResponse]` (`{data, total, page, limit}`)
// while the executions endpoint emits a legacy
// `{executions, total, page, limit}` shape — see `execution_types.go`.

export type EvaluatorStatus = 'active' | 'inactive' | 'paused'
export type EvaluatorTrigger = 'automatic' | 'manual'
export type TargetScope = 'span' | 'trace'
export type ScorerType = 'llm' | 'builtin' | 'regex'

// Operator vocabulary mirrors the evaluator filter clause domain
// (internal/core/domain/evaluation/rule.go FilterOperator constants).
export type FilterOperator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'gt'
  | 'lt'
  | 'gte'
  | 'lte'
  | 'is_empty'
  | 'is_not_empty'

export interface FilterClause {
  field: string
  operator: FilterOperator
  value: unknown
}

export interface VariableMap {
  variable_name: string
  source: 'span_input' | 'span_output' | 'span_metadata' | 'trace_input'
  json_path: string
}

export interface LLMMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
}

export interface OutputField {
  name: string
  type: 'numeric' | 'categorical' | 'boolean'
  description?: string
  min_value?: number
  max_value?: number
  categories?: string[]
}

export interface LLMScorerConfig {
  credential_id: string
  model: string
  messages: LLMMessage[]
  temperature: number
  response_format: 'json' | 'text'
  output_schema: OutputField[]
}

export interface BuiltinScorerConfig {
  scorer_name:
    | 'contains'
    | 'json_valid'
    | 'length_check'
    | 'sentiment'
    | 'toxicity'
  config: Record<string, unknown>
}

export interface RegexScorerConfig {
  pattern: string
  score_name: string
  match_score?: number
  no_match_score?: number
  capture_group?: number
}

export type ScorerConfig =
  | LLMScorerConfig
  | BuiltinScorerConfig
  | RegexScorerConfig

export interface EvaluatorListItem {
  id: string
  project_id: string
  name: string
  description?: string
  status: EvaluatorStatus
  trigger_type: EvaluatorTrigger
  target_scope: TargetScope
  filter: FilterClause[]
  span_names: string[]
  sampling_rate: number
  scorer_type: ScorerType
  scorer_config: Record<string, unknown>
  variable_mapping: VariableMap[]
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
// Shares wire shape with list items.
export type EvaluatorDetail = EvaluatorListItem

// Backward-compat alias for the LLM scorer config shape used by
// existing detail-view code paths.
export type LLMScorerConfigShape = LLMScorerConfig

// Create/update request shapes.
export interface CreateEvaluatorRequest {
  name: string
  description?: string
  status?: EvaluatorStatus
  trigger_type?: EvaluatorTrigger
  target_scope?: TargetScope
  filter?: FilterClause[]
  span_names?: string[]
  sampling_rate?: number
  scorer_type: ScorerType
  scorer_config: Record<string, unknown>
  variable_mapping?: VariableMap[]
}

export interface UpdateEvaluatorRequest {
  name?: string
  description?: string
  status?: EvaluatorStatus
  trigger_type?: EvaluatorTrigger
  target_scope?: TargetScope
  filter?: FilterClause[]
  span_names?: string[]
  sampling_rate?: number
  scorer_type?: ScorerType
  scorer_config?: Record<string, unknown>
  variable_mapping?: VariableMap[]
}

// Test-evaluator endpoint shapes — POST /evaluators/{id}/test.
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

// ============================================================================
// Execution types — GET /evaluators/{id}/executions[/...].
// Domain shapes from internal/core/domain/evaluation/{execution,analytics}.go.
// ============================================================================

export type ExecutionStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'

export type TriggerType = 'automatic' | 'manual'

export interface EvaluatorExecution {
  id: string
  evaluator_id: string
  project_id: string
  status: ExecutionStatus
  trigger_type: TriggerType
  spans_matched: number
  spans_scored: number
  errors_count: number
  error_message?: string
  started_at?: string
  completed_at?: string
  duration_ms?: number
  metadata?: Record<string, unknown>
  created_at: string
}

// Backend list endpoint emits a legacy envelope, not the canonical
// pageList shape — see internal/transport/http/handlers/evaluation/
// execution_types.go ExecutionListResponse.
export interface ExecutionListResponse {
  executions: EvaluatorExecution[]
  total: number
  page: number
  limit: number
}

export interface ExecutionListParams {
  page?: number
  limit?: number
  status?: ExecutionStatus
  trigger_type?: TriggerType
}

// Span-level execution detail — only present on the `/detail` endpoint.
export interface ResolvedVariable {
  variable_name: string
  source: string
  json_path?: string
  resolved_value: unknown
}

export interface ExecutionScoreResult {
  score_name: string
  value: number | string | boolean
  reasoning?: string
  confidence?: number
  raw_output?: unknown
}

export interface SpanExecutionDetail {
  span_id: string
  trace_id: string
  span_name: string
  status: 'success' | 'failed' | 'skipped'
  score_results: ExecutionScoreResult[]
  prompt_sent?: LLMMessage[]
  llm_response_raw?: string
  llm_response_parsed?: Record<string, unknown>
  variables_resolved: ResolvedVariable[]
  error_message?: string
  error_stack?: string
  latency_ms?: number
  created_at: string
}

export interface EvaluatorSnapshot {
  name: string
  scorer_type: ScorerType
  scorer_config: Record<string, unknown>
  variable_mapping: VariableMap[]
  filter: FilterClause[]
}

export interface EvaluatorExecutionDetail extends EvaluatorExecution {
  spans: SpanExecutionDetail[]
  evaluator_snapshot?: EvaluatorSnapshot
}

// ============================================================================
// Analytics — GET /evaluators/{id}/analytics.
// ============================================================================

export type AnalyticsPeriod = '24h' | '7d' | '30d' | '90d'

export interface EvaluatorAnalyticsParams {
  period?: AnalyticsPeriod
  from_timestamp?: string
  to_timestamp?: string
}

export interface DistributionBucket {
  bin_start: number
  bin_end: number
  count: number
  percentage?: number
}

export interface TimeSeriesPoint {
  timestamp: string
  count: number
  // success_rate is 0..1 from the backend.
  success_rate: number
  // avg_score is nullable on the wire — domain `*float64` JSON-marshals
  // to `null` when absent.
  avg_score?: number | null
}

export interface LatencyStats {
  p50: number
  p90: number
  p99: number
  avg: number
  max?: number
  min?: number
}

export interface ErrorSummary {
  error_type: string
  message: string
  count: number
  last_occurred: string
}

export interface CostEstimate {
  total_cost: number
  input_tokens: number
  output_tokens: number
  estimated_monthly: number
}

export interface EvaluatorAnalyticsResponse {
  evaluator_id: string
  period: string
  total_executions: number
  total_spans_scored?: number
  success_rate: number
  // Backend domain emits float64 (zero when no scores) — we keep
  // `| null` for forward-compat with potential nullable upgrade.
  average_score: number | null
  score_distribution: DistributionBucket[]
  execution_trend: TimeSeriesPoint[]
  score_trend?: TimeSeriesPoint[]
  latency_percentiles: LatencyStats
  top_errors: ErrorSummary[]
  cost_estimate?: CostEstimate
}

// ============================================================================
// Manual trigger — POST /evaluators/{id}/trigger.
// ============================================================================

export interface TriggerOptions {
  time_range_start?: string
  time_range_end?: string
  span_ids?: string[]
  sample_limit?: number
}

export interface TriggerResponse {
  execution_id: string
  message: string
}
