import type { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'

// Core API configuration
export interface APIClientConfig {
  baseURL: string
  timeout?: number
  headers?: Record<string, string>
}

// Enhanced configuration for production use
export interface BrokleClientConfig extends APIClientConfig {
  retries?: number
  retryDelay?: number
  enableRequestId?: boolean
  enableLogging?: boolean
  logLevel?: 'debug' | 'info' | 'warn' | 'error'
  customHeaders?: Record<string, string>
  enablePerformanceLogging?: boolean
  maxConcurrentRequests?: number
}

// Request options extending axios config
export interface RequestOptions extends Omit<AxiosRequestConfig, 'url' | 'method' | 'data'> {
  skipAuth?: boolean
  skipRefreshInterceptor?: boolean  // Bypass refresh interceptor (for /auth/refresh itself)
  _retry?: boolean  // Internal flag for tracking retry attempts
  retries?: number
  // Context header options (opt-in only)
  includeOrgContext?: boolean
  includeProjectContext?: boolean
  includeEnvironmentContext?: boolean
  customOrgId?: string
  customProjectId?: string
  customEnvironmentId?: string
}

// Error response shape. Matches Stripe/OpenAI: `{"error": {type, code,
// message, details?, param?, errors?}}`. No `success` field — HTTP
// status is the success signal (per RFC 9110 §15). The frontend sees
// this shape ONLY on 4xx/5xx; 2xx bodies are the raw resource.
//
// Mirrors pkg/response.APIError on the backend.
export interface APIErrorBody {
  type: string
  code?: string
  message: string
  details?: string
  param?: string
  errors?: Array<{
    location?: string
    message: string
    value?: unknown
  }>
}

export interface APIErrorResponse {
  error: APIErrorBody
}

// Custom API Error class that preserves full response data
export class BrokleAPIError extends Error {
  public readonly statusCode: number
  public readonly code: string
  public readonly requestId?: string
  public readonly details?: Record<string, any>
  public readonly timestamp: string
  public readonly originalError: AxiosError
  public readonly response?: AxiosResponse  // CRITICAL: Preserve response for downstream handlers

  constructor(axiosError: AxiosError) {
    const response = axiosError.response
    const errorData = response?.data as APIErrorResponse | undefined

    // Error body is always the Stripe/OpenAI-style `{error: {...}}`
    // envelope. Falls back to axiosError.message only for network
    // failures (no response body at all).
    const message =
      errorData?.error?.message ??
      axiosError.message ??
      'API request failed'

    super(message)

    this.name = 'BrokleAPIError'
    this.statusCode = response?.status || 0
    this.code = errorData?.error?.code || errorData?.error?.type || axiosError.code || 'UNKNOWN_ERROR'
    // Request IDs live in the X-Request-Id response header, not the
    // body. Kept case-insensitive because HTTP headers are; axios
    // lowercases them but be defensive.
    this.requestId =
      (response?.headers?.['x-request-id'] as string | undefined) ??
      (response?.headers?.['X-Request-Id'] as string | undefined)
    // `details` is the structured error-body field (string explanation
    // alongside message). Preserved as the error-body value for
    // downstream UIs that surface it.
    this.details = errorData?.error?.details
      ? { details: errorData.error.details }
      : undefined
    this.timestamp = new Date().toISOString()
    this.originalError = axiosError
    this.response = response

    Object.setPrototypeOf(this, BrokleAPIError.prototype)
  }

  // Helper methods
  isNetworkError(): boolean {
    return this.originalError.code === 'NETWORK_ERROR' || this.statusCode === 0
  }

  isAuthError(): boolean {
    return this.statusCode === 401
  }

  isServerError(): boolean {
    return this.statusCode >= 500
  }

  isRetryable(): boolean {
    return this.isServerError() || this.isNetworkError()
  }

  isForbidden(): boolean {
    return this.statusCode === 403
  }

  isValidationError(): boolean {
    return this.statusCode === 422 || this.statusCode === 400
  }

  toJSON(): Record<string, any> {
    return {
      name: this.name,
      message: this.message,
      statusCode: this.statusCode,
      code: this.code,
      requestId: this.requestId,
      details: this.details,
      timestamp: this.timestamp,
    }
  }
}

// Backend pagination format (matches pkg/response/response.go Pagination struct)
export interface BackendPagination {
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

// Frontend pagination format (normalized for UI)
export interface Pagination {
  page: number
  limit: number
  total: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

// Paginated response interface
export interface PaginatedResponse<T = any> {
  data: T[]
  pagination: Pagination
}

// Common query parameters
export interface QueryParams {
  [key: string]: string | number | boolean | string[] | undefined
}

// Request interceptor function type
export type RequestInterceptor = (
  config: AxiosRequestConfig
) => AxiosRequestConfig | Promise<AxiosRequestConfig>

// Response interceptor function types
export type ResponseInterceptor = (response: AxiosResponse) => AxiosResponse | Promise<AxiosResponse>
export type ResponseErrorInterceptor = (error: AxiosError) => Promise<never>

// Token refresh callback type
export type TokenRefreshCallback = () => Promise<string | null>

// Extended Axios config with custom performance tracking properties
export interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
  _requestStartTime?: number
  _retry?: boolean
}