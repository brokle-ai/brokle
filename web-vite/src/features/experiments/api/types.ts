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

// Detail response — GET /api/v1/projects/{projectId}/experiments/{experimentId}
// returns the same `evaluationDomain.ExperimentResponse` as the list.
export type ExperimentDetail = ExperimentListItem

// Experiment items live under `.../experiments/{id}/items` and return
// `{items, total}` (flat — this endpoint predates the pageList wrapper).
// `limit` / `offset` drive pagination on this endpoint, unlike the
// page/limit convention everywhere else in evaluation.
export interface ExperimentItem {
  id: string
  experiment_id: string
  dataset_item_id?: string
  trace_id?: string
  input: Record<string, unknown>
  output?: unknown
  expected?: unknown
  trial_number: number
  metadata?: Record<string, unknown>
  error?: string
  created_at: string
}

export interface ExperimentItemListResponse {
  items: ExperimentItem[]
  total: number
}

// Metrics response — GET .../experiments/{id}/metrics. Provides progress
// rollups (success/error rate) and per-score aggregations we surface on
// the detail summary row.
export interface ExperimentProgressMetrics {
  total_items: number
  completed_items: number
  failed_items: number
  pending_items: number
  progress_pct: number
  success_rate: number
  error_rate: number
}

export interface ExperimentPerformanceMetrics {
  started_at?: string
  completed_at?: string
  elapsed_seconds?: number
  eta_seconds?: number
}

export interface ScoreMetrics {
  mean: number
  std_dev: number
  min: number
  max: number
  count: number
}

export interface ExperimentMetricsResponse {
  experiment_id: string
  status: ExperimentStatus
  progress: ExperimentProgressMetrics
  performance: ExperimentPerformanceMetrics
  scores?: Record<string, ScoreMetrics>
}

// Create request shape. Mirrors
// `evaluationDomain.CreateExperimentRequest` at
// internal/core/domain/evaluation/entities.go:463. The full-fat wizard
// path (CreateExperimentFromWizardRequest — prompt_id, version_id,
// evaluators, variable_mapping) is NOT wired on the dashboard plane
// today — no `create-experiment-wizard` Huma operation exists — so
// the v1 form is intentionally minimal: name + optional dataset +
// optional description. Evaluator binding, model, and "run
// immediately" are deferred until the wizard endpoint lands.
export interface CreateExperimentRequest {
  name: string
  dataset_id?: string
  description?: string
  metadata?: Record<string, unknown>
}

// Rerun body — matches `RerunExperimentRequest` in entities.go:472.
// All fields optional; empty body re-runs with the same name.
export interface RerunExperimentRequest {
  name?: string
  description?: string
  metadata?: Record<string, unknown>
}
