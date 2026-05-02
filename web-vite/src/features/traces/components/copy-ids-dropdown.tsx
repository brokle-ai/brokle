import * as React from 'react'
import { Check, ChevronDown, Copy, FileJson, Link as LinkIcon } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { TraceDetail } from '../api/types'

interface CopyIdsDropdownProps {
  trace: TraceDetail
  orgId: string
  projectId: string
  selectedSpanId?: string
}

type CopyState = 'idle' | 'trace' | 'span' | 'url' | 'json'

/**
 * CopyIdsDropdown — bulk-copy menu for the trace detail header.
 *
 * Surfaces:
 *  - Copy trace ID
 *  - Copy span ID (only when a span is selected)
 *  - Copy shareable trace URL (uses the same `/o/$orgId/p/$projectId`
 *    deep-link the router renders)
 *  - Copy a compact JSON snapshot for paste-into-bug-report use
 */
export function CopyIdsDropdown({
  trace,
  orgId,
  projectId,
  selectedSpanId,
}: CopyIdsDropdownProps) {
  const [copied, setCopied] = React.useState<CopyState>('idle')

  const handleCopy = async (value: string, type: CopyState, label: string) => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(type)
      toast.success(`${label} copied to clipboard`)
      window.setTimeout(() => setCopied('idle'), 2000)
    } catch {
      toast.error('Failed to copy to clipboard')
    }
  }

  const getTraceUrl = () => {
    const baseUrl =
      typeof window !== 'undefined' ? window.location.origin : ''
    return `${baseUrl}/o/${orgId}/p/${projectId}/traces/${trace.trace_id}`
  }

  const getTraceJson = () => {
    return JSON.stringify(
      {
        trace_id: trace.trace_id,
        name: trace.name,
        start_time: trace.start_time,
        end_time: trace.end_time,
        duration: trace.duration,
        status_code: trace.status_code,
        total_tokens: trace.total_tokens,
        total_cost: trace.total_cost,
        model_name: trace.model_name,
        provider_name: trace.provider_name,
      },
      null,
      2,
    )
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 gap-1 px-2">
          {copied !== 'idle' ? (
            <Check className="h-3.5 w-3.5 text-green-500" />
          ) : (
            <Copy className="h-3.5 w-3.5" />
          )}
          <ChevronDown className="h-3 w-3" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-48">
        <DropdownMenuItem
          onClick={() => handleCopy(trace.trace_id, 'trace', 'Trace ID')}
        >
          <Copy className="h-4 w-4" />
          <span>Copy Trace ID</span>
          {copied === 'trace' && (
            <Check className="ml-auto h-4 w-4 text-green-500" />
          )}
        </DropdownMenuItem>

        {selectedSpanId && (
          <DropdownMenuItem
            onClick={() => handleCopy(selectedSpanId, 'span', 'Span ID')}
          >
            <Copy className="h-4 w-4" />
            <span>Copy Span ID</span>
            {copied === 'span' && (
              <Check className="ml-auto h-4 w-4 text-green-500" />
            )}
          </DropdownMenuItem>
        )}

        <DropdownMenuSeparator />

        <DropdownMenuItem
          onClick={() => handleCopy(getTraceUrl(), 'url', 'Trace URL')}
        >
          <LinkIcon className="h-4 w-4" />
          <span>Copy Trace URL</span>
          {copied === 'url' && (
            <Check className="ml-auto h-4 w-4 text-green-500" />
          )}
        </DropdownMenuItem>

        <DropdownMenuItem
          onClick={() => handleCopy(getTraceJson(), 'json', 'Trace JSON')}
        >
          <FileJson className="h-4 w-4" />
          <span>Copy as JSON</span>
          {copied === 'json' && (
            <Check className="ml-auto h-4 w-4 text-green-500" />
          )}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
