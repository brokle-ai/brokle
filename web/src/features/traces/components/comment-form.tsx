'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Send, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

const MAX_CHARS = 10000

interface CommentFormProps {
  initialContent?: string
  onSubmit: (content: string) => void
  onCancel?: () => void
  isSubmitting?: boolean
  placeholder?: string
  submitLabel?: string
  className?: string
  autoFocus?: boolean
}

/**
 * CommentForm - Form for creating or editing comments
 *
 * Features:
 * - Textarea with placeholder
 * - Character counter (max 10,000)
 * - Cmd/Ctrl+Enter to submit
 * - Clear form after successful submission
 */
export function CommentForm({
  initialContent = '',
  onSubmit,
  onCancel,
  isSubmitting = false,
  placeholder = 'Add a comment...',
  submitLabel = 'Send',
  className,
  autoFocus = false,
}: CommentFormProps) {
  const [content, setContent] = React.useState(initialContent)
  const textareaRef = React.useRef<HTMLTextAreaElement>(null)

  const canSubmit = content.trim().length > 0 && content.length <= MAX_CHARS && !isSubmitting

  const handleSubmit = () => {
    if (!canSubmit) return
    onSubmit(content.trim())
    if (!initialContent) {
      // Only clear if this is a new comment form, not an edit form
      setContent('')
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      handleSubmit()
    }
    // Escape to cancel in edit mode
    if (e.key === 'Escape' && onCancel) {
      e.preventDefault()
      onCancel()
    }
  }

  // Auto-resize textarea based on content
  React.useEffect(() => {
    const textarea = textareaRef.current
    if (textarea) {
      textarea.style.height = 'auto'
      textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`
    }
  }, [content])

  return (
    <div className={cn('space-y-2', className)}>
      <Textarea
        ref={textareaRef}
        value={content}
        onChange={(e) => setContent(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        className='min-h-[80px] resize-none'
        disabled={isSubmitting}
        autoFocus={autoFocus}
      />
      <div className='flex items-center justify-between'>
        <span
          className={cn(
            'text-xs text-muted-foreground',
            content.length > MAX_CHARS && 'text-destructive'
          )}
        >
          {content.length.toLocaleString()}/{MAX_CHARS.toLocaleString()}
        </span>
        <div className='flex items-center gap-2'>
          {onCancel && (
            <Button
              type='button'
              variant='ghost'
              size='sm'
              onClick={onCancel}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
          )}
          <Button
            type='button'
            size='sm'
            onClick={handleSubmit}
            disabled={!canSubmit}
          >
            {isSubmitting ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <>
                <Send className='h-4 w-4 mr-1' />
                {submitLabel}
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
