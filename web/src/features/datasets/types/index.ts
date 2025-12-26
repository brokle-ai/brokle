export interface Dataset {
  id: string
  project_id: string
  name: string
  description?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface DatasetItem {
  id: string
  dataset_id: string
  input: Record<string, unknown>
  expected?: Record<string, unknown>
  metadata?: Record<string, unknown>
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
