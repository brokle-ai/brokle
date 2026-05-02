import { useCallback, useEffect, useState } from 'react'
import { RotateCcw, Search, X } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { QueueStatus } from '../api/types'

interface QueuesToolbarProps {
  search: string | null
  status: QueueStatus | null
  onSearchChange: (next: string) => void
  onStatusChange: (next: QueueStatus | null) => void
  onReset: () => void
  isLoading?: boolean
}

/**
 * Search + status filter strip for the queues list view. Search is
 * locally controlled so typing doesn't push a route navigation per
 * keystroke; we debounce to 300ms and forward to the parent via
 * `onSearchChange`. Status changes navigate immediately.
 */
export function QueuesToolbar({
  search,
  status,
  onSearchChange,
  onStatusChange,
  onReset,
  isLoading = false,
}: QueuesToolbarProps) {
  const [local, setLocal] = useState(search ?? '')

  useEffect(() => {
    setLocal(search ?? '')
  }, [search])

  useEffect(() => {
    const t = setTimeout(() => {
      if (local !== (search ?? '')) {
        onSearchChange(local)
      }
    }, 300)
    return () => clearTimeout(t)
  }, [local, search, onSearchChange])

  const clearSearch = useCallback(() => {
    setLocal('')
    onSearchChange('')
  }, [onSearchChange])

  const handleStatusChange = useCallback(
    (value: string) => {
      onStatusChange(value === 'all' ? null : (value as QueueStatus))
    },
    [onStatusChange],
  )

  const handleReset = useCallback(() => {
    setLocal('')
    onReset()
  }, [onReset])

  const hasActiveFilters = (search ?? '').length > 0 || status !== null
  const filterCount = [search, status].filter((v) => v != null && v !== '')
    .length

  return (
    <div className="flex flex-wrap items-center gap-3">
      <div className="relative">
        <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search queues..."
          aria-label="Search queues by name"
          value={local}
          onChange={(e) => setLocal(e.target.value)}
          className="h-9 w-[200px] pl-8 pr-8"
          disabled={isLoading}
        />
        {local && (
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-1 top-1/2 h-6 w-6 -translate-y-1/2"
            onClick={clearSearch}
          >
            <X className="h-3.5 w-3.5" />
            <span className="sr-only">Clear search</span>
          </Button>
        )}
      </div>

      <Select
        value={status ?? 'all'}
        onValueChange={handleStatusChange}
        disabled={isLoading}
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

      {hasActiveFilters && (
        <>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleReset}
            disabled={isLoading}
            className="h-9 gap-1.5"
          >
            <RotateCcw className="h-3.5 w-3.5" />
            Reset
          </Button>
          <Badge variant="secondary" className="text-xs">
            {filterCount} filter{filterCount !== 1 ? 's' : ''}
          </Badge>
        </>
      )}
    </div>
  )
}

function StatusLabel({ status }: { status: QueueStatus }) {
  return (
    <span className="flex items-center gap-2">
      <span className={`inline-block w-2 h-2 rounded-full ${dotClass(status)}`} />
      {label(status)}
    </span>
  )
}

function dotClass(status: QueueStatus): string {
  switch (status) {
    case 'active':
      return 'bg-green-500'
    case 'paused':
      return 'bg-yellow-500'
    case 'archived':
      return 'bg-gray-400'
  }
}

function label(status: QueueStatus): string {
  switch (status) {
    case 'active':
      return 'Active'
    case 'paused':
      return 'Paused'
    case 'archived':
      return 'Archived'
  }
}
