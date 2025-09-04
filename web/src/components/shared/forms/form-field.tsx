'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { AlertCircle, Eye, EyeOff } from 'lucide-react'

interface FormFieldProps extends React.HTMLAttributes<HTMLDivElement> {
  label?: string
  description?: string
  error?: string
  required?: boolean
  children: React.ReactNode
}

export function FormField({
  label,
  description,
  error,
  required,
  children,
  className,
  ...props
}: FormFieldProps) {
  const fieldId = React.useId()

  return (
    <div className={cn('space-y-2', className)} {...props}>
      {label && (
        <Label htmlFor={fieldId} className='text-sm font-medium'>
          {label}
          {required && <span className='text-destructive ml-1'>*</span>}
        </Label>
      )}
      
      {description && (
        <p className='text-sm text-muted-foreground'>{description}</p>
      )}

      <div className='relative'>
        {React.cloneElement(children as React.ReactElement, {
          id: fieldId,
          className: cn(
            error && 'border-destructive focus-visible:ring-destructive',
            (children as React.ReactElement).props.className
          ),
        })}
      </div>

      {error && (
        <div className='flex items-center gap-2 text-sm text-destructive'>
          <AlertCircle className='h-4 w-4' />
          <span>{error}</span>
        </div>
      )}
    </div>
  )
}

interface EnhancedInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  description?: string
  error?: string
  showPasswordToggle?: boolean
}

export function EnhancedInput({
  label,
  description,
  error,
  showPasswordToggle = false,
  type,
  className,
  ...props
}: EnhancedInputProps) {
  const [showPassword, setShowPassword] = React.useState(false)
  const [inputType, setInputType] = React.useState(type)

  React.useEffect(() => {
    if (showPasswordToggle && type === 'password') {
      setInputType(showPassword ? 'text' : 'password')
    }
  }, [showPassword, showPasswordToggle, type])

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword)
  }

  return (
    <FormField
      label={label}
      description={description}
      error={error}
      required={props.required}
      className={className}
    >
      <div className='relative'>
        <Input
          {...props}
          type={inputType}
          className={cn(
            showPasswordToggle && 'pr-10'
          )}
        />
        {showPasswordToggle && type === 'password' && (
          <button
            type='button'
            onClick={togglePasswordVisibility}
            className='absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors'
          >
            {showPassword ? (
              <EyeOff className='h-4 w-4' />
            ) : (
              <Eye className='h-4 w-4' />
            )}
          </button>
        )}
      </div>
    </FormField>
  )
}

interface EnhancedTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  description?: string
  error?: string
  showCharCount?: boolean
  maxLength?: number
}

export function EnhancedTextarea({
  label,
  description,
  error,
  showCharCount = false,
  maxLength,
  value,
  className,
  ...props
}: EnhancedTextareaProps) {
  const charCount = value ? String(value).length : 0

  return (
    <FormField
      label={label}
      description={description}
      error={error}
      required={props.required}
      className={className}
    >
      <div className='space-y-2'>
        <Textarea
          {...props}
          value={value}
          maxLength={maxLength}
        />
        {showCharCount && (
          <div className='flex justify-between text-xs text-muted-foreground'>
            <span>
              {charCount}{maxLength && ` / ${maxLength}`} characters
            </span>
            {maxLength && charCount > maxLength * 0.9 && (
              <span className={cn(
                charCount >= maxLength ? 'text-destructive' : 'text-yellow-600'
              )}>
                {charCount >= maxLength ? 'Limit reached' : 'Near limit'}
              </span>
            )}
          </div>
        )}
      </div>
    </FormField>
  )
}