'use client'

import * as React from 'react'
import { CalendarIcon, Clock } from 'lucide-react'
import { format } from 'date-fns'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Input } from '@/components/ui/input'
import { FormField } from './form-field'

interface DateTimePickerProps {
  value?: Date
  onValueChange?: (date: Date | undefined) => void
  placeholder?: string
  showTime?: boolean
  dateFormat?: string
  label?: string
  description?: string
  error?: string
  required?: boolean
  disabled?: boolean
  className?: string
}

export function DateTimePicker({
  value,
  onValueChange,
  placeholder = 'Pick a date',
  showTime = false,
  dateFormat = 'PPP',
  label,
  description,
  error,
  required,
  disabled,
  className,
}: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [timeValue, setTimeValue] = React.useState(() => {
    if (value && showTime) {
      return format(value, 'HH:mm')
    }
    return '12:00'
  })

  const handleDateSelect = (selectedDate: Date | undefined) => {
    if (!selectedDate) {
      onValueChange?.(undefined)
      return
    }

    if (showTime && value) {
      // Preserve the time when changing date
      const newDate = new Date(selectedDate)
      newDate.setHours(value.getHours())
      newDate.setMinutes(value.getMinutes())
      onValueChange?.(newDate)
    } else {
      onValueChange?.(selectedDate)
    }

    if (!showTime) {
      setOpen(false)
    }
  }

  const handleTimeChange = (time: string) => {
    setTimeValue(time)
    
    if (value) {
      const [hours, minutes] = time.split(':').map(Number)
      const newDate = new Date(value)
      newDate.setHours(hours, minutes)
      onValueChange?.(newDate)
    }
  }

  const formatDisplayValue = () => {
    if (!value) return placeholder
    
    if (showTime) {
      return `${format(value, dateFormat)} ${format(value, 'HH:mm')}`
    }
    
    return format(value, dateFormat)
  }

  const content = (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant='outline'
          className={cn(
            'w-full justify-start text-left font-normal',
            !value && 'text-muted-foreground',
            error && 'border-destructive',
            className
          )}
          disabled={disabled}
        >
          <CalendarIcon className='mr-2 h-4 w-4' />
          {formatDisplayValue()}
        </Button>
      </PopoverTrigger>
      <PopoverContent className='w-auto p-0' align='start'>
        <Calendar
          mode='single'
          selected={value}
          onSelect={handleDateSelect}
          initialFocus
        />
        {showTime && (
          <div className='border-t p-3'>
            <div className='flex items-center gap-2'>
              <Clock className='h-4 w-4' />
              <Input
                type='time'
                value={timeValue}
                onChange={(e) => handleTimeChange(e.target.value)}
                className='w-auto'
              />
            </div>
          </div>
        )}
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