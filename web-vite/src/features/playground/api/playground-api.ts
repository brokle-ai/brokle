import { rawFetch } from '@/lib/api/client'
import type {
  PlaygroundSession,
  PlaygroundSessionSummary,
  CreateSessionRequest,
  UpdateSessionRequest,
} from '../types'

// Dashboard-plane playground session CRUD. All endpoints return the
// raw resource per the Stripe/OpenAI wire contract (no envelope).

export const createSession = async (
  projectId: string,
  data: CreateSessionRequest,
): Promise<PlaygroundSession> => {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/playground/sessions`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as PlaygroundSession
}

export const getSession = async (
  projectId: string,
  sessionId: string,
): Promise<PlaygroundSession> => {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/playground/sessions/${sessionId}`,
    { method: 'GET' },
  )
  return (await resp.json()) as PlaygroundSession
}

export const listSessions = async (
  projectId: string,
  params?: { limit?: number; tags?: string[] },
): Promise<PlaygroundSessionSummary[]> => {
  const search = new URLSearchParams()
  if (params?.limit) search.set('limit', String(params.limit))
  if (params?.tags && params.tags.length > 0) {
    search.set('tags', params.tags.join(','))
  }
  const qs = search.toString()
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/playground/sessions${qs ? `?${qs}` : ''}`,
    { method: 'GET' },
  )
  return (await resp.json()) as PlaygroundSessionSummary[]
}

export const updateSession = async (
  projectId: string,
  sessionId: string,
  data: UpdateSessionRequest,
): Promise<PlaygroundSession> => {
  const resp = await rawFetch(
    `/api/v1/projects/${projectId}/playground/sessions/${sessionId}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as PlaygroundSession
}

export const deleteSession = async (
  projectId: string,
  sessionId: string,
): Promise<void> => {
  await rawFetch(
    `/api/v1/projects/${projectId}/playground/sessions/${sessionId}`,
    { method: 'DELETE' },
  )
}
