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

// API Response wrapper (backend format)
export interface APIResponse<T = any> {
  success: boolean
  data: T
  message?: string
  meta?: {
    request_id?: string
    timestamp?: string
    [key: string]: any
  }
}

// API Error structure (backend format)  
export interface APIErrorResponse {
  success: false
  error: string
  message: string
  code?: string
  details?: Record<string, any>
  meta?: {
    request_id?: string
    timestamp?: string
    status_code?: number
    [key: string]: any
  }
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
    const errorData = response?.data as any

    // Extract error message - handle both formats
    let message = 'API request failed'
    if (errorData?.error?.message) {
      // New format: { success: false, error: { message: "..." } }
      message = errorData.error.message
    } else if (errorData?.message) {
      // Old format: { success: false, message: "..." }
      message = errorData.message
    } else {
      message = axiosError.message
    }
    super(message)

    this.name = 'BrokleAPIError'
    this.statusCode = response?.status || 0
    this.code = errorData?.error?.code || errorData?.code || axiosError.code || 'UNKNOWN_ERROR'
    this.requestId = errorData?.meta?.request_id
    this.details = errorData?.details
    this.timestamp = errorData?.meta?.timestamp || new Date().toISOString()
    this.originalError = axiosError
    this.response = response  // Preserve full response for downstream error handling

    // Maintain proper stack trace
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