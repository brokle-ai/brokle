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

export class AnalyticsAPIClient extends BrokleAPIClient {
  constructor() {
    super('/analytics') // All analytics endpoints prefixed with /analytics
  }

  // Dashboard analytics
  async getDashboardStats(
    timeRange: string = '24h', 
    organizationId?: string
  ): Promise<DashboardStats> {
    return this.get<DashboardStats>('/v1/analytics/dashboard', {
      timeRange,
      organizationId,
    })
  }

  // Cost analytics
  async getCostAnalytics(query: Partial<AnalyticsQuery> = {}): Promise<CostAnalytics> {
    return this.post<CostAnalytics>('/v1/analytics/costs', query)
  }

  // Usage analytics
  async getUsageAnalytics(query: AnalyticsQuery): Promise<{
    metrics: AnalyticsMetric[]
    summary: {
      total: number
      average: number
      change: number
      changePercent: number
    }
  }> {
    return this.post('/v1/analytics/usage', query)
  }

  // Model performance
  async getModelAnalytics(
    modelId?: string,
    timeRange: string = '7d'
  ): Promise<{
    models: ModelUsage[]
    trends: Record<string, AnalyticsMetric[]>
  }> {
    return this.get('/v1/analytics/models', {
      modelId,
      timeRange,
    })
  }

  // Provider performance
  async getProviderAnalytics(
    providerId?: string,
    timeRange: string = '7d'
  ): Promise<{
    providers: ProviderUsage[]
    healthScores: Record<string, number>
    latencyTrends: Record<string, AnalyticsMetric[]>
  }> {
    return this.get('/v1/analytics/providers', {
      providerId,
      timeRange,
    })
  }

  // Request analytics
  async getRequestAnalytics(
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
  }> {
    return this.post('/v1/analytics/requests', query)
  }

  // Quality analytics
  async getQualityAnalytics(
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
  }> {
    return this.get('/v1/analytics/quality', { timeRange })
  }

  // Custom queries
  async runCustomQuery(query: {
    query: string
    parameters?: Record<string, any>
    timeRange?: string
  }): Promise<{
    results: any[]
    columns: string[]
    rowCount: number
    executionTime: number
  }> {
    return this.post('/v1/analytics/query', query)
  }

  // Export data
  async exportAnalytics(
    query: AnalyticsQuery,
    format: 'csv' | 'json' | 'excel' = 'csv'
  ): Promise<{
    downloadUrl: string
    filename: string
    expiresAt: string
  }> {
    return this.post('/v1/analytics/export', { ...query, format })
  }

  // Real-time metrics
  async getRealTimeMetrics(): Promise<{
    activeRequests: number
    requestsPerSecond: number
    averageLatency: number
    errorRate: number
    topModels: Array<{
      modelId: string
      requestCount: number
    }>
    lastUpdated: string
  }> {
    return this.get('/v1/analytics/realtime')
  }
}