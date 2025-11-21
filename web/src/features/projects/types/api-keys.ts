/**
 * API Keys Types
 * TypeScript definitions for API key management matching backend exactly
 */

/**
 * API Key entity returned from backend
 * Matches backend struct from internal/core/domain/auth/auth.go
 */
export interface APIKey {
  id: string
  name: string
  key?: string // Only present on creation (one-time view)
  key_preview: string // Format: bk_AbCd...XyZa
  project_id: string
  status: 'active' | 'expired' // inactive removed - delete is now the only action
  last_used?: string // ISO 8601 timestamp
  created_at: string // ISO 8601 timestamp
  expires_at?: string // ISO 8601 timestamp
  created_by: string // User ID who created the key
}

/**
 * Request payload for creating a new API key
 * Matches CreateAPIKeyRequest from internal/transport/http/handlers/apikey/apikey.go
 */
export interface CreateAPIKeyRequest {
  name: string // 2-100 characters
  expiry_option: '30days' | '90days' | 'never'
}

/**
 * Filter options for listing API keys
 * Used as query parameters for GET /api/v1/projects/:projectId/api-keys
 */
export interface APIKeyFilters {
  status?: 'active' | 'expired' // inactive removed
  page?: number
  limit?: number // 10, 25, 50, 100
  sort_by?: 'created_at' | 'name' | 'last_used_at'
  sort_dir?: 'asc' | 'desc'
}

/**
 * Backend response format (snake_case)
 * Used for API response mapping
 */
export interface BackendAPIKey {
  id: string
  name: string
  key?: string
  key_preview: string
  project_id: string
  status: 'active' | 'expired' // inactive removed
  last_used?: string
  created_at: string
  expires_at?: string
  created_by: string
}

/**
 * Paginated list response from backend
 */
export interface APIKeyListResponse {
  data: APIKey[]
  meta: {
    pagination: {
      page: number
      limit: number
      total: number
      total_pages: number
      has_next: boolean
      has_prev: boolean
    }
    request_id: string
    timestamp: string
  }
}

/**
 * Single API key response from backend
 */
export interface APIKeyResponse {
  data: APIKey
  meta: {
    request_id: string
    timestamp: string
  }
}
