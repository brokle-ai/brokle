export interface Dataset {
  id: string
  project_id: string
  name: string
  description?: string
  metadata?: Record<string, unknown>
  current_version_id?: string
  created_at: string
  updated_at: string
}

// Dataset Versioning Types
export interface DatasetVersion {
  id: string
  dataset_id: string
  version: number
  item_count: number
  description?: string
  metadata?: Record<string, unknown>
  created_by?: string
  created_at: string
}

export interface DatasetVersionResponse {
  id: string
  dataset_id: string
  version: number
  item_count: number
  description?: string
  metadata?: Record<string, unknown>
  created_by?: string
  created_at: string
}

export interface DatasetWithVersionInfo {
  id: string
  project_id: string
  name: string
  description?: string
  metadata?: Record<string, unknown>
  current_version_id?: string
  current_version?: number
  latest_version?: number
  created_at: string
  updated_at: string
}

export interface CreateDatasetVersionRequest {
  description?: string
  metadata?: Record<string, unknown>
}

export interface PinDatasetVersionRequest {
  version_id?: string | null
}

export type DatasetItemSource = 'manual' | 'trace' | 'span' | 'csv' | 'json' | 'sdk'

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

export interface CreateDatasetRequest {
  name: string
  description?: string
  metadata?: Record<string, unknown>
}

export interface UpdateDatasetRequest {
  name?: string
  description?: string
  metadata?: Record<string, unknown>
}

export interface CreateDatasetItemRequest {
  input: Record<string, unknown>
  expected?: Record<string, unknown>
  metadata?: Record<string, unknown>
}

export interface DatasetItemListResponse {
  items: DatasetItem[]
  total: number
}

// Import/Export Types
export interface KeysMapping {
  input_keys?: string[]
  expected_keys?: string[]
  metadata_keys?: string[]
}

export interface BulkImportResult {
  created: number
  skipped: number
  errors?: string[]
}

export interface ImportFromJsonRequest {
  items: Record<string, unknown>[]
  keys_mapping?: KeysMapping
  deduplicate?: boolean
}

export interface ImportFromTracesRequest {
  trace_ids: string[]
  keys_mapping?: KeysMapping
  deduplicate?: boolean
}

export interface ImportFromSpansRequest {
  span_ids: string[]
  keys_mapping?: KeysMapping
  deduplicate?: boolean
}

// CSV Import Types
export interface CSVColumnMapping {
  input_column: string
  expected_column?: string
  metadata_columns?: string[]
}

export interface ImportFromCsvRequest {
  content: string
  column_mapping: CSVColumnMapping
  has_header: boolean
  deduplicate: boolean
}

// For client-side CSV preview
export interface CsvPreview {
  headers: string[]
  rows: string[][]
  rowCount: number
}
