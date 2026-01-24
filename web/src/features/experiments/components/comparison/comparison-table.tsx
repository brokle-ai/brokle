'use client'

import { Star } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { formatScoreStats } from '../../lib/calculate-diff'
import { ScoreProgressBar } from './score-progress-bar'
import { DeltaPercentage } from './delta-percentage'
import type { ScoreComparisonRow, ExperimentComparisonSummary } from '../../types'

interface ComparisonTableProps {
  scoreRows: ScoreComparisonRow[]
  experiments: Record<string, ExperimentComparisonSummary>
  experimentIds: string[]
  baselineId?: string
  onBaselineChange: (id: string) => void
  preferNegativeDiff?: boolean
  className?: string
}

/**
 * Table-based comparison view (Opik/Langfuse pattern)
 * Scores as rows, experiments as columns
 */
export function ComparisonTable({
  scoreRows,
  experiments,
  experimentIds,
  baselineId,
  onBaselineChange,
  preferNegativeDiff = false,
  className,
}: ComparisonTableProps) {
  // Get baseline data for percentage calculations
  const getBaselineStats = (row: ScoreComparisonRow) => {
    if (!baselineId) return null
    return row.experiments[baselineId]?.stats ?? null
  }

  return (
    <TooltipProvider>
      <div className={cn('rounded-md border', className)}>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[180px]">Metric</TableHead>
              {experimentIds.map((expId) => {
                const experiment = experiments[expId]
                const isBaseline = expId === baselineId
                return (
                  <TableHead key={expId} className="min-w-[200px]">
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
                                ? `${experiment?.name} is baseline`
                                : `Set ${experiment?.name} as baseline`
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
                      <span className="font-medium truncate">
                        {experiment?.name ?? expId}
                      </span>
                      {isBaseline && (
                        <span className="text-xs text-muted-foreground">(baseline)</span>
                      )}
                    </div>
                  </TableHead>
                )
              })}
            </TableRow>
          </TableHeader>
          <TableBody>
            {scoreRows.map((row) => {
              const baselineStats = getBaselineStats(row)

              return (
                <TableRow key={row.scoreName}>
                  <TableCell className="font-medium">{row.scoreName}</TableCell>
                  {experimentIds.map((expId) => {
                    const data = row.experiments[expId]
                    const isBaseline = expId === baselineId

                    if (!data) {
                      return (
                        <TableCell key={expId} className="text-muted-foreground">
                          -
                        </TableCell>
                      )
                    }

                    return (
                      <TableCell key={expId}>
                        <div className="space-y-2">
                          {/* Stats row */}
                          <div className="flex items-center justify-between gap-2">
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

                            {/* Show delta percentage for non-baseline */}
                            {!isBaseline && baselineStats && (
                              <DeltaPercentage
                                baseline={baselineStats.mean}
                                current={data.stats.mean}
                                preferNegativeDiff={preferNegativeDiff}
                              />
                            )}
                          </div>

                          {/* Progress bar */}
                          <ScoreProgressBar
                            value={data.stats.mean}
                            min={data.stats.min}
                            max={data.stats.max}
                          />
                        </div>
                      </TableCell>
                    )
                  })}
                </TableRow>
              )
            })}

            {scoreRows.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={experimentIds.length + 1}
                  className="h-24 text-center text-muted-foreground"
                >
                  No scores to compare
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </TooltipProvider>
  )
}
