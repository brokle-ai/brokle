'use client'

import * as React from 'react'
import ReactMarkdown from 'react-markdown'
import { User, Bot, Settings, Wrench, ChevronDown, ChevronRight, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'

// ============================================================================
// Types
// ============================================================================

interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'tool' | 'function'
  content: string | null
  name?: string
  tool_calls?: ToolCall[]
  tool_call_id?: string
  function_call?: FunctionCall
}

interface ToolCall {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

interface FunctionCall {
  name: string
  arguments: string
}

interface ChatMLViewProps {
  messages: ChatMessage[]
  className?: string
  collapseAfter?: number // Collapse messages after this count (default: 3)
  maxMarkdownChars?: number // Max chars before disabling markdown (default: 20000)
}

// ============================================================================
// Detection Function (export for use in InputOutputSection)
// ============================================================================

/**
 * Detect if content is in ChatML format (messages array with role/content)
 */
export function isChatMLFormat(content: unknown): content is ChatMessage[] {
  if (!Array.isArray(content)) return false
  if (content.length === 0) return false

  // Check if at least one item looks like a chat message
  return content.some((item) => {
    if (typeof item !== 'object' || item === null) return false
    const msg = item as Record<string, unknown>
    return (
      typeof msg.role === 'string' &&
      ['system', 'user', 'assistant', 'tool', 'function'].includes(msg.role)
    )
  })
}

/**
 * Try to extract ChatML from various formats
 */
export function extractChatML(content: unknown): ChatMessage[] | null {
  // Direct array of messages
  if (isChatMLFormat(content)) return content

  // Object with messages property
  if (typeof content === 'object' && content !== null) {
    const obj = content as Record<string, unknown>
    if (isChatMLFormat(obj.messages)) return obj.messages
  }

  return null
}

// ============================================================================
// Helper Components
// ============================================================================

function CopyButton({ value }: { value: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant='ghost'
      size='icon'
      className='h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity'
      onClick={(e) => {
        e.stopPropagation()
        handleCopy()
      }}
      title='Copy message'
    >
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3' />
      )}
    </Button>
  )
}

// ============================================================================
// Role Configuration
// ============================================================================

const roleConfig: Record<
  string,
  {
    icon: React.ElementType
    label: string
    bgClass: string
    borderClass: string
    iconClass: string
  }
> = {
  system: {
    icon: Settings,
    label: 'System',
    bgClass: 'bg-slate-50 dark:bg-slate-900/50',
    borderClass: 'border-slate-200 dark:border-slate-700',
    iconClass: 'text-slate-600 dark:text-slate-400',
  },
  user: {
    icon: User,
    label: 'User',
    bgClass: 'bg-blue-50 dark:bg-blue-900/30',
    borderClass: 'border-blue-200 dark:border-blue-800',
    iconClass: 'text-blue-600 dark:text-blue-400',
  },
  assistant: {
    icon: Bot,
    label: 'Assistant',
    bgClass: 'bg-purple-50 dark:bg-purple-900/30',
    borderClass: 'border-purple-200 dark:border-purple-800',
    iconClass: 'text-purple-600 dark:text-purple-400',
  },
  tool: {
    icon: Wrench,
    label: 'Tool',
    bgClass: 'bg-orange-50 dark:bg-orange-900/30',
    borderClass: 'border-orange-200 dark:border-orange-800',
    iconClass: 'text-orange-600 dark:text-orange-400',
  },
  function: {
    icon: Wrench,
    label: 'Function',
    bgClass: 'bg-orange-50 dark:bg-orange-900/30',
    borderClass: 'border-orange-200 dark:border-orange-800',
    iconClass: 'text-orange-600 dark:text-orange-400',
  },
}

// ============================================================================
// MessageBubble Component
// ============================================================================

interface MessageBubbleProps {
  message: ChatMessage
  enableMarkdown: boolean
}

function MessageBubble({ message, enableMarkdown }: MessageBubbleProps) {
  const config = roleConfig[message.role] || roleConfig.user
  const Icon = config.icon

  // Determine label (use name for tool/function calls)
  const label = message.name || config.label

  // Get content to display
  const content = message.content || ''
  const hasContent = content.length > 0

  // Handle tool calls in assistant messages
  const hasToolCalls = message.tool_calls && message.tool_calls.length > 0
  const hasFunctionCall = message.function_call !== undefined

  return (
    <div
      className={cn(
        'group rounded-lg border p-3 space-y-2',
        config.bgClass,
        config.borderClass
      )}
    >
      {/* Header */}
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Icon className={cn('h-4 w-4', config.iconClass)} />
          <span className='text-xs font-medium text-muted-foreground'>{label}</span>
          {message.tool_call_id && (
            <span className='text-xs text-muted-foreground font-mono'>
              ({message.tool_call_id.slice(0, 8)}...)
            </span>
          )}
        </div>
        <CopyButton value={content || JSON.stringify(message, null, 2)} />
      </div>

      {/* Content */}
      {hasContent && (
        <div className='text-sm'>
          {enableMarkdown && message.role === 'assistant' ? (
            <div className='prose prose-sm dark:prose-invert max-w-none prose-p:my-1 prose-pre:my-2 prose-pre:bg-muted prose-pre:text-muted-foreground'>
              <ReactMarkdown>{content}</ReactMarkdown>
            </div>
          ) : (
            <div className='whitespace-pre-wrap break-words'>{content}</div>
          )}
        </div>
      )}

      {/* Tool Calls */}
      {hasToolCalls && (
        <div className='space-y-2 pt-2 border-t border-dashed border-current/20'>
          <div className='text-xs font-medium text-muted-foreground'>Tool Calls:</div>
          {message.tool_calls!.map((call, idx) => (
            <ToolCallCard key={call.id || idx} call={call} />
          ))}
        </div>
      )}

      {/* Legacy Function Call */}
      {hasFunctionCall && (
        <div className='space-y-2 pt-2 border-t border-dashed border-current/20'>
          <div className='text-xs font-medium text-muted-foreground'>Function Call:</div>
          <FunctionCallCard call={message.function_call!} />
        </div>
      )}
    </div>
  )
}

// ============================================================================
// ToolCallCard Component
// ============================================================================

function ToolCallCard({ call }: { call: ToolCall }) {
  const [isExpanded, setIsExpanded] = React.useState(false)

  let parsedArgs: unknown = null
  try {
    parsedArgs = JSON.parse(call.function.arguments)
  } catch {
    // Keep as string
  }

  return (
    <div className='bg-background/50 rounded-md border p-2'>
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <CollapsibleTrigger asChild>
          <button className='flex items-center gap-2 w-full text-left hover:bg-muted/50 rounded p-1 -m-1'>
            {isExpanded ? (
              <ChevronDown className='h-3 w-3 text-muted-foreground' />
            ) : (
              <ChevronRight className='h-3 w-3 text-muted-foreground' />
            )}
            <Wrench className='h-3 w-3 text-orange-500' />
            <span className='text-xs font-medium'>{call.function.name}</span>
            <span className='text-xs text-muted-foreground ml-auto font-mono'>
              {call.id.slice(0, 8)}...
            </span>
          </button>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <pre className='mt-2 text-xs font-mono bg-muted rounded p-2 overflow-x-auto'>
            {parsedArgs
              ? JSON.stringify(parsedArgs, null, 2)
              : call.function.arguments}
          </pre>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

// ============================================================================
// FunctionCallCard Component (Legacy)
// ============================================================================

function FunctionCallCard({ call }: { call: FunctionCall }) {
  const [isExpanded, setIsExpanded] = React.useState(false)

  let parsedArgs: unknown = null
  try {
    parsedArgs = JSON.parse(call.arguments)
  } catch {
    // Keep as string
  }

  return (
    <div className='bg-background/50 rounded-md border p-2'>
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <CollapsibleTrigger asChild>
          <button className='flex items-center gap-2 w-full text-left hover:bg-muted/50 rounded p-1 -m-1'>
            {isExpanded ? (
              <ChevronDown className='h-3 w-3 text-muted-foreground' />
            ) : (
              <ChevronRight className='h-3 w-3 text-muted-foreground' />
            )}
            <Wrench className='h-3 w-3 text-orange-500' />
            <span className='text-xs font-medium'>{call.name}</span>
          </button>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <pre className='mt-2 text-xs font-mono bg-muted rounded p-2 overflow-x-auto'>
            {parsedArgs ? JSON.stringify(parsedArgs, null, 2) : call.arguments}
          </pre>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

// ============================================================================
// ChatMLView Component (Main Export)
// ============================================================================

/**
 * ChatMLView - Renders OpenAI-style chat messages with role-based styling
 *
 * Features:
 * - Role-based message styling (system, user, assistant, tool)
 * - Markdown rendering for assistant messages (with char limit)
 * - Tool call visualization
 * - Message collapsing for long conversations
 * - Copy functionality
 */
export function ChatMLView({
  messages,
  className,
  collapseAfter = 3,
  maxMarkdownChars = 20000,
}: ChatMLViewProps) {
  const [showAll, setShowAll] = React.useState(false)

  // Calculate total content length for markdown toggle
  const totalChars = React.useMemo(
    () =>
      messages.reduce((sum, msg) => sum + (msg.content?.length || 0), 0),
    [messages]
  )
  const enableMarkdown = totalChars <= maxMarkdownChars

  // Determine which messages to show
  const shouldCollapse = messages.length > collapseAfter && !showAll
  const visibleMessages = shouldCollapse
    ? [messages[0], ...messages.slice(-2)]
    : messages
  const hiddenCount = messages.length - 3

  return (
    <div className={cn('space-y-3', className)}>
      {visibleMessages.map((msg, idx) => (
        <React.Fragment key={idx}>
          {/* Show "N more messages" button */}
          {shouldCollapse && idx === 1 && hiddenCount > 0 && (
            <Button
              variant='outline'
              size='sm'
              className='w-full text-xs'
              onClick={() => setShowAll(true)}
            >
              Show {hiddenCount} more message{hiddenCount !== 1 ? 's' : ''}
            </Button>
          )}
          <MessageBubble message={msg} enableMarkdown={enableMarkdown} />
        </React.Fragment>
      ))}

      {/* Show "Collapse" button */}
      {showAll && messages.length > collapseAfter && (
        <Button
          variant='ghost'
          size='sm'
          className='w-full text-xs text-muted-foreground'
          onClick={() => setShowAll(false)}
        >
          Collapse messages
        </Button>
      )}

      {/* Markdown disabled warning */}
      {!enableMarkdown && (
        <div className='text-xs text-muted-foreground text-center py-1'>
          Markdown rendering disabled (content exceeds {maxMarkdownChars.toLocaleString()} chars)
        </div>
      )}
    </div>
  )
}
