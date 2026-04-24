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
