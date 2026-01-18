import { BrokleAPIClient } from '@/lib/api/core/client'
import type { PaginatedResponse } from '@/lib/api/core/types'

const client = new BrokleAPIClient('/api')

// ============================================================================
// Sort Field Mapping (camelCase → snake_case for API)
// ============================================================================

/**
 * Maps frontend column accessorKey (camelCase) to backend API field names (snake_case)
 * Backend validates against whitelist: ["last_trace", "first_trace", "trace_count",
 * "total_tokens", "total_cost", "total_duration", "error_count"]
 */
const SORT_FIELD_MAP: Record<string, string> = {
  traceCount: 'trace_count',
  totalTokens: 'total_tokens',
  totalCost: 'total_cost',
  lastTrace: 'last_trace',
  firstTrace: 'first_trace',
  totalDuration: 'total_duration',
  errorCount: 'error_count',
}

// ============================================================================
// Session Types (matching backend domain)
// ============================================================================

/**
 * Session summary from the backend
 * Sessions are identified by session_id attribute on root spans
 */
export interface SessionSummary {
  session_id: string
  trace_count: number
  first_trace: string // ISO timestamp
  last_trace: string // ISO timestamp
  total_duration: number // nanoseconds
  total_tokens: number
  total_cost: number
  error_count: number
  user_ids: string[]
}

/**
 * Transformed session for frontend use
 */
export interface Session {
  sessionId: string
  traceCount: number
  firstTrace: Date
  lastTrace: Date
  totalDuration: number // nanoseconds
  totalTokens: number
  totalCost: number
  errorCount: number
  userIds: string[]
}

/**
 * Session list response from the backend
 */
export interface SessionListResponse {
  sessions: SessionSummary[]
  total_count: number
  page: number
  page_size: number
  total_pages: number
}

// ============================================================================
// API Parameter Types
// ============================================================================

export interface GetSessionsParams {
  projectId: string
  page?: number
  pageSize?: number
  search?: string
  userId?: string
  startTime?: Date
  endTime?: Date
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

// ============================================================================
// Transform Functions
// ============================================================================

/**
 * Transform backend session summary to frontend Session format
 */
function transformSession(raw: SessionSummary): Session {
  return {
    sessionId: raw.session_id,
    traceCount: raw.trace_count,
    firstTrace: new Date(raw.first_trace),
    lastTrace: new Date(raw.last_trace),
    totalDuration: raw.total_duration,
    totalTokens: raw.total_tokens,
    totalCost: raw.total_cost,
    errorCount: raw.error_count,
    userIds: raw.user_ids || [],
  }
}

// ============================================================================
// API Functions
// ============================================================================

/**
 * Fetch paginated sessions for a project
 */
export async function getProjectSessions(
  params: GetSessionsParams
): Promise<PaginatedResponse<Session>> {
  const {
    projectId,
    page = 1,
    pageSize = 20,
    search,
    userId,
    startTime,
    endTime,
    sortBy,
    sortOrder,
  } = params

  const searchParams: Record<string, string | number | undefined> = {
    page,
    limit: pageSize,
  }

  if (search) searchParams.search = search
  if (userId) searchParams.user_id = userId
  if (startTime) searchParams.start_time = Math.floor(startTime.getTime() / 1000)
  if (endTime) searchParams.end_time = Math.floor(endTime.getTime() / 1000)
  if (sortBy && SORT_FIELD_MAP[sortBy]) {
    searchParams.sort_by = SORT_FIELD_MAP[sortBy]
  }
  if (sortOrder) searchParams.sort_dir = sortOrder

  const response = await client.get<SessionListResponse>(
    `/v1/projects/${projectId}/sessions`,
    searchParams
  )

  return {
    data: (response.sessions ?? []).map(transformSession),
    pagination: {
      page: response.page,
      pageSize: response.page_size,
      totalCount: response.total_count,
      totalPages: response.total_pages,
    },
  }
}
