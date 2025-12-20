import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  PlaygroundSession,
  PlaygroundSessionSummary,
  CreateSessionRequest,
  UpdateSessionRequest,
} from '../types'

const client = new BrokleAPIClient('/api')

// ============================================================================
// Session API
// ============================================================================

/**
 * Create a new playground session
 *
 * All sessions are saved to the database with a name.
 *
 * Backend endpoint: POST /api/v1/projects/:projectId/playground/sessions
 */
export const createSession = async (
  projectId: string,
  data: CreateSessionRequest
): Promise<PlaygroundSession> => {
  return client.post<PlaygroundSession>(
    `/v1/projects/${projectId}/playground/sessions`,
    data
  )
}

/**
 * Get a session by ID
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/playground/sessions/:sessionId
 */
export const getSession = async (
  projectId: string,
  sessionId: string
): Promise<PlaygroundSession> => {
  return client.get<PlaygroundSession>(
    `/v1/projects/${projectId}/playground/sessions/${sessionId}`
  )
}

/**
 * List sessions for a project
 *
 * Backend endpoint: GET /api/v1/projects/:projectId/playground/sessions
 */
export const listSessions = async (
  projectId: string,
  params?: { limit?: number; tags?: string[] }
): Promise<PlaygroundSessionSummary[]> => {
  const queryParams: Record<string, string | number> = {}

  if (params?.limit) queryParams.limit = params.limit
  if (params?.tags && params.tags.length > 0) queryParams.tags = params.tags.join(',')

  return client.get<PlaygroundSessionSummary[]>(
    `/v1/projects/${projectId}/playground/sessions`,
    queryParams
  )
}

/**
 * Update a session's content and metadata
 *
 * Backend endpoint: PUT /api/v1/projects/:projectId/playground/sessions/:sessionId
 */
export const updateSession = async (
  projectId: string,
  sessionId: string,
  data: UpdateSessionRequest
): Promise<PlaygroundSession> => {
  return client.put<PlaygroundSession>(
    `/v1/projects/${projectId}/playground/sessions/${sessionId}`,
    data
  )
}

/**
 * Delete a session
 *
 * Backend endpoint: DELETE /api/v1/projects/:projectId/playground/sessions/:sessionId
 */
export const deleteSession = async (
  projectId: string,
  sessionId: string
): Promise<void> => {
  await client.delete(`/v1/projects/${projectId}/playground/sessions/${sessionId}`)
}
