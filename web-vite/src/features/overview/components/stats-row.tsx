import { Activity, AlertTriangle, Clock, DollarSign } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { OverviewStats } from '../api/types'

// The four cards the Phase-1.5 overview ships with. Charts
// (trace-volume, cost-by-model, top-errors) are deferred to the second
// port; keeping them out of the default view mirrors how Langfuse's
// open-source dashboard opens — stats first, drilldowns on click.

interface StatsRowProps {
  stats: OverviewStats
  className?: string
}

function formatCurrency(value: number): string {
  if (value === 0) return '$0.00'
  if (value < 0.01) return `$${value.toFixed(4)}`
  if (value < 1) return `$${value.toFixed(3)}`
  if (value < 100) return `$${value.toFixed(2)}`
  if (value < 1000) return `$${value.toFixed(1)}`
  if (value < 10000) return `$${(value / 1000).toFixed(1)}k`
  return `$${(value / 1000).toFixed(0)}k`
}

function formatLatency(ms: number): string {
  if (ms === 0) return '—'
  if (ms < 1) return '<1ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatPercentage(value: number): string {
  if (value === 0) return '0%'
  if (value < 0.1) return '<0.1%'
  if (value < 10) return `${value.toFixed(1)}%`
  return `${Math.round(value)}%`
}

function formatCount(value: number): string {
  if (value < 1000) return value.toString()
  if (value < 10000) return `${(value / 1000).toFixed(1)}k`
  if (value < 1_000_000) return `${Math.round(value / 1000)}k`
  return `${(value / 1_000_000).toFixed(1)}M`
}

interface TrendProps {
  value: number
  // When true, a *decrease* is good (cost, latency, error-rate).
  goodWhenDown?: boolean
}

function Trend({ value, goodWhenDown }: TrendProps) {
  if (value === 0) {
    return <span className="text-xs text-muted-foreground">no change</span>
  }
  const up = value > 0
  const positive = goodWhenDown ? !up : up
  const cls = positive ? 'text-emerald-600' : 'text-destructive'
  const arrow = up ? '↑' : '↓'
  return (
    <span className={cn('text-xs font-medium', cls)}>
      {arrow} {Math.abs(value).toFixed(1)}% vs prev
    </span>
  )
}

interface StatCardProps {
  title: string
  value: string
  trend: number
  goodWhenDown?: boolean
  icon: LucideIcon
}

function StatCard({ title, value, trend, goodWhenDown, icon: Icon }: StatCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-2 space-y-0">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-semibold tabular-nums">{value}</div>
        <div className="mt-1">
          <Trend value={trend} goodWhenDown={goodWhenDown} />
        </div>
      </CardContent>
    </Card>
  )
}

export function StatsRow({ stats, className }: StatsRowProps) {
  return (
    <div className={cn('grid gap-4 md:grid-cols-2 lg:grid-cols-4', className)}>
      <StatCard
        title="Traces"
        value={formatCount(stats.traces_count)}
        trend={stats.traces_trend}
        icon={Activity}
      />
      <StatCard
        title="Total cost"
        value={formatCurrency(stats.total_cost)}
        trend={stats.cost_trend}
        goodWhenDown
        icon={DollarSign}
      />
      <StatCard
        title="Avg latency"
        value={formatLatency(stats.avg_latency_ms)}
        trend={stats.latency_trend}
        goodWhenDown
        icon={Clock}
      />
      <StatCard
        title="Error rate"
        value={formatPercentage(stats.error_rate)}
        trend={stats.error_rate_trend}
        goodWhenDown
        icon={AlertTriangle}
      />
    </div>
  )
}
