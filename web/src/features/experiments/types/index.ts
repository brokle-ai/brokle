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

export interface ExperimentItemListResponse {
  items: ExperimentItem[]
  total: number
}
