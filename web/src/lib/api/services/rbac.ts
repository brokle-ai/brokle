/**
 * RBAC API Service
 * Handles scope-based authorization API calls
 */

import { BrokleAPIClient } from '../core/client'

const client = new BrokleAPIClient('/api/v1')

// ========================================
// Types
// ========================================

export interface CheckUserScopesRequest {
  userId: string
  organizationId: string
  projectId: string | null
  scopes: string[]
}

export interface CheckUserScopesResponse {
  [scope: string]: boolean  // scope name → has access
}

export interface UserScopesResponse {
  userId: string
  organizationId: string | null
  projectId: string | null
  globalScopes: string[]
  organizationScopes: string[]
  projectScopes: string[]
  effectiveScopes: string[]
}

export interface ScopeCategory {
  name: string
  displayName: string
  description: string
  level: 'organization' | 'project' | 'global'
  scopes: string[]
}

// ========================================
// API Functions
// ========================================

/**
 * Check if user has specific scopes in the given context
 *
 * POST /api/v1/rbac/users/scopes/check
 *
 * @param request - User, context, and scopes to check
 * @returns Map of scope → boolean (true if user has scope)
 */
export async function checkUserScopes(
  request: CheckUserScopesRequest
): Promise<CheckUserScopesResponse> {
  return await client.post<CheckUserScopesResponse>(
    `/rbac/users/${request.userId}/scopes/check`,
    {
      organization_id: request.organizationId,
      project_id: request.projectId,
      scopes: request.scopes,
    }
  )
}

/**
 * Get all effective scopes for a user in the given context
 *
 * GET /api/v1/rbac/users/:userId/scopes
 *
 * @param userId - User ID
 * @param organizationId - Organization ID (optional)
 * @param projectId - Project ID (optional)
 * @returns User's scopes broken down by level
 */
export async function getUserScopes(
  userId: string,
  organizationId?: string,
  projectId?: string
): Promise<UserScopesResponse> {
  const params: Record<string, string> = {}
  if (organizationId) params.organization_id = organizationId
  if (projectId) params.project_id = projectId

  return await client.get<UserScopesResponse>(`/rbac/users/${userId}/scopes`, params)
}

/**
 * Get available scopes grouped by category (for UI)
 *
 * GET /api/v1/rbac/scopes/categories
 *
 * @returns List of scope categories with their scopes
 */
export async function getScopeCategories(): Promise<ScopeCategory[]> {
  return await client.get<ScopeCategory[]>('/rbac/scopes/categories')
}

/**
 * Get all available scopes for a specific level
 *
 * GET /api/v1/rbac/scopes?level=organization
 *
 * @param level - Scope level (organization, project, global)
 * @returns List of scope names
 */
export async function getAvailableScopes(
  level?: 'organization' | 'project' | 'global'
): Promise<string[]> {
  const params = level ? { level } : {}
  return await client.get<string[]>('/rbac/scopes', params)
}

// ========================================
// Helper Functions
// ========================================

/**
 * Batch check multiple users' scopes (for admin dashboards)
 */
export interface BatchCheckScopesRequest {
  checks: Array<{
    userId: string
    organizationId: string
    projectId: string | null
    scopes: string[]
  }>
}

export interface BatchCheckScopesResponse {
  results: Array<{
    userId: string
    scopes: Record<string, boolean>
  }>
}

export async function batchCheckScopes(
  request: BatchCheckScopesRequest
): Promise<BatchCheckScopesResponse> {
  return await client.post<BatchCheckScopesResponse>('/rbac/scopes/batch-check', request)
}
