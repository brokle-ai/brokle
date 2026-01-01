'use client'

import { useState, useEffect, useCallback } from 'react'
import { Search, X, ChevronDown } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

export type SearchType = 'id' | 'content' | 'all'

interface SearchBarProps {
  value: string
  onChange: (value: string) => void
  searchTypes: SearchType[]
  onSearchTypesChange: (types: SearchType[]) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  debounceMs?: number
}

const searchTypeLabels: Record<SearchType, string> = {
  id: 'IDs',
  content: 'Content',
  all: 'All fields',
}

const searchTypeDescriptions: Record<SearchType, string> = {
  id: 'Search trace ID, span ID, span name',
  content: 'Search input/output content',
  all: 'Search all text fields',
}

export function SearchBar({
  value,
  onChange,
  searchTypes,
  onSearchTypesChange,
  placeholder = 'Search traces...',
  disabled = false,
  className,
  debounceMs = 300,
}: SearchBarProps) {
  const [localValue, setLocalValue] = useState(value)

  // Sync local value with prop
  useEffect(() => {
    setLocalValue(value)
  }, [value])

  // Debounced onChange
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localValue !== value) {
        onChange(localValue)
      }
    }, debounceMs)

    return () => clearTimeout(timer)
  }, [localValue, value, onChange, debounceMs])

  // Handle search type toggle
  const handleSearchTypeToggle = useCallback(
    (type: SearchType) => {
      // If 'all' is toggled, replace with 'all' or remove it
      if (type === 'all') {
        if (searchTypes.includes('all')) {
          onSearchTypesChange(['id']) // Default fallback
        } else {
          onSearchTypesChange(['all'])
        }
        return
      }

      // If selecting id or content, remove 'all'
      let newTypes: SearchType[] = searchTypes.filter((t) => t !== 'all')

      if (newTypes.includes(type)) {
        newTypes = newTypes.filter((t) => t !== type)
        // Ensure at least one type is selected
        if (newTypes.length === 0) {
          newTypes = ['id']
        }
      } else {
        newTypes = [...newTypes, type]
      }

      // If both id and content are selected, that's equivalent to 'all'
      if (newTypes.includes('id') && newTypes.includes('content')) {
        newTypes = ['all']
      }

      onSearchTypesChange(newTypes)
    },
    [searchTypes, onSearchTypesChange]
  )

  // Clear search
  const handleClear = useCallback(() => {
    setLocalValue('')
    onChange('')
  }, [onChange])

  // Get display label for current search types
  const searchTypesLabel = searchTypes.includes('all')
    ? 'All'
    : searchTypes.map((t) => searchTypeLabels[t]).join(', ')

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className="relative flex-1">
        <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          placeholder={placeholder}
          className="h-8 pl-8 pr-8"
          disabled={disabled}
        />
        {localValue && (
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-0.5 top-1/2 h-7 w-7 -translate-y-1/2"
            onClick={handleClear}
            disabled={disabled}
          >
            <X className="h-3.5 w-3.5" />
            <span className="sr-only">Clear search</span>
          </Button>
        )}
      </div>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className="h-8 gap-1"
            disabled={disabled}
          >
            <Badge variant="secondary" className="text-xs font-normal">
              {searchTypesLabel}
            </Badge>
            <ChevronDown className="h-3.5 w-3.5" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-[200px]">
          <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">
            Search in
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuCheckboxItem
            checked={searchTypes.includes('id')}
            onCheckedChange={() => handleSearchTypeToggle('id')}
          >
            <div>
              <div className="font-medium">{searchTypeLabels.id}</div>
              <div className="text-xs text-muted-foreground">
                {searchTypeDescriptions.id}
              </div>
            </div>
          </DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem
            checked={searchTypes.includes('content')}
            onCheckedChange={() => handleSearchTypeToggle('content')}
          >
            <div>
              <div className="font-medium">{searchTypeLabels.content}</div>
              <div className="text-xs text-muted-foreground">
                {searchTypeDescriptions.content}
              </div>
            </div>
          </DropdownMenuCheckboxItem>
          <DropdownMenuSeparator />
          <DropdownMenuCheckboxItem
            checked={searchTypes.includes('all')}
            onCheckedChange={() => handleSearchTypeToggle('all')}
          >
            <div>
              <div className="font-medium">{searchTypeLabels.all}</div>
              <div className="text-xs text-muted-foreground">
                {searchTypeDescriptions.all}
              </div>
            </div>
          </DropdownMenuCheckboxItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
