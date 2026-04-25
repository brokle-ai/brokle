import * as React from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import type { ChatMessage } from '../../types'

interface ChatMessageListProps {
  messages: ChatMessage[]
  className?: string
  /**
   * Default open state for each message. Set false to render the
   * list collapsed by default for long histories.
   */
  defaultOpen?: boolean
}

function roleColor(message: ChatMessage): string {
  if (message.type === 'placeholder') {
    return 'bg-amber-100 text-amber-900 dark:bg-amber-900/40 dark:text-amber-100'
  }
  switch (message.role) {
    case 'system':
      return 'bg-purple-100 text-purple-900 dark:bg-purple-900/40 dark:text-purple-100'
    case 'assistant':
      return 'bg-blue-100 text-blue-900 dark:bg-blue-900/40 dark:text-blue-100'
    case 'user':
      return 'bg-muted text-foreground'
    default:
      return 'bg-muted text-foreground'
  }
}

function roleLabel(message: ChatMessage): string {
  if (message.type === 'placeholder') {
    return message.name ? `Placeholder · ${message.name}` : 'Placeholder'
  }
  return message.role ?? 'message'
}

function ChatMessageRow({
  message,
  defaultOpen,
}: {
  message: ChatMessage
  defaultOpen: boolean
}) {
  const [open, setOpen] = React.useState(defaultOpen)
  const content = message.content ?? ''
  const lineCount = content.length === 0 ? 0 : content.split('\n').length
  const previewLine = content.split('\n', 1)[0] ?? ''

  return (
    <div className={cn('rounded-md border overflow-hidden', roleColor(message))}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left"
      >
        <div className="flex items-center gap-2 min-w-0">
          {open ? (
            <ChevronDown className="h-3.5 w-3.5 shrink-0" />
          ) : (
            <ChevronRight className="h-3.5 w-3.5 shrink-0" />
          )}
          <Badge
            variant="outline"
            className="bg-background/60 font-mono text-[10px] uppercase tracking-wide"
          >
            {roleLabel(message)}
          </Badge>
          {!open && previewLine.length > 0 && (
            <span className="truncate text-xs text-muted-foreground">
              {previewLine}
            </span>
          )}
        </div>
        <span className="shrink-0 text-[10px] text-muted-foreground">
          {lineCount} {lineCount === 1 ? 'line' : 'lines'}
        </span>
      </button>
      {open && (
        <pre className="whitespace-pre-wrap break-words border-t bg-background/40 px-3 py-2 font-mono text-sm">
          {content || ' '}
        </pre>
      )}
    </div>
  )
}

export function ChatMessageList({
  messages,
  className,
  defaultOpen = true,
}: ChatMessageListProps) {
  if (messages.length === 0) {
    return (
      <div className="rounded-md border border-dashed p-4 text-center text-sm text-muted-foreground">
        No messages
      </div>
    )
  }

  return (
    <div className={cn('space-y-2', className)}>
      {messages.map((msg, i) => (
        <ChatMessageRow
          key={i}
          message={msg}
          defaultOpen={defaultOpen}
        />
      ))}
    </div>
  )
}
