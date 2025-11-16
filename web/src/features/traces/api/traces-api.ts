import { BrokleAPIClient } from '@/lib/api/core/client'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type { Trace, Span, Score } from '../data/schema'
import {
  transformTrace,
  transformTraceResponse,
  transformSpan,
  transformScore,
  stringToStatusCode,
} from '../utils/transform'

// Create API client instance
const client = new BrokleAPIClient('/api')

// ============================================================================
// API Parameter Types
// ============================================================================

export interface GetTracesParams {
  projectId: string
  page?: number
  pageSize?: number
  status?: string[] // ['ok', 'error', 'unset']
  search?: string
  sessionId?: string
  userId?: string
  startTime?: Date
  endTime?: Date
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface GetSpansParams {
  projectId: string
  traceId?: string
  type?: string
  model?: string
  level?: string
  page?: number
  pageSize?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface GetScoresParams {
  projectId: string
  traceId?: string
  spanId?: string
  sessionId?: string
  name?: string
  source?: string
  dataType?: string
  page?: number
  pageSize?: number
}

export interface UpdateTraceData {
  name?: string
  tags?: string[]
  bookmarked?: boolean
  public?: boolean
}

export interface UpdateSpanData {
  span_name?: string
}

export interface UpdateScoreData {
  name?: string
  value?: number
  string_value?: string
  comment?: string
}

// ============================================================================
// Traces API
// ============================================================================

/**
 * Get all traces for a project with filtering and pagination
 *
 * Backend endpoint: GET /api/v1/traces
 *
 * @param params - Filter and pagination parameters
 * @returns Traces array with pagination metadata
 */
export const getProjectTraces = async (params: GetTracesParams): Promise<{
  traces: Trace[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
}> => {
  const {
    projectId,
    page = 1,
    pageSize = 20,
    status,
    search,
    sessionId,
    userId,
    startTime,
    endTime,
    sortBy,
    sortOrder,
  } = params

  // Build query parameters
  const queryParams: Record<string, any> = {
    project_id: projectId,
    page,
    limit: pageSize,
  }

  // Add optional filters
  if (search) queryParams.search = search
  if (sessionId) queryParams.session_id = sessionId
  if (userId) queryParams.user_id = userId
  if (startTime) queryParams.start_time = startTime.toISOString()
  if (endTime) queryParams.end_time = endTime.toISOString()
  if (sortBy) queryParams.sort_by = sortBy
  if (sortOrder) queryParams.sort_dir = sortOrder

  // Convert status strings to codes (e.g., ['ok', 'error'] → '1,2')
  if (status && status.length > 0) {
    const statusCodes = status.map(stringToStatusCode)
    queryParams.status = statusCodes.join(',')
  }

  // Make API request using getPaginated (preserves pagination metadata)
  const response = await client.getPaginated<any>('/v1/traces', queryParams)

  // Transform trace data
  return {
    traces: response.data.map(transformTrace),
    totalCount: response.pagination.total,
    page: response.pagination.page,
    pageSize: response.pagination.limit,
    totalPages: response.pagination.totalPages,
  }
}

/**
 * Get a single trace by ID
 *
 * Backend endpoint: GET /api/v1/traces/:id
 *
 * @param projectId - Project ULID
 * @param traceId - Trace ID (32 hex characters)
 * @returns Single trace object
 */
export const getTraceById = async (
  projectId: string,
  traceId: string
): Promise<Trace> => {
  const response = await client.get(`/v1/traces/${traceId}`, {
    project_id: projectId,
  })
  return transformTraceResponse(response)
}

/**
 * Get trace with hierarchical spans tree
 *
 * Backend endpoint: GET /api/v1/traces/:id/spans
 *
 * @param projectId - Project ULID
 * @param traceId - Trace ID
 * @returns Trace with nested spans
 */
export const getTraceWithSpans = async (
  projectId: string,
  traceId: string
): Promise<Trace> => {
  const response = await client.get(`/v1/traces/${traceId}/spans`, {
    project_id: projectId,
  })
  return transformTraceResponse(response)
}

/**
 * Get trace with quality scores
 *
 * Backend endpoint: GET /api/v1/traces/:id/scores
 *
 * @param projectId - Project ULID
 * @param traceId - Trace ID
 * @returns Trace with quality scores
 */
export const getTraceWithScores = async (
  projectId: string,
  traceId: string
): Promise<Trace> => {
  const response = await client.get(`/v1/traces/${traceId}/scores`, {
    project_id: projectId,
  })
  return transformTraceResponse(response)
}

/**
 * Update trace metadata
 *
 * Backend endpoint: PUT /api/v1/traces/:id
 *
 * @param projectId - Project ULID
 * @param traceId - Trace ID
 * @param data - Updated trace data
 * @returns Updated trace
 */
export const updateTrace = async (
  projectId: string,
  traceId: string,
  data: UpdateTraceData
): Promise<Trace> => {
  const response = await client.put(`/v1/traces/${traceId}`, {
    project_id: projectId,
    ...data,
  })
  return transformTraceResponse(response)
}

/**
 * Delete a trace (NOT IMPLEMENTED IN BACKEND)
 *
 * @deprecated Backend endpoint not yet implemented
 * @throws Error indicating feature is not available
 */
export const deleteTrace = async (
  projectId: string,
  traceId: string
): Promise<void> => {
  throw new Error('Delete functionality is not yet implemented on the backend')
  // Future implementation:
  // await client.delete(`/v1/traces/${traceId}`, {
  //   project_id: projectId,
  // })
}

/**
 * Delete multiple traces (NOT IMPLEMENTED IN BACKEND)
 *
 * @deprecated Backend endpoint not yet implemented
 * @throws Error indicating feature is not available
 */
export const deleteMultipleTraces = async (
  projectId: string,
  traceIds: string[]
): Promise<void> => {
  throw new Error('Bulk delete functionality is not yet implemented on the backend')
  // Future implementation:
  // await client.post(`/v1/traces/bulk-delete`, {
  //   project_id: projectId,
  //   trace_ids: traceIds,
  // })
}

/**
 * Export traces to CSV (NOT IMPLEMENTED IN BACKEND)
 *
 * @deprecated Backend endpoint not yet implemented
 * @throws Error indicating feature is not available
 */
export const exportTraces = async (
  projectId: string,
  traceIds?: string[]
): Promise<Blob> => {
  throw new Error('Export functionality is not yet implemented on the backend')
  // Future implementation:
  // const response = await client.get('/v1/traces/export', {
  //   project_id: projectId,
  //   trace_ids: traceIds?.join(','),
  //   format: 'csv',
  // })
  // return response as Blob
}

// ============================================================================
// Spans API
// ============================================================================

/**
 * Get spans with filtering
 *
 * Backend endpoint: GET /api/v1/spans
 *
 * @param params - Filter and pagination parameters
 * @returns Spans array with pagination
 */
export const getSpans = async (params: GetSpansParams): Promise<{
  spans: Span[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
}> => {
  const {
    projectId,
    traceId,
    type,
    model,
    level,
    page = 1,
    pageSize = 20,
    sortBy,
    sortOrder,
  } = params

  const queryParams: Record<string, any> = {
    project_id: projectId,
    page,
    limit: pageSize,
  }

  if (traceId) queryParams.trace_id = traceId
  if (type) queryParams.type = type
  if (model) queryParams.model = model
  if (level) queryParams.level = level
  if (sortBy) queryParams.sort_by = sortBy
  if (sortOrder) queryParams.sort_dir = sortOrder

  // Make API request using getPaginated (preserves pagination metadata)
  const response = await client.getPaginated<any>('/v1/spans', queryParams)

  // Transform span data
  return {
    spans: response.data.map(transformSpan),
    totalCount: response.pagination.total,
    page: response.pagination.page,
    pageSize: response.pagination.limit,
    totalPages: response.pagination.totalPages,
  }
}

/**
 * Get a single span by ID
 *
 * Backend endpoint: GET /api/v1/spans/:id
 */
export const getSpanById = async (
  projectId: string,
  spanId: string
): Promise<Span> => {
  const response = await client.get<any>(`/v1/spans/${spanId}`, {
    project_id: projectId,
  })
  return transformSpan(response)
}

/**
 * Update span metadata
 *
 * Backend endpoint: PUT /api/v1/spans/:id
 */
export const updateSpan = async (
  projectId: string,
  spanId: string,
  data: UpdateSpanData
): Promise<Span> => {
  const response = await client.put<any>(`/v1/spans/${spanId}`, {
    project_id: projectId,
    ...data,
  })
  return transformSpan(response)
}

// ============================================================================
// Quality Scores API
// ============================================================================

/**
 * Get quality scores with filtering
 *
 * Backend endpoint: GET /api/v1/scores
 */
export const getScores = async (params: GetScoresParams): Promise<{
  scores: Score[]
  totalCount: number
}> => {
  const {
    projectId,
    traceId,
    spanId,
    sessionId,
    name,
    source,
    dataType,
    page = 1,
    pageSize = 20,
  } = params

  const queryParams: Record<string, any> = {
    project_id: projectId,
    page,
    limit: pageSize,
  }

  if (traceId) queryParams.trace_id = traceId
  if (spanId) queryParams.span_id = spanId
  if (sessionId) queryParams.session_id = sessionId
  if (name) queryParams.name = name
  if (source) queryParams.source = source
  if (dataType) queryParams.data_type = dataType

  const response = await client.getPaginated<Score>('/v1/scores', queryParams)

  return {
    scores: response.data.map(transformScore),
    totalCount: response.pagination.total,
  }
}

/**
 * Get a single score by ID
 *
 * Backend endpoint: GET /api/v1/scores/:id
 */
export const getScoreById = async (
  projectId: string,
  scoreId: string
): Promise<Score> => {
  const response = await client.get<any>(`/v1/scores/${scoreId}`, {
    project_id: projectId,
  })
  return transformScore(response)
}

/**
 * Update a quality score
 *
 * Backend endpoint: PUT /api/v1/scores/:id
 */
export const updateScore = async (
  projectId: string,
  scoreId: string,
  data: UpdateScoreData
): Promise<Score> => {
  const response = await client.put<any>(`/v1/scores/${scoreId}`, {
    project_id: projectId,
    ...data,
  })
  return transformScore(response)
}
