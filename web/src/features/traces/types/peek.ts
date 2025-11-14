export type PeekContext = 'peek' | 'full'

export type TraceTab = 'details' | 'spans' | 'metadata'

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
