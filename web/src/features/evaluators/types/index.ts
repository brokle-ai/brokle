export type EvaluatorStatus = 'active' | 'inactive' | 'paused'
export type EvaluatorTrigger = 'on_span_complete'
export type TargetScope = 'span' | 'trace'
export type ScorerType = 'llm' | 'builtin' | 'regex'

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
  scorer_name: 'contains' | 'json_valid' | 'length_check' | 'sentiment' | 'toxicity'
  config: Record<string, unknown>
}

export interface RegexScorerConfig {
  pattern: string
  score_name: string
  match_score?: number
  no_match_score?: number
  capture_group?: number
}

export type ScorerConfig = LLMScorerConfig | BuiltinScorerConfig | RegexScorerConfig

export interface Evaluator {
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
  scorer_config: ScorerConfig
  variable_mapping: VariableMap[]
  created_by?: string
  created_at: string
  updated_at: string
}

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
  scorer_config: ScorerConfig
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
  scorer_config?: ScorerConfig
  variable_mapping?: VariableMap[]
}

// Note: EvaluatorListResponse removed - use PaginatedResponse<Evaluator> from @/lib/api/core/types

export interface EvaluatorListParams {
  page?: number
  limit?: number
  status?: EvaluatorStatus
  scorer_type?: ScorerType
  search?: string
  sort_by?: 'name' | 'status' | 'sampling_rate' | 'created_at' | 'updated_at'
  sort_dir?: 'asc' | 'desc'
}

// ============================================================================
// Evaluator Execution Types (for execution history tracking)
// ============================================================================

export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
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

// Note: ExecutionListResponse removed - use PaginatedResponse<EvaluatorExecution> from @/lib/api/core/types

export interface ExecutionListParams {
  page?: number
  limit?: number
  status?: ExecutionStatus
  trigger_type?: TriggerType
}

// ============================================================================
// Manual Trigger Types
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

// ============================================================================
// Evaluator Analytics Types
// ============================================================================

export interface EvaluatorAnalyticsParams {
  period?: '24h' | '7d' | '30d' | '90d'
  from_timestamp?: string
  to_timestamp?: string
}

export interface DistributionBucket {
  bin_start: number
  bin_end: number
  count: number
  percentage?: number // Optional: provided by backend for histogram display
}

export interface TimeSeriesPoint {
  timestamp: string
  count: number
  avg_score?: number
  success_rate: number
}

export interface LatencyStats {
  p50: number
  p90: number
  p99: number
  avg: number
  max?: number // Optional: provided by backend
  min?: number // Optional: provided by backend
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
  total_spans_scored?: number // Optional: provided by backend
  success_rate: number
  average_score: number | null
  score_distribution: DistributionBucket[]
  execution_trend: TimeSeriesPoint[]
  score_trend?: TimeSeriesPoint[] // Optional: provided by backend
  latency_percentiles: LatencyStats
  top_errors: ErrorSummary[]
  cost_estimate?: CostEstimate // Optional: only for LLM evaluators
}

// ============================================================================
// Execution Detail Types (for debugging and inspection)
// ============================================================================

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

export interface EvaluatorExecutionDetail extends EvaluatorExecution {
  spans: SpanExecutionDetail[]
  evaluator_snapshot?: {
    name: string
    scorer_type: ScorerType
    scorer_config: ScorerConfig
    variable_mapping: VariableMap[]
    filter: FilterClause[]
  }
}

// ============================================================================
// Test Evaluator Types (for testing evaluators before activation)
// ============================================================================

export interface TestSampleInput {
  input?: string
  output?: string
  metadata?: Record<string, unknown>
}

export interface TestEvaluatorRequest {
  trace_id?: string
  span_id?: string
  span_ids?: string[]
  limit?: number
  time_range?: string
  sample_input?: TestSampleInput
}

export interface TestScoreResult {
  score_name: string
  value: number | string | boolean
  reasoning?: string
  confidence?: number
}

export interface TestExecution {
  span_id: string
  trace_id: string
  span_name: string
  matched_filter: boolean
  status: 'success' | 'failed' | 'skipped' | 'filtered'
  score_results: TestScoreResult[]
  prompt_sent?: LLMMessage[]
  llm_response?: string
  llm_response_parsed?: Record<string, unknown>
  variables_resolved: ResolvedVariable[]
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

export interface EvaluatorPreview {
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
  evaluator_preview: EvaluatorPreview
}
