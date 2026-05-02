import {
  createContext,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import type { TraceListItem } from '../api/types'

// TracesContext carries the *current page* of traces down to the peek
// sheet's navigation hooks. Web's server-side pagination constrains
// prev/next navigation to the page the user is currently viewing —
// there's no cross-page prefetch because the API doesn't expose a
// "next after trace X" cursor today.
//
// Keeping this colocated (not lifted into a route loader) mirrors
// web/'s original shape. The peek sheet pulls `currentPageTraceIds`
// from here via `useDetailNavigation`.

interface TracesContextValue {
  currentPageTraceIds: string[]
  setCurrentPageTraceIds: (ids: string[]) => void
  currentPageTraces: TraceListItem[]
  setCurrentPageTraces: (traces: TraceListItem[]) => void
}

const TracesContext = createContext<TracesContextValue | null>(null)

interface TracesProviderProps {
  children: ReactNode
}

export function TracesProvider({ children }: TracesProviderProps) {
  const [currentPageTraceIds, setCurrentPageTraceIds] = useState<string[]>([])
  const [currentPageTraces, setCurrentPageTraces] = useState<TraceListItem[]>(
    [],
  )

  const value = useMemo(
    () => ({
      currentPageTraceIds,
      setCurrentPageTraceIds,
      currentPageTraces,
      setCurrentPageTraces,
    }),
    [currentPageTraceIds, currentPageTraces],
  )

  return (
    <TracesContext.Provider value={value}>{children}</TracesContext.Provider>
  )
}

export function useTraces(): TracesContextValue {
  const ctx = useContext(TracesContext)
  if (!ctx) {
    throw new Error('useTraces must be used within <TracesProvider>')
  }
  return ctx
}
