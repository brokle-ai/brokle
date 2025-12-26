/**
 * Color scale utilities for score analytics visualizations
 * Uses OKLCH color space for perceptually uniform gradients
 */

export interface OklchColor {
  l: number // Lightness (0-100)
  c: number // Chroma (0-0.4 typical)
  h: number // Hue (0-360)
}

/**
 * Base colors for score visualizations
 * Using OKLCH for perceptually uniform interpolation
 */
export const SCORE_COLORS = {
  primary: { l: 66.2, c: 0.225, h: 25.9 }, // Blue
  secondary: { l: 60.4, c: 0.26, h: 302 }, // Purple
  success: { l: 74.6, c: 0.182, h: 142.5 }, // Green
  warning: { l: 79.5, c: 0.178, h: 85 }, // Yellow
  danger: { l: 62.8, c: 0.258, h: 29.2 }, // Red
}

/**
 * Convert OKLCH color to CSS string
 */
export function oklchToCss(color: OklchColor): string {
  return `oklch(${color.l}% ${color.c} ${color.h})`
}

/**
 * Interpolate between two OKLCH colors
 */
export function interpolateOklch(
  start: OklchColor,
  end: OklchColor,
  t: number
): OklchColor {
  // Clamp t to [0, 1]
  const clampedT = Math.max(0, Math.min(1, t))

  return {
    l: start.l + (end.l - start.l) * clampedT,
    c: start.c + (end.c - start.c) * clampedT,
    h: start.h + (end.h - start.h) * clampedT,
  }
}

/**
 * Generate a monochromatic color scale from a base color
 * Varies lightness while keeping chroma and hue
 */
export function generateMonoColorScale(
  baseColor: OklchColor,
  steps: number = 10
): string[] {
  const colors: string[] = []
  const lightStart = 95 // Very light
  const lightEnd = 35 // Darker

  for (let i = 0; i < steps; i++) {
    const t = i / (steps - 1)
    const color: OklchColor = {
      l: lightStart + (lightEnd - lightStart) * t,
      c: baseColor.c * (0.3 + 0.7 * t), // Increase chroma with density
      h: baseColor.h,
    }
    colors.push(oklchToCss(color))
  }

  return colors
}

/**
 * Get heatmap cell color based on value
 * Uses blue color scale by default
 */
export function getHeatmapCellColor(
  value: number,
  min: number,
  max: number,
  baseColor: OklchColor = SCORE_COLORS.primary
): string {
  if (max === min) {
    // Single value, use mid-tone
    return oklchToCss({ ...baseColor, l: 65 })
  }

  const t = (value - min) / (max - min)

  // Light background for low values, darker for high values
  const color: OklchColor = {
    l: 95 - 60 * t, // 95% -> 35%
    c: baseColor.c * (0.3 + 0.7 * t), // Low chroma when light
    h: baseColor.h,
  }

  return oklchToCss(color)
}

/**
 * Get diverging color scale for correlation values
 * Red for negative, blue for positive
 */
export function getCorrelationColor(value: number): string {
  const absValue = Math.abs(value)
  const t = absValue // 0 to 1

  if (value >= 0) {
    // Positive: white to blue
    return oklchToCss({
      l: 95 - 55 * t,
      c: 0.225 * t,
      h: 25.9,
    })
  } else {
    // Negative: white to red
    return oklchToCss({
      l: 95 - 50 * t,
      c: 0.258 * t,
      h: 29.2,
    })
  }
}

/**
 * Chart color palette for multiple data series
 */
export const CHART_COLORS = {
  series1: 'hsl(220, 70%, 50%)', // Blue
  series2: 'hsl(280, 60%, 50%)', // Purple
  series3: 'hsl(150, 60%, 40%)', // Green
  series4: 'hsl(45, 85%, 45%)', // Yellow
  series5: 'hsl(0, 70%, 50%)', // Red
}

/**
 * Get chart color for a data series by index
 */
export function getChartColor(index: number): string {
  const colors = Object.values(CHART_COLORS)
  return colors[index % colors.length]
}

/**
 * Color for score value based on normalized value (0-1)
 */
export function getScoreValueColor(normalizedValue: number): string {
  // Red (0) -> Yellow (0.5) -> Green (1)
  if (normalizedValue <= 0.5) {
    const t = normalizedValue * 2 // 0-1 for first half
    return oklchToCss(interpolateOklch(SCORE_COLORS.danger, SCORE_COLORS.warning, t))
  } else {
    const t = (normalizedValue - 0.5) * 2 // 0-1 for second half
    return oklchToCss(interpolateOklch(SCORE_COLORS.warning, SCORE_COLORS.success, t))
  }
}

/**
 * Get text color that contrasts with a background lightness
 */
export function getContrastTextColor(backgroundLightness: number): string {
  return backgroundLightness > 55 ? '#1f2937' : '#f9fafb'
}

/**
 * Tailwind-compatible color classes for score visualization
 */
export const SCORE_COLOR_CLASSES = {
  high: 'bg-green-100 text-green-800 border-green-200',
  medium: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  low: 'bg-red-100 text-red-800 border-red-200',
  neutral: 'bg-gray-100 text-gray-800 border-gray-200',
}

/**
 * Get Tailwind color class based on score value
 */
export function getScoreColorClass(
  value: number,
  min: number = 0,
  max: number = 1
): string {
  const normalized = (value - min) / (max - min || 1)

  if (normalized >= 0.7) return SCORE_COLOR_CLASSES.high
  if (normalized >= 0.4) return SCORE_COLOR_CLASSES.medium
  if (normalized > 0) return SCORE_COLOR_CLASSES.low
  return SCORE_COLOR_CLASSES.neutral
}
