// Wire types for the annotation-queues list endpoint at
// GET /api/v1/projects/{projectId}/annotation-queues. The list view
// returns queue+stats pairs so we can show item counts without a
// per-row stats roundtrip. Assignments are a separate endpoint; we
// surface assignee count only when a follow-up port wires that in.
//
// Envelope: flat page-based — `{data, total, page, limit}`. No
// has_next/has_prev; derive at the render layer.

export type QueueStatus = 'active' | 'paused' | 'archived'

export interface AnnotationQueue {
  id: string
  project_id: string
  name: string
  description?: string
  instructions?: string
  score_config_ids: string[]
  status: QueueStatus
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
// Returns `QueueWithStatsResponse` per the annotation handler. The plain
// `/...{queueId}` endpoint exists too but returns only the queue; we
// prefer the /stats variant so detail and list share the same aggregate
// shape.
export type QueueDetail = QueueWithStats

// Items endpoint — GET /api/v1/projects/{projectId}/annotation-queues/{queueId}/items.
// Returns a flat pageList envelope `{data, total, page, limit}`. `status`
// is a server-side filter (pending|completed|skipped per the backend
// enum) — enforce at the route search-schema layer.
export type QueueItemStatus = 'pending' | 'completed' | 'skipped'

export interface QueueItem {
  id: string
  queue_id: string
  // W3C hex trace/span ID — NOT a uuid.
  object_id: string
  object_type: 'trace' | 'span'
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

// Claim / complete / skip — item lifecycle during review.
// Endpoints:
//   POST /api/v1/projects/{projectId}/annotation-queues/{queueId}/items/claim
//   POST .../items/{itemId}/complete
//   POST .../items/{itemId}/skip
//
// ClaimNextRequest.seen_item_ids carries IDs the reviewer has already
// touched in this session so the server picks a different one. Empty
// array on first claim, grows as the reviewer progresses.
export interface ClaimNextRequest {
  seen_item_ids?: string[]
}

// Polymorphic value per the backend — number for NUMERIC/BOOLEAN
// (BOOLEAN encodes true=1/false=0 at the server; we pass the raw
// boolean here and let the wire serializer honour the polymorphism),
// string for CATEGORICAL. Validated server-side against the referenced
// score config's data type.
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
