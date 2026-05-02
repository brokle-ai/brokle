/**
 * Statistics interpretation utilities for score analytics.
 * Provides human-readable interpretations of correlation, error,
 * and agreement metrics.
 */

import type { InterpretationResult } from '../api/types'

export function interpretCorrelation(r: number): InterpretationResult {
  const absR = Math.abs(r)
  const direction = r >= 0 ? 'positive' : 'negative'

  if (absR >= 0.9)
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `Very strong ${direction} correlation`,
    }
  if (absR >= 0.7)
    return {
      strength: 'Strong',
      color: 'blue',
      description: `Strong ${direction} correlation`,
    }
  if (absR >= 0.5)
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `Moderate ${direction} correlation`,
    }
  if (absR >= 0.3)
    return {
      strength: 'Weak',
      color: 'orange',
      description: `Weak ${direction} correlation`,
    }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: 'Very weak or no correlation',
  }
}

export function interpretCohensKappa(kappa: number): InterpretationResult {
  if (kappa >= 0.81)
    return {
      strength: 'Very Strong',
      color: 'green',
      description: 'Almost perfect agreement',
    }
  if (kappa >= 0.61)
    return {
      strength: 'Strong',
      color: 'blue',
      description: 'Substantial agreement',
    }
  if (kappa >= 0.41)
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: 'Moderate agreement',
    }
  if (kappa >= 0.21)
    return { strength: 'Weak', color: 'orange', description: 'Fair agreement' }
  if (kappa >= 0)
    return {
      strength: 'Very Weak',
      color: 'red',
      description: 'Slight agreement',
    }
  return {
    strength: 'None',
    color: 'gray',
    description: 'Less than chance agreement',
  }
}

export function interpretMAE(mae: number, range: number): InterpretationResult {
  if (range === 0)
    return {
      strength: 'None',
      color: 'gray',
      description: 'Cannot interpret (zero range)',
    }
  const errorPercent = (mae / range) * 100
  if (errorPercent <= 5)
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `Excellent accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  if (errorPercent <= 10)
    return {
      strength: 'Strong',
      color: 'blue',
      description: `Good accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  if (errorPercent <= 20)
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `Moderate accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  if (errorPercent <= 35)
    return {
      strength: 'Weak',
      color: 'orange',
      description: `Fair accuracy (${errorPercent.toFixed(1)}% error)`,
    }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: `Poor accuracy (${errorPercent.toFixed(1)}% error)`,
  }
}

export function interpretRMSE(
  rmse: number,
  range: number,
): InterpretationResult {
  if (range === 0)
    return {
      strength: 'None',
      color: 'gray',
      description: 'Cannot interpret (zero range)',
    }
  const errorPercent = (rmse / range) * 100
  if (errorPercent <= 7)
    return {
      strength: 'Very Strong',
      color: 'green',
      description: `Excellent precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  if (errorPercent <= 15)
    return {
      strength: 'Strong',
      color: 'blue',
      description: `Good precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  if (errorPercent <= 25)
    return {
      strength: 'Moderate',
      color: 'yellow',
      description: `Moderate precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  if (errorPercent <= 40)
    return {
      strength: 'Weak',
      color: 'orange',
      description: `Fair precision (${errorPercent.toFixed(1)}% RMSE)`,
    }
  return {
    strength: 'Very Weak',
    color: 'red',
    description: `Poor precision (${errorPercent.toFixed(1)}% RMSE)`,
  }
}

export function formatNumber(value: number, precision = 4): string {
  if (Number.isInteger(value)) return value.toLocaleString()
  return value.toFixed(precision)
}

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
