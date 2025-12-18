'use client'

import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface VariableInputsProps {
  variables: string[]
  values: Record<string, string>
  onChange: (values: Record<string, string>) => void
}

export function VariableInputs({ variables, values, onChange }: VariableInputsProps) {
  if (variables.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Variables</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground italic">
            No variables required for this prompt
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Variable Inputs</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {variables.map((variable) => (
          <div key={variable} className="space-y-2">
            <Label htmlFor={variable} className="font-mono text-sm">
              {`{{${variable}}}`}
            </Label>
            <Textarea
              id={variable}
              value={values[variable] || ''}
              onChange={(e) =>
                onChange({ ...values, [variable]: e.target.value })
              }
              placeholder={`Enter value for ${variable}...`}
              className="min-h-[80px]"
            />
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
