// Analytics API - Latest endpoints for dashboard application
// Direct functions using optimal ML-powered backend endpoints

import { BrokleAPIClient } from '../core/client'
import type { 
  PaginatedResponse,
  QueryParams
} from '../core/types'

// Analytics types
export interface AnalyticsMetric {
  timestamp: string
  value: number
  label?: string
}

export interface ModelUsage {
  modelId: string
  modelName: string
  providerId: string
  requests: number
  cost: number
  averageLatency: number
  errorRate: number
}

export interface ProviderUsage {
  providerId: string
  providerName: string
  requests: number
  percentage: number
  cost: number
  averageLatency: number
}

export interface DashboardStats {
  totalRequests: number
  totalCost: number
  averageLatency: number
  errorRate: number
  topModels: ModelUsage[]
  costTrend: AnalyticsMetric[]
  requestTrend: AnalyticsMetric[]
  providerDistribution: ProviderUsage[]
}

export interface AnalyticsQuery {
  metric: 'requests' | 'cost' | 'latency' | 'errors' | 'tokens' | 'quality_score'
  timeRange: '1h' | '24h' | '7d' | '30d' | '90d' | 'custom'
  granularity: 'minute' | 'hour' | 'day' | 'week' | 'month'
  filters?: {
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
  groupBy?: string[]
}

export interface CostAnalytics {
  totalCost: number
  costByProvider: Array<{
    providerId: string
    providerName: string
    cost: number
    percentage: number
  }>
  costByModel: Array<{
    modelId: string
    modelName: string
    cost: number
    requests: number
    averageCostPerRequest: number
  }>
  costTrend: AnalyticsMetric[]
  projectedMonthlyCost: number
  budgetUtilization?: number
}

// Flexible base client - versions specified per endpoint
const client = new BrokleAPIClient('/api')

// Direct analytics functions - latest ML-powered endpoints
export const getDashboardStats = async (
    timeRange: string = '24h', 
    organizationId?: string
  ): Promise<DashboardStats> => {
    return client.get<DashboardStats>('/v3/analytics/overview', {
      timeRange,
      organizationId,
    })
  }

export const getCostAnalytics = async (query: Partial<AnalyticsQuery> = {}): Promise<CostAnalytics> => {
    return client.post<CostAnalytics>('/v3/analytics/costs', query)
  }

export const getUsageAnalytics = async (query: AnalyticsQuery): Promise<{
    metrics: AnalyticsMetric[]
    summary: {
      total: number
      average: number
      change: number
      changePercent: number
    }
  }> => {
    return client.post('/v3/analytics/requests', query)
  }

export const getModelAnalytics = async (
    modelId?: string,
    timeRange: string = '7d'
  ): Promise<{
    models: ModelUsage[]
    trends: Record<string, AnalyticsMetric[]>
  }> => {
    return client.get('/v3/analytics/models', {
      modelId,
      timeRange,
    })
  }

export const getProviderAnalytics = async (
    providerId?: string,
    timeRange: string = '7d'
  ): Promise<{
    providers: ProviderUsage[]
    healthScores: Record<string, number>
    latencyTrends: Record<string, AnalyticsMetric[]>
  }> => {
    return client.get('/v3/analytics/providers', {
      providerId,
      timeRange,
    })
  }

export const getRequestAnalytics = async (
    query: Partial<AnalyticsQuery> = {}
  ): Promise<{
    totalRequests: number
    successRate: number
    errorRate: number
    averageLatency: number
    requestsByStatus: Record<string, number>
    latencyPercentiles: {
      p50: number
      p95: number
      p99: number
    }
  }> => {
    return client.post('/v3/analytics/requests', query)
  }

export const getQualityAnalytics = async (
    timeRange: string = '7d'
  ): Promise<{
    averageQualityScore: number
    qualityTrend: AnalyticsMetric[]
    qualityByModel: Array<{
      modelId: string
      modelName: string
      averageScore: number
      totalEvaluations: number
    }>
    qualityDistribution: Record<string, number>
  }> => {
    return client.get('/v3/analytics/quality', { timeRange })
  }

export const runCustomQuery = async (query: {
    query: string
    parameters?: Record<string, any>
    timeRange?: string
  }): Promise<{
    results: any[]
    columns: string[]
    rowCount: number
    executionTime: number
  }> => {
    return client.post('/v3/analytics/query', query)
  }

export const exportAnalytics = async (
    query: AnalyticsQuery,
    format: 'csv' | 'json' | 'excel' = 'csv'
  ): Promise<{
    downloadUrl: string
    filename: string
    expiresAt: string
  }> => {
    return client.post('/v3/analytics/export', { ...query, format })
  }

export const getRealTimeMetrics = async (): Promise<{
    activeRequests: number
    requestsPerSecond: number
    averageLatency: number
    errorRate: number
    topModels: Array<{
      modelId: string
      requestCount: number
    }>
    lastUpdated: string
  }> => {
    return client.get('/v3/analytics/realtime')
  }

