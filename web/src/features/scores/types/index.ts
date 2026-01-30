export * from './analytics'

export type ScoreType = 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN'
export type ScoreSource = 'code' | 'llm' | 'human'

export interface ScoreConfig {
  id: string
  project_id: string
  name: string
  description?: string
  type: ScoreType
  min_value?: number
  max_value?: number
  categories?: string[]
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface Score {
  id: string
  project_id: string
  trace_id?: string
  span_id?: string
  name: string
  value?: number
  string_value?: string
  type: ScoreType
  source: ScoreSource
  reason?: string
  metadata?: Record<string, unknown>
  experiment_id?: string
  experiment_item_id?: string
  timestamp: string
}

export interface CreateScoreConfigRequest {
  name: string
  description?: string
  type: ScoreType
  min_value?: number
  max_value?: number
  categories?: string[]
  metadata?: Record<string, unknown>
}

export interface UpdateScoreConfigRequest {
  name?: string
  description?: string
  type?: ScoreType
  min_value?: number
  max_value?: number
  categories?: string[]
  metadata?: Record<string, unknown>
}

export interface ScoreListParams {
  trace_id?: string
  span_id?: string
  name?: string
  source?: ScoreSource
  type?: ScoreType
  page?: number
  limit?: number
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export interface ScoreConfigListParams {
  page?: number
  limit?: number
}
