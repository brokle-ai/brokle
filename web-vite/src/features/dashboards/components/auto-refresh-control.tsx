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

type RefreshInterval = {
  value: number | null
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
  interval: number | null
  onChange: (interval: number | null) => void
  isRefreshing?: boolean
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

  const currentOption = REFRESH_INTERVALS.find((opt) => opt.value === interval)
  const displayLabel = currentOption?.label || 'Off'

  useEffect(() => {
    if (!interval) {
      setCountdown(null)
      return
    }

    setCountdown(Math.floor(interval / 1000))

    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev === null || prev <= 1) return Math.floor(interval / 1000)
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(timer)
  }, [interval])

  const handleSelect = useCallback(
    (newInterval: number | null) => {
      onChange(newInterval)
      setOpen(false)
    },
    [onChange],
  )

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
      {onRefresh && (
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
          onClick={onRefresh}
          disabled={isRefreshing}
          title="Refresh now"
        >
          <RefreshCw className={cn('h-4 w-4', isRefreshing && 'animate-spin')} />
        </Button>
      )}

      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className={cn(
              'h-9 gap-2 text-sm font-normal',
              isActive && 'border-primary/50 text-primary',
            )}
          >
            {isActive && isRefreshing ? (
              <RefreshCw className="h-3.5 w-3.5 animate-spin" />
            ) : isActive ? (
              <span className="flex items-center gap-1.5">
                <span
                  className={cn(
                    'h-2 w-2 rounded-full',
                    isActive
                      ? 'bg-green-500 animate-pulse'
                      : 'bg-muted-foreground',
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

export function useAutoRefresh(initialInterval: number | null = null) {
  const [interval, setIntervalState] = useState<number | null>(initialInterval)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())

  const triggerRefresh = useCallback(() => {
    setLastRefresh(new Date())
  }, [])

  return {
    interval,
    setInterval: setIntervalState,
    lastRefresh,
    triggerRefresh,
  }
}
