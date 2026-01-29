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
import type { ScorerType, EvaluatorStatus } from '../types'

interface EvaluatorsToolbarProps {
  // Current values
  search: string | null
  scorerType: ScorerType | null
  status: EvaluatorStatus | null

  // Setters
  onSearchChange: (search: string) => void
  onScorerTypeChange: (scorerType: ScorerType | null) => void
  onStatusChange: (status: EvaluatorStatus | null) => void
  onReset: () => void

  // State
  hasActiveFilters: boolean
  isLoading?: boolean
}

/**
 * Toolbar component for the Evaluators table
 *
 * Features:
 * - Debounced search input (300ms)
 * - Scorer type filter dropdown
 * - Status filter dropdown
 * - Reset filters button
 * - Loading state indicator
 */
export function EvaluatorsToolbar({
  search,
  scorerType,
  status,
  onSearchChange,
  onScorerTypeChange,
  onStatusChange,
  onReset,
  hasActiveFilters,
  isLoading = false,
}: EvaluatorsToolbarProps) {
  const [isPending, startTransition] = useTransition()
  const [localSearch, setLocalSearch] = useState(search ?? '')

  // Sync local search with prop changes (controlled component pattern)
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    setLocalSearch(search ?? '')
  }, [search])
  /* eslint-enable react-hooks/set-state-in-effect */

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

  const handleScorerTypeChange = useCallback(
    (value: string) => {
      startTransition(() => {
        onScorerTypeChange(value === 'all' ? null : (value as ScorerType))
      })
    },
    [onScorerTypeChange]
  )

  const handleStatusChange = useCallback(
    (value: string) => {
      startTransition(() => {
        onStatusChange(value === 'all' ? null : (value as EvaluatorStatus))
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
          placeholder="Search by name..."
          aria-label="Search evaluators by name"
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

      {/* Scorer Type Filter */}
      <Select
        value={scorerType ?? 'all'}
        onValueChange={handleScorerTypeChange}
        disabled={isTransitioning}
      >
        <SelectTrigger className="h-9 w-[140px]" aria-label="Filter by scorer type">
          <SelectValue placeholder="Scorer Type" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Types</SelectItem>
          <SelectItem value="llm">
            <span className="flex items-center gap-2">
              <span className="text-muted-foreground">🤖</span>
              LLM
            </span>
          </SelectItem>
          <SelectItem value="builtin">
            <span className="flex items-center gap-2">
              <span className="text-muted-foreground">📊</span>
              Builtin
            </span>
          </SelectItem>
          <SelectItem value="regex">
            <span className="flex items-center gap-2">
              <span className="text-muted-foreground">🔤</span>
              Regex
            </span>
          </SelectItem>
        </SelectContent>
      </Select>

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
          <SelectItem value="inactive">
            <StatusLabel status="inactive" />
          </SelectItem>
          <SelectItem value="paused">
            <StatusLabel status="paused" />
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
          {[search, scorerType, status].filter(Boolean).length} filter
          {[search, scorerType, status].filter(Boolean).length !== 1 ? 's' : ''}
        </Badge>
      )}
    </div>
  )
}

/**
 * Status label with colored indicator
 */
function StatusLabel({ status }: { status: EvaluatorStatus }) {
  return (
    <span className="flex items-center gap-2">
      <span className={`inline-block w-2 h-2 rounded-full ${getStatusDotClass(status)}`} />
      {getStatusLabel(status)}
    </span>
  )
}

function getStatusDotClass(status: EvaluatorStatus): string {
  switch (status) {
    case 'active':
      return 'bg-green-500'
    case 'inactive':
      return 'bg-gray-400'
    case 'paused':
      return 'bg-yellow-500'
    default:
      return 'bg-gray-500'
  }
}

function getStatusLabel(status: EvaluatorStatus): string {
  switch (status) {
    case 'active':
      return 'Active'
    case 'inactive':
      return 'Inactive'
    case 'paused':
      return 'Paused'
    default:
      return status
  }
}
