'use client'

import { useCallback, useState, useEffect, useTransition } from 'react'
import { Search, X, RotateCcw } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import type { QueueStatus } from '../types'

interface QueuesToolbarProps {
  // Current values
  search: string | null
  status: QueueStatus | null

  // Setters
  onSearchChange: (search: string) => void
  onStatusChange: (status: QueueStatus | null) => void
  onReset: () => void

  // State
  hasActiveFilters: boolean
  isLoading?: boolean
}

/**
 * Toolbar component for the Annotation Queues table
 *
 * Features:
 * - Debounced search input (300ms)
 * - Status filter dropdown
 * - Reset filters button
 * - Loading state indicator
 */
export function QueuesToolbar({
  search,
  status,
  onSearchChange,
  onStatusChange,
  onReset,
  hasActiveFilters,
  isLoading = false,
}: QueuesToolbarProps) {
  const [isPending, startTransition] = useTransition()
  const [localSearch, setLocalSearch] = useState(search ?? '')

  // Sync local search with prop changes (controlled component pattern)
  useEffect(() => {
    setLocalSearch(search ?? '')
  }, [search])

  // Debounced search handler
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== (search ?? '')) {
        startTransition(() => {
          onSearchChange(localSearch)
        })
      }
    }, 300)

    return () => clearTimeout(timer)
  }, [localSearch, search, onSearchChange])

  const handleSearchClear = useCallback(() => {
    setLocalSearch('')
    startTransition(() => {
      onSearchChange('')
    })
  }, [onSearchChange])

  const handleStatusChange = useCallback(
    (value: string) => {
      startTransition(() => {
        onStatusChange(value === 'all' ? null : (value as QueueStatus))
      })
    },
    [onStatusChange]
  )

  const handleReset = useCallback(() => {
    setLocalSearch('')
    startTransition(() => {
      onReset()
    })
  }, [onReset])

  const isTransitioning = isPending || isLoading

  return (
    <div className="flex flex-wrap items-center gap-3">
      {/* Search Input */}
      <div className="relative">
        <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search queues..."
          aria-label="Search queues by name"
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
          className="h-9 w-[200px] pl-8 pr-8"
          disabled={isTransitioning}
        />
        {localSearch && (
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-1 top-1/2 h-6 w-6 -translate-y-1/2"
            onClick={handleSearchClear}
          >
            <X className="h-3.5 w-3.5" />
            <span className="sr-only">Clear search</span>
          </Button>
        )}
      </div>

      {/* Status Filter */}
      <Select
        value={status ?? 'all'}
        onValueChange={handleStatusChange}
        disabled={isTransitioning}
      >
        <SelectTrigger className="h-9 w-[130px]" aria-label="Filter by status">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Statuses</SelectItem>
          <SelectItem value="active">
            <StatusLabel status="active" />
          </SelectItem>
          <SelectItem value="paused">
            <StatusLabel status="paused" />
          </SelectItem>
          <SelectItem value="archived">
            <StatusLabel status="archived" />
          </SelectItem>
        </SelectContent>
      </Select>

      {/* Reset Filters Button */}
      {hasActiveFilters && (
        <Button
          variant="ghost"
          size="sm"
          onClick={handleReset}
          disabled={isTransitioning}
          className="h-9 gap-1.5"
        >
          <RotateCcw className="h-3.5 w-3.5" />
          Reset
        </Button>
      )}

      {/* Active Filters Count Badge */}
      {hasActiveFilters && (
        <Badge variant="secondary" className="text-xs">
          {[search, status].filter(Boolean).length} filter
          {[search, status].filter(Boolean).length !== 1 ? 's' : ''}
        </Badge>
      )}
    </div>
  )
}

/**
 * Status label with colored indicator
 */
function StatusLabel({ status }: { status: QueueStatus }) {
  return (
    <span className="flex items-center gap-2">
      <span className={`inline-block w-2 h-2 rounded-full ${getStatusDotClass(status)}`} />
      {getStatusLabel(status)}
    </span>
  )
}

function getStatusDotClass(status: QueueStatus): string {
  switch (status) {
    case 'active':
      return 'bg-green-500'
    case 'paused':
      return 'bg-yellow-500'
    case 'archived':
      return 'bg-gray-400'
    default:
      return 'bg-gray-500'
  }
}

function getStatusLabel(status: QueueStatus): string {
  switch (status) {
    case 'active':
      return 'Active'
    case 'paused':
      return 'Paused'
    case 'archived':
      return 'Archived'
    default:
      return status
  }
}
