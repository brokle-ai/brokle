import { Link } from '@tanstack/react-router'
import { ArrowRight, TrendingUp, TrendingDown, Minus } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { ScoreSummary } from '../api/types'

interface ScoreOverviewProps {
  data: ScoreSummary[] | null | undefined
  orgId: string
  projectId: string
  className?: string
}

function formatScoreValue(value: number): string {
  if (value < 0.01) return '0.00'
  if (value < 1) return value.toFixed(2)
  if (value < 10) return value.toFixed(1)
  return Math.round(value).toString()
}

function formatTrend(trend: number): string {
  const abs = Math.abs(trend)
  if (abs < 0.1) return '0%'
  if (abs < 10) return `${abs.toFixed(1)}%`
  return `${Math.round(abs)}%`
}

function TrendIndicator({ trend }: { trend: number }) {
  if (Math.abs(trend) < 0.1)
    return <Minus className="h-3 w-3 text-muted-foreground" />
  if (trend > 0) return <TrendingUp className="h-3 w-3 text-green-500" />
  return <TrendingDown className="h-3 w-3 text-red-500" />
}

function MiniSparkline({
  data,
}: {
  data: { timestamp: string; value: number }[]
}) {
  if (!data || data.length < 2)
    return <div className="h-8 w-16 bg-muted rounded" />

  const values = data.map((d) => d.value)
  const min = Math.min(...values)
  const max = Math.max(...values)
  const range = max - min || 1
  const points = data
    .map((d, i) => {
      const x = (i / (data.length - 1)) * 64
      const y = 32 - ((d.value - min) / range) * 28
      return `${x},${y}`
    })
    .join(' ')

  return (
    <svg width="64" height="32" className="text-primary">
      <polyline
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        points={points}
      />
    </svg>
  )
}

export function ScoreOverview({
  data,
  orgId,
  projectId,
  className,
}: ScoreOverviewProps) {
  if (!data || data.length === 0) return null

  const displayScores = data.slice(0, 3)

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base font-medium">Score Overview</CardTitle>
        <Button asChild variant="ghost" size="sm" className="gap-1 text-xs">
          <Link
            to="/o/$orgId/p/$projectId/scores"
            params={{ orgId, projectId }}
            search={{
              page: 1,
              limit: 20,
              name: undefined,
              source: undefined,
              type: undefined,
              traceId: undefined,
              spanId: undefined,
            }}
          >
            View All
            <ArrowRight className="h-3 w-3" />
          </Link>
        </Button>
      </CardHeader>
      <CardContent>
        <div className="grid gap-4 md:grid-cols-3">
          {displayScores.map((score) => (
            <Link
              key={score.name}
              to="/o/$orgId/p/$projectId/scores"
              params={{ orgId, projectId }}
              search={{
                page: 1,
                limit: 20,
                name: score.name,
                source: undefined,
                type: undefined,
                traceId: undefined,
                spanId: undefined,
              }}
              className="p-3 rounded-lg bg-muted/30 hover:bg-muted/50 transition-colors block"
            >
              <div
                className="text-sm text-muted-foreground mb-1 truncate"
                title={score.name}
              >
                {score.name}
              </div>
              <div className="flex items-center gap-2 mb-2">
                <span className="text-xl font-semibold">
                  {formatScoreValue(score.avg_value)}
                </span>
                <div className="flex items-center gap-1 text-xs">
                  <TrendIndicator trend={score.trend} />
                  <span
                    className={cn(
                      Math.abs(score.trend) < 0.1
                        ? 'text-muted-foreground'
                        : score.trend > 0
                          ? 'text-green-500'
                          : 'text-red-500',
                    )}
                  >
                    {formatTrend(score.trend)}
                  </span>
                </div>
              </div>
              <MiniSparkline data={score.sparkline} />
            </Link>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
