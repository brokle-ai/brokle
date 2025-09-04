import { BrokleAPIClient } from '../core/client'
import type { QueryParams } from '../core/types'

// Dashboard-specific types
export interface QuickStat {
  label: string
  value: string | number
  change?: number
  changeType?: 'increase' | 'decrease' | 'neutral'
  icon?: string
}

export interface ChartData {
  labels: string[]
  datasets: Array<{
    label: string
    data: number[]
    backgroundColor?: string
    borderColor?: string
  }>
}

export interface DashboardOverview {
  quickStats: QuickStat[]
  requestTrend: ChartData
  costTrend: ChartData
  topModels: Array<{
    name: string
    requests: number
    cost: number
    change: number
  }>
  recentActivity: Array<{
    id: string
    type: 'request' | 'error' | 'cost_alert' | 'model_change'
    message: string
    timestamp: string
    severity?: 'info' | 'warning' | 'error'
  }>
  alerts: Array<{
    id: string
    type: 'budget' | 'error_rate' | 'latency' | 'quota'
    message: string
    severity: 'info' | 'warning' | 'error'
    timestamp: string
    acknowledged: boolean
  }>
}

export class DashboardAPIClient extends BrokleAPIClient {
  constructor() {
    super('/dashboard')
  }

  // Main dashboard overview
  async getOverview(timeRange: string = '24h'): Promise<DashboardOverview> {
    return this.get<DashboardOverview>('/v1/dashboard/overview', { timeRange })
  }

  // Quick stats
  async getQuickStats(timeRange: string = '24h'): Promise<QuickStat[]> {
    return this.get<QuickStat[]>('/v1/dashboard/stats', { timeRange })
  }

  // Recent activity
  async getRecentActivity(limit: number = 10): Promise<DashboardOverview['recentActivity']> {
    return this.get('/v1/dashboard/activity', { limit })
  }

  // System alerts
  async getAlerts(acknowledged?: boolean): Promise<DashboardOverview['alerts']> {
    return this.get('/v1/dashboard/alerts', { acknowledged })
  }

  // Acknowledge alert
  async acknowledgeAlert(alertId: string): Promise<void> {
    await this.patch(`/v1/dashboard/alerts/${alertId}`, { acknowledged: true })
  }

  // Dismiss alert
  async dismissAlert(alertId: string): Promise<void> {
    await this.delete(`/v1/dashboard/alerts/${alertId}`)
  }

  // Dashboard widgets data
  async getWidgetData(widgetType: string, config?: QueryParams): Promise<any> {
    return this.get(`/v1/dashboard/widgets/${widgetType}`, config)
  }

  // Save dashboard configuration
  async saveDashboardConfig(config: {
    widgets: Array<{
      id: string
      type: string
      position: { x: number; y: number }
      size: { width: number; height: number }
      config: Record<string, any>
    }>
    layout: string
  }): Promise<void> {
    await this.post('/v1/dashboard/config', config)
  }

  // Load dashboard configuration
  async getDashboardConfig(): Promise<{
    widgets: any[]
    layout: string
    lastUpdated: string
  }> {
    return this.get('/v1/dashboard/config')
  }
}