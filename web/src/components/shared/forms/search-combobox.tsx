'use client'

import * as React from 'react'
import { Check, ChevronsUpDown, Search } from 'lucide-react'
import { cn } from '@/lib/utils'
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

interface SearchComboboxProps {
  options: Option[]
  value?: string
  onValueChange?: (value: string) => void
  placeholder?: string
  searchPlaceholder?: string
  emptyText?: string
  label?: string
  description?: string
  error?: string
  required?: boolean
  disabled?: boolean
  loading?: boolean
  className?: string
}

export function SearchCombobox({
  options,
  value,
  onValueChange,
  placeholder = 'Select option...',
  searchPlaceholder = 'Search options...',
  emptyText = 'No option found.',
  label,
  description,
  error,
  required,
  disabled,
  loading,
  className,
}: SearchComboboxProps) {
  const [open, setOpen] = React.useState(false)
  const [searchValue, setSearchValue] = React.useState('')

  const selectedOption = options.find((option) => option.value === value)

  const filteredOptions = React.useMemo(() => {
    if (!searchValue) return options
    
    return options.filter((option) =>
      option.label.toLowerCase().includes(searchValue.toLowerCase()) ||
      option.description?.toLowerCase().includes(searchValue.toLowerCase())
    )
  }, [options, searchValue])

  const handleSelect = (selectedValue: string) => {
    if (selectedValue === value) {
      onValueChange?.('')
    } else {
      onValueChange?.(selectedValue)
    }
    setOpen(false)
    setSearchValue('')
  }

  const content = (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant='outline'
          role='combobox'
          aria-expanded={open}
          className={cn(
            'w-full justify-between',
            !selectedOption && 'text-muted-foreground',
            error && 'border-destructive',
            className
          )}
          disabled={disabled || loading}
        >
          <span className='truncate'>
            {loading ? 'Loading...' : selectedOption?.label ?? placeholder}
          </span>
          <ChevronsUpDown className='ml-2 h-4 w-4 shrink-0 opacity-50' />
        </Button>
      </PopoverTrigger>
      <PopoverContent className='w-full p-0' align='start'>
        <Command shouldFilter={false}>
          <div className='flex items-center border-b px-3'>
            <Search className='mr-2 h-4 w-4 shrink-0 opacity-50' />
            <CommandInput
              placeholder={searchPlaceholder}
              value={searchValue}
              onValueChange={setSearchValue}
              className='flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50'
            />
          </div>
          <CommandEmpty className='py-6 text-center text-sm'>
            {emptyText}
          </CommandEmpty>
          <CommandGroup className='max-h-64 overflow-auto'>
            {filteredOptions.map((option) => (
              <CommandItem
                key={option.value}
                value={option.value}
                disabled={option.disabled}
                onSelect={handleSelect}
                className='flex items-start gap-2 p-3'
              >
                <Check
                  className={cn(
                    'mt-0.5 h-4 w-4',
                    value === option.value ? 'opacity-100' : 'opacity-0'
                  )}
                />
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