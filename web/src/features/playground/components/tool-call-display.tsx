'use client'

import { useState, useCallback, useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { ChevronDown, ChevronRight, Wrench, Plus, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ToolCall, ChatMessage } from '../types'
import { formatToolCallArguments, createMessage } from '../types'

interface ToolCallDisplayProps {
  toolCalls: ToolCall[]
  /** Called when user clicks "Add to messages" - receives new assistant and tool result messages */
  onAddToMessages?: (messages: ChatMessage[]) => void
  className?: string
}

export function ToolCallDisplay({
  toolCalls,
  onAddToMessages,
  className,
}: ToolCallDisplayProps) {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set())
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const toggleExpand = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }, [])

  const handleCopyArgs = useCallback(async (toolCall: ToolCall) => {
    try {
      const formatted = formatToolCallArguments(toolCall.function.arguments)
      await navigator.clipboard.writeText(formatted)
      setCopiedId(toolCall.id)
      setTimeout(() => setCopiedId(null), 2000)
    } catch {
      // Clipboard API failed
    }
  }, [])

  // Create messages for multi-turn conversation (Langfuse pattern)
  const handleAddToMessages = useCallback(() => {
    if (!onAddToMessages) return

    // Create an assistant message containing the tool calls
    const assistantMessage = createMessage(
      'assistant',
      `[Tool calls: ${toolCalls.map((tc) => tc.function.name).join(', ')}]`
    )

    // For each tool call, the user would typically add tool results as user messages
    // Following Langfuse pattern: just add assistant message acknowledging tool calls
    // User can then manually add tool results
    onAddToMessages([assistantMessage])
  }, [toolCalls, onAddToMessages])

  if (toolCalls.length === 0) {
    return null
  }

  return (
    <Card className={cn('border-amber-500/30 bg-amber-50/30 dark:bg-amber-950/10', className)}>
      <CardHeader className="py-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Wrench className="h-4 w-4 text-amber-600 dark:text-amber-400" />
            <CardTitle className="text-sm font-medium">
              Tool Calls ({toolCalls.length})
            </CardTitle>
          </div>
          {onAddToMessages && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleAddToMessages}
              className="h-7 text-xs"
            >
              <Plus className="mr-1 h-3 w-3" />
              Add to Messages
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-0 space-y-2">
        {toolCalls.map((toolCall) => {
          const isExpanded = expandedIds.has(toolCall.id)
          const formattedArgs = formatToolCallArguments(toolCall.function.arguments)
          const isCopied = copiedId === toolCall.id

          return (
            <Collapsible
              key={toolCall.id}
              open={isExpanded}
              onOpenChange={() => toggleExpand(toolCall.id)}
            >
              <Card>
                <CollapsibleTrigger asChild>
                  <div className="flex items-center gap-2 p-3 cursor-pointer hover:bg-muted/50 transition-colors">
                    {isExpanded ? (
                      <ChevronDown className="h-4 w-4 flex-shrink-0" />
                    ) : (
                      <ChevronRight className="h-4 w-4 flex-shrink-0" />
                    )}
                    <span className="font-mono text-sm font-medium">
                      {toolCall.function.name}
                    </span>
                    <Badge variant="outline" className="text-xs">
                      {toolCall.id.slice(0, 8)}...
                    </Badge>
                  </div>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <div className="px-3 pb-3 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">Arguments:</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs"
                        onClick={() => handleCopyArgs(toolCall)}
                      >
                        {isCopied ? (
                          <>
                            <Check className="mr-1 h-3 w-3" />
                            Copied
                          </>
                        ) : (
                          <>
                            <Copy className="mr-1 h-3 w-3" />
                            Copy
                          </>
                        )}
                      </Button>
                    </div>
                    <pre className="text-xs font-mono bg-muted p-2 rounded overflow-x-auto max-h-48">
                      {formattedArgs}
                    </pre>
                  </div>
                </CollapsibleContent>
              </Card>
            </Collapsible>
          )
        })}
      </CardContent>
    </Card>
  )
}

interface ToolCallsSummaryProps {
  toolCalls: ToolCall[]
  className?: string
}

/**
 * Compact summary view for tool calls (shown in streaming output header)
 */
export function ToolCallsSummary({ toolCalls, className }: ToolCallsSummaryProps) {
  if (toolCalls.length === 0) return null

  return (
    <div className={cn('flex items-center gap-2 text-xs', className)}>
      <Wrench className="h-3 w-3 text-amber-600 dark:text-amber-400" />
      <span className="text-muted-foreground">
        {toolCalls.length} tool call{toolCalls.length !== 1 ? 's' : ''}:
      </span>
      <div className="flex flex-wrap gap-1">
        {toolCalls.map((tc) => (
          <Badge key={tc.id} variant="secondary" className="text-xs">
            {tc.function.name}
          </Badge>
        ))}
      </div>
    </div>
  )
}
