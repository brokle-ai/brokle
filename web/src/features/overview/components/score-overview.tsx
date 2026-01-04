'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ArrowRight, TrendingUp, TrendingDown, Minus } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ScoreSummary } from '../types'

interface ScoreOverviewProps {
  data: ScoreSummary[] | null
  isLoading?: boolean
  error?: string | null
  className?: string
  projectSlug?: string
}

function formatScoreValue(value: number): string {
  if (value < 0.01) return '0.00'
  if (value < 1) return value.toFixed(2)
  if (value < 10) return value.toFixed(1)
  return Math.round(value).toString()
}

function formatTrend(trend: number): string {
  const absValue = Math.abs(trend)
  if (absValue < 0.1) return '0%'
  if (absValue < 10) return `${absValue.toFixed(1)}%`
  return `${Math.round(absValue)}%`
}

function TrendIndicator({ trend }: { trend: number }) {
  if (Math.abs(trend) < 0.1) {
    return <Minus className="h-3 w-3 text-muted-foreground" />
  }
  if (trend > 0) {
    return <TrendingUp className="h-3 w-3 text-green-500" />
  }
  return <TrendingDown className="h-3 w-3 text-red-500" />
}

function MiniSparkline({ data }: { data: { timestamp: string; value: number }[] }) {
  if (!data || data.length < 2) {
    return <div className="h-8 w-16 bg-muted rounded" />
  }

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
  isLoading,
  error,
  className,
  projectSlug,
}: ScoreOverviewProps) {
  const router = useRouter()

  // Don't render if no scores exist
  if (!isLoading && (!data || data.length === 0)) {
    return null
  }

  const handleViewAll = () => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/evaluations/scores`)
    }
  }

  const handleScoreClick = (scoreName: string) => {
    if (projectSlug) {
      router.push(`/projects/${projectSlug}/evaluations/scores?score=${encodeURIComponent(scoreName)}`)
    }
  }

  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium">Score Overview</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="p-3 rounded-lg bg-muted/30">
                <Skeleton className="h-4 w-20 mb-2" />
                <Skeleton className="h-6 w-12 mb-2" />
                <Skeleton className="h-8 w-16" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className={cn('border-destructive', className)}>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base font-medium text-destructive">
            Score Overview
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">{error}</p>
        </CardContent>
      </Card>
    )
  }

  // Take top 3 scores
  const displayScores = data?.slice(0, 3) ?? []

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base font-medium">Score Overview</CardTitle>
        <Button
          variant="ghost"
          size="sm"
          className="gap-1 text-xs"
          onClick={handleViewAll}
        >
          View All
          <ArrowRight className="h-3 w-3" />
        </Button>
      </CardHeader>
      <CardContent>
        <div className="grid gap-4 md:grid-cols-3">
          {displayScores.map((score) => (
            <div
              key={score.name}
              className="p-3 rounded-lg bg-muted/30 hover:bg-muted/50 cursor-pointer transition-colors"
              onClick={() => handleScoreClick(score.name)}
              role="button"
              tabIndex={0}
            >
              <div className="text-sm text-muted-foreground mb-1 truncate" title={score.name}>
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
                          : 'text-red-500'
                    )}
                  >
                    {formatTrend(score.trend)}
                  </span>
                </div>
              </div>
              <MiniSparkline data={score.sparkline} />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
