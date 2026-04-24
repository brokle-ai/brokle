import { useCallback, useMemo } from 'react'
import { useQueryStates, parseAsString } from 'nuqs'
import type { TimeRange, RelativeTimeRange } from '../types'

const VALID_RELATIVE_OPTIONS: RelativeTimeRange[] = [
  '15m',
  '30m',
  '1h',
  '3h',
  '6h',
  '12h',
  '24h',
  '7d',
  '14d',
  '30d',
  'custom',
]

const DEFAULT_TIME_RANGE: TimeRange = {
  relative: '24h',
}

function parseRelativeTimeRange(
  value: string | null,
): RelativeTimeRange | undefined {
  if (!value) return undefined
  return VALID_RELATIVE_OPTIONS.includes(value as RelativeTimeRange)
    ? (value as RelativeTimeRange)
    : undefined
}

export interface UseDashboardTimeRangeReturn {
  timeRange: TimeRange
  setTimeRange: (range: TimeRange) => void
  resetTimeRange: () => void
  isCustomized: boolean
}

export function useDashboardTimeRange(
  dashboardDefault?: TimeRange,
): UseDashboardTimeRangeReturn {
  const [urlParams, setUrlParams] = useQueryStates({
    time_rel: parseAsString,
    time_from: parseAsString,
    time_to: parseAsString,
  })

  const timeRange = useMemo((): TimeRange => {
    const relative = parseRelativeTimeRange(urlParams.time_rel)
    const from = urlParams.time_from ?? undefined
    const to = urlParams.time_to ?? undefined

    if (relative) {
      if (relative === 'custom' && from && to) {
        return { relative: 'custom', from, to }
      }
      return { relative }
    }
    return dashboardDefault ?? DEFAULT_TIME_RANGE
  }, [urlParams, dashboardDefault])

  const setTimeRange = useCallback(
    (range: TimeRange) => {
      if (range.relative === 'custom') {
        setUrlParams({
          time_rel: 'custom',
          time_from: range.from ?? null,
          time_to: range.to ?? null,
        })
      } else {
        setUrlParams({
          time_rel: range.relative ?? null,
          time_from: null,
          time_to: null,
        })
      }
    },
    [setUrlParams],
  )

  const resetTimeRange = useCallback(() => {
    setUrlParams({ time_rel: null, time_from: null, time_to: null })
  }, [setUrlParams])

  const isCustomized = useMemo(() => {
    return (
      urlParams.time_rel !== null ||
      urlParams.time_from !== null ||
      urlParams.time_to !== null
    )
  }, [urlParams])

  return { timeRange, setTimeRange, resetTimeRange, isCustomized }
}
