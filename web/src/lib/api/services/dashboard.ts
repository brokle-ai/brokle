// Dashboard API - Direct functions for dashboard data

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

export interface DashboardConfig {
  widgets: Array<{
    id: string
    type: string
    position: { x: number; y: number }
    size: { width: number; height: number }
    config: Record<string, any>
  }>
  layout: string
}

// Flexible base client - versions specified per endpoint
const client = new BrokleAPIClient('/api')

// Direct dashboard functions
export const getOverview = async (timeRange: string = '24h'): Promise<DashboardOverview> => {
    return client.get<DashboardOverview>('/v2/analytics/overview', { timeRange })
  }

export const getQuickStats = async (timeRange: string = '24h'): Promise<QuickStat[]> => {
    return client.get<QuickStat[]>('/v2/analytics/overview', { timeRange })
  }

export const getRecentActivity = async (limit: number = 10): Promise<DashboardOverview['recentActivity']> => {
    return client.get('/logs/requests', { limit })
  }

export const getAlerts = async (acknowledged?: boolean): Promise<DashboardOverview['alerts']> => {
    const params = acknowledged !== undefined ? { acknowledged } : {}
    return client.get('/alerts', params)
  }

export const acknowledgeAlert = async (alertId: string): Promise<void> => {
    return client.patch<void>(`/alerts/${alertId}/acknowledge`, {})
  }

export const dismissAlert = async (alertId: string): Promise<void> => {
    return client.delete<void>(`/alerts/${alertId}`)
  }

export const getWidgetData = async (widgetType: string, config?: QueryParams): Promise<any> => {
    return client.get(`/dashboard/widgets/${widgetType}`, config)
  }

export const saveDashboardConfig = async (config: DashboardConfig): Promise<void> => {
    return client.post<void>('/dashboard/config', config)
  }

export const getDashboardConfig = async (): Promise<{
    widgets: any[]
    layout: string
    lastUpdated: string
  }> => {
    return client.get('/dashboard/config')
  }

