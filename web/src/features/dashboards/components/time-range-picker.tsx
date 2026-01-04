'use client'

import { useState, useMemo } from 'react'
import { CalendarIcon, Clock, ChevronDown, Check, ChevronLeft } from 'lucide-react'
import { format, subDays, subHours } from 'date-fns'
import type { DateRange } from 'react-day-picker'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import type { TimeRange, RelativeTimeRange } from '../types'

// Get browser timezone abbreviation (e.g., "EST", "PST", "UTC")
function getTimezoneAbbr(): string {
  return (
    new Date()
      .toLocaleTimeString('en-US', { timeZoneName: 'short' })
      .split(' ')
      .pop() || 'UTC'
  )
}

type RelativeOption = {
  value: RelativeTimeRange
  label: string
  duration: number
}

const RELATIVE_OPTIONS: RelativeOption[] = [
  { value: '15m', label: 'Last 15 minutes', duration: 1000 * 60 * 15 },
  { value: '30m', label: 'Last 30 minutes', duration: 1000 * 60 * 30 },
  { value: '1h', label: 'Last 1 hour', duration: 1000 * 60 * 60 },
  { value: '3h', label: 'Last 3 hours', duration: 1000 * 60 * 60 * 3 },
  { value: '6h', label: 'Last 6 hours', duration: 1000 * 60 * 60 * 6 },
  { value: '12h', label: 'Last 12 hours', duration: 1000 * 60 * 60 * 12 },
  { value: '24h', label: 'Last 24 hours', duration: 1000 * 60 * 60 * 24 },
  { value: '7d', label: 'Last 7 days', duration: 1000 * 60 * 60 * 24 * 7 },
  { value: '14d', label: 'Last 14 days', duration: 1000 * 60 * 60 * 24 * 14 },
  { value: '30d', label: 'Last 30 days', duration: 1000 * 60 * 60 * 24 * 30 },
]

interface TimeRangePickerProps {
  value: TimeRange
  onChange: (range: TimeRange) => void
  className?: string
}

export function TimeRangePicker({ value, onChange, className }: TimeRangePickerProps) {
  const [open, setOpen] = useState(false)
  const [showCustom, setShowCustom] = useState(false)
  const [customFrom, setCustomFrom] = useState<Date | undefined>(() => {
    if (value.relative === 'custom' && value.from) {
      return new Date(value.from)
    }
    return undefined
  })
  const [customTo, setCustomTo] = useState<Date | undefined>(() => {
    if (value.relative === 'custom' && value.to) {
      return new Date(value.to)
    }
    return undefined
  })
  const [customFromTime, setCustomFromTime] = useState('00:00')
  const [customToTime, setCustomToTime] = useState('23:59')

  const isCustom = value.relative === 'custom'

  const displayLabel = useMemo(() => {
    if (value.relative && value.relative !== 'custom') {
      const option = RELATIVE_OPTIONS.find((o) => o.value === value.relative)
      return option?.label || value.relative
    }

    if (isCustom && value.from && value.to) {
      const from = new Date(value.from)
      const to = new Date(value.to)
      return `${format(from, 'MMM d, HH:mm')} - ${format(to, 'MMM d, HH:mm')}`
    }

    return 'Select time range'
  }, [value, isCustom])

  const handleRelativeSelect = (option: RelativeOption) => {
    onChange({
      relative: option.value,
      from: undefined,
      to: undefined,
    })
    setOpen(false)
    setShowCustom(false)
  }

  const handleApplyCustom = () => {
    if (!customFrom || !customTo) return

    const fromDate = new Date(customFrom)
    const [fromHours, fromMinutes] = customFromTime.split(':').map(Number)
    fromDate.setHours(fromHours, fromMinutes, 0, 0)

    const toDate = new Date(customTo)
    const [toHours, toMinutes] = customToTime.split(':').map(Number)
    toDate.setHours(toHours, toMinutes, 59, 999)

    onChange({
      relative: 'custom',
      from: fromDate.toISOString(),
      to: toDate.toISOString(),
    })
    setOpen(false)
    setShowCustom(false)
  }

  const handleOpenChange = (isOpen: boolean) => {
    setOpen(isOpen)
    if (!isOpen) {
      setShowCustom(false)
    }
  }

  return (
    <div className={className}>
      <Popover open={open} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className={cn(
              'h-9 gap-2 text-sm font-normal',
              !value.relative && 'text-muted-foreground'
            )}
          >
            <Clock className="h-4 w-4" />
            <span>{displayLabel}</span>
            <ChevronDown className="h-3.5 w-3.5 opacity-50" />
          </Button>
        </PopoverTrigger>

        <PopoverContent className="w-auto p-0" align="end">
          {showCustom ? (
            // Custom date picker view - inline calendar with time inputs
            <div className="p-3 space-y-3">
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2"
                  onClick={() => setShowCustom(false)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <span className="text-sm font-medium">Custom Range</span>
              </div>

              {/* Inline Calendar - no nested popover */}
              <Calendar
                mode="range"
                selected={
                  customFrom && customTo
                    ? { from: customFrom, to: customTo }
                    : customFrom
                      ? { from: customFrom, to: undefined }
                      : undefined
                }
                onSelect={(range: DateRange | undefined) => {
                  setCustomFrom(range?.from)
                  setCustomTo(range?.to)
                }}
                numberOfMonths={1}
                className="rounded-md border"
              />

              {/* Time inputs with timezone */}
              <div className="space-y-3 border-t pt-3">
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <Label className="text-xs text-muted-foreground">Start time</Label>
                    <div className="flex items-center gap-2">
                      <Input
                        type="time"
                        value={customFromTime}
                        onChange={(e) => setCustomFromTime(e.target.value)}
                        className="h-8 text-sm"
                      />
                    </div>
                  </div>
                  <div className="space-y-1.5">
                    <Label className="text-xs text-muted-foreground">End time</Label>
                    <div className="flex items-center gap-2">
                      <Input
                        type="time"
                        value={customToTime}
                        onChange={(e) => setCustomToTime(e.target.value)}
                        className="h-8 text-sm"
                      />
                    </div>
                  </div>
                </div>
                <div className="text-xs text-muted-foreground text-center">
                  Timezone: {getTimezoneAbbr()}
                </div>
              </div>

              <Button
                size="sm"
                className="w-full"
                onClick={handleApplyCustom}
                disabled={!customFrom || !customTo}
              >
                Apply
              </Button>
            </div>
          ) : (
            // Preset options view
            <div className="w-48 p-2">
              <div className="text-xs font-medium text-muted-foreground px-2 py-1.5">
                Time Range
              </div>
              <div className="space-y-0.5">
                {RELATIVE_OPTIONS.map((option) => (
                  <button
                    key={option.value}
                    className={cn(
                      'flex items-center justify-between w-full px-2 py-1.5 text-sm rounded-md hover:bg-muted transition-colors',
                      value.relative === option.value && 'bg-muted'
                    )}
                    onClick={() => handleRelativeSelect(option)}
                  >
                    <span>{option.label}</span>
                    {value.relative === option.value && (
                      <Check className="h-4 w-4 text-primary" />
                    )}
                  </button>
                ))}
              </div>

              <Separator className="my-2" />

              <button
                className={cn(
                  'flex items-center justify-between w-full px-2 py-1.5 text-sm rounded-md hover:bg-muted transition-colors',
                  isCustom && 'bg-muted'
                )}
                onClick={() => {
                  if (!customFrom) setCustomFrom(subDays(new Date(), 1))
                  if (!customTo) setCustomTo(new Date())
                  setShowCustom(true)
                }}
              >
                <span className="flex items-center gap-2">
                  <CalendarIcon className="h-3.5 w-3.5" />
                  Custom range
                </span>
                {isCustom && <Check className="h-4 w-4 text-primary" />}
              </button>
            </div>
          )}
        </PopoverContent>
      </Popover>
    </div>
  )
}

export function getTimeRangeDates(range: TimeRange): { from: Date; to: Date } {
  const now = new Date()

  if (range.relative === 'custom' && range.from && range.to) {
    return {
      from: new Date(range.from),
      to: new Date(range.to),
    }
  }

  const option = RELATIVE_OPTIONS.find((o) => o.value === range.relative)
  if (option) {
    return {
      from: new Date(now.getTime() - option.duration),
      to: now,
    }
  }

  return {
    from: subHours(now, 24),
    to: now,
  }
}
