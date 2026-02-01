'use client'

import { useCallback, useState, useEffect, useTransition } from 'react'
import { Search, X, RotateCcw, SlidersHorizontal } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import type { ScoreType, ScoreSource } from '../types'
import { getDataTypeIndicator, getSourceIndicator } from '../lib/score-colors'

interface ScoresToolbarProps {
  // Current values
  search: string | null
  dataType: ScoreType | null
  source: ScoreSource | null

  // Setters
  onSearchChange: (search: string) => void
  onDataTypeChange: (dataType: ScoreType | null) => void
  onSourceChange: (source: ScoreSource | null) => void
  onReset: () => void

  // State
  hasActiveFilters: boolean
  isLoading?: boolean
}

/**
 * Toolbar component for the Scores table
 *
 * Features:
 * - Debounced search input (300ms)
 * - Data type filter dropdown
 * - Source filter dropdown
 * - Reset filters button
 * - Loading state indicator
 */
export function ScoresToolbar({
  search,
  dataType,
  source,
  onSearchChange,
  onDataTypeChange,
  onSourceChange,
  onReset,
  hasActiveFilters,
  isLoading = false,
}: ScoresToolbarProps) {
  const [isPending, startTransition] = useTransition()
  const [localSearch, setLocalSearch] = useState(search ?? '')

  // Sync local search with prop changes
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

  const handleDataTypeChange = useCallback(
    (value: string) => {
      startTransition(() => {
        onDataTypeChange(value === 'all' ? null : (value as ScoreType))
      })
    },
    [onDataTypeChange]
  )

  const handleSourceChange = useCallback(
    (value: string) => {
      startTransition(() => {
        onSourceChange(value === 'all' ? null : (value as ScoreSource))
      })
    },
    [onSourceChange]
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
          aria-label="Search scores by name"
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

      {/* Data Type Filter */}
      <Select
        value={dataType ?? 'all'}
        onValueChange={handleDataTypeChange}
        disabled={isTransitioning}
      >
        <SelectTrigger className="h-9 w-[140px]" aria-label="Filter by data type">
          <SelectValue placeholder="Data Type" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Types</SelectItem>
          <SelectItem value="NUMERIC">
            <span className="flex items-center gap-2">
              <span className="text-muted-foreground">{getDataTypeIndicator('NUMERIC').symbol}</span>
              Numeric
            </span>
          </SelectItem>
          <SelectItem value="BOOLEAN">
            <span className="flex items-center gap-2">
              <span className="text-muted-foreground">{getDataTypeIndicator('BOOLEAN').symbol}</span>
              Boolean
            </span>
          </SelectItem>
          <SelectItem value="CATEGORICAL">
            <span className="flex items-center gap-2">
              <span className="text-muted-foreground">{getDataTypeIndicator('CATEGORICAL').symbol}</span>
              Categorical
            </span>
          </SelectItem>
        </SelectContent>
      </Select>

      {/* Source Filter */}
      <Select
        value={source ?? 'all'}
        onValueChange={handleSourceChange}
        disabled={isTransitioning}
      >
        <SelectTrigger className="h-9 w-[120px]" aria-label="Filter by source">
          <SelectValue placeholder="Source" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Sources</SelectItem>
          <SelectItem value="code">
            <SourceLabel source="code" />
          </SelectItem>
          <SelectItem value="llm">
            <SourceLabel source="llm" />
          </SelectItem>
          <SelectItem value="human">
            <SourceLabel source="human" />
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
          {[search, dataType, source].filter(Boolean).length} filter
          {[search, dataType, source].filter(Boolean).length !== 1 ? 's' : ''}
        </Badge>
      )}
    </div>
  )
}

/**
 * Source label with colored indicator
 */
function SourceLabel({ source }: { source: ScoreSource }) {
  const { label, className } = getSourceIndicator(source)
  return (
    <span className="flex items-center gap-2">
      <span className={`inline-block w-2 h-2 rounded-full ${getSourceDotClass(source)}`} />
      {label}
    </span>
  )
}

function getSourceDotClass(source: ScoreSource): string {
  switch (source) {
    case 'code':
      return 'bg-blue-500'
    case 'llm':
      return 'bg-purple-500'
    case 'human':
      return 'bg-green-500'
    default:
      return 'bg-gray-500'
  }
}
