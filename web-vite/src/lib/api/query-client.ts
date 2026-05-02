import { QueryClient } from '@tanstack/react-query'
import { BrokleError, ServerError } from './errors'

// Query-wide defaults. Two invariants worth calling out:
//
// 1. `staleTime: 30_000` — cuts navigation-triggered refetches when
//    TanStack Router's `ensureQueryData` runs a route loader
//    (otherwise every navigation refires the queryFn; see #3997 +
//    plan's re-architecture §2).
//
// 2. `throwOnError` predicate — only unrecoverable errors (5xx or
//    `type:"api_error"`) bubble to React Error Boundaries; 4xx is
//    surfaced via the `error` field so forms can render inline
//    validation (`err.fieldIssues`). 401 is intercepted by the HTTP
//    middleware before React Query sees it.
export function createAppQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        refetchOnWindowFocus: false,
        retry: (count, err) => count < 1 && err instanceof ServerError,
        throwOnError: (err) => shouldBubbleToBoundary(err),
      },
      mutations: {
        // Surface mutation errors inline; caller decides via onError
        // whether to escalate. Never blanket-bubble mutations.
        throwOnError: false,
        retry: 0,
      },
    },
  })
}

function shouldBubbleToBoundary(err: unknown): boolean {
  if (err instanceof ServerError) return true
  if (err instanceof BrokleError) return err.body.error.type === 'api_error'
  return false
}
