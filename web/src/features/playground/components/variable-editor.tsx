'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { RotateCcw } from 'lucide-react'

interface VariableEditorProps {
  variables: Record<string, string>
  onChange: (variables: Record<string, string>) => void
  extractedVariables?: string[]
}

export function VariableEditor({
  variables,
  onChange,
  extractedVariables = [],
}: VariableEditorProps) {
  const handleVariableChange = (name: string, value: string) => {
    onChange({ ...variables, [name]: value })
  }

  const handleClearAll = () => {
    const cleared = Object.keys(variables).reduce(
      (acc, key) => ({ ...acc, [key]: '' }),
      {}
    )
    onChange(cleared)
  }

  if (extractedVariables.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium">Variables</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground italic">
            No variables found in this prompt. Use <code className="text-xs bg-muted px-1 py-0.5 rounded">{'{{  }}'}</code> syntax to add variables.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
        <CardTitle className="text-sm font-medium">Variables</CardTitle>
        <Button variant="ghost" size="sm" onClick={handleClearAll}>
          <RotateCcw className="mr-2 h-3 w-3" />
          Clear All
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {extractedVariables.map((variable) => (
          <div key={variable} className="space-y-2">
            <Label htmlFor={variable} className="font-mono text-sm">
              {`{{${variable}}}`}
            </Label>
            <Textarea
              id={variable}
              value={variables[variable] || ''}
              onChange={(e) => handleVariableChange(variable, e.target.value)}
              placeholder={`Enter value for ${variable}...`}
              className="min-h-[60px] font-mono text-sm"
              rows={2}
            />
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

// Utility function to extract variables from messages (chat-only now)
export function extractVariablesFromMessages(messages: Array<{ content: string }>): string[] {
  const variableSet = new Set<string>()
  const regex = /\{\{([^}]+)\}\}/g

  for (const message of messages) {
    let match
    // Reset regex lastIndex for each message
    regex.lastIndex = 0
    while ((match = regex.exec(message.content)) !== null) {
      const varName = match[1].trim()
      if (varName) {
        variableSet.add(varName)
      }
    }
  }

  return Array.from(variableSet).sort()
}
