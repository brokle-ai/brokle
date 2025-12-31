'use client'

import { useMemo } from 'react'
import { EyeIcon, Loader2Icon, AlertCircleIcon, CheckCircleIcon } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { TextTemplate, ChatTemplate, PromptType, TemplateDialect } from '@/features/prompts/types'

interface TemplatePreviewProps {
  compiled: TextTemplate | ChatTemplate | null
  type: PromptType
  dialect?: TemplateDialect
  isLoading?: boolean
  error?: string | null
  className?: string
}

/**
 * TemplatePreview - Shows the compiled template result.
 *
 * Features:
 * - Displays compiled text or chat messages
 * - Shows loading state during compilation
 * - Displays error messages
 * - Syntax highlighting for the output
 */
export function TemplatePreview({
  compiled,
  type,
  dialect,
  isLoading = false,
  error = null,
  className,
}: TemplatePreviewProps) {
  return (
    <div className={cn('flex flex-col gap-2', className)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <EyeIcon className="size-4" />
          <span>Preview</span>
          {isLoading && <Loader2Icon className="size-3 animate-spin" />}
        </div>
        {dialect && dialect !== 'auto' && (
          <Badge variant="outline" className="text-xs">
            {dialect}
          </Badge>
        )}
      </div>

      <div className="rounded-md border bg-muted/50 min-h-[100px]">
        {error ? (
          <PreviewError message={error} />
        ) : isLoading ? (
          <PreviewLoading />
        ) : compiled ? (
          type === 'text' ? (
            <TextPreview content={(compiled as TextTemplate).content} />
          ) : (
            <ChatPreview messages={(compiled as ChatTemplate).messages} />
          )
        ) : (
          <PreviewEmpty />
        )}
      </div>
    </div>
  )
}

function PreviewLoading() {
  return (
    <div className="flex items-center justify-center h-[100px] text-muted-foreground text-sm">
      <Loader2Icon className="size-4 animate-spin mr-2" />
      <span>Compiling...</span>
    </div>
  )
}

function PreviewEmpty() {
  return (
    <div className="flex items-center justify-center h-[100px] text-muted-foreground text-sm">
      Enter variable values to see preview
    </div>
  )
}

function PreviewError({ message }: { message: string }) {
  return (
    <div className="flex items-start gap-2 p-3 text-destructive">
      <AlertCircleIcon className="size-4 mt-0.5 shrink-0" />
      <div className="text-sm">{message}</div>
    </div>
  )
}

function TextPreview({ content }: { content: string }) {
  return (
    <ScrollArea className="h-[200px]">
      <pre className="p-3 text-sm font-mono whitespace-pre-wrap break-words">
        {content || <span className="text-muted-foreground italic">Empty content</span>}
      </pre>
    </ScrollArea>
  )
}

interface ChatMessage {
  type?: string
  role?: string
  content?: string
  name?: string
}

function ChatPreview({ messages }: { messages: ChatMessage[] }) {
  if (!messages || messages.length === 0) {
    return (
      <div className="flex items-center justify-center h-[100px] text-muted-foreground text-sm">
        No messages
      </div>
    )
  }

  return (
    <ScrollArea className="h-[200px]">
      <div className="p-3 space-y-3">
        {messages.map((message, index) => (
          <ChatMessagePreview key={index} message={message} />
        ))}
      </div>
    </ScrollArea>
  )
}

function ChatMessagePreview({ message }: { message: ChatMessage }) {
  const roleColor = useMemo(() => {
    switch (message.role) {
      case 'system':
        return 'text-purple-600 dark:text-purple-400'
      case 'user':
        return 'text-blue-600 dark:text-blue-400'
      case 'assistant':
        return 'text-green-600 dark:text-green-400'
      default:
        return 'text-muted-foreground'
    }
  }, [message.role])

  const roleBg = useMemo(() => {
    switch (message.role) {
      case 'system':
        return 'bg-purple-50 dark:bg-purple-950/20 border-purple-200 dark:border-purple-800'
      case 'user':
        return 'bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800'
      case 'assistant':
        return 'bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800'
      default:
        return 'bg-muted border-border'
    }
  }, [message.role])

  // Handle placeholder messages
  if (message.type === 'placeholder') {
    return (
      <div className="flex items-center gap-2 p-2 rounded border border-dashed border-muted-foreground/30 text-muted-foreground text-sm">
        <Badge variant="outline" className="text-xs">
          placeholder
        </Badge>
        <span className="font-mono">{message.name || 'unnamed'}</span>
      </div>
    )
  }

  return (
    <div className={cn('rounded-md border p-2', roleBg)}>
      <div className={cn('text-xs font-medium mb-1 uppercase', roleColor)}>
        {message.role || 'unknown'}
      </div>
      <div className="text-sm whitespace-pre-wrap break-words">
        {message.content || (
          <span className="text-muted-foreground italic">Empty message</span>
        )}
      </div>
    </div>
  )
}

interface ValidationStatusProps {
  isValid: boolean | null
  isValidating: boolean
  errorCount?: number
  warningCount?: number
  className?: string
}

export function ValidationStatus({
  isValid,
  isValidating,
  errorCount = 0,
  warningCount = 0,
  className,
}: ValidationStatusProps) {
  if (isValidating) {
    return (
      <div className={cn('flex items-center gap-1.5 text-muted-foreground text-xs', className)}>
        <Loader2Icon className="size-3 animate-spin" />
        <span>Validating...</span>
      </div>
    )
  }

  if (isValid === null) {
    return null
  }

  if (isValid) {
    return (
      <div className={cn('flex items-center gap-1.5 text-green-600 dark:text-green-400 text-xs', className)}>
        <CheckCircleIcon className="size-3" />
        <span>Valid</span>
        {warningCount > 0 && (
          <Badge variant="outline" className="text-xs text-amber-600 dark:text-amber-400 border-amber-300">
            {warningCount} warning{warningCount !== 1 ? 's' : ''}
          </Badge>
        )}
      </div>
    )
  }

  return (
    <div className={cn('flex items-center gap-1.5 text-destructive text-xs', className)}>
      <AlertCircleIcon className="size-3" />
      <span>
        {errorCount} error{errorCount !== 1 ? 's' : ''}
      </span>
      {warningCount > 0 && (
        <span className="text-amber-600 dark:text-amber-400">
          , {warningCount} warning{warningCount !== 1 ? 's' : ''}
        </span>
      )}
    </div>
  )
}

interface SyntaxError {
  line: number
  column: number
  message: string
  code: string
}

interface SyntaxErrorListProps {
  errors: SyntaxError[]
  onErrorClick?: (error: SyntaxError) => void
  className?: string
}

export function SyntaxErrorList({
  errors,
  onErrorClick,
  className,
}: SyntaxErrorListProps) {
  if (errors.length === 0) {
    return null
  }

  return (
    <div className={cn('flex flex-col gap-1', className)}>
      <div className="text-xs font-medium text-destructive flex items-center gap-1">
        <AlertCircleIcon className="size-3" />
        Errors
      </div>
      <div className="space-y-1">
        {errors.map((error, index) => (
          <div
            key={index}
            className={cn(
              'text-xs p-2 rounded bg-destructive/10 text-destructive',
              onErrorClick && 'cursor-pointer hover:bg-destructive/20'
            )}
            onClick={() => onErrorClick?.(error)}
          >
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="text-[10px] px-1">
                {error.code}
              </Badge>
              <span className="text-muted-foreground">
                Line {error.line}, Col {error.column}
              </span>
            </div>
            <div className="mt-1">{error.message}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
