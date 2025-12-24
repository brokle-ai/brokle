'use client'

import { useState, useMemo } from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Loader2, X } from 'lucide-react'
import { useScoreAnalyticsQuery, useScoreNamesQuery } from '../../hooks/use-score-analytics'
import type { ScoreAnalyticsParams } from '../../types'
import { StatisticsCard } from './statistics-card'
import { TimelineChartCard } from './timeline-chart-card'
import { DistributionCard } from './distribution-card'
import { HeatmapCard } from './heatmap-card'

interface ScoreAnalyticsDashboardProps {
  projectId: string
}

export function ScoreAnalyticsDashboard({ projectId }: ScoreAnalyticsDashboardProps) {
  const [selectedScore, setSelectedScore] = useState<string>('')
  const [compareScore, setCompareScore] = useState<string>('')
  const [interval, setInterval] = useState<'hour' | 'day' | 'week'>('day')

  const { data: scoreNames, isLoading: isLoadingNames } = useScoreNamesQuery(projectId)

  const analyticsParams: ScoreAnalyticsParams | undefined = useMemo(() => {
    if (!selectedScore) return undefined
    return {
      score_name: selectedScore,
      compare_score_name: compareScore || undefined,
      interval,
    }
  }, [selectedScore, compareScore, interval])

  const {
    data: analytics,
    isLoading: isLoadingAnalytics,
    error,
  } = useScoreAnalyticsQuery(projectId, analyticsParams)

  const clearCompareScore = () => {
    setCompareScore('')
  }

  if (!isLoadingNames && (!scoreNames || scoreNames.length === 0)) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="text-muted-foreground mb-2">No scores found in this project</p>
        <p className="text-sm text-muted-foreground">
          Scores will appear here once you start recording quality evaluations
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Controls */}
      <div className="flex flex-wrap items-center gap-4">
        {/* Primary Score Selector */}
        <div className="flex flex-col gap-1">
          <label htmlFor="score-select" className="text-xs text-muted-foreground">Score</label>
          <Select
            value={selectedScore}
            onValueChange={setSelectedScore}
            disabled={isLoadingNames}
          >
            <SelectTrigger id="score-select" className="w-[200px]" aria-label="Select primary score">
              <SelectValue placeholder="Select a score" />
            </SelectTrigger>
            <SelectContent>
              {scoreNames?.map((name) => (
                <SelectItem key={name} value={name}>
                  {name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Compare Score Selector */}
        <div className="flex flex-col gap-1">
          <label htmlFor="compare-select" className="text-xs text-muted-foreground">Compare with</label>
          <div className="flex items-center gap-2">
            <Select
              value={compareScore}
              onValueChange={setCompareScore}
              disabled={isLoadingNames || !selectedScore}
            >
              <SelectTrigger id="compare-select" className="w-[200px]" aria-label="Select comparison score">
                <SelectValue placeholder="Optional comparison" />
              </SelectTrigger>
              <SelectContent>
                {scoreNames
                  ?.filter((name) => name !== selectedScore)
                  .map((name) => (
                    <SelectItem key={name} value={name}>
                      {name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
            {compareScore && (
              <Button
                variant="ghost"
                size="icon"
                onClick={clearCompareScore}
                className="h-8 w-8"
                aria-label="Clear comparison score"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </Button>
            )}
          </div>
        </div>

        {/* Interval Selector */}
        <div className="flex flex-col gap-1">
          <label htmlFor="interval-select" className="text-xs text-muted-foreground">Interval</label>
          <Select
            value={interval}
            onValueChange={(v) => setInterval(v as 'hour' | 'day' | 'week')}
            disabled={!selectedScore}
          >
            <SelectTrigger id="interval-select" className="w-[120px]" aria-label="Select time interval">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="hour">Hourly</SelectItem>
              <SelectItem value="day">Daily</SelectItem>
              <SelectItem value="week">Weekly</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {!selectedScore && (
        <div className="flex flex-col items-center justify-center py-16 text-center border rounded-lg bg-muted/30">
          <p className="text-muted-foreground">
            Select a score to view analytics
          </p>
        </div>
      )}

      {/* Loading state */}
      {selectedScore && isLoadingAnalytics && (
        <div
          className="flex items-center justify-center py-16"
          role="status"
          aria-live="polite"
        >
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" aria-hidden="true" />
          <span className="sr-only">Loading analytics data...</span>
        </div>
      )}

      {/* Error state */}
      {error && (
        <div className="flex flex-col items-center justify-center py-16 text-center border rounded-lg border-destructive/20 bg-destructive/5">
          <p className="text-destructive mb-2">Failed to load analytics</p>
          <p className="text-sm text-muted-foreground">
            {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        </div>
      )}

      {/* Analytics Dashboard Grid */}
      {analytics && !isLoadingAnalytics && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Statistics Card */}
          <StatisticsCard
            statistics={analytics.statistics}
            compareStatistics={analytics.compare_statistics}
            comparison={analytics.comparison}
            scoreName={selectedScore}
            compareScoreName={compareScore || undefined}
          />

          {/* Timeline Chart */}
          <TimelineChartCard
            timeSeries={analytics.time_series}
            compareTimeSeries={analytics.compare_time_series}
            scoreName={selectedScore}
            compareScoreName={compareScore || undefined}
            mean={analytics.statistics.mean}
            compareMean={analytics.compare_statistics?.mean}
          />

          {/* Distribution Card */}
          <DistributionCard
            distribution={analytics.distribution}
            compareDistribution={analytics.compare_distribution}
            scoreName={selectedScore}
            compareScoreName={compareScore || undefined}
          />

          {/* Heatmap Card */}
          <HeatmapCard
            heatmap={analytics.heatmap}
            scoreName={selectedScore}
            compareScoreName={compareScore || undefined}
          />
        </div>
      )}
    </div>
  )
}
