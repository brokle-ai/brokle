'use client'

import * as React from 'react'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { ChevronDown, ChevronUp, MessageSquare } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ReasonEditorProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  maxLength?: number
  disabled?: boolean
  collapsible?: boolean
  defaultExpanded?: boolean
  className?: string
}

/**
 * Expandable textarea for score reasons/explanations
 *
 * Features:
 * - Optional collapsible mode
 * - Character count with max length
 * - Expandable height
 * - Placeholder with hint
 */
export function ReasonEditor({
  value,
  onChange,
  placeholder = 'Explain why you gave this score...',
  maxLength = 1000,
  disabled = false,
  collapsible = true,
  defaultExpanded = false,
  className,
}: ReasonEditorProps) {
  const [isExpanded, setIsExpanded] = React.useState(defaultExpanded || !!value)
  const textareaRef = React.useRef<HTMLTextAreaElement>(null)

  // Auto-expand when value is set
  React.useEffect(() => {
    if (value && !isExpanded) {
      setIsExpanded(true)
    }
  }, [value, isExpanded])

  // Focus textarea when expanded
  React.useEffect(() => {
    if (isExpanded && textareaRef.current) {
      textareaRef.current.focus()
    }
  }, [isExpanded])

  const charCount = value.length
  const isNearLimit = charCount > maxLength * 0.9

  if (collapsible && !isExpanded) {
    return (
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={() => setIsExpanded(true)}
        disabled={disabled}
        className={cn('w-full justify-start gap-2 text-muted-foreground', className)}
      >
        <MessageSquare className="h-4 w-4" />
        Add explanation (optional)
        <ChevronDown className="h-4 w-4 ml-auto" />
      </Button>
    )
  }

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between">
        <Label htmlFor="reason" className="text-sm">
          Explanation
          <span className="text-muted-foreground ml-1">(optional)</span>
        </Label>
        {collapsible && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded(false)}
            disabled={disabled}
            className="h-6 px-2"
          >
            <ChevronUp className="h-4 w-4" />
          </Button>
        )}
      </div>
      <Textarea
        ref={textareaRef}
        id="reason"
        value={value}
        onChange={(e) => onChange(e.target.value.slice(0, maxLength))}
        placeholder={placeholder}
        disabled={disabled}
        rows={3}
        className="resize-y min-h-[80px]"
      />
      <div className="flex justify-end">
        <span
          className={cn(
            'text-xs',
            isNearLimit ? 'text-warning' : 'text-muted-foreground',
            charCount >= maxLength && 'text-destructive'
          )}
        >
          {charCount}/{maxLength}
        </span>
      </div>
    </div>
  )
}

/**
 * Compact reason display with expand capability
 */
interface ReasonDisplayProps {
  reason: string
  maxPreviewLength?: number
  className?: string
}

export function ReasonDisplay({
  reason,
  maxPreviewLength = 100,
  className,
}: ReasonDisplayProps) {
  const [isExpanded, setIsExpanded] = React.useState(false)
  const shouldTruncate = reason.length > maxPreviewLength

  if (!reason) return null

  return (
    <div className={cn('space-y-1', className)}>
      <p className={cn('text-sm', !isExpanded && shouldTruncate && 'line-clamp-2')}>
        {isExpanded ? reason : reason.slice(0, maxPreviewLength)}
        {!isExpanded && shouldTruncate && '...'}
      </p>
      {shouldTruncate && (
        <Button
          variant="link"
          size="sm"
          onClick={() => setIsExpanded(!isExpanded)}
          className="h-auto p-0 text-xs"
        >
          {isExpanded ? 'Show less' : 'Show more'}
        </Button>
      )}
    </div>
  )
}
