// Wire types for the annotation-queues list endpoint at
// GET /api/v1/projects/{projectId}/annotation-queues. The list view
// returns queue+stats pairs so we can show item counts without a
// per-row stats roundtrip. Assignments are a separate endpoint.
//
// Envelope: flat page-based — `{data, total, page, limit}`. No
// has_next/has_prev; derive at the render layer.

export type QueueStatus = 'active' | 'paused' | 'archived'
export type ObjectType = 'trace' | 'span'
export type AssignmentRole = 'annotator' | 'reviewer' | 'admin'

export interface QueueSettings {
  lock_timeout_seconds?: number
  auto_assignment?: boolean
}

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

export interface QueueStats {
  total_items: number
  pending_items: number
  in_progress_items: number
  completed_items: number
  skipped_items: number
}

export interface QueueWithStats {
  queue: AnnotationQueue
  stats: QueueStats
}

export interface QueueListResponse {
  data: QueueWithStats[]
  total: number
  page: number
  limit: number
}

// Detail endpoint — GET /api/v1/projects/{projectId}/annotation-queues/{queueId}/stats.
export type QueueDetail = QueueWithStats

// Items endpoint — GET .../items. Flat page envelope.
export type QueueItemStatus = 'pending' | 'completed' | 'skipped'

export interface QueueItem {
  id: string
  queue_id: string
  // W3C hex trace/span ID — NOT a uuid.
  object_id: string
  object_type: ObjectType
  status: string
  priority: number
  locked_at?: string
  locked_by_user_id?: string
  annotator_user_id?: string
  completed_at?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface QueueItemsListResponse {
  data: QueueItem[]
  total: number
  page: number
  limit: number
}

// Item lifecycle requests.
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

// Queue CRUD requests.
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

// Bulk add — POST /items with a `{items}` body.
export interface AddQueueItemRequest {
  object_id: string
  object_type: ObjectType
  priority?: number
  metadata?: Record<string, unknown>
}

export interface AddItemsBatchRequest {
  items: AddQueueItemRequest[]
}

export interface BatchAddItemsResponse {
  created: number
}

// Assignments.
export interface QueueAssignment {
  id: string
  queue_id: string
  user_id: string
  role: AssignmentRole
  assigned_at: string
  assigned_by?: string
}

export interface AssignUserRequest {
  user_id: string
  role: AssignmentRole
}
