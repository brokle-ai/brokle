import { AIRequest, AIProvider, AIModel } from './ai'
import { TimeSeries, ModelUsage, ProviderUsage } from './api'

export interface DashboardData {
  overview: DashboardOverview
  metrics: DashboardMetrics
  charts: DashboardCharts
  recentActivity: RecentActivity[]
  alerts: DashboardAlert[]
}

export interface DashboardOverview {
  totalRequests: number
  totalCost: number
  averageLatency: number
  errorRate: number
  activeModels: number
  activeProviders: number
  cacheHitRate: number
  costOptimization: number
}

export interface DashboardMetrics {
  requestsToday: MetricChange
  costToday: MetricChange
  latencyToday: MetricChange
  errorsToday: MetricChange
  qualityScore: MetricChange
  uptime: MetricChange
}

export interface MetricChange {
  current: number
  previous: number
  change: number
  trend: 'up' | 'down' | 'stable'
  unit: string
}

export interface DashboardCharts {
  requestsOverTime: TimeSeries[]
  costOverTime: TimeSeries[]
  latencyOverTime: TimeSeries[]
  errorRateOverTime: TimeSeries[]
  modelUsage: ModelUsage[]
  providerDistribution: ProviderUsage[]
  geographicDistribution: GeographicUsage[]
}

export interface GeographicUsage {
  region: string
  country: string
  requests: number
  percentage: number
  avgLatency: number
}

export interface RecentActivity {
  id: string
  type: ActivityType
  title: string
  description: string
  timestamp: string
  severity: ActivitySeverity
  metadata?: Record<string, any>
}

export interface DashboardAlert {
  id: string
  type: AlertType
  title: string
  message: string
  severity: AlertSeverity
  timestamp: string
  acknowledged: boolean
  actions?: AlertAction[]
}

export interface AlertAction {
  label: string
  action: string
  url?: string
}

export type ActivityType = 
  | 'request' 
  | 'error' 
  | 'provider_change' 
  | 'model_update' 
  | 'cost_alert' 
  | 'quality_degradation'
  | 'rate_limit'
  | 'cache_miss'

export type ActivitySeverity = 'info' | 'warning' | 'error' | 'critical'

export type AlertType = 
  | 'cost_budget' 
  | 'error_rate' 
  | 'latency_spike' 
  | 'provider_down' 
  | 'quality_drop'
  | 'rate_limit_hit'
  | 'token_limit'

export type AlertSeverity = 'low' | 'medium' | 'high' | 'critical'

// Widget types for customizable dashboard
export interface DashboardWidget {
  id: string
  type: WidgetType
  title: string
  position: WidgetPosition
  size: WidgetSize
  config: WidgetConfig
  data?: any
}

export interface WidgetPosition {
  x: number
  y: number
}

export interface WidgetSize {
  width: number
  height: number
}

export interface WidgetConfig {
  refreshInterval?: number
  timeRange?: string
  filters?: Record<string, any>
  visualization?: VisualizationType
}

export type WidgetType = 
  | 'metric_card' 
  | 'line_chart' 
  | 'bar_chart' 
  | 'pie_chart' 
  | 'table' 
  | 'heatmap'
  | 'activity_feed'
  | 'alert_list'

export type VisualizationType = 
  | 'line' 
  | 'area' 
  | 'bar' 
  | 'pie' 
  | 'donut' 
  | 'scatter' 
  | 'heatmap'

// Dashboard customization
export interface DashboardLayout {
  id: string
  name: string
  isDefault: boolean
  widgets: DashboardWidget[]
  createdAt: string
  updatedAt: string
}

export interface DashboardPreferences {
  defaultLayout: string
  refreshInterval: number
  timezone: string
  dateFormat: string
  currency: string
  theme: 'light' | 'dark' | 'system'
}