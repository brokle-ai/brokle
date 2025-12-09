export interface ApiResponse<T = any> {
  data: T
  message?: string
  success: boolean
  timestamp: string
  requestId?: string
}

export interface ApiError {
  error: string
  message: string
  code: string
  statusCode: number
  timestamp: string
  path?: string
  details?: Record<string, any>
  requestId?: string
}

export interface RequestOptions extends RequestInit {
  timeout?: number
  retries?: number
  idempotent?: boolean
  skipAuth?: boolean
  baseURL?: string
}

export interface RetryOptions {
  maxRetries?: number
  baseDelay?: number
  maxDelay?: number
  factor?: number
  jitter?: boolean
}

export interface CircuitBreakerState {
  failures: number
  lastFailure: number
  state: 'closed' | 'open' | 'half-open'
  threshold: number
  timeout: number
}

export interface PaginatedResponse<T = any> {
  data: T[]
  pagination: Pagination
  filters?: Record<string, any>
  sort?: SortOptions
}

export interface Pagination {
  page: number
  limit: number
  total: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

export interface SortOptions {
  field: string
  order: 'asc' | 'desc'
}

export interface FilterOptions {
  [key: string]: string | number | boolean | string[] | Date
}

export interface SearchParams {
  query?: string
  filters?: FilterOptions
  sort?: SortOptions
  pagination?: {
    page: number
    limit: number
  }
}

export interface ApiKeyValidation {
  valid: boolean
  keyId?: string
  permissions: string[]
  rateLimit: RateLimit
  organization: string
  project?: string
}

export interface RateLimit {
  limit: number
  remaining: number
  reset: number
  window: string
}

// Request/Response types for specific endpoints
export interface DashboardStats {
  totalRequests: number
  totalCost: number
  averageLatency: number
  errorRate: number
  costTrend: TimeSeries[]
  requestTrend: TimeSeries[]
}

export interface TimeSeries {
  timestamp: string
  value: number
  label?: string
}

export interface AnalyticsQuery {
  metric: AnalyticsMetric
  timeRange: TimeRange
  granularity: TimeGranularity
  filters?: AnalyticsFilters
  groupBy?: string[]
}

export interface AnalyticsFilters {
  organizationId?: string
  projectId?: string
  environment?: string
  providerId?: string[]
  modelId?: string[]
  status?: string[]
  dateRange?: {
    start: string
    end: string
  }
}

export type AnalyticsMetric = 
  | 'requests' 
  | 'cost' 
  | 'latency' 
  | 'errors' 
  | 'tokens' 
  | 'quality_score'

export type TimeRange = 
  | '1h' 
  | '24h' 
  | '7d' 
  | '30d' 
  | '90d' 
  | 'custom'

export type TimeGranularity = 
  | 'minute' 
  | 'hour' 
  | 'day' 
  | 'week' 
  | 'month'