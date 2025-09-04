'use client'

import * as React from 'react'
import { X, Check, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { FormField } from './form-field'

interface Option {
  value: string
  label: string
  description?: string
  disabled?: boolean
}

interface MultiSelectProps {
  options: Option[]
  value?: string[]
  onValueChange?: (value: string[]) => void
  placeholder?: string
  searchPlaceholder?: string
  emptyText?: string
  maxDisplayed?: number
  label?: string
  description?: string
  error?: string
  required?: boolean
  disabled?: boolean
  className?: string
}

export function MultiSelect({
  options,
  value = [],
  onValueChange,
  placeholder = 'Select options...',
  searchPlaceholder = 'Search options...',
  emptyText = 'No options found.',
  maxDisplayed = 3,
  label,
  description,
  error,
  required,
  disabled,
  className,
}: MultiSelectProps) {
  const [open, setOpen] = React.useState(false)
  const [searchValue, setSearchValue] = React.useState('')

  const selectedOptions = options.filter((option) => value.includes(option.value))

  const filteredOptions = React.useMemo(() => {
    if (!searchValue) return options
    
    return options.filter((option) =>
      option.label.toLowerCase().includes(searchValue.toLowerCase()) ||
      option.description?.toLowerCase().includes(searchValue.toLowerCase())
    )
  }, [options, searchValue])

  const handleSelect = (selectedValue: string) => {
    const newValue = value.includes(selectedValue)
      ? value.filter((v) => v !== selectedValue)
      : [...value, selectedValue]
    
    onValueChange?.(newValue)
  }

  const handleRemove = (valueToRemove: string, event: React.MouseEvent) => {
    event.stopPropagation()
    const newValue = value.filter((v) => v !== valueToRemove)
    onValueChange?.(newValue)
  }

  const handleClear = (event: React.MouseEvent) => {
    event.stopPropagation()
    onValueChange?.([])
  }

  const displayedOptions = selectedOptions.slice(0, maxDisplayed)
  const remainingCount = selectedOptions.length - maxDisplayed

  const content = (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant='outline'
          role='combobox'
          aria-expanded={open}
          className={cn(
            'w-full justify-between min-h-10 h-auto p-1',
            error && 'border-destructive',
            className
          )}
          disabled={disabled}
        >
          <div className='flex flex-wrap items-center gap-1 flex-1'>
            {selectedOptions.length === 0 ? (
              <span className='text-muted-foreground px-2 py-1'>{placeholder}</span>
            ) : (
              <>
                {displayedOptions.map((option) => (
                  <Badge
                    key={option.value}
                    variant='secondary'
                    className='text-xs h-6'
                  >
                    {option.label}
                    <button
                      onClick={(e) => handleRemove(option.value, e)}
                      className='ml-1 ring-offset-background rounded-full outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2'
                    >
                      <X className='h-3 w-3' />
                    </button>
                  </Badge>
                ))}
                {remainingCount > 0 && (
                  <Badge variant='secondary' className='text-xs h-6'>
                    +{remainingCount} more
                  </Badge>
                )}
              </>
            )}
          </div>
          <div className='flex items-center gap-1 px-2'>
            {selectedOptions.length > 0 && (
              <button
                onClick={handleClear}
                className='text-muted-foreground hover:text-foreground'
              >
                <X className='h-4 w-4' />
              </button>
            )}
            <ChevronDown className='h-4 w-4 opacity-50' />
          </div>
        </Button>
      </PopoverTrigger>
      <PopoverContent className='w-full p-0' align='start'>
        <Command shouldFilter={false}>
          <CommandInput
            placeholder={searchPlaceholder}
            value={searchValue}
            onValueChange={setSearchValue}
          />
          <CommandEmpty>{emptyText}</CommandEmpty>
          <CommandGroup className='max-h-64 overflow-auto'>
            {filteredOptions.map((option) => (
              <CommandItem
                key={option.value}
                value={option.value}
                disabled={option.disabled}
                onSelect={handleSelect}
                className='flex items-start gap-2'
              >
                <div className={cn(
                  'border border-primary mt-1 flex h-4 w-4 items-center justify-center rounded-sm',
                  value.includes(option.value)
                    ? 'bg-primary text-primary-foreground'
                    : 'opacity-50 [&_svg]:invisible'
                )}>
                  <Check className='h-3 w-3' />
                </div>
                <div className='flex-1 space-y-1'>
                  <div className='text-sm font-medium leading-none'>
                    {option.label}
                  </div>
                  {option.description && (
                    <div className='text-xs text-muted-foreground'>
                      {option.description}
                    </div>
                  )}
                </div>
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  )

  if (label || description || error) {
    return (
      <FormField
        label={label}
        description={description}
        error={error}
        required={required}
      >
        {content}
      </FormField>
    )
  }

  return content
}