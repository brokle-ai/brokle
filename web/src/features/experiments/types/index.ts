export type ExperimentStatus = 'pending' | 'running' | 'completed' | 'failed'

export interface Experiment {
  id: string
  project_id: string
  dataset_id: string | null
  name: string
  description: string | null
  status: ExperimentStatus
  metadata: Record<string, unknown>
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface ExperimentItem {
  id: string
  experiment_id: string
  dataset_item_id: string | null
  trace_id: string | null
  input: Record<string, unknown>
  output: Record<string, unknown> | null
  expected: Record<string, unknown> | null
  trial_number: number
  metadata: Record<string, unknown>
  created_at: string
}

export interface CreateExperimentRequest {
  name: string
  dataset_id?: string
  description?: string
  metadata?: Record<string, unknown>
}

export interface UpdateExperimentRequest {
  name?: string
  description?: string
  status?: ExperimentStatus
  metadata?: Record<string, unknown>
}

export interface RerunExperimentRequest {
  name?: string
  description?: string
  metadata?: Record<string, unknown>
}

// Pagination params for experiments list
export interface ExperimentListParams {
  page?: number
  limit?: number
  dataset_id?: string
  status?: string
  search?: string
  ids?: string // Comma-separated experiment IDs
}

export interface ExperimentItemListResponse {
  items: ExperimentItem[]
  total: number
}

export interface ExperimentScoreStats {
  mean: number
  std_dev: number
  min: number
  max: number
  count: number
}

export interface ExperimentScoreDiff {
  type: 'NUMERIC' | 'CATEGORICAL'
  difference?: number // Absolute difference for numeric
  direction?: '+' | '-' // Direction of change
  isDifferent?: boolean // For categorical comparisons
}

export interface ExperimentComparisonSummary {
  name: string
  status: string
}

export interface ExperimentComparisonResponse {
  experiments: Record<string, ExperimentComparisonSummary>
  scores: Record<string, Record<string, ExperimentScoreStats>>
  diffs?: Record<string, Record<string, ExperimentScoreDiff>>
}

export interface CompareExperimentsRequest {
  experiment_ids: string[]
  baseline_id?: string
}

export interface ScoreComparisonRow {
  scoreName: string
  experiments: Record<
    string,
    {
      stats: ExperimentScoreStats
      diff?: ExperimentScoreDiff
    }
  >
}
