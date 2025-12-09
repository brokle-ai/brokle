'use client'

import * as React from 'react'
import {
  Wrench,
  ChevronDown,
  ChevronRight,
  Code,
  CheckCircle,
  XCircle,
  Circle,
  Copy,
  Check,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'

// ============================================================================
// Types
// ============================================================================

interface ToolDefinition {
  type?: 'function'
  function: {
    name: string
    description?: string
    parameters?: Record<string, unknown>
  }
}

interface FunctionDefinition {
  name: string
  description?: string
  parameters?: Record<string, unknown>
}

interface ToolInvocation {
  id: string
  type: 'function'
  function: {
    name: string
    arguments: string
  }
}

interface ToolCallViewProps {
  tools?: ToolDefinition[] | FunctionDefinition[]
  toolCalls?: ToolInvocation[]
  className?: string
}

// ============================================================================
// Detection Functions
// ============================================================================

/**
 * Check if content contains tool definitions
 */
export function hasToolDefinitions(content: unknown): content is ToolDefinition[] | FunctionDefinition[] {
  if (!Array.isArray(content)) return false
  if (content.length === 0) return false

  return content.some((item) => {
    if (typeof item !== 'object' || item === null) return false
    const tool = item as Record<string, unknown>
    // OpenAI tools format: { type: 'function', function: { name, ... } }
    if (tool.type === 'function' && typeof tool.function === 'object') return true
    // Legacy functions format: { name, description?, parameters? }
    if (typeof tool.name === 'string') return true
    return false
  })
}

/**
 * Extract tools from various input formats
 */
export function extractTools(content: unknown): ToolDefinition[] | FunctionDefinition[] | null {
  // Direct array of tools
  if (hasToolDefinitions(content)) return content

  // Object with tools property
  if (typeof content === 'object' && content !== null) {
    const obj = content as Record<string, unknown>
    if (hasToolDefinitions(obj.tools)) return obj.tools
    if (hasToolDefinitions(obj.functions)) return obj.functions
  }

  return null
}

/**
 * Check if content contains tool calls
 */
export function hasToolCalls(content: unknown): content is ToolInvocation[] {
  if (!Array.isArray(content)) return false
  if (content.length === 0) return false

  return content.some((item) => {
    if (typeof item !== 'object' || item === null) return false
    const call = item as Record<string, unknown>
    return (
      typeof call.id === 'string' &&
      call.type === 'function' &&
      typeof call.function === 'object'
    )
  })
}

/**
 * Extract tool calls from various output formats
 */
export function extractToolCalls(content: unknown): ToolInvocation[] | null {
  // Direct array of tool calls
  if (hasToolCalls(content)) return content

  // Object with tool_calls property
  if (typeof content === 'object' && content !== null) {
    const obj = content as Record<string, unknown>
    if (hasToolCalls(obj.tool_calls)) return obj.tool_calls
    // From assistant message
    if (typeof obj.message === 'object' && obj.message !== null) {
      const msg = obj.message as Record<string, unknown>
      if (hasToolCalls(msg.tool_calls)) return msg.tool_calls
    }
    // From choices[0].message.tool_calls
    if (Array.isArray(obj.choices) && obj.choices.length > 0) {
      const choice = obj.choices[0] as Record<string, unknown>
      if (typeof choice.message === 'object' && choice.message !== null) {
        const msg = choice.message as Record<string, unknown>
        if (hasToolCalls(msg.tool_calls)) return msg.tool_calls
      }
    }
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
      title='Copy'
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
// ToolDefinitionCard Component
// ============================================================================

interface ToolDefinitionCardProps {
  tool: ToolDefinition | FunctionDefinition
  wasCalled?: boolean
}

function ToolDefinitionCard({ tool, wasCalled }: ToolDefinitionCardProps) {
  const [isExpanded, setIsExpanded] = React.useState(false)

  // Normalize tool format
  const functionDef = 'function' in tool ? tool.function : tool

  return (
    <div
      className={cn(
        'group rounded-lg border p-3',
        wasCalled
          ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'
          : 'bg-muted/50 border-border'
      )}
    >
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <CollapsibleTrigger asChild>
          <button className='flex items-center gap-2 w-full text-left'>
            {isExpanded ? (
              <ChevronDown className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            ) : (
              <ChevronRight className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            )}
            <Wrench className='h-4 w-4 text-orange-500 flex-shrink-0' />
            <span className='font-medium text-sm'>{functionDef.name}</span>

            {wasCalled !== undefined && (
              <Badge
                variant={wasCalled ? 'default' : 'secondary'}
                className={cn(
                  'ml-auto text-xs',
                  wasCalled && 'bg-green-500 hover:bg-green-600'
                )}
              >
                {wasCalled ? (
                  <>
                    <CheckCircle className='h-3 w-3 mr-1' />
                    Called
                  </>
                ) : (
                  <>
                    <Circle className='h-3 w-3 mr-1' />
                    Not called
                  </>
                )}
              </Badge>
            )}

            <CopyButton value={JSON.stringify(functionDef, null, 2)} />
          </button>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <div className='mt-3 space-y-3'>
            {/* Description */}
            {functionDef.description && (
              <div className='text-sm text-muted-foreground'>
                {functionDef.description}
              </div>
            )}

            {/* Parameters Schema */}
            {functionDef.parameters && (
              <div className='space-y-1'>
                <div className='flex items-center gap-1.5 text-xs font-medium text-muted-foreground'>
                  <Code className='h-3 w-3' />
                  Parameters Schema
                </div>
                <pre className='text-xs font-mono bg-background rounded-md border p-2 overflow-x-auto'>
                  {JSON.stringify(functionDef.parameters, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

// ============================================================================
// ToolInvocationCard Component
// ============================================================================

interface ToolInvocationCardProps {
  invocation: ToolInvocation
}

function ToolInvocationCard({ invocation }: ToolInvocationCardProps) {
  const [isExpanded, setIsExpanded] = React.useState(true)

  // Try to parse arguments
  let parsedArgs: unknown = null
  try {
    parsedArgs = JSON.parse(invocation.function.arguments)
  } catch {
    // Keep as string
  }

  return (
    <div className='group rounded-lg border p-3 bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800'>
      <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
        <CollapsibleTrigger asChild>
          <button className='flex items-center gap-2 w-full text-left'>
            {isExpanded ? (
              <ChevronDown className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            ) : (
              <ChevronRight className='h-4 w-4 text-muted-foreground flex-shrink-0' />
            )}
            <Wrench className='h-4 w-4 text-orange-500 flex-shrink-0' />
            <span className='font-medium text-sm'>{invocation.function.name}</span>
            <span className='text-xs text-muted-foreground font-mono ml-auto mr-2'>
              {invocation.id.slice(0, 12)}...
            </span>
            <CopyButton value={JSON.stringify(invocation, null, 2)} />
          </button>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <div className='mt-3 space-y-1'>
            <div className='flex items-center gap-1.5 text-xs font-medium text-muted-foreground'>
              <Code className='h-3 w-3' />
              Arguments
            </div>
            <pre className='text-xs font-mono bg-background rounded-md border p-2 overflow-x-auto'>
              {parsedArgs
                ? JSON.stringify(parsedArgs, null, 2)
                : invocation.function.arguments}
            </pre>
          </div>
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

// ============================================================================
// ToolCallView Component (Main Export)
// ============================================================================

/**
 * ToolCallView - Displays tool definitions and invocations
 *
 * Features:
 * - Shows tool definitions with parameters schema
 * - Shows which tools were called vs not called
 * - Displays tool call arguments
 * - Collapsible sections
 */
export function ToolCallView({ tools, toolCalls, className }: ToolCallViewProps) {
  // Get set of called tool names
  const calledToolNames = React.useMemo(() => {
    if (!toolCalls) return new Set<string>()
    return new Set(toolCalls.map((call) => call.function.name))
  }, [toolCalls])

  const hasTools = tools && tools.length > 0
  const hasInvocations = toolCalls && toolCalls.length > 0

  if (!hasTools && !hasInvocations) {
    return null
  }

  return (
    <div className={cn('space-y-4', className)}>
      {/* Tool Definitions */}
      {hasTools && (
        <div className='space-y-2'>
          <div className='flex items-center gap-2 text-sm font-medium'>
            <Wrench className='h-4 w-4 text-orange-500' />
            Tool Definitions ({tools.length})
          </div>
          <div className='space-y-2'>
            {tools.map((tool, idx) => {
              const name = 'function' in tool ? tool.function.name : tool.name
              const wasCalled = hasInvocations ? calledToolNames.has(name) : undefined
              return (
                <ToolDefinitionCard
                  key={`${name}-${idx}`}
                  tool={tool}
                  wasCalled={wasCalled}
                />
              )
            })}
          </div>
        </div>
      )}

      {/* Tool Invocations */}
      {hasInvocations && (
        <div className='space-y-2'>
          <div className='flex items-center gap-2 text-sm font-medium'>
            <CheckCircle className='h-4 w-4 text-green-500' />
            Tool Calls ({toolCalls.length})
          </div>
          <div className='space-y-2'>
            {toolCalls.map((invocation, idx) => (
              <ToolInvocationCard
                key={invocation.id || idx}
                invocation={invocation}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// ============================================================================
// Combined Tool Section for Input/Output
// ============================================================================

interface ToolSectionProps {
  inputContent: unknown
  outputContent: unknown
  className?: string
}

/**
 * ToolSection - Combines tool definitions from input and calls from output
 */
export function ToolSection({ inputContent, outputContent, className }: ToolSectionProps) {
  const tools = extractTools(inputContent)
  const toolCalls = extractToolCalls(outputContent)

  if (!tools && !toolCalls) {
    return null
  }

  return (
    <ToolCallView
      tools={tools || undefined}
      toolCalls={toolCalls || undefined}
      className={className}
    />
  )
}
