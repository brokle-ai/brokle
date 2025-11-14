import { BrokleAPIClient } from '@/lib/api/core/client'
import type { Trace } from '../data/schema'

const client = new BrokleAPIClient('/api')

export interface GetTracesParams {
  projectSlug: string
  page?: number
  pageSize?: number
  status?: string[]
  search?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

/**
 * Get all traces for a project
 */
export const getProjectTraces = async (params: GetTracesParams): Promise<{
  traces: Trace[]
  totalCount: number
}> => {
  const { projectSlug, ...queryParams } = params
  return client.get(`/v1/projects/${projectSlug}/traces`, queryParams)
}

/**
 * Get a single trace by ID
 */
export const getTraceById = async (
  projectSlug: string,
  traceId: string
): Promise<Trace> => {
  return client.get(`/v1/projects/${projectSlug}/traces/${traceId}`)
}

/**
 * Delete a trace
 */
export const deleteTrace = async (
  projectSlug: string,
  traceId: string
): Promise<void> => {
  return client.delete(`/v1/projects/${projectSlug}/traces/${traceId}`)
}

/**
 * Delete multiple traces
 */
export const deleteMultipleTraces = async (
  projectSlug: string,
  traceIds: string[]
): Promise<void> => {
  return client.post(`/v1/projects/${projectSlug}/traces/bulk-delete`, { traceIds })
}

/**
 * Export traces to CSV
 */
export const exportTraces = async (
  projectSlug: string,
  traceIds?: string[]
): Promise<Blob> => {
  const params = traceIds ? { traceIds } : {}
  return client.get(`/v1/projects/${projectSlug}/traces/export`, params)
}
