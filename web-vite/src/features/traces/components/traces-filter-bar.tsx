import { useState, useEffect } from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { TraceRange, TraceStatus } from '../api/queries'

export interface TracesFilterValue {
  q?: string
  status?: TraceStatus
  range: TraceRange
  model?: string
}

interface TracesFilterBarProps {
  value: TracesFilterValue
  onChange: (next: TracesFilterValue) => void
}

// Sentinel for "no filter" inside the <Select> component. Radix's Select
// rejects `value=""` — we map "any" <-> undefined at the boundary.
const STATUS_ANY = 'any'

export function TracesFilterBar({ value, onChange }: TracesFilterBarProps) {
  // Local draft for the free-text inputs so each keystroke doesn't
  // trigger a new loader run. Committed on Enter or blur.
  const [qDraft, setQDraft] = useState(value.q ?? '')
  const [modelDraft, setModelDraft] = useState(value.model ?? '')

  // Re-sync when the search URL is mutated externally (e.g. pagination
  // reset).
  useEffect(() => {
    setQDraft(value.q ?? '')
  }, [value.q])
  useEffect(() => {
    setModelDraft(value.model ?? '')
  }, [value.model])

  const commitQ = () => {
    const next = qDraft.trim()
    if ((value.q ?? '') === next) return
    onChange({ ...value, q: next.length > 0 ? next : undefined })
  }
  const commitModel = () => {
    const next = modelDraft.trim()
    if ((value.model ?? '') === next) return
    onChange({ ...value, model: next.length > 0 ? next : undefined })
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        placeholder="Search traces…"
        className="h-8 w-64"
        value={qDraft}
        onChange={(e) => setQDraft(e.target.value)}
        onBlur={commitQ}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            commitQ()
          }
        }}
      />

      <Select
        value={value.status ?? STATUS_ANY}
        onValueChange={(v) =>
          onChange({
            ...value,
            status: v === STATUS_ANY ? undefined : (v as TraceStatus),
          })
        }
      >
        <SelectTrigger className="h-8 w-32">
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={STATUS_ANY}>Any status</SelectItem>
          <SelectItem value="ok">OK</SelectItem>
          <SelectItem value="error">Error</SelectItem>
          <SelectItem value="unset">Unset</SelectItem>
        </SelectContent>
      </Select>

      <Select
        value={value.range}
        onValueChange={(v) => onChange({ ...value, range: v as TraceRange })}
      >
        <SelectTrigger className="h-8 w-32">
          <SelectValue placeholder="Time range" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="24h">Last 24 hours</SelectItem>
          <SelectItem value="7d">Last 7 days</SelectItem>
          <SelectItem value="30d">Last 30 days</SelectItem>
          <SelectItem value="all">All time</SelectItem>
        </SelectContent>
      </Select>

      <Input
        placeholder="Model…"
        className="h-8 w-40"
        value={modelDraft}
        onChange={(e) => setModelDraft(e.target.value)}
        onBlur={commitModel}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            commitModel()
          }
        }}
      />

      {(value.q || value.status || value.model || value.range !== 'all') && (
        <Button
          variant="ghost"
          size="sm"
          className="h-8"
          onClick={() => {
            setQDraft('')
            setModelDraft('')
            onChange({ range: 'all' })
          }}
        >
          Clear
        </Button>
      )}
    </div>
  )
}
