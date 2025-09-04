'use client'

import * as React from 'react'
import { Search, X, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

interface SearchInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  value?: string
  onValueChange?: (value: string) => void
  onSearch?: (value: string) => void
  loading?: boolean
  debounceMs?: number
  showClearButton?: boolean
  searchIcon?: boolean
}

export function SearchInput({
  value = '',
  onValueChange,
  onSearch,
  loading = false,
  debounceMs = 300,
  showClearButton = true,
  searchIcon = true,
  placeholder = 'Search...',
  className,
  ...props
}: SearchInputProps) {
  const [internalValue, setInternalValue] = React.useState(value)
  const debounceRef = React.useRef<NodeJS.Timeout>()

  // Sync external value changes
  React.useEffect(() => {
    setInternalValue(value)
  }, [value])

  // Debounced search
  React.useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    debounceRef.current = setTimeout(() => {
      if (onSearch && internalValue !== value) {
        onSearch(internalValue)
      }
    }, debounceMs)

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [internalValue, debounceMs, onSearch, value])

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = event.target.value
    setInternalValue(newValue)
    onValueChange?.(newValue)
  }

  const handleClear = () => {
    setInternalValue('')
    onValueChange?.('')
    onSearch?.('')
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      event.preventDefault()
      onSearch?.(internalValue)
    }
    props.onKeyDown?.(event)
  }

  return (
    <div className='relative'>
      {searchIcon && (
        <Search className='absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
      )}
      
      <Input
        {...props}
        value={internalValue}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        className={cn(
          searchIcon && 'pl-9',
          (showClearButton && internalValue) || loading ? 'pr-9' : '',
          className
        )}
      />

      <div className='absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1'>
        {loading && (
          <Loader2 className='h-4 w-4 animate-spin text-muted-foreground' />
        )}
        
        {!loading && showClearButton && internalValue && (
          <Button
            type='button'
            variant='ghost'
            size='sm'
            onClick={handleClear}
            className='h-6 w-6 p-0 text-muted-foreground hover:text-foreground'
          >
            <X className='h-3 w-3' />
          </Button>
        )}
      </div>
    </div>
  )
}