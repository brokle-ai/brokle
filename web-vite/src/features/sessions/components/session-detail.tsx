import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { SessionDetail } from '../api/types'

interface SessionDetailHeaderProps {
  session: SessionDetail
}

// Duration arrives in nanoseconds (summed OTLP span durations). Match
// the formatter in `sessions-table.tsx` so totals render identically
// on list vs detail views.
function formatDurationNs(ns: number | undefined): string {
  if (ns === undefined || ns === null || ns === 0) return '—'
  const ms = ns / 1_000_000
  if (ms < 1_000) return `${ms.toFixed(1)}ms`
  const s = ms / 1_000
  if (s < 60) return `${s.toFixed(2)}s`
  const m = s / 60
  if (m < 60) return `${m.toFixed(1)}m`
  const h = m / 60
  return `${h.toFixed(1)}h`
}

function formatCost(cost: number | undefined): string {
  if (cost === undefined || cost === null) return '—'
  if (cost === 0) return '$0.00'
  if (cost < 0.01) return `$${cost.toFixed(6)}`
  return `$${cost.toFixed(4)}`
}

function formatTokens(tokens: number | undefined): string {
  if (tokens === undefined || tokens === null || tokens === 0) return '—'
  return tokens.toLocaleString()
}

function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}

export function SessionDetailHeader({ session }: SessionDetailHeaderProps) {
  // The backend aggregates user_ids across all traces in the session —
  // a single session may span multiple user identities when the same
  // session_id is attached to requests from different users. Render
  // them inline; fall back to an explicit "anonymous" pill when empty.
  const userIds = session.user_ids ?? []

  return (
    <Card>
      <CardHeader>
        <CardTitle className="font-mono text-base break-all">
          {session.session_id}
        </CardTitle>
        <div className="flex flex-wrap items-center gap-1.5 pt-1">
          {userIds.length === 0 ? (
            <Badge variant="outline" className="font-normal">
              No user identifier
            </Badge>
          ) : (
            userIds.map((uid) => (
              <Badge key={uid} variant="secondary" className="font-mono">
                {uid}
              </Badge>
            ))
          )}
        </div>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-4 text-sm md:grid-cols-3 lg:grid-cols-6">
        <Stat label="Traces" value={session.trace_count.toLocaleString()} />
        <Stat
          label="Errors"
          value={
            session.error_count > 0 ? (
              <span className="text-destructive">
                {session.error_count.toLocaleString()}
              </span>
            ) : (
              '0'
            )
          }
        />
        <Stat label="Tokens" value={formatTokens(session.total_tokens)} />
        <Stat label="Cost" value={formatCost(session.total_cost)} />
        <Stat label="Duration" value={formatDurationNs(session.total_duration)} />
        <Stat
          label="First seen"
          value={
            <span className="text-muted-foreground">
              {formatTimestamp(session.first_trace)}
            </span>
          }
        />
        <Stat
          label="Last seen"
          value={
            <span className="text-muted-foreground">
              {formatTimestamp(session.last_trace)}
            </span>
          }
        />
      </CardContent>
    </Card>
  )
}

function Stat({
  label,
  value,
}: {
  label: string
  value: React.ReactNode
}) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs uppercase tracking-wide text-muted-foreground">
        {label}
      </span>
      <span className="font-medium">{value}</span>
    </div>
  )
}
