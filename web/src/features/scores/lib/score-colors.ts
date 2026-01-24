/**
 * Deterministic color generation for scores
 * Based on Phoenix pattern: generates consistent colors from score names
 */

/**
 * Hash a string to a number (for deterministic color generation)
 */
function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32-bit integer
  }
  return Math.abs(hash)
}

/**
 * Color palette for score tags
 * Using HSL for easy manipulation and distinct colors
 */
const SCORE_TAG_COLORS = [
  { hue: 220, saturation: 70, name: 'blue' },
  { hue: 280, saturation: 65, name: 'purple' },
  { hue: 150, saturation: 60, name: 'green' },
  { hue: 30, saturation: 80, name: 'orange' },
  { hue: 340, saturation: 70, name: 'pink' },
  { hue: 190, saturation: 70, name: 'cyan' },
  { hue: 45, saturation: 85, name: 'amber' },
  { hue: 260, saturation: 60, name: 'indigo' },
  { hue: 0, saturation: 70, name: 'red' },
  { hue: 170, saturation: 60, name: 'teal' },
]

export interface ScoreTagColor {
  bg: string
  text: string
  border: string
  indicator: string
  hue: number
}

/**
 * Get a deterministic color for a score name
 * Same name always produces the same color
 */
export function getScoreTagColor(scoreName: string): ScoreTagColor {
  const hash = hashString(scoreName)
  const colorIndex = hash % SCORE_TAG_COLORS.length
  const { hue, saturation } = SCORE_TAG_COLORS[colorIndex]

  return {
    // Light mode optimized
    bg: `hsl(${hue}, ${saturation}%, 95%)`,
    text: `hsl(${hue}, ${saturation}%, 30%)`,
    border: `hsl(${hue}, ${saturation}%, 85%)`,
    indicator: `hsl(${hue}, ${saturation}%, 55%)`,
    hue,
  }
}

/**
 * Get Tailwind-compatible classes for a score tag
 * Uses CSS variables for dark mode support
 */
export function getScoreTagClasses(scoreName: string): {
  containerClass: string
  indicatorClass: string
  textClass: string
} {
  const hash = hashString(scoreName)
  const colorIndex = hash % SCORE_TAG_COLORS.length
  const { name } = SCORE_TAG_COLORS[colorIndex]

  // Map to Tailwind color classes
  const colorMap: Record<string, { container: string; indicator: string; text: string }> = {
    blue: {
      container: 'bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800',
      indicator: 'bg-blue-500',
      text: 'text-blue-700 dark:text-blue-300',
    },
    purple: {
      container: 'bg-purple-50 border-purple-200 dark:bg-purple-950/30 dark:border-purple-800',
      indicator: 'bg-purple-500',
      text: 'text-purple-700 dark:text-purple-300',
    },
    green: {
      container: 'bg-green-50 border-green-200 dark:bg-green-950/30 dark:border-green-800',
      indicator: 'bg-green-500',
      text: 'text-green-700 dark:text-green-300',
    },
    orange: {
      container: 'bg-orange-50 border-orange-200 dark:bg-orange-950/30 dark:border-orange-800',
      indicator: 'bg-orange-500',
      text: 'text-orange-700 dark:text-orange-300',
    },
    pink: {
      container: 'bg-pink-50 border-pink-200 dark:bg-pink-950/30 dark:border-pink-800',
      indicator: 'bg-pink-500',
      text: 'text-pink-700 dark:text-pink-300',
    },
    cyan: {
      container: 'bg-cyan-50 border-cyan-200 dark:bg-cyan-950/30 dark:border-cyan-800',
      indicator: 'bg-cyan-500',
      text: 'text-cyan-700 dark:text-cyan-300',
    },
    amber: {
      container: 'bg-amber-50 border-amber-200 dark:bg-amber-950/30 dark:border-amber-800',
      indicator: 'bg-amber-500',
      text: 'text-amber-700 dark:text-amber-300',
    },
    indigo: {
      container: 'bg-indigo-50 border-indigo-200 dark:bg-indigo-950/30 dark:border-indigo-800',
      indicator: 'bg-indigo-500',
      text: 'text-indigo-700 dark:text-indigo-300',
    },
    red: {
      container: 'bg-red-50 border-red-200 dark:bg-red-950/30 dark:border-red-800',
      indicator: 'bg-red-500',
      text: 'text-red-700 dark:text-red-300',
    },
    teal: {
      container: 'bg-teal-50 border-teal-200 dark:bg-teal-950/30 dark:border-teal-800',
      indicator: 'bg-teal-500',
      text: 'text-teal-700 dark:text-teal-300',
    },
  }

  const colors = colorMap[name] || colorMap.blue

  return {
    containerClass: colors.container,
    indicatorClass: colors.indicator,
    textClass: colors.text,
  }
}

/**
 * Get data type indicator symbol (Langfuse pattern)
 */
export function getDataTypeIndicator(dataType: 'NUMERIC' | 'CATEGORICAL' | 'BOOLEAN'): {
  symbol: string
  label: string
} {
  switch (dataType) {
    case 'NUMERIC':
      return { symbol: '#', label: 'Numeric' }
    case 'CATEGORICAL':
      return { symbol: 'Ⓒ', label: 'Categorical' }
    case 'BOOLEAN':
      return { symbol: '◉', label: 'Boolean' }
    default:
      return { symbol: '?', label: 'Unknown' }
  }
}

/**
 * Get source indicator styling
 */
export function getSourceIndicator(source: 'code' | 'llm' | 'human'): {
  label: string
  className: string
} {
  switch (source) {
    case 'code':
      return {
        label: 'SDK',
        className: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
      }
    case 'llm':
      return {
        label: 'LLM',
        className: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
      }
    case 'human':
      return {
        label: 'Human',
        className: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
      }
    default:
      return {
        label: source,
        className: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-300',
      }
  }
}
