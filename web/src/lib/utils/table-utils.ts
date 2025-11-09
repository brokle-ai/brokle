import type { ReadonlyURLSearchParams } from 'next/navigation'

/**
 * Build a new URL query string by merging current params with updates
 * Removes params with null/undefined values
 */
export function buildTableUrl(
  currentParams: ReadonlyURLSearchParams | Record<string, string>,
  updates: Record<string, string | null | undefined>
): string {
  // Convert to URLSearchParams
  const params = new URLSearchParams(
    currentParams instanceof URLSearchParams
      ? currentParams.toString()
      : Object.entries(currentParams)
          .filter(([_, value]) => value !== null && value !== undefined)
          .map(([key, value]) => [key, String(value)])
  )

  // Apply updates
  Object.entries(updates).forEach(([key, value]) => {
    if (value === null || value === undefined || value === '') {
      params.delete(key)
    } else {
      params.set(key, value)
    }
  })

  const queryString = params.toString()
  return queryString ? `?${queryString}` : ''
}

/**
 * Debounce function with cancel support
 * Delays execution until after delay milliseconds
 * Returns cancelable function to prevent stale executions
 */
export function debounce<T extends (...args: any[]) => void>(
  func: T,
  delay: number
): ((...args: Parameters<T>) => void) & { cancel: () => void } {
  let timeoutId: NodeJS.Timeout | null = null

  const debounced = function (...args: Parameters<T>) {
    if (timeoutId !== null) {
      clearTimeout(timeoutId)
    }
    timeoutId = setTimeout(() => {
      func(...args)
      timeoutId = null
    }, delay)
  }

  debounced.cancel = () => {
    if (timeoutId !== null) {
      clearTimeout(timeoutId)
      timeoutId = null
    }
  }

  return debounced as typeof debounced & { cancel: () => void }
}
