// Annotation Queue Types

// Enums
export type QueueStatus = 'active' | 'paused' | 'archived'
export type ItemStatus = 'pending' | 'completed' | 'skipped'
export type ObjectType = 'trace' | 'span'
export type AssignmentRole = 'annotator' | 'reviewer' | 'admin'

// Queue Settings
export interface QueueSettings {
  lock_timeout_seconds?: number
  auto_assignment?: boolean
}

// Annotation Queue
export interface AnnotationQueue {
  id: string
  project_id: string
  name: string
  description?: string
  instructions?: string
  score_config_ids: string[]
  status: QueueStatus
  settings?: QueueSettings
  created_by?: string
  created_at: string
  updated_at: string
}

// Queue Statistics
export interface QueueStats {
  total_items: number
  pending_items: number
  in_progress_items: number
  completed_items: number
  skipped_items: number
}

// Queue with Stats (for list views)
export interface QueueWithStats {
  queue: AnnotationQueue
  stats: QueueStats
}

// Queue Item
export interface QueueItem {
  id: string
  queue_id: string
  object_id: string
  object_type: ObjectType
  status: ItemStatus
  priority: number
  locked_at?: string
  locked_by_user_id?: string
  annotator_user_id?: string
  completed_at?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

// Queue Assignment
export interface QueueAssignment {
  id: string
  queue_id: string
  user_id: string
  role: AssignmentRole
  assigned_at: string
  assigned_by?: string
}

// Request Types

export interface CreateQueueRequest {
  name: string
  description?: string
  instructions?: string
  score_config_ids?: string[]
  settings?: QueueSettings
}

export interface UpdateQueueRequest {
  name?: string
  description?: string
  instructions?: string
  score_config_ids?: string[]
  status?: QueueStatus
  settings?: QueueSettings
}

export interface AddItemRequest {
  object_id: string
  object_type: ObjectType
  priority?: number
  metadata?: Record<string, unknown>
}

export interface AddItemsBatchRequest {
  items: AddItemRequest[]
}

export interface ClaimNextRequest {
  seen_item_ids?: string[]
}

export interface ScoreSubmission {
  score_config_id: string
  value: number | string | boolean
  comment?: string
}

export interface CompleteItemRequest {
  scores?: ScoreSubmission[]
}

export interface SkipItemRequest {
  reason?: string
}

export interface AssignUserRequest {
  user_id: string
  role: AssignmentRole
}

// Response Types

export interface BatchAddItemsResponse {
  created: number
}

// Filter Types for List Operations

export interface QueueListFilter {
  status?: QueueStatus
  page?: number
  limit?: number
  search?: string
}

export interface ItemListFilter {
  status?: ItemStatus
  page?: number
  limit?: number
}
