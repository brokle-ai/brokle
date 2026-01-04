'use client'

import { useState, useEffect, useCallback } from 'react'
import { RefreshCw, ChevronDown, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

/**
 * Auto-refresh interval options
 */
type RefreshInterval = {
  value: number | null // null = off, number = ms
  label: string
}

const REFRESH_INTERVALS: RefreshInterval[] = [
  { value: null, label: 'Off' },
  { value: 10 * 1000, label: '10 seconds' },
  { value: 30 * 1000, label: '30 seconds' },
  { value: 60 * 1000, label: '1 minute' },
  { value: 5 * 60 * 1000, label: '5 minutes' },
]

interface AutoRefreshControlProps {
  /** Current refresh interval in milliseconds (null = off) */
  interval: number | null
  /** Callback when interval changes */
  onChange: (interval: number | null) => void
  /** Whether a refresh is currently in progress */
  isRefreshing?: boolean
  /** Callback to trigger a manual refresh */
  onRefresh?: () => void
  className?: string
}

export function AutoRefreshControl({
  interval,
  onChange,
  isRefreshing = false,
  onRefresh,
  className,
}: AutoRefreshControlProps) {
  const [open, setOpen] = useState(false)
  const [countdown, setCountdown] = useState<number | null>(null)

  const isActive = interval !== null

  // Get display label for current interval
  const currentOption = REFRESH_INTERVALS.find((opt) => opt.value === interval)
  const displayLabel = currentOption?.label || 'Off'

  // Countdown timer when auto-refresh is active
  useEffect(() => {
    if (!interval) {
      setCountdown(null)
      return
    }

    setCountdown(Math.floor(interval / 1000))

    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev === null || prev <= 1) {
          return Math.floor(interval / 1000)
        }
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(timer)
  }, [interval])

  // Handle interval selection
  const handleSelect = useCallback((newInterval: number | null) => {
    onChange(newInterval)
    setOpen(false)
  }, [onChange])

  // Format countdown for display
  const formatCountdown = (seconds: number): string => {
    if (seconds >= 60) {
      const mins = Math.floor(seconds / 60)
      const secs = seconds % 60
      return `${mins}:${secs.toString().padStart(2, '0')}`
    }
    return `${seconds}s`
  }

  return (
    <div className={cn('flex items-center gap-1', className)}>
      {/* Manual refresh button */}
      {onRefresh && (
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
          onClick={onRefresh}
          disabled={isRefreshing}
          title="Refresh now"
        >
          <RefreshCw
            className={cn('h-4 w-4', isRefreshing && 'animate-spin')}
          />
        </Button>
      )}

      {/* Auto-refresh dropdown */}
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className={cn(
              'h-9 gap-2 text-sm font-normal',
              isActive && 'border-primary/50 text-primary'
            )}
          >
            {isActive && isRefreshing ? (
              <RefreshCw className="h-3.5 w-3.5 animate-spin" />
            ) : isActive ? (
              <span className="flex items-center gap-1.5">
                <span
                  className={cn(
                    'h-2 w-2 rounded-full',
                    isActive ? 'bg-green-500 animate-pulse' : 'bg-muted-foreground'
                  )}
                />
                <span className="text-xs tabular-nums">
                  {countdown !== null ? formatCountdown(countdown) : ''}
                </span>
              </span>
            ) : (
              <RefreshCw className="h-3.5 w-3.5" />
            )}
            <span className="hidden sm:inline">{displayLabel}</span>
            <ChevronDown className="h-3.5 w-3.5 opacity-50" />
          </Button>
        </DropdownMenuTrigger>

        <DropdownMenuContent align="end" className="w-40">
          <div className="px-2 py-1.5 text-xs font-medium text-muted-foreground">
            Auto Refresh
          </div>
          <DropdownMenuSeparator />
          {REFRESH_INTERVALS.map((option) => (
            <DropdownMenuItem
              key={option.label}
              onClick={() => handleSelect(option.value)}
              className="justify-between"
            >
              <span>{option.label}</span>
              {interval === option.value && (
                <Check className="h-4 w-4 text-primary" />
              )}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

/**
 * Hook to manage auto-refresh state
 */
export function useAutoRefresh(initialInterval: number | null = null) {
  const [interval, setInterval] = useState<number | null>(initialInterval)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const triggerRefresh = useCallback(() => {
    setLastRefresh(new Date())
  }, [])

  return {
    interval,
    setInterval,
    lastRefresh,
    triggerRefresh,
  }
}
