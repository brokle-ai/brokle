// Wire types for the datasets list endpoint. Narrow shape — only the
// fields the list view renders. Detail view / item-level DTOs land in a
// follow-up port.
//
// Shape comes from the backend handler at
// internal/transport/http/handlers/evaluation/dataset.go (dashListDatasets)
// which emits `pageList[*DatasetWithItemCountResponse]` → `{data, total,
// page, limit}`. Timestamps arrive as RFC 3339 strings (no reviver on
// rawFetch), so we keep them string-typed and format at render time.

export interface DatasetListItem {
  id: string
  project_id: string
  name: string
  description?: string
  metadata?: Record<string, unknown>
  current_version_id?: string
  item_count: number
  created_at: string
  updated_at: string
}

// Canonical evaluation-plane list envelope. NOTE: this differs from the
// observability (traces) envelope which nests pagination under
// `pagination: {page, limit, total, total_pages, has_next, has_prev}`.
// The evaluation handlers flatten `{data, total, page, limit}` via the
// shared `pageList[T]` wrapper in
// internal/transport/http/handlers/evaluation/types.go.
export interface EvaluationPageList<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

export type DatasetListResponse = EvaluationPageList<DatasetListItem>

// Detail response at GET /api/v1/projects/{projectId}/datasets/{datasetId}.
// Shape is `evaluationDomain.DatasetResponse` — no item_count (that's a
// list-side rollup), no `updated_at` is omitted here since the endpoint
// always returns it.
export interface DatasetDetail {
  id: string
  project_id: string
  name: string
  description?: string
  metadata?: Record<string, unknown>
  current_version_id?: string
  created_at: string
  updated_at: string
}

// Dataset item source — matches the backend enum. `sdk` appears in
// ingestion paths but the list handler preserves whatever the repo
// persisted, so keep the union open-enough.
export type DatasetItemSource = 'manual' | 'trace' | 'span' | 'csv' | 'json' | 'sdk'

// Dataset item as emitted by the handler DTO in
// internal/transport/http/handlers/evaluation/types.go. `source` is
// pre-stringified server-side.
export interface DatasetItem {
  id: string
  dataset_id: string
  input: Record<string, unknown>
  expected?: Record<string, unknown>
  metadata?: Record<string, unknown>
  source: DatasetItemSource
  source_trace_id?: string
  source_span_id?: string
  created_at: string
}

export type DatasetItemsListResponse = EvaluationPageList<DatasetItem>
