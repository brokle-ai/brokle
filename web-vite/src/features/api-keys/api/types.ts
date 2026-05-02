// Wire types for the project API-keys list / create / delete endpoints.
// Shapes come from internal/transport/http/handlers/apikey/handlers.go —
// `apiKey`, `listAPIKeysResponse`, `createAPIKeyBody`.
//
// The full key value (`key`) is ONLY populated on the create response.
// It's never re-served by list or get — subsequent rows carry only the
// `bk_AbCd…XyZa` preview.

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

// The create response uses the same shape plus the one-shot `key`.
export interface ApiKey extends ApiKeyListItem {
  key: string
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

// Expiry is an enum bucket (30days / 90days / never) — the backend
// resolves it into an actual timestamp. See `createAPIKeyBody`.
export type ApiKeyExpiryOption = '30days' | '90days' | 'never'

export interface CreateApiKeyRequest {
  name: string
  expiry_option: ApiKeyExpiryOption
}
