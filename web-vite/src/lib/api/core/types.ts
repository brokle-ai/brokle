// Minimal shared wire-shape types for feature hooks ported from web/.
// Keep this surface intentionally small — web-vite's canonical API
// surface is openapi-fetch typed clients + rawFetch. This file exists
// to give ported feature code a stable `PaginatedResponse<T>` shape.

export interface Pagination {
  page: number
  limit: number
  total: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

export interface PaginatedResponse<T = unknown> {
  data: T[]
  pagination: Pagination
}
