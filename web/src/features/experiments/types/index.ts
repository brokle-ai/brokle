export type ExperimentStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'partial'
  | 'cancelled'

export type ExperimentSource = 'sdk' | 'dashboard'

export interface Experiment {
  id: string
  project_id: string
  dataset_id: string | null
  name: string
  description: string | null
  status: ExperimentStatus
  source: ExperimentSource
  config_id: string | null
  metadata: Record<string, unknown>
  total_items: number
  completed_items: number
  failed_items: number
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface ExperimentProgress {
  id: string
  status: ExperimentStatus
  total_items: number
  completed_items: number
  failed_items: number
  pending_items: number
  progress_pct: number
  started_at?: string
  completed_at?: string
  elapsed_seconds?: number
  eta_seconds?: number
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

// Re-export wizard types
export * from './wizard'
