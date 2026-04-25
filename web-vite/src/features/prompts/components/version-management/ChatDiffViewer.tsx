import { useMemo } from 'react'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { DiffViewer } from './DiffViewer'
import type { ChatMessage } from '../../types'

interface ChatDiffViewerProps {
  fromMessages: ChatMessage[]
  toMessages: ChatMessage[]
  oldLabel?: string
  newLabel?: string
  oldSubLabel?: string
  newSubLabel?: string
}

type ChangeKind = 'unchanged' | 'modified' | 'added' | 'removed'

interface AlignedRow {
  index: number
  kind: ChangeKind
  from: ChatMessage | null
  to: ChatMessage | null
}

function messageKey(m: ChatMessage): string {
  // Group by role for alignment; placeholders use their name.
  if (m.type === 'placeholder') return `placeholder:${m.name ?? ''}`
  return `msg:${m.role ?? ''}`
}

function messageContent(m: ChatMessage | null): string {
  if (!m) return ''
  return m.content ?? ''
}

function isUnchanged(a: ChatMessage, b: ChatMessage): boolean {
  return (
    a.type === b.type &&
    (a.role ?? '') === (b.role ?? '') &&
    (a.name ?? '') === (b.name ?? '') &&
    (a.content ?? '') === (b.content ?? '')
  )
}

/**
 * Greedy positional alignment: walks both lists in order and pairs
 * messages slot-by-slot. When the slot count differs the tail is
 * marked added/removed. This matches how engineers visually expect
 * a chat-template diff (system/user/assistant ordering matters).
 */
function alignMessages(
  fromMessages: ChatMessage[],
  toMessages: ChatMessage[],
): AlignedRow[] {
  const rows: AlignedRow[] = []
  const max = Math.max(fromMessages.length, toMessages.length)
  for (let i = 0; i < max; i++) {
    const from = fromMessages[i] ?? null
    const to = toMessages[i] ?? null
    if (from && to) {
      if (isUnchanged(from, to)) {
        rows.push({ index: i, kind: 'unchanged', from, to })
      } else if (messageKey(from) === messageKey(to)) {
        rows.push({ index: i, kind: 'modified', from, to })
      } else {
        // Role/type mismatch — treat as removed + added pair so the
        // visual makes the role flip obvious.
        rows.push({ index: i, kind: 'removed', from, to: null })
        rows.push({ index: i, kind: 'added', from: null, to })
      }
    } else if (from) {
      rows.push({ index: i, kind: 'removed', from, to: null })
    } else if (to) {
      rows.push({ index: i, kind: 'added', from: null, to })
    }
  }
  return rows
}

function roleLabel(m: ChatMessage | null): string {
  if (!m) return '—'
  if (m.type === 'placeholder') {
    return m.name ? `Placeholder · ${m.name}` : 'Placeholder'
  }
  return m.role ?? 'message'
}

function badgeStyles(kind: ChangeKind): string {
  switch (kind) {
    case 'added':
      return 'bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-200 border-green-300 dark:border-green-800'
    case 'removed':
      return 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-200 border-red-300 dark:border-red-800'
    case 'modified':
      return 'bg-amber-100 text-amber-900 dark:bg-amber-900/40 dark:text-amber-100 border-amber-300 dark:border-amber-800'
    default:
      return 'bg-muted text-muted-foreground'
  }
}

function kindLabel(kind: ChangeKind): string {
  switch (kind) {
    case 'added':
      return 'Added'
    case 'removed':
      return 'Removed'
    case 'modified':
      return 'Modified'
    default:
      return 'Unchanged'
  }
}

export function ChatDiffViewer({
  fromMessages,
  toMessages,
  oldLabel = 'Original Version',
  newLabel = 'New Version',
  oldSubLabel,
  newSubLabel,
}: ChatDiffViewerProps) {
  const rows = useMemo(
    () => alignMessages(fromMessages, toMessages),
    [fromMessages, toMessages],
  )

  if (rows.length === 0) {
    return (
      <div className="rounded-lg border p-6 text-center text-sm text-muted-foreground">
        No messages
      </div>
    )
  }

  const allUnchanged = rows.every((r) => r.kind === 'unchanged')
  if (allUnchanged) {
    return (
      <div className="text-sm text-muted-foreground py-2">No changes</div>
    )
  }

  return (
    <div className="space-y-3">
      {/* Top label bar mirroring DiffViewer's layout. */}
      <div className="rounded-lg border overflow-hidden">
        <div className="grid grid-cols-2 bg-muted/60">
          <div className="border-r border-b px-4 py-3">
            <span className="text-sm font-medium">{oldLabel}</span>
            {oldSubLabel && (
              <span className="ml-2 text-sm text-muted-foreground">
                {oldSubLabel}
              </span>
            )}
          </div>
          <div className="border-b px-4 py-3">
            <span className="text-sm font-medium">{newLabel}</span>
            {newSubLabel && (
              <span className="ml-2 text-sm text-muted-foreground">
                {newSubLabel}
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="space-y-3">
        {rows.map((row, idx) => (
          <div
            key={`${row.index}-${row.kind}-${idx}`}
            className="rounded-lg border overflow-hidden"
          >
            <div className="flex items-center justify-between gap-2 border-b bg-background px-3 py-2">
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono text-muted-foreground">
                  #{row.index + 1}
                </span>
                <Badge
                  variant="outline"
                  className={cn('text-xs', badgeStyles(row.kind))}
                >
                  {kindLabel(row.kind)}
                </Badge>
              </div>
            </div>
            {row.kind === 'unchanged' ? (
              <pre className="whitespace-pre-wrap break-words bg-muted/30 px-4 py-2.5 font-mono text-sm">
                {messageContent(row.from) || ' '}
              </pre>
            ) : (
              <DiffViewer
                oldString={messageContent(row.from)}
                newString={messageContent(row.to)}
                oldLabel={roleLabel(row.from)}
                newLabel={roleLabel(row.to)}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
