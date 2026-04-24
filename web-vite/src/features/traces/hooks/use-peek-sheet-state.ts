import { useSearch } from '@tanstack/react-router'
import type { TraceTab } from '../types/peek'
import { PEEK_TABS } from '../types/peek'

// Peek sheet URL-state adapter. Reads from the traces list route's
// search params via `useSearch({ strict: false })` so the hook can be
// used inside child components without threading the typed route ID.
//
// URL params (all optional, Zod-validated at the route):
// - peek: trace ID currently shown in the sheet
// - peekTab: which tab is active (io | tree | metadata | scores | comments)
//
// Writes are handled by `use-peek-navigation.ts` — splitting read vs
// write matches web/'s two-hook layout and keeps the call sites
// tight (a header only needs the setters; the content only needs the
// current values).

interface PeekSearchShape {
  peek?: string
  peekTab?: TraceTab
}

function isTraceTab(value: unknown): value is TraceTab {
  return (
    typeof value === 'string' &&
    (PEEK_TABS as readonly string[]).includes(value)
  )
}

export function usePeekSheetState(): {
  peekId: string | null
  selectedTab: TraceTab
} {
  const search = useSearch({ strict: false }) as PeekSearchShape
  const peekId = typeof search.peek === 'string' ? search.peek : null
  const selectedTab: TraceTab = isTraceTab(search.peekTab)
    ? search.peekTab
    : 'io'
  return { peekId, selectedTab }
}
