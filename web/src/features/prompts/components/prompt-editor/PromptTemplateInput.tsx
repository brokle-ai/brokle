'use client'

import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import type { TextTemplate } from '../../types'

interface PromptTemplateInputProps {
  value: TextTemplate
  onChange: (value: TextTemplate) => void
  variables: string[]
  readOnly?: boolean
}

export function PromptTemplateInput({
  value,
  onChange,
  variables,
  readOnly,
}: PromptTemplateInputProps) {
  if (readOnly) {
    return (
      <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-sm">
        {value.content}
      </pre>
    )
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label>Template Content</Label>
        <Textarea
          value={value.content}
          onChange={(e) => onChange({ content: e.target.value })}
          placeholder="Enter your prompt template here. Use {{variable}} for interpolation..."
          className="min-h-[300px] font-mono text-sm"
        />
      </div>
      <div className="space-y-2">
        <Label>Detected Variables</Label>
        <div className="flex flex-wrap gap-1">
          {variables.length === 0 ? (
            <span className="text-sm text-muted-foreground italic">
              No variables detected
            </span>
          ) : (
            variables.map((v) => (
              <code
                key={v}
                className="px-2 py-1 rounded-md bg-muted text-xs font-mono"
              >
                {`{{${v}}}`}
              </code>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
