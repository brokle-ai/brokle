import type { TimeRange } from './types'
import { RELATIVE_OPTIONS } from './types'

export function getTimezoneAbbr(): string {
  return (
    new Date()
      .toLocaleTimeString('en-US', { timeZoneName: 'short' })
      .split(' ')
      .pop() || 'UTC'
  )
}

export function formatTimeRangeLabel(range: TimeRange): string {
  if (range.relative && range.relative !== 'custom') {
    const option = RELATIVE_OPTIONS.find((o) => o.value === range.relative)
    return option?.label || range.relative
  }
  if (range.relative === 'custom' && range.from && range.to) {
    const from = new Date(range.from)
    const to = new Date(range.to)
    const opts: Intl.DateTimeFormatOptions = {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }
    return `${from.toLocaleDateString('en-US', opts)} - ${to.toLocaleDateString('en-US', opts)}`
  }
  return 'Select time range'
}
