'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { HelpCircle } from 'lucide-react'
import type { ScoreStatistics, ComparisonMetrics } from '../../types'
import {
  formatNumber,
  interpretCorrelation,
  interpretCohensKappa,
  interpretMAE,
  interpretRMSE,
  getBadgeColor,
} from '../../lib/statistics-utils'

interface StatisticsCardProps {
  statistics: ScoreStatistics
  compareStatistics?: ScoreStatistics
  comparison?: ComparisonMetrics
  scoreName: string
  compareScoreName?: string
}

function StatItem({
  label,
  value,
  tooltip,
}: {
  label: string
  value: string | number
  tooltip?: string
}) {
  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-1">
        <span className="text-xs text-muted-foreground">{label}</span>
        {tooltip && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <HelpCircle className="h-3 w-3 text-muted-foreground cursor-help" />
              </TooltipTrigger>
              <TooltipContent className="max-w-xs">
                <p className="text-sm">{tooltip}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
      </div>
      <span className="text-sm font-medium">{typeof value === 'number' ? formatNumber(value) : value}</span>
    </div>
  )
}

function MetricWithBadge({
  label,
  value,
  interpretation,
  tooltip,
}: {
  label: string
  value: number
  interpretation: ReturnType<typeof interpretCorrelation>
  tooltip?: string
}) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-1">
        <span className="text-xs text-muted-foreground">{label}</span>
        {tooltip && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <HelpCircle className="h-3 w-3 text-muted-foreground cursor-help" />
              </TooltipTrigger>
              <TooltipContent className="max-w-xs">
                <p className="text-sm">{tooltip}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
      </div>
      <div className="flex items-center gap-2">
        <span className="text-sm font-medium">{formatNumber(value)}</span>
        <Badge variant="outline" className={getBadgeColor(interpretation.color)}>
          {interpretation.strength}
        </Badge>
      </div>
    </div>
  )
}

export function StatisticsCard({
  statistics,
  compareStatistics,
  comparison,
  scoreName,
  compareScoreName,
}: StatisticsCardProps) {
  const range = statistics.max - statistics.min

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-medium">Score Statistics</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Primary Score Statistics */}
        <div>
          <h4 className="text-sm font-medium mb-2">{scoreName}</h4>
          <div className="grid grid-cols-3 gap-4">
            <StatItem
              label="Count"
              value={statistics.count}
              tooltip="Total number of scores recorded"
            />
            <StatItem
              label="Mean"
              value={statistics.mean}
              tooltip="Average score value"
            />
            <StatItem
              label="Std Dev"
              value={statistics.std_dev}
              tooltip="Standard deviation - measures score variability"
            />
            <StatItem
              label="Min"
              value={statistics.min}
            />
            <StatItem
              label="Max"
              value={statistics.max}
            />
            <StatItem
              label="Median"
              value={statistics.median}
              tooltip="Middle value when scores are sorted"
            />
          </div>
        </div>

        {/* Compare Score Statistics */}
        {compareStatistics && compareScoreName && (
          <div className="border-t pt-4">
            <h4 className="text-sm font-medium mb-2">{compareScoreName}</h4>
            <div className="grid grid-cols-3 gap-4">
              <StatItem label="Count" value={compareStatistics.count} />
              <StatItem label="Mean" value={compareStatistics.mean} />
              <StatItem label="Std Dev" value={compareStatistics.std_dev} />
              <StatItem label="Min" value={compareStatistics.min} />
              <StatItem label="Max" value={compareStatistics.max} />
              <StatItem label="Median" value={compareStatistics.median} />
            </div>
          </div>
        )}

        {/* Comparison Metrics */}
        {comparison && (
          <div className="border-t pt-4 space-y-3">
            <h4 className="text-sm font-medium">Comparison Metrics</h4>
            <div className="text-xs text-muted-foreground">
              {comparison.matched_count.toLocaleString()} matched traces
            </div>
            <div className="space-y-2">
              <MetricWithBadge
                label="Pearson"
                value={comparison.pearson_correlation}
                interpretation={interpretCorrelation(comparison.pearson_correlation)}
                tooltip="Linear correlation coefficient (-1 to 1)"
              />
              <MetricWithBadge
                label="Spearman"
                value={comparison.spearman_correlation}
                interpretation={interpretCorrelation(comparison.spearman_correlation)}
                tooltip="Rank-based correlation coefficient (-1 to 1)"
              />
              <MetricWithBadge
                label="MAE"
                value={comparison.mae}
                interpretation={interpretMAE(comparison.mae, range)}
                tooltip="Mean Absolute Error - average absolute difference"
              />
              <MetricWithBadge
                label="RMSE"
                value={comparison.rmse}
                interpretation={interpretRMSE(comparison.rmse, range)}
                tooltip="Root Mean Square Error - penalizes large errors more"
              />
              {comparison.cohens_kappa !== undefined && (
                <MetricWithBadge
                  label="Cohen's Kappa"
                  value={comparison.cohens_kappa}
                  interpretation={interpretCohensKappa(comparison.cohens_kappa)}
                  tooltip="Inter-rater agreement beyond chance"
                />
              )}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
