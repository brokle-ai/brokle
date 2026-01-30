'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { datasetsApi } from '../api/datasets-api'
import type { PaginatedResponse } from '@/lib/api/core/types'
import type {
  CreateDatasetRequest,
  UpdateDatasetRequest,
  CreateDatasetItemRequest,
  CreateDatasetVersionRequest,
  PinDatasetVersionRequest,
  Dataset,
  DatasetItem,
  ImportFromJsonRequest,
  ImportFromTracesRequest,
  ImportFromSpansRequest,
  ImportFromCsvRequest,
  DatasetListParams,
  DatasetItemListParams,
} from '../types'

export const datasetQueryKeys = {
  all: ['datasets'] as const,
  lists: () => [...datasetQueryKeys.all, 'list'] as const,
  list: (projectId: string) => [...datasetQueryKeys.lists(), projectId] as const,
  listFiltered: (projectId: string, filters?: DatasetListParams) =>
    [...datasetQueryKeys.list(projectId), filters] as const,
  details: () => [...datasetQueryKeys.all, 'detail'] as const,
  detail: (projectId: string, datasetId: string) =>
    [...datasetQueryKeys.details(), projectId, datasetId] as const,
  items: (projectId: string, datasetId: string) =>
    [...datasetQueryKeys.all, 'items', projectId, datasetId] as const,
  // Version-related query keys
  versionInfo: (projectId: string, datasetId: string) =>
    [...datasetQueryKeys.all, 'versionInfo', projectId, datasetId] as const,
  versions: (projectId: string, datasetId: string) =>
    [...datasetQueryKeys.all, 'versions', projectId, datasetId] as const,
  versionDetail: (projectId: string, datasetId: string, versionId: string) =>
    [...datasetQueryKeys.all, 'version', projectId, datasetId, versionId] as const,
  versionItems: (projectId: string, datasetId: string, versionId: string) =>
    [...datasetQueryKeys.all, 'versionItems', projectId, datasetId, versionId] as const,
}

export function useDatasetsQuery(
  projectId: string | undefined,
  params?: DatasetListParams
) {
  return useQuery({
    queryKey: [...datasetQueryKeys.list(projectId ?? ''), params?.search, params?.page, params?.limit],
    queryFn: () => datasetsApi.listDatasets(projectId!, params),
    enabled: !!projectId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Query hook to list datasets with pagination, search, and sorting
 */
export function useDatasetsWithPaginationQuery(
  projectId: string | undefined,
  params?: DatasetListParams,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: datasetQueryKeys.listFiltered(projectId ?? '', params),
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return datasetsApi.listDatasetsWithPagination(projectId, params)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useDatasetQuery(
  projectId: string | undefined,
  datasetId: string | undefined
) {
  return useQuery({
    queryKey: datasetQueryKeys.detail(projectId ?? '', datasetId ?? ''),
    queryFn: () => datasetsApi.getDataset(projectId!, datasetId!),
    enabled: !!projectId && !!datasetId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useDatasetItemsQuery(
  projectId: string | undefined,
  datasetId: string | undefined,
  params?: DatasetItemListParams
) {
  return useQuery({
    queryKey: [
      ...datasetQueryKeys.items(projectId ?? '', datasetId ?? ''),
      params?.page,
      params?.limit,
    ],
    queryFn: () => datasetsApi.listDatasetItems(projectId!, datasetId!, params),
    enabled: !!projectId && !!datasetId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useCreateDatasetMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateDatasetRequest) =>
      datasetsApi.createDataset(projectId, data),
    onSuccess: (newDataset) => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.list(projectId),
      })
      toast.success('Dataset Created', {
        description: `"${newDataset.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Dataset', {
        description: apiError?.message || 'Could not create dataset. Please try again.',
      })
    },
  })
}

export function useUpdateDatasetMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateDatasetRequest) =>
      datasetsApi.updateDataset(projectId, datasetId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.all,
      })
      toast.success('Dataset Updated', {
        description: 'Dataset has been updated.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Dataset', {
        description: apiError?.message || 'Could not update dataset. Please try again.',
      })
    },
  })
}

export function useDeleteDatasetMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ datasetId, datasetName }: { datasetId: string; datasetName: string }) => {
      await datasetsApi.deleteDataset(projectId, datasetId)
      return { datasetId, datasetName }
    },
    onMutate: async ({ datasetId }) => {
      await queryClient.cancelQueries({
        queryKey: datasetQueryKeys.list(projectId),
      })

      // Get ALL matching queries (prefix match for paginated queries)
      const previousQueries = queryClient.getQueriesData<PaginatedResponse<Dataset>>({
        queryKey: datasetQueryKeys.list(projectId),
      })

      // Optimistic update - update ALL matching queries
      queryClient.setQueriesData<PaginatedResponse<Dataset>>(
        { queryKey: datasetQueryKeys.list(projectId) },
        (old) => old ? {
          data: old.data.filter((d) => d.id !== datasetId),
          pagination: {
            ...old.pagination,
            total: old.pagination.total - 1,
          },
        } : old
      )

      return { previousQueries }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.list(projectId),
      })
      toast.success('Dataset Deleted', {
        description: `"${variables.datasetName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      // Rollback ALL affected queries
      context?.previousQueries?.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data)
      })
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Dataset', {
        description: apiError?.message || 'Could not delete dataset. Please try again.',
      })
    },
  })
}

export function useCreateDatasetItemMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateDatasetItemRequest) =>
      datasetsApi.createDatasetItem(projectId, datasetId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      toast.success('Item Added', {
        description: 'Dataset item has been added successfully.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Add Item', {
        description: apiError?.message || 'Could not add item. Please try again.',
      })
    },
  })
}

export function useDeleteDatasetItemMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (itemId: string) => {
      await datasetsApi.deleteDatasetItem(projectId, datasetId, itemId)
      return itemId
    },
    onMutate: async (itemId) => {
      await queryClient.cancelQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })

      // Get ALL matching queries (prefix match for paginated queries)
      const previousQueries = queryClient.getQueriesData<PaginatedResponse<DatasetItem>>({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })

      // Optimistic update - update ALL matching queries
      queryClient.setQueriesData<PaginatedResponse<DatasetItem>>(
        { queryKey: datasetQueryKeys.items(projectId, datasetId) },
        (old) => old ? {
          data: old.data.filter((i) => i.id !== itemId),
          pagination: {
            ...old.pagination,
            total: old.pagination.total - 1,
          },
        } : old
      )

      return { previousQueries }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      toast.success('Item Deleted', {
        description: 'Dataset item has been deleted.',
      })
    },
    onError: (error: unknown, _itemId, context) => {
      // Rollback ALL affected queries
      context?.previousQueries?.forEach(([queryKey, data]) => {
        queryClient.setQueryData(queryKey, data)
      })
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Item', {
        description: apiError?.message || 'Could not delete item. Please try again.',
      })
    },
  })
}

// Import Mutations
export function useImportFromJsonMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: ImportFromJsonRequest) =>
      datasetsApi.importFromJson(projectId, datasetId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      toast.success('Import Complete', {
        description: `Created ${result.created} items, skipped ${result.skipped} duplicates.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Import Failed', {
        description: apiError?.message || 'Could not import items. Please try again.',
      })
    },
  })
}

export function useImportFromTracesMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: ImportFromTracesRequest) =>
      datasetsApi.importFromTraces(projectId, datasetId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      toast.success('Import Complete', {
        description: `Created ${result.created} items from traces, skipped ${result.skipped} duplicates.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Import Failed', {
        description: apiError?.message || 'Could not import from traces. Please try again.',
      })
    },
  })
}

export function useImportFromSpansMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: ImportFromSpansRequest) =>
      datasetsApi.importFromSpans(projectId, datasetId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      toast.success('Import Complete', {
        description: `Created ${result.created} items from spans, skipped ${result.skipped} duplicates.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Import Failed', {
        description: apiError?.message || 'Could not import from spans. Please try again.',
      })
    },
  })
}

export function useImportFromCsvMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: ImportFromCsvRequest) =>
      datasetsApi.importFromCsv(projectId, datasetId, data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.items(projectId, datasetId),
      })
      toast.success('CSV Import Complete', {
        description: `Created ${result.created} items, skipped ${result.skipped} duplicates.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('CSV Import Failed', {
        description: apiError?.message || 'Could not import from CSV. Please try again.',
      })
    },
  })
}

// Export Query
export function useExportDatasetQuery(
  projectId: string | undefined,
  datasetId: string | undefined,
  enabled = false
) {
  return useQuery({
    queryKey: [...datasetQueryKeys.items(projectId ?? '', datasetId ?? ''), 'export'],
    queryFn: () => datasetsApi.exportDataset(projectId!, datasetId!),
    enabled: enabled && !!projectId && !!datasetId,
    staleTime: 0,
    gcTime: 0,
  })
}

// ============================================================================
// Dataset Versioning Hooks
// ============================================================================

export function useDatasetWithVersionInfoQuery(
  projectId: string | undefined,
  datasetId: string | undefined
) {
  return useQuery({
    queryKey: datasetQueryKeys.versionInfo(projectId ?? '', datasetId ?? ''),
    queryFn: () => datasetsApi.getDatasetWithVersionInfo(projectId!, datasetId!),
    enabled: !!projectId && !!datasetId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useDatasetVersionsQuery(
  projectId: string | undefined,
  datasetId: string | undefined
) {
  return useQuery({
    queryKey: datasetQueryKeys.versions(projectId ?? '', datasetId ?? ''),
    queryFn: () => datasetsApi.listVersions(projectId!, datasetId!),
    enabled: !!projectId && !!datasetId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useDatasetVersionQuery(
  projectId: string | undefined,
  datasetId: string | undefined,
  versionId: string | undefined
) {
  return useQuery({
    queryKey: datasetQueryKeys.versionDetail(projectId ?? '', datasetId ?? '', versionId ?? ''),
    queryFn: () => datasetsApi.getVersion(projectId!, datasetId!, versionId!),
    enabled: !!projectId && !!datasetId && !!versionId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useDatasetVersionItemsQuery(
  projectId: string | undefined,
  datasetId: string | undefined,
  versionId: string | undefined,
  params?: DatasetItemListParams
) {
  return useQuery({
    queryKey: [
      ...datasetQueryKeys.versionItems(projectId ?? '', datasetId ?? '', versionId ?? ''),
      params?.page,
      params?.limit,
    ],
    queryFn: () => datasetsApi.getVersionItems(projectId!, datasetId!, versionId!, params),
    enabled: !!projectId && !!datasetId && !!versionId,
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

export function useCreateDatasetVersionMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data?: CreateDatasetVersionRequest) =>
      datasetsApi.createVersion(projectId, datasetId, data),
    onSuccess: (newVersion) => {
      // Invalidate versions list and version info
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.versions(projectId, datasetId),
      })
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.versionInfo(projectId, datasetId),
      })
      toast.success('Version Created', {
        description: `Version ${newVersion.version} has been created with ${newVersion.item_count} items.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Version', {
        description: apiError?.message || 'Could not create version. Please try again.',
      })
    },
  })
}

export function usePinDatasetVersionMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: PinDatasetVersionRequest) =>
      datasetsApi.pinVersion(projectId, datasetId, data),
    onSuccess: (dataset, variables) => {
      // Invalidate version info and dataset detail
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.versionInfo(projectId, datasetId),
      })
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.detail(projectId, datasetId),
      })

      if (variables.version_id) {
        toast.success('Version Pinned', {
          description: 'Dataset is now pinned to the selected version.',
        })
      } else {
        toast.success('Version Unpinned', {
          description: 'Dataset will now use the latest items.',
        })
      }
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Version Pin', {
        description: apiError?.message || 'Could not update version pin. Please try again.',
      })
    },
  })
}

export function useUnpinDatasetVersionMutation(projectId: string, datasetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => datasetsApi.unpinVersion(projectId, datasetId),
    onSuccess: () => {
      // Invalidate version info and dataset detail
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.versionInfo(projectId, datasetId),
      })
      queryClient.invalidateQueries({
        queryKey: datasetQueryKeys.detail(projectId, datasetId),
      })
      toast.success('Version Unpinned', {
        description: 'Dataset will now use the latest items.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Unpin Version', {
        description: apiError?.message || 'Could not unpin version. Please try again.',
      })
    },
  })
}
