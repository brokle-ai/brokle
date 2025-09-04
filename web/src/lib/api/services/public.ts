import { BrokleAPIClient } from '../core/client'

// Public API response types
export interface HealthStatus {
  status: 'ok' | 'degraded' | 'down'
  timestamp: string
  version: string
  services: ServiceHealth[]
  uptime: number
}

export interface ServiceHealth {
  name: string
  status: 'healthy' | 'unhealthy' | 'unknown'
  latency?: number
  lastChecked: string
}

export interface PublicStats {
  totalRequests: number
  activeUsers: number
  supportedModels: number
  supportedProviders: number
  lastUpdated: string
}

export interface SystemStatus {
  overall: 'operational' | 'degraded' | 'outage'
  services: {
    api: 'operational' | 'degraded' | 'outage'
    gateway: 'operational' | 'degraded' | 'outage'
    dashboard: 'operational' | 'degraded' | 'outage'
    monitoring: 'operational' | 'degraded' | 'outage'
  }
  incidents: SystemIncident[]
  lastUpdated: string
}

export interface SystemIncident {
  id: string
  title: string
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved'
  severity: 'minor' | 'major' | 'critical'
  startedAt: string
  resolvedAt?: string
  affectedServices: string[]
  updates: IncidentUpdate[]
}

export interface IncidentUpdate {
  id: string
  message: string
  timestamp: string
  status: string
}

export class PublicAPIClient extends BrokleAPIClient {
  constructor() {
    super('') // No base path - public endpoints are at root level
    
    // Override all requests to skip authentication
    this.axiosInstance.interceptors.request.use(
      (config) => {
        // Force all public API requests to skip auth
        (config as any).skipAuth = true
        return config
      },
      (error) => Promise.reject(error)
    )
  }

  // Health check endpoints
  async getHealthCheck(): Promise<HealthStatus> {
    return this.get<HealthStatus>('/health')
  }

  async getReadinessCheck(): Promise<{ ready: boolean; timestamp: string }> {
    return this.get('/ready')
  }

  async getLivenessCheck(): Promise<{ alive: boolean; timestamp: string }> {
    return this.get('/live')
  }

  // System status endpoints
  async getSystemStatus(): Promise<SystemStatus> {
    return this.get<SystemStatus>('/status')
  }

  async getServiceHealth(serviceName: string): Promise<ServiceHealth> {
    return this.get<ServiceHealth>(`/health/${serviceName}`)
  }

  // Public statistics (non-sensitive data)
  async getPublicStats(): Promise<PublicStats> {
    return this.get<PublicStats>('/stats/public')
  }

  // Version and build info
  async getVersionInfo(): Promise<{
    version: string
    buildId: string
    commitHash: string
    buildDate: string
    environment: string
  }> {
    return this.get('/version')
  }

  // Supported models and providers (public info)
  async getSupportedModels(): Promise<{
    models: Array<{
      id: string
      name: string
      provider: string
      type: 'text' | 'image' | 'audio' | 'multimodal'
      capabilities: string[]
    }>
  }> {
    return this.get('/models/supported')
  }

  async getSupportedProviders(): Promise<{
    providers: Array<{
      id: string
      name: string
      status: 'active' | 'deprecated' | 'maintenance'
      supportedTypes: string[]
      regions: string[]
    }>
  }> {
    return this.get('/providers/supported')
  }

  // Pricing information (public)
  async getPricingInfo(): Promise<{
    plans: Array<{
      id: string
      name: string
      price: number
      currency: string
      features: string[]
      limits: Record<string, number>
    }>
  }> {
    return this.get('/pricing')
  }

  // Documentation endpoints
  async getAPIDocumentation(): Promise<{
    openapi: string
    version: string
    title: string
    description: string
    endpoints: Array<{
      path: string
      method: string
      summary: string
      tags: string[]
    }>
  }> {
    return this.get('/docs/api')
  }

  // Contact and support endpoints (public forms)
  async submitContactForm(data: {
    name: string
    email: string
    company?: string
    subject: string
    message: string
  }): Promise<{ success: boolean; ticketId: string }> {
    return this.post('/contact', data)
  }

  async submitFeedback(data: {
    email?: string
    type: 'bug' | 'feature' | 'improvement' | 'other'
    message: string
    page?: string
    userAgent?: string
  }): Promise<{ success: boolean; feedbackId: string }> {
    return this.post('/feedback', data)
  }

  // Newsletter subscription
  async subscribeNewsletter(email: string): Promise<{ success: boolean }> {
    return this.post('/newsletter/subscribe', { email })
  }

  async unsubscribeNewsletter(token: string): Promise<{ success: boolean }> {
    return this.post('/newsletter/unsubscribe', { token })
  }

  // Rate limiting info (public)
  async getRateLimitInfo(): Promise<{
    limits: {
      unauthenticated: {
        requests: number
        window: string
      }
      authenticated: {
        [tier: string]: {
          requests: number
          window: string
        }
      }
    }
  }> {
    return this.get('/limits')
  }

  // Service announcements and notices
  async getServiceAnnouncements(): Promise<{
    announcements: Array<{
      id: string
      title: string
      message: string
      type: 'info' | 'warning' | 'maintenance' | 'feature'
      startDate: string
      endDate?: string
      affectedServices?: string[]
    }>
  }> {
    return this.get('/announcements')
  }
}