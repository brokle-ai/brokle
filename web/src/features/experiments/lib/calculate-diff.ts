import type { ExperimentScoreDiff } from '../types'

export interface DiffDisplayConfig {
  preferNegativeDiff?: boolean // For metrics where lower is better (latency, toxicity)
}

export type DiffStyle = 'positive' | 'negative' | 'neutral'

export interface DiffDisplayResult {
  style: DiffStyle
  label: string
}

/**
 * Get display style for a diff value
 * - Green (+): improvement (higher value, or lower if preferNegativeDiff)
 * - Red (-): regression
 * - Neutral: baseline or no diff
 */
export function getDiffDisplay(
  diff: ExperimentScoreDiff | undefined,
  config: DiffDisplayConfig = {}
): DiffDisplayResult | null {
  if (!diff) return null

  if (diff.type === 'CATEGORICAL') {
    return diff.isDifferent ? { style: 'neutral', label: 'Varies' } : null
  }

  const { difference = 0, direction } = diff
  const { preferNegativeDiff = false } = config

  if (difference === 0) return null

  const isPositiveDirection = direction === '+'
  const isImprovement = preferNegativeDiff
    ? !isPositiveDirection // Lower is better
    : isPositiveDirection // Higher is better

  const sign = isPositiveDirection ? '+' : '-'
  const formattedValue =
    difference < 0.01 ? difference.toExponential(2) : difference.toFixed(2)

  return {
    style: isImprovement ? 'positive' : 'negative',
    label: `${sign}${formattedValue}`,
  }
}

/**
 * Format score stats for display (mean ± std_dev)
 */
export function formatScoreStats(
  mean: number,
  stdDev: number,
  precision: number = 2
): string {
  return `${mean.toFixed(precision)} ± ${stdDev.toFixed(precision)}`
}
