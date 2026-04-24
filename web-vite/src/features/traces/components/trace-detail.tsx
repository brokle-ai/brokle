import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Span, TraceDetail as TraceDetailType } from '../api/types'
import { SpanTree } from './span-tree'
import { IoPreview } from './io-preview'

interface TraceDetailProps {
  trace: TraceDetailType
  spans: Span[]
}

function formatDuration(ns: number | undefined): string {
  if (ns === undefined || ns === null) return '—'
  if (ns < 1_000) return `${ns}ns`
  const us = ns / 1_000
  if (us < 1_000) return `${us.toFixed(1)}µs`
  const ms = us / 1_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  return `${(ms / 1_000).toFixed(2)}s`
}

function formatCost(costStr: string | undefined): string {
  if (!costStr) return '—'
  const n = Number(costStr)
  if (!Number.isFinite(n)) return '—'
  if (n === 0) return '$0.00'
  if (n < 0.01) return `$${n.toFixed(6)}`
  return `$${n.toFixed(4)}`
}

function formatTimestamp(iso: string | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

function formatTokens(tokens: number | undefined): string {
  if (tokens === undefined || tokens === null) return '—'
  return tokens.toLocaleString()
}

function StatusBadge({ trace }: { trace: TraceDetailType }) {
  if (trace.has_error || trace.status_code === 2) {
    return <Badge variant="destructive">Error</Badge>
  }
  if (trace.status_code === 1) {
    return <Badge variant="secondary">OK</Badge>
  }
  return <Badge variant="outline">Unset</Badge>
}

function CopyIdButton({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1200)
    } catch {
      // clipboard permission denied — silent, same as Langfuse UX
    }
  }
  return (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      className="h-7 gap-1 px-2 font-mono text-xs"
      onClick={copy}
      aria-label="Copy trace ID"
    >
      <span className="truncate">{value}</span>
      {copied ? (
        <Check className="h-3.5 w-3.5 text-green-600" />
      ) : (
        <Copy className="h-3.5 w-3.5" />
      )}
    </Button>
  )
}

function MetaPair({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="text-sm">{value}</span>
    </div>
  )
}

export function TraceDetail({ trace, spans }: TraceDetailProps) {
  const [selectedSpanId, setSelectedSpanId] = useState<string | undefined>(
    () => spans.find((s) => !s.parent_span_id)?.span_id ?? spans[0]?.span_id,
  )
  const [spanTreeOpen, setSpanTreeOpen] = useState(true)

  const selectedSpan = spans.find((s) => s.span_id === selectedSpanId)

  return (
    <div className="space-y-6">
      <section className="space-y-3">
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-semibold">{trace.name}</h1>
          <StatusBadge trace={trace} />
        </div>

        <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
          <MetaPair label="Duration" value={formatDuration(trace.duration)} />
          <MetaPair label="Tokens" value={formatTokens(trace.total_tokens)} />
          <MetaPair label="Cost" value={formatCost(trace.total_cost)} />
          <MetaPair label="Spans" value={trace.span_count.toLocaleString()} />
          <MetaPair label="Start" value={formatTimestamp(trace.start_time)} />
          <MetaPair
            label="Trace ID"
            value={<CopyIdButton value={trace.trace_id} />}
          />
        </div>
      </section>

      <Collapsible open={spanTreeOpen} onOpenChange={setSpanTreeOpen}>
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium">Spans ({spans.length})</h2>
          <CollapsibleTrigger asChild>
            <Button variant="ghost" size="sm" className="h-7 gap-1">
              <ChevronDown
                className={cn(
                  'h-4 w-4 transition-transform',
                  !spanTreeOpen && '-rotate-90',
                )}
              />
              {spanTreeOpen ? 'Hide' : 'Show'}
            </Button>
          </CollapsibleTrigger>
        </div>
        <CollapsibleContent className="mt-2">
          <SpanTree
            spans={spans}
            selectedSpanId={selectedSpanId}
            onSpanSelect={(s) => setSelectedSpanId(s.span_id)}
          />
        </CollapsibleContent>
      </Collapsible>

      {selectedSpan && (
        <section className="space-y-4">
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
            <MetaPair
              label="Model"
              value={selectedSpan.model_name ?? '—'}
            />
            <MetaPair
              label="Provider"
              value={selectedSpan.provider_name ?? '—'}
            />
            <MetaPair
              label="Latency"
              value={formatDuration(selectedSpan.duration)}
            />
            <MetaPair
              label="Tokens"
              value={formatTokens(
                sumTokens(selectedSpan.usage_details),
              )}
            />
            <MetaPair
              label="Cost"
              value={formatCost(selectedSpan.total_cost)}
            />
          </div>

          <IoPreview value={selectedSpan.input} label="Input" />
          <IoPreview value={selectedSpan.output} label="Output" />
        </section>
      )}
    </div>
  )
}

// Small helper: collapse usage_details (a map of token-type -> count)
// into a single total. Undefined if the map is absent or empty.
function sumTokens(
  usage: Record<string, number> | undefined,
): number | undefined {
  if (!usage) return undefined
  let total = 0
  let sawAny = false
  for (const v of Object.values(usage)) {
    if (typeof v === 'number' && Number.isFinite(v)) {
      total += v
      sawAny = true
    }
  }
  return sawAny ? total : undefined
}
