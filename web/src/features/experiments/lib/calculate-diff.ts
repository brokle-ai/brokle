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

// ============================================================================
// Percentage Calculations (Phoenix pattern)
// ============================================================================

export interface DiffPercentageResult {
  percentage: number
  direction: '+' | '-'
  isSignificant: boolean // > 1% change
  isUndefined: boolean // true when baseline=0 but current≠0
}

/**
 * Calculate percentage change between baseline and current value
 * Returns null only when both baseline and current are 0 (truly no change)
 * Returns isUndefined=true when baseline is 0 but current is not
 */
export function calculateDiffPercentage(
  baseline: number,
  current: number
): DiffPercentageResult | null {
  // Both zero → truly no change
  if (baseline === 0 && current === 0) {
    return null
  }

  // Baseline is zero but current is not → undefined percentage
  if (baseline === 0) {
    return {
      percentage: Infinity,
      direction: current > 0 ? '+' : '-',
      isSignificant: true,
      isUndefined: true,
    }
  }

  // Normal calculation
  const percentChange = ((current - baseline) / Math.abs(baseline)) * 100
  const direction: '+' | '-' = percentChange >= 0 ? '+' : '-'

  return {
    percentage: Math.abs(percentChange),
    direction,
    isSignificant: Math.abs(percentChange) >= 1,
    isUndefined: false,
  }
}

/**
 * Calculate percentile position within min/max range
 * Used for progress bar visualization (Phoenix pattern)
 */
export function calculateScorePercentile(
  value: number,
  min: number,
  max: number
): number {
  if (max === min) return 50 // Avoid division by zero
  const percentile = ((value - min) / (max - min)) * 100
  return Math.max(0, Math.min(100, percentile))
}

// ============================================================================
// Diff Classification (Phoenix pattern)
// ============================================================================

export interface DiffClassification {
  improved: number
  regressed: number
  unchanged: number
}

/**
 * Classify all score diffs for summary display
 * Returns counts of improved, regressed, and unchanged scores
 */
export function classifyDiffs(
  scoreRows: import('../types').ScoreComparisonRow[],
  baselineId: string,
  config: DiffDisplayConfig = {}
): DiffClassification {
  const result: DiffClassification = {
    improved: 0,
    regressed: 0,
    unchanged: 0,
  }

  const { preferNegativeDiff = false } = config

  for (const row of scoreRows) {
    // Check each non-baseline experiment for this score
    for (const [expId, data] of Object.entries(row.experiments)) {
      if (expId === baselineId || !data.diff) continue

      if (data.diff.type === 'CATEGORICAL') {
        // Categorical comparisons: both matches and differences count as unchanged
        // (since we don't classify categorical as improved/regressed)
        result.unchanged++
        continue
      }

      const { difference = 0, direction } = data.diff

      if (difference === 0) {
        result.unchanged++
        continue
      }

      const isPositiveDirection = direction === '+'
      const isImprovement = preferNegativeDiff
        ? !isPositiveDirection
        : isPositiveDirection

      if (isImprovement) {
        result.improved++
      } else {
        result.regressed++
      }
    }
  }

  return result
}
