// Thin queryOptions shims for call sites outside the datasets feature
// (currently the experiments form dataset-picker). The datasets
// feature itself uses the full `datasetsApi` surface in
// `./datasets-api.ts` — anything new inside datasets should go there.

import { queryOptions } from '@tanstack/react-query'
import { datasetsApi } from './datasets-api'

export interface DatasetListParams {
  page: number
  limit: number
  q?: string
}

export const datasetListQueryOptions = (
  projectId: string,
  params: DatasetListParams,
) =>
  queryOptions({
    queryKey: ['datasets', 'list', projectId, params] as const,
    queryFn: () =>
      datasetsApi.listDatasets(projectId, {
        page: params.page,
        limit: params.limit,
        search: params.q,
      }),
    staleTime: 30 * 1000,
  })
