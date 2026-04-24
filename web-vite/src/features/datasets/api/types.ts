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
