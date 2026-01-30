'use client'

import * as React from 'react'
import { Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'

interface JsonConfigViewerProps {
  config: Record<string, unknown> | null
  className?: string
}

export function JsonConfigViewer({ config, className }: JsonConfigViewerProps) {
  const [copied, setCopied] = React.useState(false)

  const jsonText = config ? JSON.stringify(config, null, 2) : '{}'
  const isEmpty = !config || Object.keys(config).length === 0

  const handleCopy = async () => {
    await navigator.clipboard.writeText(jsonText)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className={cn('space-y-2', className)}>
      <div className="flex items-center justify-between">
        <Label className="text-sm font-medium">Config</Label>
        {!isEmpty && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 px-2 text-xs"
            onClick={handleCopy}
          >
            {copied ? (
              <>
                <Check className="h-3 w-3 mr-1 text-green-500" />
                Copied
              </>
            ) : (
              <>
                <Copy className="h-3 w-3 mr-1" />
                Copy
              </>
            )}
          </Button>
        )}
      </div>

      <p className="text-xs text-muted-foreground">
        Arbitrary JSON configuration available on the prompt.
      </p>

      <pre
        className={cn(
          'rounded-lg bg-muted p-4 font-mono text-sm overflow-x-auto whitespace-pre',
          isEmpty && 'text-muted-foreground'
        )}
      >
        {isEmpty ? 'No configuration' : jsonText}
      </pre>
    </div>
  )
}
