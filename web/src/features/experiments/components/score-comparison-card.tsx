'use client'

import { Star } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DiffLabel } from './diff-label'
import { ScoreProgressBar, DeltaPercentage } from './comparison'
import { formatScoreStats } from '../lib/calculate-diff'
import type { ScoreComparisonRow, ExperimentComparisonSummary } from '../types'

interface ScoreComparisonCardProps {
  row: ScoreComparisonRow
  experiments: Record<string, ExperimentComparisonSummary>
  experimentIds: string[]
  baselineId?: string
  onBaselineChange: (id: string) => void
  preferNegativeDiff?: boolean
}

export function ScoreComparisonCard({
  row,
  experiments,
  experimentIds,
  baselineId,
  onBaselineChange,
  preferNegativeDiff = false,
}: ScoreComparisonCardProps) {
  // Get baseline stats for delta percentage calculation
  const baselineStats = baselineId ? row.experiments[baselineId]?.stats : null

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">{row.scoreName}</CardTitle>
      </CardHeader>
      <CardContent>
        <TooltipProvider>
          <div className="space-y-3">
            {experimentIds.map((expId) => {
              const data = row.experiments[expId]
              const experiment = experiments[expId]
              const isBaseline = expId === baselineId

              if (!data || !experiment) return null

              return (
                <div
                  key={expId}
                  className={cn(
                    'p-2 rounded-md space-y-2',
                    isBaseline && 'bg-muted/50'
                  )}
                >
                  {/* Header row with name and stats */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6"
                            onClick={() => onBaselineChange(expId)}
                            aria-pressed={isBaseline}
                            aria-label={
                              isBaseline
                                ? `${experiment.name} is baseline`
                                : `Set ${experiment.name} as baseline`
                            }
                          >
                            <Star
                              className={cn(
                                'h-4 w-4',
                                isBaseline
                                  ? 'fill-yellow-400 text-yellow-400'
                                  : 'text-muted-foreground'
                              )}
                            />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          {isBaseline ? 'Current baseline' : 'Set as baseline'}
                        </TooltipContent>
                      </Tooltip>
                      <span className="text-sm font-medium">{experiment.name}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="text-sm font-mono">
                            {formatScoreStats(data.stats.mean, data.stats.std_dev)}
                          </span>
                        </TooltipTrigger>
                        <TooltipContent>
                          <div className="text-xs space-y-1">
                            <div>Min: {data.stats.min.toFixed(3)}</div>
                            <div>Max: {data.stats.max.toFixed(3)}</div>
                            <div>Count: {data.stats.count}</div>
                          </div>
                        </TooltipContent>
                      </Tooltip>

                      {/* Delta percentage and diff label for non-baseline */}
                      {!isBaseline && baselineStats && (
                        <DeltaPercentage
                          baseline={baselineStats.mean}
                          current={data.stats.mean}
                          preferNegativeDiff={preferNegativeDiff}
                        />
                      )}

                      {!isBaseline && (
                        <DiffLabel
                          diff={data.diff}
                          preferNegativeDiff={preferNegativeDiff}
                        />
                      )}
                    </div>
                  </div>

                  {/* Progress bar */}
                  <ScoreProgressBar
                    value={data.stats.mean}
                    min={data.stats.min}
                    max={data.stats.max}
                  />
                </div>
              )
            })}
          </div>
        </TooltipProvider>
      </CardContent>
    </Card>
  )
}
