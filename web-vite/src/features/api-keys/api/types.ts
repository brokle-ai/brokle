// Wire types for the project API-keys list endpoint. Shape comes from
// internal/transport/http/handlers/apikey/handlers.go `apiKey` and
// `listAPIKeysResponse` at GET /api/v1/projects/{projectId}/api-keys.
//
// The full key value (`key`) is NEVER present on list responses — it
// only appears once on create. List rows carry only the preview
// (`bk_AbCd...XyZa`) and metadata.

export interface ApiKeyListItem {
  id: string
  name: string
  key_preview: string
  project_id: string
  status: 'active' | 'expired'
  // RFC 3339 timestamps on the wire; components format on render.
  last_used?: string
  created_at: string
  expires_at?: string
  created_by: string
}

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface ApiKeyListResponse {
  data: ApiKeyListItem[]
  pagination: Pagination
}
