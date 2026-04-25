/**
 * Color scale utilities for score analytics visualizations.
 * Uses OKLCH for perceptually uniform interpolation.
 */

export interface OklchColor {
  l: number // Lightness (0-100)
  c: number // Chroma (0-0.4 typical)
  h: number // Hue (0-360)
}

export const SCORE_COLORS = {
  primary: { l: 66.2, c: 0.225, h: 25.9 },
  secondary: { l: 60.4, c: 0.26, h: 302 },
  success: { l: 74.6, c: 0.182, h: 142.5 },
  warning: { l: 79.5, c: 0.178, h: 85 },
  danger: { l: 62.8, c: 0.258, h: 29.2 },
}

export function oklchToCss(color: OklchColor): string {
  return `oklch(${color.l}% ${color.c} ${color.h})`
}

export function interpolateOklch(
  start: OklchColor,
  end: OklchColor,
  t: number,
): OklchColor {
  const clampedT = Math.max(0, Math.min(1, t))
  return {
    l: start.l + (end.l - start.l) * clampedT,
    c: start.c + (end.c - start.c) * clampedT,
    h: start.h + (end.h - start.h) * clampedT,
  }
}

export function getHeatmapCellColor(
  value: number,
  min: number,
  max: number,
  baseColor: OklchColor = SCORE_COLORS.primary,
): string {
  if (max === min) {
    return oklchToCss({ ...baseColor, l: 65 })
  }
  const t = (value - min) / (max - min)
  const color: OklchColor = {
    l: 95 - 60 * t,
    c: baseColor.c * (0.3 + 0.7 * t),
    h: baseColor.h,
  }
  return oklchToCss(color)
}

export function getContrastTextColor(backgroundLightness: number): string {
  return backgroundLightness > 55 ? '#1f2937' : '#f9fafb'
}

export const CHART_COLORS = {
  series1: 'hsl(220, 70%, 50%)',
  series2: 'hsl(280, 60%, 50%)',
  series3: 'hsl(150, 60%, 40%)',
  series4: 'hsl(45, 85%, 45%)',
  series5: 'hsl(0, 70%, 50%)',
}
