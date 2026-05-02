// Peek-sheet types. The peek surface is intentionally thinner than
// the full detail route (`traces/$traceId`): it wraps the same span
// tree + detail panel + IO preview in a side sheet so users can triage
// without losing their place in the list.
//
// `PeekContext` is retained for parity with web/ even though the
// visible consumers today only need the binary "peek vs full"
// distinction — shared across future per-feature overrides.

export type PeekContext = 'peek' | 'full'

export type TraceTab = 'io' | 'tree' | 'metadata' | 'scores' | 'comments'

export const PEEK_TABS: readonly TraceTab[] = [
  'io',
  'tree',
  'metadata',
  'scores',
  'comments',
] as const

export interface PeekNavigationState {
  peekId: string | null
  selectedTab: TraceTab | null
  canNavigate: boolean
}

export interface DetailNavigationState {
  canGoPrev: boolean
  canGoNext: boolean
  position: number
  totalInPage: number
}
