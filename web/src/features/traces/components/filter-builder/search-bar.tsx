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
  onChange: (value: string, searchType?: string) => void
  searchType?: string | null
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
  searchType = 'all',
  placeholder = 'Search traces...',
  disabled = false,
  className,
  debounceMs = 300,
}: SearchBarProps) {
  const [localValue, setLocalValue] = useState(value)
  const [localSearchType, setLocalSearchType] = useState<SearchType>(
    (searchType as SearchType) || 'all'
  )

  // Sync local value with prop
  useEffect(() => {
    setLocalValue(value)
  }, [value])

  // Sync search type with prop
  useEffect(() => {
    setLocalSearchType((searchType as SearchType) || 'all')
  }, [searchType])

  // Debounced onChange
  useEffect(() => {
    const timer = setTimeout(() => {
      if (localValue !== value) {
        onChange(localValue, localSearchType)
      }
    }, debounceMs)

    return () => clearTimeout(timer)
  }, [localValue, value, onChange, debounceMs, localSearchType])

  const handleSearchTypeChange = useCallback(
    (type: SearchType) => {
      setLocalSearchType(type)
      // Immediately trigger search with new type if there's a value
      if (localValue) {
        onChange(localValue, type)
      }
    },
    [localValue, onChange]
  )

  const handleClear = useCallback(() => {
    setLocalValue('')
    onChange('', localSearchType)
  }, [onChange, localSearchType])

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
              {searchTypeLabels[localSearchType]}
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
            checked={localSearchType === 'id'}
            onCheckedChange={() => handleSearchTypeChange('id')}
          >
            <div>
              <div className="font-medium">{searchTypeLabels.id}</div>
              <div className="text-xs text-muted-foreground">
                {searchTypeDescriptions.id}
              </div>
            </div>
          </DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem
            checked={localSearchType === 'content'}
            onCheckedChange={() => handleSearchTypeChange('content')}
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
            checked={localSearchType === 'all'}
            onCheckedChange={() => handleSearchTypeChange('all')}
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
