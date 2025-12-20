'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Copy, ExternalLink, Loader2, StopCircle } from 'lucide-react'
import { toast } from 'sonner'

import type { StreamMetrics } from '../types'

interface StreamingOutputProps {
  content: string
  isStreaming: boolean
  error?: string | null
  metrics?: StreamMetrics
  traceId?: string
  onStop?: () => void
}

export function StreamingOutput({
  content,
  isStreaming,
  error,
  metrics,
  traceId,
  onStop,
}: StreamingOutputProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(content)
    setCopied(true)
    toast.success('Copied to clipboard')
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Card className="flex flex-col h-full">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">Output</span>
          {isStreaming && (
            <Badge variant="secondary" className="animate-pulse">
              <Loader2 className="mr-1 h-3 w-3 animate-spin" />
              Streaming...
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          {isStreaming && onStop && (
            <Button variant="ghost" size="sm" onClick={onStop}>
              <StopCircle className="mr-2 h-4 w-4" />
              Stop
            </Button>
          )}
          {!isStreaming && content && (
            <Button variant="ghost" size="sm" onClick={handleCopy}>
              <Copy className="mr-2 h-4 w-4" />
              {copied ? 'Copied!' : 'Copy'}
            </Button>
          )}
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col gap-3">
        <div className="flex-1 border rounded-md p-4 bg-muted/30 overflow-auto">
          {error ? (
            <div className="text-sm text-destructive">{error}</div>
          ) : content ? (
            <div className="text-sm font-mono whitespace-pre-wrap prose prose-sm max-w-none dark:prose-invert">
              {content}
            </div>
          ) : (
            <div className="text-sm text-muted-foreground italic text-center py-8">
              No output yet. Click Execute to run the prompt.
            </div>
          )}
        </div>

        {!isStreaming && content && metrics && (
          <>
            <Separator />
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-4 text-muted-foreground flex-wrap">
                {metrics.model && (
                  <span>
                    Model: <strong>{metrics.model}</strong>
                  </span>
                )}
                {metrics.total_duration_ms !== undefined && (
                  <span>
                    Duration: <strong>{metrics.total_duration_ms}ms</strong>
                  </span>
                )}
                {metrics.ttft_ms !== undefined && (
                  <span>
                    TTFT: <strong>{metrics.ttft_ms.toFixed(0)}ms</strong>
                  </span>
                )}
                {metrics.total_tokens !== undefined && (
                  <span>
                    Tokens: <strong>{metrics.total_tokens}</strong>
                    {metrics.prompt_tokens !== undefined && metrics.completion_tokens !== undefined && (
                      <span className="text-muted-foreground/70">
                        {' '}({metrics.prompt_tokens}+{metrics.completion_tokens})
                      </span>
                    )}
                  </span>
                )}
                {metrics.cost !== undefined && (
                  <span>
                    Cost: <strong>${metrics.cost.toFixed(4)}</strong>
                  </span>
                )}
              </div>
              {traceId && (
                <Button variant="link" size="sm" className="h-auto p-0" asChild>
                  <a href={`/traces/${traceId}`} target="_blank" rel="noopener noreferrer">
                    View Trace
                    <ExternalLink className="ml-1 h-3 w-3" />
                  </a>
                </Button>
              )}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
