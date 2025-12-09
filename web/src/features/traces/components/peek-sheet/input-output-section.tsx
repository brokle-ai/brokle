'use client'

import * as React from 'react'
import { Button } from '@/components/ui/button'
import { Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { PathValueTree } from './path-value-tree'
import { ChatMLView, extractChatML } from './chatml-view'
import { extractTools, extractToolCalls } from './tool-call-view'
import { CollapsibleSection } from './collapsible-section'
import type { ViewMode } from './detail-panel'

interface InputOutputSectionProps {
  title: 'Input' | 'Output'
  content: string | object | null | undefined
  viewMode: ViewMode
  defaultExpanded?: boolean
  className?: string
  /** Content from the other section (for tool call matching) */
  counterpartContent?: string | object | null | undefined
}

type ContentFormat = 'plain' | 'json' | 'chatml' | 'tools'

interface ParsedContent {
  parsed: unknown
  isJson: boolean
  formatted: string
  format: ContentFormat
  chatMessages?: ReturnType<typeof extractChatML>
  tools?: ReturnType<typeof extractTools>
  toolCalls?: ReturnType<typeof extractToolCalls>
}

/**
 * Try to parse content as JSON and detect special formats
 * Returns { parsed, isJson, formatted, format, chatMessages?, tools?, toolCalls? }
 */
function parseContent(
  content: string | object | null | undefined,
  detectChatML: boolean = true,
  isInput: boolean = true
): ParsedContent {
  if (content === null || content === undefined) {
    return { parsed: null, isJson: false, formatted: 'null', format: 'plain' }
  }

  let parsed: unknown = content
  let isJson = false
  let formatted = String(content)

  // Parse JSON if string
  if (typeof content === 'string') {
    try {
      const jsonParsed = JSON.parse(content)
      if (typeof jsonParsed === 'object' && jsonParsed !== null) {
        parsed = jsonParsed
        isJson = true
        formatted = JSON.stringify(jsonParsed, null, 2)
      }
    } catch {
      // Not JSON, keep as plain text
      return { parsed: content, isJson: false, formatted: content, format: 'plain' }
    }
  } else if (typeof content === 'object') {
    parsed = content
    isJson = true
    formatted = JSON.stringify(content, null, 2)
  }

  // Detect ChatML format
  if (detectChatML && isJson) {
    const chatMessages = extractChatML(parsed)
    if (chatMessages) {
      return {
        parsed,
        isJson: true,
        formatted,
        format: 'chatml',
        chatMessages,
      }
    }
  }

  // Detect tool definitions (in input) or tool calls (in output)
  if (isJson) {
    if (isInput) {
      const tools = extractTools(parsed)
      if (tools && tools.length > 0) {
        return {
          parsed,
          isJson: true,
          formatted,
          format: 'tools',
          tools,
        }
      }
    } else {
      const toolCalls = extractToolCalls(parsed)
      if (toolCalls && toolCalls.length > 0) {
        return {
          parsed,
          isJson: true,
          formatted,
          format: 'tools',
          toolCalls,
        }
      }
    }
  }

  return { parsed, isJson, formatted, format: isJson ? 'json' : 'plain' }
}

/**
 * CopyButton - Small copy button with feedback
 */
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
      className='h-6 w-6'
      onClick={(e) => {
        e.stopPropagation()
        handleCopy()
      }}
      title='Copy to clipboard'
    >
      {copied ? (
        <Check className='h-3 w-3 text-green-500' />
      ) : (
        <Copy className='h-3 w-3' />
      )}
    </Button>
  )
}

/**
 * InputOutputSection - Collapsible section for displaying Input/Output data
 *
 * ViewMode is controlled from parent (DetailPanel) and applies globally.
 * Uses CollapsibleSection for consistent styling across all sections.
 * Features:
 * - PathValueTree for formatted view (when isJson and viewMode='formatted')
 * - ChatML view for message arrays (auto-detected)
 * - Tool definitions/invocations view (auto-detected)
 * - Raw JSON/text view (when viewMode='json' or non-JSON content)
 * - Copy functionality
 */
export function InputOutputSection({
  title,
  content,
  viewMode,
  defaultExpanded = true,
  className,
}: InputOutputSectionProps) {
  const isInput = title === 'Input'
  const parsedData = React.useMemo(
    () => parseContent(content, true, isInput),
    [content, isInput]
  )
  const { parsed, isJson, formatted, format, chatMessages } = parsedData

  const hasContent = content !== null && content !== undefined && content !== ''
  const isEmpty = !hasContent

  // Get type badge text based on format
  const getTypeBadge = () => {
    if (format === 'chatml') return 'ChatML'
    if (format === 'tools') return 'Tools'
    return undefined
  }

  // Render content based on format and viewMode
  const renderContent = () => {
    // Always show raw JSON if viewMode is 'json'
    if (viewMode === 'json') {
      return (
        <pre
          className={cn(
            'bg-muted/50 rounded-md p-3 text-xs font-mono',
            'overflow-x-auto max-h-[400px]',
            'whitespace-pre-wrap break-words'
          )}
        >
          {formatted}
        </pre>
      )
    }

    // ChatML format - render as message bubbles
    if (format === 'chatml' && chatMessages) {
      return (
        <div className='bg-muted/30 rounded-md p-3 max-h-[400px] overflow-auto'>
          <ChatMLView messages={chatMessages} />
        </div>
      )
    }

    // Default: PathValueTree for JSON, plain text otherwise
    if (isJson) {
      return (
        <div className='bg-muted/50 rounded-md p-3 max-h-[400px] overflow-auto'>
          <PathValueTree
            data={parsed as Record<string, unknown>}
            maxInitialDepth={2}
            showCopyButtons={true}
          />
        </div>
      )
    }

    return (
      <pre
        className={cn(
          'bg-muted/50 rounded-md p-3 text-xs font-mono',
          'overflow-x-auto max-h-[400px]',
          'whitespace-pre-wrap break-words'
        )}
      >
        {formatted}
      </pre>
    )
  }

  return (
    <CollapsibleSection
      title={title}
      typeBadge={getTypeBadge()}
      defaultExpanded={defaultExpanded}
      className={className}
      emptyMessage={`No ${title.toLowerCase()} data`}
    >
      {!isEmpty && (
        <div className='relative'>
          {/* Copy button */}
          <div className='absolute top-2 right-2 z-10'>
            <CopyButton value={formatted} />
          </div>
          {renderContent()}
        </div>
      )}
    </CollapsibleSection>
  )
}
