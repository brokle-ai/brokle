/**
 * Statistics interpretation utilities for score analytics
 * Provides human-readable interpretations of statistical metrics
 */

import type { InterpretationResult } from '../types'

/**
 * Interpret Pearson or Spearman correlation coefficient
 * Based on widely accepted thresholds for correlation strength
 */
export function interpretCorrelation(r: number): InterpretationResult {
  const absR = Math.abs(r)
  const direction = r >= 0 ? 'positive' : 'negative'

  if (absR >= 0.9) {
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `Very strong ${direction} correlation`,
    }
  }
  if (absR >= 0.7) {
    return {
      strength: 'Strong',
      color: 'blue',
      description: `Strong ${direction} correlation`,
    }
  }
  if (absR >= 0.5) {
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `Moderate ${direction} correlation`,
    }
  }
  if (absR >= 0.3) {
    return {
      strength: 'Weak',
      color: 'orange',
      description: `Weak ${direction} correlation`,
    }
  }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: `Very weak or no correlation`,
  }
}

/**
 * Interpret Cohen's Kappa coefficient for inter-rater agreement
 * Based on Landis & Koch (1977) interpretation guidelines
 */
export function interpretCohensKappa(kappa: number): InterpretationResult {
  if (kappa >= 0.81) {
    return {
      strength: 'Very Strong',
      color: 'green',
      description: 'Almost perfect agreement',
    }
  }
  if (kappa >= 0.61) {
    return {
      strength: 'Strong',
      color: 'blue',
      description: 'Substantial agreement',
    }
  }
  if (kappa >= 0.41) {
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: 'Moderate agreement',
    }
  }
  if (kappa >= 0.21) {
    return {
      strength: 'Weak',
      color: 'orange',
      description: 'Fair agreement',
    }
  }
  if (kappa >= 0) {
    return {
      strength: 'Very Weak',
      color: 'red',
      description: 'Slight agreement',
    }
  }
  return {
    strength: 'None',
    color: 'gray',
    description: 'Less than chance agreement',
  }
}

/**
 * Interpret MAE (Mean Absolute Error) relative to the value range
 * Returns interpretation based on error as percentage of range
 */
export function interpretMAE(mae: number, range: number): InterpretationResult {
  if (range === 0) {
    return {
      strength: 'None',
      color: 'gray',
      description: 'Cannot interpret (zero range)',
    }
  }

  const errorPercent = (mae / range) * 100

  if (errorPercent <= 5) {
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `Excellent accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  }
  if (errorPercent <= 10) {
    return {
      strength: 'Strong',
      color: 'blue',
      description: `Good accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  }
  if (errorPercent <= 20) {
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `Moderate accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  }
  if (errorPercent <= 35) {
    return {
      strength: 'Weak',
      color: 'orange',
      description: `Fair accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: `Poor accuracy (${errorPercent.toFixed(1)}% error)`,
  }
}

/**
 * Interpret RMSE (Root Mean Square Error) relative to the value range
 * RMSE penalizes larger errors more than MAE
 */
export function interpretRMSE(rmse: number, range: number): InterpretationResult {
  if (range === 0) {
    return {
      strength: 'None',
      color: 'gray',
      description: 'Cannot interpret (zero range)',
    }
  }

  const errorPercent = (rmse / range) * 100

  if (errorPercent <= 7) {
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `Excellent precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  }
  if (errorPercent <= 15) {
    return {
      strength: 'Strong',
      color: 'blue',
      description: `Good precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  }
  if (errorPercent <= 25) {
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `Moderate precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  }
  if (errorPercent <= 40) {
    return {
      strength: 'Weak',
      color: 'orange',
      description: `Fair precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: `Poor precision (${errorPercent.toFixed(1)}% RMSE)`,
  }
}

/**
 * Interpret overall agreement percentage
 */
export function interpretAgreement(agreement: number): InterpretationResult {
  const percent = agreement * 100

  if (percent >= 90) {
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `${percent.toFixed(1)}% agreement`,
    }
  }
  if (percent >= 75) {
    return {
      strength: 'Strong',
      color: 'blue',
      description: `${percent.toFixed(1)}% agreement`,
    }
  }
  if (percent >= 60) {
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `${percent.toFixed(1)}% agreement`,
    }
  }
  if (percent >= 45) {
    return {
      strength: 'Weak',
      color: 'orange',
      description: `${percent.toFixed(1)}% agreement`,
    }
  }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: `${percent.toFixed(1)}% agreement`,
  }
}

/**
 * Format a number for display with appropriate precision
 */
export function formatNumber(value: number, precision: number = 4): string {
  if (Number.isInteger(value)) {
    return value.toLocaleString()
  }
  return value.toFixed(precision)
}

/**
 * Format a percentage value
 */
export function formatPercent(value: number, precision: number = 1): string {
  return `${(value * 100).toFixed(precision)}%`
}

/**
 * Get color class based on interpretation color
 */
export function getColorClass(color: InterpretationResult['color']): string {
  const colorMap: Record<InterpretationResult['color'], string> = {
    green: 'text-green-600 bg-green-50 border-green-200',
    blue: 'text-blue-600 bg-blue-50 border-blue-200',
    yellow: 'text-yellow-600 bg-yellow-50 border-yellow-200',
    orange: 'text-orange-600 bg-orange-50 border-orange-200',
    red: 'text-red-600 bg-red-50 border-red-200',
    gray: 'text-gray-600 bg-gray-50 border-gray-200',
  }
  return colorMap[color] ?? colorMap.gray
}

/**
 * Get badge color class based on interpretation color
 */
export function getBadgeColor(color: InterpretationResult['color']): string {
  const colorMap: Record<InterpretationResult['color'], string> = {
    green: 'bg-green-100 text-green-800',
    blue: 'bg-blue-100 text-blue-800',
    yellow: 'bg-yellow-100 text-yellow-800',
    orange: 'bg-orange-100 text-orange-800',
    red: 'bg-red-100 text-red-800',
    gray: 'bg-gray-100 text-gray-800',
  }
  return colorMap[color] ?? colorMap.gray
}
