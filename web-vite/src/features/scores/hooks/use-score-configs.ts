import { useQuery } from '@tanstack/react-query'
import { scoreConfigsQueryOptions } from '../api/queries'

/**
 * Read hook for the project score-config catalog.
 *
 * Thin wrapper over `scoreConfigsQueryOptions` so callers in the traces
 * feature (and anywhere else) can drop a projectId and consume the
 * query without knowing the query-key/staleTime layout.
 */
export function useScoreConfigsQuery(projectId: string | undefined) {
  return useQuery({
    ...scoreConfigsQueryOptions(projectId ?? ''),
    enabled: !!projectId,
  })
}
