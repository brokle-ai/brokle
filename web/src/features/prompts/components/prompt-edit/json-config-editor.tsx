'use client'

import * as React from 'react'
import { AlertCircle, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'

interface JsonConfigEditorProps {
  config: Record<string, unknown> | null
  onChange: (config: Record<string, unknown> | null, isValid: boolean) => void
  className?: string
}

export function JsonConfigEditor({
  config,
  onChange,
  className,
}: JsonConfigEditorProps) {
  const [jsonText, setJsonText] = React.useState(() =>
    config ? JSON.stringify(config, null, 2) : '{}'
  )
  const [error, setError] = React.useState<string | null>(null)
  const [isValid, setIsValid] = React.useState(true)

  // Sync jsonText when config prop changes (e.g., version switch)
  React.useEffect(() => {
    setJsonText(config ? JSON.stringify(config, null, 2) : '{}')
    setError(null)
    setIsValid(true)
  }, [config])

  const handleChange = (value: string) => {
    setJsonText(value)

    if (!value.trim() || value.trim() === '{}') {
      setError(null)
      setIsValid(true)
      onChange(null, true)
      return
    }

    try {
      const parsed = JSON.parse(value)
      if (typeof parsed !== 'object' || Array.isArray(parsed) || parsed === null) {
        throw new Error('Config must be a JSON object')
      }
      setError(null)
      setIsValid(true)
      onChange(parsed as Record<string, unknown>, true)
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Invalid JSON'
      setError(message)
      setIsValid(false)
      onChange(null, false)
    }
  }

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between">
        <Label className="text-sm font-medium">Config</Label>
        {jsonText.trim() && jsonText.trim() !== '{}' && (
          <span
            className={cn(
              'flex items-center gap-1 text-xs',
              isValid ? 'text-green-600' : 'text-destructive'
            )}
          >
            {isValid ? (
              <>
                <Check className="h-3 w-3" />
                Valid JSON
              </>
            ) : (
              <>
                <AlertCircle className="h-3 w-3" />
                Invalid JSON
              </>
            )}
          </span>
        )}
      </div>

      <p className="text-xs text-muted-foreground">
        Arbitrary JSON configuration that is available on the prompt. Use this to
        track LLM parameters, function definitions, or any other metadata.
      </p>

      <Textarea
        value={jsonText}
        onChange={(e) => handleChange(e.target.value)}
        placeholder="{}"
        className={cn(
          'font-mono text-sm min-h-[120px] resize-y',
          error && 'border-destructive focus-visible:ring-destructive'
        )}
      />

      {error && (
        <p className="text-xs text-destructive flex items-center gap-1">
          <AlertCircle className="h-3 w-3" />
          {error}
        </p>
      )}
    </div>
  )
}
