import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
  Clock,
  Loader2,
  Target,
  TrendingUp,
} from 'lucide-react'
import { format, parseISO } from 'date-fns'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from 'recharts'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { evaluatorAnalyticsQueryOptions } from '../api/queries'
import type {
  AnalyticsPeriod,
  DistributionBucket,
  ErrorSummary,
  EvaluatorAnalyticsParams,
  TimeSeriesPoint,
} from '../api/types'

interface EvaluatorAnalyticsTabProps {
  projectId: string
  evaluatorId: string
}

const CHART_COLORS = {
  primary: 'hsl(220, 70%, 50%)',
  secondary: 'hsl(150, 60%, 40%)',
} as const

function formatNumber(value: number, precision = 2): string {
  if (Number.isInteger(value)) return value.toLocaleString()
  return value.toFixed(precision)
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms.toFixed(0)}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  return `${(ms / 60_000).toFixed(1)}m`
}

interface StatCardProps {
  title: string
  value: string | number
  icon: React.ReactNode
  subtitle?: string
}

function StatCard({ title, value, icon, subtitle }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{title}</span>
            <span className="text-2xl font-bold">
              {typeof value === 'number' ? formatNumber(value) : value}
            </span>
            {subtitle ? (
              <span className="text-xs text-muted-foreground">{subtitle}</span>
            ) : null}
          </div>
          <div className="text-muted-foreground">{icon}</div>
        </div>
      </CardContent>
    </Card>
  )
}

interface TrendChartProps {
  data: TimeSeriesPoint[]
}

function TrendChart({ data }: TrendChartProps) {
  const chartData = useMemo(
    () =>
      data.map((p) => ({
        formattedDate: format(parseISO(p.timestamp), 'MMM d'),
        count: p.count,
        successRate: p.success_rate * 100,
      })),
    [data],
  )

  if (chartData.length === 0) {
    return (
      <div className="flex h-[250px] items-center justify-center text-muted-foreground">
        No execution data available
      </div>
    )
  }

  return (
    <div className="h-[250px]">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart
          data={chartData}
          margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis
            dataKey="formattedDate"
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            className="text-muted-foreground"
          />
          <YAxis
            yAxisId="left"
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            className="text-muted-foreground"
          />
          <YAxis
            yAxisId="right"
            orientation="right"
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            domain={[0, 100]}
            className="text-muted-foreground"
          />
          <RechartsTooltip
            content={({ active, payload, label }) => {
              if (!active || !payload || payload.length === 0) return null
              return (
                <div className="rounded-lg border bg-popover p-3 shadow-lg">
                  <p className="mb-2 text-sm font-medium">{label}</p>
                  {payload.map((entry, idx) => (
                    <div
                      key={idx}
                      className="flex items-center gap-2 text-sm"
                    >
                      <div
                        className="h-3 w-3 rounded-full"
                        style={{ backgroundColor: entry.color }}
                      />
                      <span className="text-muted-foreground">
                        {entry.name}:
                      </span>
                      <span className="font-medium">
                        {entry.name === 'Success Rate'
                          ? `${(entry.value as number).toFixed(1)}%`
                          : entry.value}
                      </span>
                    </div>
                  ))}
                </div>
              )
            }}
          />
          <Line
            yAxisId="left"
            type="monotone"
            dataKey="count"
            name="Executions"
            stroke={CHART_COLORS.primary}
            strokeWidth={2}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
          <Line
            yAxisId="right"
            type="monotone"
            dataKey="successRate"
            name="Success Rate"
            stroke={CHART_COLORS.secondary}
            strokeWidth={2}
            dot={{ r: 3 }}
            activeDot={{ r: 5 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

interface DistributionChartProps {
  data: DistributionBucket[]
}

function DistributionChart({ data }: DistributionChartProps) {
  const chartData = useMemo(() => {
    const total = data.reduce((sum, bin) => sum + bin.count, 0)
    return data.map((bin) => ({
      label:
        bin.bin_start === bin.bin_end
          ? bin.bin_start.toFixed(2)
          : `${bin.bin_start.toFixed(2)}-${bin.bin_end.toFixed(2)}`,
      value: bin.count,
      percentage: total > 0 ? (bin.count / total) * 100 : 0,
    }))
  }, [data])

  if (chartData.length === 0) {
    return (
      <div className="flex h-[200px] items-center justify-center text-muted-foreground">
        No distribution data available
      </div>
    )
  }

  return (
    <div className="h-[200px]">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={chartData}
          margin={{ top: 10, right: 20, left: 10, bottom: 20 }}
        >
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis
            dataKey="label"
            tick={{ fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            angle={-45}
            textAnchor="end"
            height={50}
            className="text-muted-foreground"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            className="text-muted-foreground"
          />
          <RechartsTooltip
            content={({ active, payload }) => {
              if (!active || !payload || payload.length === 0) return null
              const d = payload[0].payload as (typeof chartData)[number]
              return (
                <div className="rounded-lg border bg-popover p-3 shadow-lg">
                  <p className="mb-1 text-sm font-medium">{d.label}</p>
                  <p className="text-sm text-muted-foreground">
                    Count:{' '}
                    <span className="font-medium">
                      {d.value.toLocaleString()}
                    </span>
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Percentage:{' '}
                    <span className="font-medium">
                      {d.percentage.toFixed(1)}%
                    </span>
                  </p>
                </div>
              )
            }}
          />
          <Bar dataKey="value" radius={[4, 4, 0, 0]}>
            {chartData.map((_, idx) => (
              <Cell key={idx} fill={CHART_COLORS.primary} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

interface ErrorListProps {
  errors: ErrorSummary[]
}

function ErrorList({ errors }: ErrorListProps) {
  if (errors.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No errors in this period
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {errors.map((err, idx) => (
        <div
          key={idx}
          className="flex items-start justify-between rounded-lg border border-destructive/20 bg-destructive/5 p-3"
        >
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 shrink-0 text-destructive" />
              <span className="truncate text-sm font-medium">
                {err.error_type}
              </span>
            </div>
            <p className="mt-1 truncate text-xs text-muted-foreground">
              {err.message}
            </p>
          </div>
          <div className="ml-4 flex flex-col items-end gap-1">
            <span className="text-sm font-medium">{err.count}×</span>
            <span className="text-xs text-muted-foreground">
              {format(parseISO(err.last_occurred), 'MMM d, HH:mm')}
            </span>
          </div>
        </div>
      ))}
    </div>
  )
}

export function EvaluatorAnalyticsTab({
  projectId,
  evaluatorId,
}: EvaluatorAnalyticsTabProps) {
  const [period, setPeriod] = useState<AnalyticsPeriod>('7d')

  const params: EvaluatorAnalyticsParams = useMemo(() => ({ period }), [period])

  const {
    data: analytics,
    isLoading,
    error,
  } = useQuery(evaluatorAnalyticsQueryOptions(projectId, evaluatorId, params))

  if (isLoading) {
    return (
      <div
        className="flex items-center justify-center py-16"
        role="status"
        aria-live="polite"
      >
        <Loader2
          className="h-8 w-8 animate-spin text-muted-foreground"
          aria-hidden="true"
        />
        <span className="sr-only">Loading analytics…</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="max-w-md rounded-lg bg-destructive/10 p-6">
          <h3 className="mb-2 font-semibold text-destructive">
            Failed to load analytics
          </h3>
          <p className="text-sm text-muted-foreground">
            {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        </div>
      </div>
    )
  }

  if (!analytics) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="text-muted-foreground">No analytics data available</p>
        <p className="mt-2 text-sm text-muted-foreground">
          Analytics will appear here once the evaluator starts scoring spans.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium">Evaluator Analytics</h3>
        <Select
          value={period}
          onValueChange={(v) => setPeriod(v as AnalyticsPeriod)}
        >
          <SelectTrigger
            className="w-[140px]"
            aria-label="Select time period"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="24h">Last 24 hours</SelectItem>
            <SelectItem value="7d">Last 7 days</SelectItem>
            <SelectItem value="30d">Last 30 days</SelectItem>
            <SelectItem value="90d">Last 90 days</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard
          title="Total Executions"
          value={analytics.total_executions}
          icon={<TrendingUp className="h-5 w-5" />}
        />
        <StatCard
          title="Success Rate"
          value={`${(analytics.success_rate * 100).toFixed(1)}%`}
          icon={<Target className="h-5 w-5" />}
        />
        <StatCard
          title="Avg Score"
          value={
            analytics.average_score !== null
              ? formatNumber(analytics.average_score)
              : 'N/A'
          }
          icon={<Target className="h-5 w-5" />}
        />
        <StatCard
          title="Avg Latency"
          value={formatDuration(analytics.latency_percentiles.avg)}
          icon={<Clock className="h-5 w-5" />}
          subtitle={`p90: ${formatDuration(analytics.latency_percentiles.p90)}`}
        />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-medium">
              Execution Trend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <TrendChart data={analytics.execution_trend} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-medium">
              Score Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <DistributionChart data={analytics.score_distribution} />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-medium">Recent Errors</CardTitle>
        </CardHeader>
        <CardContent>
          <ErrorList errors={analytics.top_errors} />
        </CardContent>
      </Card>
    </div>
  )
}
