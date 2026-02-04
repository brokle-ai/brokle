'use client'

import * as React from 'react'
import { Activity, DollarSign, Clock, AlertTriangle } from 'lucide-react'
import { MetricCard } from '@/components/shared/metrics/metric-card'
import { cn } from '@/lib/utils'
import type { OverviewStats } from '../types'

interface StatsRowProps {
  stats: OverviewStats | null
  isLoading?: boolean
  error?: string | null
  className?: string
}

function formatCurrency(value: number): string {
  if (value < 0.01) return '$0.00'
  if (value < 1) return `$${value.toFixed(3)}`
  if (value < 100) return `$${value.toFixed(2)}`
  if (value < 1000) return `$${value.toFixed(1)}`
  if (value < 10000) return `$${(value / 1000).toFixed(1)}k`
  return `$${(value / 1000).toFixed(0)}k`
}

function formatLatency(ms: number): string {
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatPercentage(value: number): string {
  if (value < 0.1) return '<0.1%'
  if (value < 10) return `${value.toFixed(1)}%`
  return `${Math.round(value)}%`
}

function formatNumber(value: number): string {
  if (value < 1000) return value.toString()
  if (value < 10000) return `${(value / 1000).toFixed(1)}k`
  if (value < 1000000) return `${Math.round(value / 1000)}k`
  return `${(value / 1000000).toFixed(1)}M`
}

export function StatsRow({ stats, isLoading, error, className }: StatsRowProps) {
  const metrics = React.useMemo(() => {
    if (!stats) return []

    return [
      {
        title: 'Traces',
        value: formatNumber(stats.traces_count),
        trend: stats.traces_trend !== 0 ? {
          value: Math.abs(Math.round(stats.traces_trend)),
          label: 'vs previous period',
          direction: stats.traces_trend > 0 ? 'up' as const : 'down' as const,
        } : undefined,
        icon: Activity,
      },
      {
        title: 'Total Cost',
        value: formatCurrency(stats.total_cost),
        trend: stats.cost_trend !== 0 ? {
          value: Math.abs(Math.round(stats.cost_trend)),
          label: 'vs previous period',
          direction: stats.cost_trend > 0 ? 'up' as const : 'down' as const,
          isPositiveWhenDown: true,  // cost decrease is good
        } : undefined,
        icon: DollarSign,
      },
      {
        title: 'Avg Latency',
        value: formatLatency(stats.avg_latency_ms),
        trend: stats.latency_trend !== 0 ? {
          value: Math.abs(Math.round(stats.latency_trend)),
          label: 'vs previous period',
          direction: stats.latency_trend > 0 ? 'up' as const : 'down' as const,
          isPositiveWhenDown: true,  // latency decrease is good
        } : undefined,
        icon: Clock,
      },
      {
        title: 'Error Rate',
        value: formatPercentage(stats.error_rate),
        trend: stats.error_rate_trend !== 0 ? {
          value: Math.abs(Math.round(stats.error_rate_trend)),
          label: 'vs previous period',
          direction: stats.error_rate_trend > 0 ? 'up' as const : 'down' as const,
          isPositiveWhenDown: true,  // error rate decrease is good
        } : undefined,
        icon: AlertTriangle,
      },
    ]
  }, [stats])

  if (isLoading) {
    return (
      <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-4', className)}>
        {[1, 2, 3, 4].map((i) => (
          <MetricCard
            key={i}
            title=""
            value=""
            loading
          />
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-4', className)}>
        {[1, 2, 3, 4].map((i) => (
          <MetricCard
            key={i}
            title="Error"
            value="-"
            error={error}
          />
        ))}
      </div>
    )
  }

  return (
    <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-4', className)}>
      {metrics.map((metric) => (
        <MetricCard
          key={metric.title}
          title={metric.title}
          value={metric.value}
          icon={metric.icon}
          trend={metric.trend}
        />
      ))}
    </div>
  )
}
