// Public API - Direct functions for public/status operations

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

export interface ContactFormData {
  name: string
  email: string
  company?: string
  subject: string
  message: string
}

export interface FeedbackData {
  email?: string
  type: 'bug' | 'feature' | 'improvement' | 'other'
  message: string
  page?: string
  userAgent?: string
}

// Public client (no auth, no base path)
const client = new BrokleAPIClient('')

// Override all requests to skip authentication
client.axiosInstance.interceptors.request.use(
  (config) => {
    // Force all public API requests to skip auth
    (config as any).skipAuth = true
    return config
  },
  (error) => Promise.reject(error)
)

// Direct public API functions
export const getHealthCheck = async (): Promise<HealthStatus> => {
    return client.get<HealthStatus>('/health')
  }

export const getReadinessCheck = async (): Promise<{ ready: boolean; timestamp: string }> => {
    return client.get('/ready')
  }

export const getLivenessCheck = async (): Promise<{ alive: boolean; timestamp: string }> => {
    return client.get('/live')
  }

export const getSystemStatus = async (): Promise<SystemStatus> => {
    return client.get<SystemStatus>('/status')
  }

export const getServiceHealth = async (serviceName: string): Promise<ServiceHealth> => {
    return client.get<ServiceHealth>(`/health/${serviceName}`)
  }

export const getPublicStats = async (): Promise<PublicStats> => {
    return client.get<PublicStats>('/stats/public')
  }

export const getVersionInfo = async (): Promise<{
    version: string
    buildId: string
    commitHash: string
    buildDate: string
    environment: string
  }> => {
    return client.get('/version')
  }

export const getSupportedModels = async (): Promise<{
    models: Array<{
      id: string
      name: string
      provider: string
      type: 'text' | 'image' | 'audio' | 'multimodal'
      capabilities: string[]
    }>
  }> => {
    return client.get('/models/supported')
  }

export const getSupportedProviders = async (): Promise<{
    providers: Array<{
      id: string
      name: string
      status: 'active' | 'deprecated' | 'maintenance'
      supportedTypes: string[]
      regions: string[]
    }>
  }> => {
    return client.get('/providers/supported')
  }

export const getPricingInfo = async (): Promise<{
    plans: Array<{
      id: string
      name: string
      price: number
      currency: string
      features: string[]
      limits: Record<string, number>
    }>
  }> => {
    return client.get('/pricing')
  }

export const getAPIDocumentation = async (): Promise<{
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
  }> => {
    return client.get('/docs/api')
  }

export const submitContactForm = async (data: ContactFormData): Promise<{ success: boolean; ticketId: string }> => {
    return client.post('/contact', data)
  }

export const submitFeedback = async (data: FeedbackData): Promise<{ success: boolean; feedbackId: string }> => {
    return client.post('/feedback', data)
  }

export const subscribeNewsletter = async (email: string): Promise<{ success: boolean }> => {
    return client.post('/newsletter/subscribe', { email })
  }

export const unsubscribeNewsletter = async (token: string): Promise<{ success: boolean }> => {
    return client.post('/newsletter/unsubscribe', { token })
  }

export const getRateLimitInfo = async (): Promise<{
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
  }> => {
    return client.get('/limits')
  }

export const getServiceAnnouncements = async (): Promise<{
    announcements: Array<{
      id: string
      title: string
      message: string
      type: 'info' | 'warning' | 'maintenance' | 'feature'
      startDate: string
      endDate?: string
      affectedServices?: string[]
    }>
  }> => {
    return client.get('/announcements')
  }

