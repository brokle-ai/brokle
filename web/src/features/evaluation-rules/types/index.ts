export type RuleStatus = 'active' | 'inactive' | 'paused'
export type RuleTrigger = 'on_span_complete'
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

export interface EvaluationRule {
  id: string
  project_id: string
  name: string
  description?: string
  status: RuleStatus
  trigger_type: RuleTrigger
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

export interface CreateEvaluationRuleRequest {
  name: string
  description?: string
  status?: RuleStatus
  trigger_type?: RuleTrigger
  target_scope?: TargetScope
  filter?: FilterClause[]
  span_names?: string[]
  sampling_rate?: number
  scorer_type: ScorerType
  scorer_config: ScorerConfig
  variable_mapping?: VariableMap[]
}

export interface UpdateEvaluationRuleRequest {
  name?: string
  description?: string
  status?: RuleStatus
  trigger_type?: RuleTrigger
  target_scope?: TargetScope
  filter?: FilterClause[]
  span_names?: string[]
  sampling_rate?: number
  scorer_type?: ScorerType
  scorer_config?: ScorerConfig
  variable_mapping?: VariableMap[]
}

export interface RuleListResponse {
  rules: EvaluationRule[]
  total: number
  page: number
  limit: number
}

export interface RuleListParams {
  page?: number
  limit?: number
  status?: RuleStatus
  scorer_type?: ScorerType
  search?: string
}
