import { useEffect, useState } from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { ScoreDataType, ScoreSource } from '../api/types'

export interface ScoresFilterValue {
  name?: string
  source?: ScoreSource
  type?: ScoreDataType
  traceId?: string
  spanId?: string
}

interface ScoresFilterBarProps {
  value: ScoresFilterValue
  onChange: (next: ScoresFilterValue) => void
}

// Sentinel for "no filter" inside Radix's <Select>. Radix rejects
// `value=""`, so we map "any" <-> undefined at the boundary. Same
// trick as `traces-filter-bar.tsx`.
const ANY = 'any'

export function ScoresFilterBar({ value, onChange }: ScoresFilterBarProps) {
  // Drafts for free-text inputs so keystrokes don't re-trigger the
  // loader. Committed on Enter / blur.
  const [nameDraft, setNameDraft] = useState(value.name ?? '')
  const [traceDraft, setTraceDraft] = useState(value.traceId ?? '')
  const [spanDraft, setSpanDraft] = useState(value.spanId ?? '')

  useEffect(() => {
    setNameDraft(value.name ?? '')
  }, [value.name])
  useEffect(() => {
    setTraceDraft(value.traceId ?? '')
  }, [value.traceId])
  useEffect(() => {
    setSpanDraft(value.spanId ?? '')
  }, [value.spanId])

  const commitName = () => {
    const next = nameDraft.trim()
    if ((value.name ?? '') === next) return
    onChange({ ...value, name: next.length > 0 ? next : undefined })
  }
  const commitTraceId = () => {
    const next = traceDraft.trim()
    if ((value.traceId ?? '') === next) return
    onChange({ ...value, traceId: next.length > 0 ? next : undefined })
  }
  const commitSpanId = () => {
    const next = spanDraft.trim()
    if ((value.spanId ?? '') === next) return
    onChange({ ...value, spanId: next.length > 0 ? next : undefined })
  }

  const hasAny =
    value.name || value.source || value.type || value.traceId || value.spanId

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        placeholder="Name…"
        className="h-8 w-48"
        value={nameDraft}
        onChange={(e) => setNameDraft(e.target.value)}
        onBlur={commitName}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            commitName()
          }
        }}
      />

      <Select
        value={value.source ?? ANY}
        onValueChange={(v) =>
          onChange({
            ...value,
            source: v === ANY ? undefined : (v as ScoreSource),
          })
        }
      >
        <SelectTrigger className="h-8 w-36">
          <SelectValue placeholder="Source" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ANY}>Any source</SelectItem>
          <SelectItem value="code">Code</SelectItem>
          <SelectItem value="llm">LLM</SelectItem>
          <SelectItem value="human">Human</SelectItem>
        </SelectContent>
      </Select>

      <Select
        value={value.type ?? ANY}
        onValueChange={(v) =>
          onChange({
            ...value,
            type: v === ANY ? undefined : (v as ScoreDataType),
          })
        }
      >
        <SelectTrigger className="h-8 w-40">
          <SelectValue placeholder="Type" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ANY}>Any type</SelectItem>
          <SelectItem value="NUMERIC">Numeric</SelectItem>
          <SelectItem value="CATEGORICAL">Categorical</SelectItem>
          <SelectItem value="BOOLEAN">Boolean</SelectItem>
        </SelectContent>
      </Select>

      <Input
        placeholder="Trace ID…"
        className="h-8 w-56 font-mono text-xs"
        value={traceDraft}
        onChange={(e) => setTraceDraft(e.target.value)}
        onBlur={commitTraceId}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            commitTraceId()
          }
        }}
      />

      <Input
        placeholder="Span ID…"
        className="h-8 w-56 font-mono text-xs"
        value={spanDraft}
        onChange={(e) => setSpanDraft(e.target.value)}
        onBlur={commitSpanId}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            commitSpanId()
          }
        }}
      />

      {hasAny ? (
        <Button
          variant="ghost"
          size="sm"
          className="h-8"
          onClick={() => {
            setNameDraft('')
            setTraceDraft('')
            setSpanDraft('')
            onChange({})
          }}
        >
          Clear
        </Button>
      ) : null}
    </div>
  )
}
