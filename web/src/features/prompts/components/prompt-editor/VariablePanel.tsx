'use client'

import { useCallback, useState } from 'react'
import { VariableIcon, PlusIcon, CopyIcon, CheckIcon, Loader2Icon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import type { TemplateDialect } from '@/features/prompts/types'

interface VariablePanelProps {
  variables: string[]
  isLoading?: boolean
  dialect?: TemplateDialect
  onInsertVariable?: (variable: string, syntax: string) => void
  className?: string
}

/**
 * VariablePanel - Displays extracted template variables.
 *
 * Features:
 * - Shows all extracted variables
 * - Click to copy variable syntax
 * - Click to insert into editor (if callback provided)
 * - Dialect-aware variable syntax
 */
export function VariablePanel({
  variables,
  isLoading = false,
  dialect = 'simple',
  onInsertVariable,
  className,
}: VariablePanelProps) {
  const [copiedVariable, setCopiedVariable] = useState<string | null>(null)

  const getVariableSyntax = useCallback(
    (variable: string): string => {
      switch (dialect) {
        case 'jinja2':
          return `{{ ${variable} }}`
        case 'mustache':
        case 'simple':
        default:
          return `{{${variable}}}`
      }
    },
    [dialect]
  )

  const handleCopy = useCallback(
    async (variable: string) => {
      const syntax = getVariableSyntax(variable)
      try {
        await navigator.clipboard.writeText(syntax)
        setCopiedVariable(variable)
        setTimeout(() => setCopiedVariable(null), 2000)
      } catch {
        // Fallback for older browsers
        console.error('Failed to copy to clipboard')
      }
    },
    [getVariableSyntax]
  )

  const handleInsert = useCallback(
    (variable: string) => {
      if (onInsertVariable) {
        const syntax = getVariableSyntax(variable)
        onInsertVariable(variable, syntax)
      }
    },
    [onInsertVariable, getVariableSyntax]
  )

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
        <VariableIcon className="size-4" />
        <span>Variables</span>
        {isLoading && <Loader2Icon className="size-3 animate-spin" />}
        {!isLoading && variables.length > 0 && (
          <Badge variant="secondary" className="text-xs px-1.5 py-0">
            {variables.length}
          </Badge>
        )}
      </div>

      {variables.length === 0 ? (
        <div className="text-xs text-muted-foreground py-2">
          {isLoading
            ? 'Detecting variables...'
            : 'No variables found. Use {{variable}} syntax to add variables.'}
        </div>
      ) : (
        <ScrollArea className="max-h-[200px]">
          <div className="flex flex-wrap gap-1.5">
            {variables.map((variable) => (
              <VariableChip
                key={variable}
                variable={variable}
                syntax={getVariableSyntax(variable)}
                isCopied={copiedVariable === variable}
                canInsert={!!onInsertVariable}
                onCopy={() => handleCopy(variable)}
                onInsert={() => handleInsert(variable)}
              />
            ))}
          </div>
        </ScrollArea>
      )}
    </div>
  )
}

interface VariableChipProps {
  variable: string
  syntax: string
  isCopied: boolean
  canInsert: boolean
  onCopy: () => void
  onInsert: () => void
}

function VariableChip({
  variable,
  syntax,
  isCopied,
  canInsert,
  onCopy,
  onInsert,
}: VariableChipProps) {
  return (
    <div className="group relative">
      <Badge
        variant="outline"
        className="cursor-pointer hover:bg-accent transition-colors pr-1.5"
        onClick={canInsert ? onInsert : onCopy}
        title={canInsert ? `Insert ${syntax}` : `Copy ${syntax}`}
      >
        <code className="text-xs font-mono">{variable}</code>
        <span className="ml-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {isCopied ? (
            <CheckIcon className="size-3 text-green-500" />
          ) : canInsert ? (
            <PlusIcon className="size-3" />
          ) : (
            <CopyIcon className="size-3" />
          )}
        </span>
      </Badge>
    </div>
  )
}

interface VariableInputPanelProps {
  variables: string[]
  values: Record<string, unknown>
  onChange: (values: Record<string, unknown>) => void
  className?: string
}

export function VariableInputPanel({
  variables,
  values,
  onChange,
  className,
}: VariableInputPanelProps) {
  const handleValueChange = useCallback(
    (variable: string, value: string) => {
      onChange({
        ...values,
        [variable]: value,
      })
    },
    [values, onChange]
  )

  if (variables.length === 0) {
    return null
  }

  return (
    <div className={cn('flex flex-col gap-3', className)}>
      <div className="text-sm font-medium text-muted-foreground">
        Sample Values for Preview
      </div>
      <div className="grid gap-2">
        {variables.map((variable) => (
          <div key={variable} className="flex items-center gap-2">
            <label
              htmlFor={`var-${variable}`}
              className="text-xs font-mono text-muted-foreground min-w-[100px] truncate"
              title={variable}
            >
              {variable}
            </label>
            <Input
              id={`var-${variable}`}
              type="text"
              placeholder={`Enter ${variable}...`}
              value={(values[variable] as string) || ''}
              onChange={(e) => handleValueChange(variable, e.target.value)}
              className="h-7 text-sm"
            />
          </div>
        ))}
      </div>
    </div>
  )
}

export function generateSampleValues(variables: string[]): Record<string, string> {
  const sampleValues: Record<string, string> = {}

  for (const variable of variables) {
    // Generate contextual sample values based on variable name
    const lowerVar = variable.toLowerCase()

    if (lowerVar.includes('name')) {
      sampleValues[variable] = 'John Doe'
    } else if (lowerVar.includes('email')) {
      sampleValues[variable] = 'john@example.com'
    } else if (lowerVar.includes('date') || lowerVar.includes('time')) {
      sampleValues[variable] = new Date().toLocaleDateString()
    } else if (lowerVar.includes('count') || lowerVar.includes('number')) {
      sampleValues[variable] = '42'
    } else if (lowerVar.includes('price') || lowerVar.includes('amount')) {
      sampleValues[variable] = '$99.99'
    } else if (lowerVar.includes('url') || lowerVar.includes('link')) {
      sampleValues[variable] = 'https://example.com'
    } else if (lowerVar.includes('message') || lowerVar.includes('content')) {
      sampleValues[variable] = 'Hello, this is a sample message.'
    } else if (lowerVar.includes('user')) {
      sampleValues[variable] = 'alice'
    } else if (lowerVar.includes('query') || lowerVar.includes('question')) {
      sampleValues[variable] = 'What is the meaning of life?'
    } else if (lowerVar.includes('response') || lowerVar.includes('answer')) {
      sampleValues[variable] = 'The answer is 42.'
    } else if (lowerVar.includes('system') || lowerVar.includes('persona')) {
      sampleValues[variable] = 'You are a helpful assistant.'
    } else if (lowerVar.includes('context') || lowerVar.includes('background')) {
      sampleValues[variable] = 'Background information goes here.'
    } else if (lowerVar.includes('list') || lowerVar.includes('items')) {
      sampleValues[variable] = 'item1, item2, item3'
    } else {
      // Default sample value
      sampleValues[variable] = `[${variable}]`
    }
  }

  return sampleValues
}
