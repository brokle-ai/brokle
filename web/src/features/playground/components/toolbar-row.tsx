'use client'

import { useState } from 'react'
import { Wrench, Variable, RotateCcw, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface ToolbarRowProps {
  variables: Record<string, string>
  extractedVariables: string[]
  onVariablesChange: (variables: Record<string, string>) => void
  onAddMessage: () => void
  disabled?: boolean
}

export function ToolbarRow({
  variables,
  extractedVariables,
  onVariablesChange,
  onAddMessage,
  disabled,
}: ToolbarRowProps) {
  const [variablesOpen, setVariablesOpen] = useState(false)

  const variableCount = extractedVariables.length

  const handleVariableChange = (name: string, value: string) => {
    onVariablesChange({ ...variables, [name]: value })
  }

  const handleClearAllVariables = () => {
    const cleared = Object.keys(variables).reduce(
      (acc, key) => ({ ...acc, [key]: '' }),
      {}
    )
    onVariablesChange(cleared)
  }

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <Button
        variant="outline"
        size="sm"
        onClick={onAddMessage}
        disabled={disabled}
      >
        <Plus className="mr-2 h-4 w-4" />
        Message
      </Button>

      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              disabled
              className="opacity-50 cursor-not-allowed"
            >
              <Wrench className="mr-2 h-4 w-4" />
              Tools
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Tool calling support coming soon</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <Popover open={variablesOpen} onOpenChange={setVariablesOpen}>
        <PopoverTrigger asChild>
          <Button variant="outline" size="sm" disabled={disabled}>
            <Variable className="mr-2 h-4 w-4" />
            Variables
            {variableCount > 0 && (
              <Badge variant="secondary" className="ml-2 h-5 px-1.5 text-xs">
                {variableCount}
              </Badge>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80" align="start">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h4 className="font-medium text-sm">Variables</h4>
              {variableCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleClearAllVariables}
                  className="h-7 text-xs"
                >
                  <RotateCcw className="mr-1 h-3 w-3" />
                  Clear
                </Button>
              )}
            </div>

            {variableCount === 0 ? (
              <p className="text-sm text-muted-foreground">
                No variables found. Use <code className="text-xs bg-muted px-1 py-0.5 rounded">{'{{name}}'}</code> syntax in messages.
              </p>
            ) : (
              <div className="space-y-3 max-h-[300px] overflow-y-auto">
                {extractedVariables.map((variable) => (
                  <div key={variable} className="space-y-1">
                    <Label htmlFor={`var-${variable}`} className="font-mono text-xs">
                      {`{{${variable}}}`}
                    </Label>
                    <Textarea
                      id={`var-${variable}`}
                      value={variables[variable] || ''}
                      onChange={(e) => handleVariableChange(variable, e.target.value)}
                      placeholder={`Value for ${variable}...`}
                      className="min-h-[50px] font-mono text-xs"
                      rows={2}
                    />
                  </div>
                ))}
              </div>
            )}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
