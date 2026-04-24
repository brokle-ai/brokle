// Wire types for the experiments list endpoint.
//
// Shape comes from internal/transport/http/handlers/evaluation/experiment.go
// (listExperiments) emitting `pageList[*ExperimentResponse]` →
// `{data, total, page, limit}`.

export type ExperimentStatus =
  | 'draft'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'

export type ExperimentSource = 'dataset' | 'trace' | 'manual'

export interface ExperimentListItem {
  id: string
  project_id: string
  dataset_id?: string
  name: string
  description?: string
  status: ExperimentStatus
  source: ExperimentSource
  config_id?: string
  metadata?: Record<string, unknown>
  total_items: number
  completed_items: number
  failed_items: number
  started_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export interface EvaluationPageList<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

export type ExperimentListResponse = EvaluationPageList<ExperimentListItem>
