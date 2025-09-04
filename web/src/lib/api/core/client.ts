import axios, { AxiosInstance, AxiosResponse } from 'axios'
import { getTokenManager } from '@/lib/auth/token-manager'
import { urlContextManager } from '@/lib/context/url-context-manager'
import type { 
  BrokleClientConfig,
  RequestOptions,
  APIResponse,
  QueryParams,
  PaginatedResponse,
  BackendPagination,
  Pagination,
} from './types'
import { BrokleAPIError as APIError } from './types'

export class BrokleAPIClient {
  protected axiosInstance: AxiosInstance
  private tokenManager = getTokenManager()
  private isRefreshing = false
  private refreshPromise: Promise<string | null> | null = null

  constructor(
    basePath: string = '',
    protected config: Partial<BrokleClientConfig> = {}
  ) {
    const defaultConfig: BrokleClientConfig = {
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000',
      timeout: 30000,
      retries: 3,
      retryDelay: 1000,
      enableRequestId: true,
      enableLogging: process.env.NODE_ENV === 'development',
      logLevel: 'info',
      enablePerformanceLogging: process.env.NODE_ENV === 'development',
      maxConcurrentRequests: 10,
      headers: {
        'Content-Type': 'application/json',
      },
    }

    const finalConfig = { ...defaultConfig, ...config }

    this.axiosInstance = axios.create({
      baseURL: finalConfig.baseURL,
      timeout: finalConfig.timeout,
      headers: finalConfig.headers,
    })

    // Set base path for service-specific clients
    if (basePath) {
      this.axiosInstance.defaults.baseURL += basePath
    }

    this.setupInterceptors()
  }

  // Public HTTP methods with retry support
  async get<T>(
    endpoint: string, 
    params?: QueryParams, 
    options: RequestOptions = {}
  ): Promise<T> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.get<APIResponse<T>>(endpoint, {
        params,
        ...options,
      })
      return this.extractData(response)
    }, options.retries)
  }

  async post<T>(
    endpoint: string, 
    data?: any, 
    options: RequestOptions = {}
  ): Promise<T> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.post<APIResponse<T>>(endpoint, data, options)
      return this.extractData(response)
    }, options.retries)
  }

  async put<T>(
    endpoint: string, 
    data?: any, 
    options: RequestOptions = {}
  ): Promise<T> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.put<APIResponse<T>>(endpoint, data, options)
      return this.extractData(response)
    }, options.retries)
  }

  async patch<T>(
    endpoint: string, 
    data?: any, 
    options: RequestOptions = {}
  ): Promise<T> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.patch<APIResponse<T>>(endpoint, data, options)
      return this.extractData(response)
    }, options.retries)
  }

  async delete<T>(
    endpoint: string, 
    options: RequestOptions = {}
  ): Promise<T> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.delete<APIResponse<T>>(endpoint, options)
      return this.extractData(response)
    }, options.retries)
  }

  // Paginated HTTP methods (preserve pagination metadata from meta.pagination)

  async getPaginated<T>(
    endpoint: string, 
    params?: QueryParams, 
    options: RequestOptions = {}
  ): Promise<PaginatedResponse<T>> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.get<APIResponse<T[]>>(endpoint, {
        params,
        ...options,
      })
      return this.extractPaginatedData(response)
    }, options.retries)
  }

  async postPaginated<T>(
    endpoint: string, 
    data?: any, 
    options: RequestOptions = {}
  ): Promise<PaginatedResponse<T>> {
    return this.executeWithRetry(async () => {
      const response = await this.axiosInstance.post<APIResponse<T[]>>(endpoint, data, options)
      return this.extractPaginatedData(response)
    }, options.retries)
  }

  // Setup axios interceptors
  private setupInterceptors(): void {
    // Request interceptor - Add authentication and context headers
    this.axiosInstance.interceptors.request.use(
      async (config) => {
        // Skip auth if explicitly requested
        if ((config as any).skipAuth) {
          return config
        }

        // Initialize headers
        config.headers = config.headers || {}

        // Add Bearer token for authenticated requests
        const token = await this.tokenManager.getValidAccessToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }

        // Add context headers based on explicit request options (URL-based, opt-in only)
        const requestOptions = config as RequestOptions
        try {
          // Only generate headers if explicitly requested and we have a pathname
          if (typeof window !== 'undefined' && 
              (requestOptions.includeOrgContext || requestOptions.includeProjectContext || requestOptions.includeEnvironmentContext)) {
            
            const contextHeaders = await urlContextManager.getHeadersFromURL(window.location.pathname, {
              includeOrgContext: requestOptions.includeOrgContext,
              includeProjectContext: requestOptions.includeProjectContext,
              includeEnvironmentContext: requestOptions.includeEnvironmentContext,
              customOrgId: requestOptions.customOrgId,
              customProjectId: requestOptions.customProjectId,
              customEnvironmentId: requestOptions.customEnvironmentId,
            })

            // Add context headers to request (will be empty if none requested)
            Object.assign(config.headers, contextHeaders)

            // Log context headers in development (only if headers were added)
            if (this.config.enableLogging && Object.keys(contextHeaders).length > 0) {
              console.debug('[API Context]', {
                url: config.url,
                method: config.method?.toUpperCase(),
                pathname: window.location.pathname,
                contextHeaders,
              })
            }
          }
        } catch (error) {
          // Don't fail the request if context headers fail
          console.warn('[API] Failed to add context headers:', error)
        }

        // Add request ID for tracing if enabled
        if (this.config.enableRequestId && !config.headers['X-Request-Id']) {
          const { ulid } = await import('ulid')
          config.headers['X-Request-Id'] = ulid()
        }

        // Add custom headers from config
        if (this.config.customHeaders) {
          Object.assign(config.headers, this.config.customHeaders)
        }

        // Add performance timing start
        if (this.config.enablePerformanceLogging) {
          (config as any)._requestStartTime = Date.now()
        }

        return config
      },
      (error) => {
        return Promise.reject(new APIError(error))
      }
    )

    // Response interceptor - Handle data extraction and token refresh
    this.axiosInstance.interceptors.response.use(
      (response: AxiosResponse) => {
        // Calculate performance timing if enabled
        const startTime = (response.config as any)._requestStartTime
        const duration = startTime ? Date.now() - startTime : undefined

        // Enhanced logging based on configuration
        if (this.config.enableLogging) {
          const logData = {
            method: response.config.method?.toUpperCase(),
            url: response.config.url,
            status: response.status,
            requestId: response.headers['x-request-id'] || response.config.headers?.['X-Request-Id'],
            ...(duration && { duration: `${duration}ms` }),
            ...(this.config.logLevel === 'debug' && { 
              responseData: response.data,
              responseHeaders: response.headers 
            })
          }

          const logLevel = this.config.logLevel || 'info'
          const message = `[API] ${logData.method} ${logData.url}`

          switch (logLevel) {
            case 'debug':
              console.debug(message, logData)
              break
            case 'info':
              console.info(message, { 
                status: logData.status, 
                requestId: logData.requestId,
                ...(duration && { duration: logData.duration })
              })
              break
            case 'warn':
              if (response.status >= 400) console.warn(message, logData)
              break
            case 'error':
              if (response.status >= 500) console.error(message, logData)
              break
          }

          // Performance logging if enabled
          if (this.config.enablePerformanceLogging && duration) {
            console.debug(`[PERF] ${logData.method} ${logData.url}: ${duration}ms`)
          }
        }

        return response
      },
      async (error) => {
        const originalRequest = error.config

        // Handle 401 errors with token refresh (but skip for requests that don't need auth)
        if (error.response?.status === 401 && 
            !originalRequest._retry && 
            !originalRequest.skipAuth) {
          originalRequest._retry = true

          try {
            // Prevent multiple refresh attempts
            if (!this.isRefreshing) {
              this.isRefreshing = true
              this.refreshPromise = this.tokenManager.refreshAccessToken()
            }

            const newToken = await this.refreshPromise
            
            if (newToken) {
              // Retry original request with new token
              originalRequest.headers = originalRequest.headers || {}
              originalRequest.headers.Authorization = `Bearer ${newToken}`
              return this.axiosInstance.request(originalRequest)
            }
          } catch (refreshError) {
            // Refresh failed, clear tokens and redirect to login
            this.tokenManager.clearTokens()
            
            // Broadcast session expired
            if (typeof window !== 'undefined') {
              window.dispatchEvent(new CustomEvent('auth:session-expired'))
            }
            
            return Promise.reject(new APIError(error))
          } finally {
            this.isRefreshing = false
            this.refreshPromise = null
          }
        }

        // Enhanced error logging based on configuration
        if (this.config.enableLogging) {
          const startTime = (error.config as any)?._requestStartTime
          const duration = startTime ? Date.now() - startTime : undefined

          const errorData = {
            method: error.config?.method?.toUpperCase(),
            url: error.config?.url,
            status: error.response?.status,
            statusText: error.response?.statusText,
            requestId: error.response?.headers['x-request-id'] || error.config?.headers?.['X-Request-Id'],
            errorCode: error.code,
            ...(duration && { duration: `${duration}ms` }),
            ...(this.config.logLevel === 'debug' && {
              requestData: error.config?.data,
              responseData: error.response?.data,
              stack: error.stack
            })
          }

          const logLevel = this.config.logLevel || 'info'
          const message = `[API ERROR] ${errorData.method} ${errorData.url}`

          // Log based on error severity and configuration
          if (errorData.status && errorData.status >= 500) {
            console.error(message, errorData)
          } else if (errorData.status && errorData.status >= 400) {
            if (logLevel === 'debug' || logLevel === 'info') {
              console.warn(message, errorData)
            }
          } else {
            // Network errors, timeouts, etc.
            console.error(message, errorData)
          }

          // Performance logging for failed requests
          if (this.config.enablePerformanceLogging && duration) {
            console.debug(`[PERF ERROR] ${errorData.method} ${errorData.url}: ${duration}ms (failed)`)
          }
        }

        return Promise.reject(new APIError(error))
      }
    )
  }

  // Extract data from API response wrapper
  private extractData<T>(response: AxiosResponse<APIResponse<T>>): T {
    const { data, success } = response.data

    if (!success) {
      throw new Error('API response indicates failure but was not caught by error interceptor')
    }

    return data
  }

  // Convert backend pagination format to frontend format
  private convertPagination(backendPagination: BackendPagination): Pagination {
    return {
      page: backendPagination.page,
      limit: backendPagination.page_size,        // snake_case to camelCase
      total: backendPagination.total,
      totalPages: backendPagination.total_page,  // snake_case to camelCase
      hasNext: backendPagination.has_next,       // snake_case to camelCase
      hasPrev: backendPagination.has_prev        // snake_case to camelCase
    }
  }

  // Extract paginated data from API response (preserves pagination metadata)
  private extractPaginatedData<T>(response: AxiosResponse<APIResponse<T[]>>): PaginatedResponse<T> {
    const { data, success, meta } = response.data

    if (!success) {
      throw new Error('API response indicates failure but was not caught by error interceptor')
    }

    // Check if pagination metadata exists
    const backendPagination = meta?.pagination as BackendPagination
    if (!backendPagination) {
      throw new Error('No pagination metadata found in response. Use regular get() method for non-paginated responses.')
    }

    return {
      data,
      pagination: this.convertPagination(backendPagination)
    }
  }

  // File upload with progress tracking
  async uploadFile<T = any>(
    endpoint: string,
    file: File | Blob,
    options: {
      fieldName?: string
      additionalFields?: Record<string, any>
      onProgress?: (progress: number) => void
      retries?: number
    } = {}
  ): Promise<T> {
    const {
      fieldName = 'file',
      additionalFields = {},
      onProgress,
      retries
    } = options

    return this.executeWithRetry(async () => {
      const formData = new FormData()
      formData.append(fieldName, file)
      
      // Add additional fields to form data
      Object.entries(additionalFields).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          formData.append(key, String(value))
        }
      })

      const response = await this.axiosInstance.post<APIResponse<T>>(endpoint, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          if (onProgress && progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
            onProgress(progress)
          }
        },
      })

      return this.extractData(response)
    }, retries)
  }

  // Batch file upload with progress tracking
  async uploadFiles<T = any>(
    endpoint: string,
    files: Array<File | Blob>,
    options: {
      fieldName?: string
      additionalFields?: Record<string, any>
      onProgress?: (fileIndex: number, progress: number) => void
      onComplete?: (fileIndex: number) => void
      retries?: number
    } = {}
  ): Promise<T[]> {
    const {
      fieldName = 'files',
      additionalFields = {},
      onProgress,
      onComplete,
      retries
    } = options

    const results: T[] = []
    
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      
      try {
        const result = await this.uploadFile<T>(endpoint, file, {
          fieldName: `${fieldName}[${i}]`,
          additionalFields: { ...additionalFields, fileIndex: i },
          onProgress: (progress) => onProgress?.(i, progress),
          retries
        })
        
        results.push(result)
        onComplete?.(i)
      } catch (error) {
        console.error(`[API] File upload failed for file ${i}:`, error)
        throw error
      }
    }

    return results
  }

  // Retry logic with exponential backoff
  private async executeWithRetry<T>(
    operation: () => Promise<T>,
    customRetries?: number
  ): Promise<T> {
    const maxRetries = customRetries ?? this.config.retries ?? 3
    const baseDelay = this.config.retryDelay ?? 1000

    let lastError: any
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        return await operation()
      } catch (error: any) {
        lastError = error
        
        // Don't retry on certain errors
        if (!this.shouldRetry(error, attempt) || attempt === maxRetries) {
          throw error
        }

        // Calculate exponential backoff delay
        const delay = baseDelay * Math.pow(2, attempt) + Math.random() * 1000
        
        if (this.config.enableLogging) {
          console.warn(`[API Retry] Attempt ${attempt + 1}/${maxRetries + 1} failed, retrying in ${Math.round(delay)}ms`, {
            error: error.message,
            status: error.response?.status,
            endpoint: error.config?.url
          })
        }

        // Wait before retrying
        await this.delay(delay)
      }
    }

    throw lastError
  }

  private shouldRetry(error: any, attempt: number): boolean {
    // Handle BrokleAPIError (wrapped) vs raw axios error
    const status = error.statusCode || error.response?.status
    
    // Never retry auth failures
    if (status === 401) return false

    // Never retry client errors (400-499) except specific cases
    if (status >= 400 && status < 500) {
      return status === 429 || status === 408
    }

    // Retry server errors and network issues
    return status >= 500 || (!status && !error.response)
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  // Utility methods

  getBaseURL(): string {
    return this.axiosInstance.defaults.baseURL || ''
  }

  // Development helper
  debug(): void {
    if (process.env.NODE_ENV !== 'development') return

    console.group('🌐 BrokleAPIClient Debug')
    console.log('Base URL:', this.getBaseURL())
    console.log('Default Headers:', this.axiosInstance.defaults.headers)
    console.log('Timeout:', this.axiosInstance.defaults.timeout)
    console.log('Retry Config:', {
      retries: this.config.retries,
      retryDelay: this.config.retryDelay
    })
    console.log('Token Manager:', this.tokenManager)
    console.groupEnd()
  }
}