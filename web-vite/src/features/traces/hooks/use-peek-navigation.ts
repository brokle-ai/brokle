import { useCallback } from 'react'
import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import type { TraceTab } from '../types/peek'

// Peek-sheet mutations: open / close / expand-to-full-page / switch
// tab. Uses TanStack Router's typed `useNavigate` with `to: '.'` to
// stay on the current route while updating search params — the
// equivalent of web/'s `router.push(pathname + ?...)` pattern.
//
// `expandPeek` is the "maximize to full page" affordance — it
// navigates to `/o/$orgId/p/$projectId/traces/$traceId` using the
// route's typed params.

interface PeekSearchShape {
  peek?: string
  peekTab?: TraceTab
  [key: string]: unknown
}

interface TracesRouteParams {
  orgId?: string
  projectId?: string
}

export function usePeekNavigation() {
  const navigate = useNavigate()
  const params = useParams({ strict: false }) as TracesRouteParams
  const search = useSearch({ strict: false }) as PeekSearchShape

  const openPeek = useCallback(
    (traceId: string) => {
      // `replace: false` so the Back button dismisses the peek (web/
      // parity — router.push semantics).
      navigate({
        to: '.',
        search: (prev: Record<string, unknown>) => ({
          ...prev,
          peek: traceId,
          // Reset the tab when switching to a new trace so each peek
          // opens on the IO preview regardless of the previous tab.
          peekTab: undefined,
        }),
      })
    },
    [navigate],
  )

  const closePeek = useCallback(() => {
    navigate({
      to: '.',
      search: (prev: Record<string, unknown>) => {
        const next = { ...prev }
        delete next.peek
        delete next.peekTab
        return next
      },
    })
  }, [navigate])

  const expandPeek = useCallback(
    (newTab = false) => {
      const peekId = typeof search.peek === 'string' ? search.peek : null
      if (!peekId || !params.orgId || !params.projectId) return

      if (newTab) {
        // `useNavigate` can't open a new browser tab. The `$traceId`
        // route path is the canonical URL for full-page detail.
        const url = `/o/${encodeURIComponent(params.orgId)}/p/${encodeURIComponent(
          params.projectId,
        )}/traces/${encodeURIComponent(peekId)}`
        window.open(url, '_blank', 'noopener,noreferrer')
        return
      }
      navigate({
        to: '/o/$orgId/p/$projectId/traces/$traceId',
        params: {
          orgId: params.orgId,
          projectId: params.projectId,
          traceId: peekId,
        },
      })
    },
    [navigate, params.orgId, params.projectId, search.peek],
  )

  const setTab = useCallback(
    (tab: TraceTab) => {
      navigate({
        to: '.',
        replace: true,
        search: (prev: Record<string, unknown>) => ({
          ...prev,
          peekTab: tab,
        }),
      })
    },
    [navigate],
  )

  return { openPeek, closePeek, expandPeek, setTab }
}
