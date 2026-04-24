import {
  queryOptions,
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  getPrompts,
  getPromptById,
  createPrompt as apiCreatePrompt,
  updatePrompt as apiUpdatePrompt,
  deletePrompt as apiDeletePrompt,
  getVersions,
  createVersion as apiCreateVersion,
  getVersion,
  setLabels as apiSetLabels,
  getVersionDiff,
  getProtectedLabels,
  setProtectedLabels as apiSetProtectedLabels,
} from './prompts-api'
import type {
  Prompt,
  PromptListItem,
  PromptListResponse,
  PromptVersion,
  CreatePromptRequest,
  UpdatePromptRequest,
  CreateVersionRequest,
  PromptType,
} from '../types'

// ---- Query keys --------------------------------------------------------

export interface PromptFilters {
  type?: PromptType
  tags?: string[]
  search?: string
  page?: number
  limit?: number
}

export const promptQueryKeys = {
  all: ['prompts'] as const,
  lists: () => [...promptQueryKeys.all, 'list'] as const,
  list: (projectId: string, filters?: PromptFilters) =>
    [...promptQueryKeys.lists(), projectId, filters] as const,
  details: () => [...promptQueryKeys.all, 'detail'] as const,
  detail: (projectId: string, promptId: string) =>
    [...promptQueryKeys.details(), projectId, promptId] as const,
  versions: (projectId: string, promptId: string) =>
    [...promptQueryKeys.detail(projectId, promptId), 'versions'] as const,
  version: (projectId: string, promptId: string, versionId: string) =>
    [...promptQueryKeys.versions(projectId, promptId), versionId] as const,
  diff: (projectId: string, promptId: string, from: number, to: number) =>
    [...promptQueryKeys.detail(projectId, promptId), 'diff', from, to] as const,
  protectedLabels: (projectId: string) =>
    [...promptQueryKeys.all, 'protected-labels', projectId] as const,
}

// Back-compat alias (route files still import `promptsKeys`).
export const promptsKeys = {
  all: promptQueryKeys.all,
  lists: promptQueryKeys.lists,
  list: (projectId: string, params: PromptListParams) =>
    promptQueryKeys.list(projectId, {
      page: params.page,
      limit: params.limit,
      search: params.q,
    }),
  details: promptQueryKeys.details,
  detail: promptQueryKeys.detail,
  versions: promptQueryKeys.versions,
}

export interface PromptListParams {
  page: number
  limit: number
  q?: string
}

// ---- Query option builders --------------------------------------------

export const promptListQueryOptions = (
  projectId: string,
  params: PromptListParams,
) =>
  queryOptions({
    queryKey: promptsKeys.list(projectId, params),
    queryFn: () =>
      getPrompts({
        projectId,
        page: params.page,
        limit: params.limit,
        search: params.q,
      }),
    staleTime: 30_000,
    placeholderData: keepPreviousData,
  })

export const promptDetailQueryOptions = (projectId: string, promptId: string) =>
  queryOptions({
    queryKey: promptQueryKeys.detail(projectId, promptId),
    queryFn: () => getPromptById(projectId, promptId),
    staleTime: 30_000,
  })

export const promptVersionsQueryOptions = (projectId: string, promptId: string) =>
  queryOptions({
    queryKey: promptQueryKeys.versions(projectId, promptId),
    queryFn: () => getVersions(projectId, promptId),
    staleTime: 30_000,
  })

// ---- Raw mutation fns (for route-level useMutation composition) --------

export async function createPrompt(
  projectId: string,
  data: CreatePromptRequest,
): Promise<Prompt> {
  return apiCreatePrompt(projectId, data)
}

export async function createPromptVersion(
  projectId: string,
  promptId: string,
  data: CreateVersionRequest,
): Promise<PromptVersion> {
  return apiCreateVersion(projectId, promptId, data)
}

// ---- Hooks -------------------------------------------------------------

export function usePromptsQuery(
  projectId: string | undefined,
  filters?: PromptFilters,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: promptQueryKeys.list(projectId || '', filters),
    queryFn: async () => {
      if (!projectId) throw new Error('Project ID is required')
      return getPrompts({ projectId, ...filters })
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
    placeholderData: keepPreviousData,
  })
}

export function usePromptQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: promptQueryKeys.detail(projectId || '', promptId || ''),
    queryFn: async () => {
      if (!projectId || !promptId) throw new Error('Missing IDs')
      return getPromptById(projectId, promptId)
    },
    enabled: !!projectId && !!promptId && (options.enabled ?? true),
    staleTime: 30_000,
  })
}

export function useVersionsQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: promptQueryKeys.versions(projectId || '', promptId || ''),
    queryFn: async () => {
      if (!projectId || !promptId) throw new Error('Missing IDs')
      return getVersions(projectId, promptId)
    },
    enabled: !!projectId && !!promptId && (options.enabled ?? true),
    staleTime: 30_000,
  })
}

export function useVersionQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  versionId: string | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: promptQueryKeys.version(
      projectId || '',
      promptId || '',
      versionId || '',
    ),
    queryFn: async () => {
      if (!projectId || !promptId || !versionId) throw new Error('Missing IDs')
      return getVersion(projectId, promptId, versionId)
    },
    enabled:
      !!projectId && !!promptId && !!versionId && (options.enabled ?? true),
    staleTime: 60_000,
  })
}

export function useVersionDiffQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  fromVersion: number | undefined,
  toVersion: number | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: promptQueryKeys.diff(
      projectId || '',
      promptId || '',
      fromVersion ?? 0,
      toVersion ?? 0,
    ),
    queryFn: async () => {
      if (
        !projectId ||
        !promptId ||
        fromVersion === undefined ||
        toVersion === undefined
      )
        throw new Error('Missing params')
      return getVersionDiff(projectId, promptId, fromVersion, toVersion)
    },
    enabled:
      !!projectId &&
      !!promptId &&
      fromVersion !== undefined &&
      toVersion !== undefined &&
      (options.enabled ?? true),
    staleTime: 5 * 60 * 1000,
  })
}

export function useProtectedLabelsQuery(
  projectId: string | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: promptQueryKeys.protectedLabels(projectId || ''),
    queryFn: async () => {
      if (!projectId) throw new Error('Missing project ID')
      return getProtectedLabels(projectId)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 60_000,
  })
}

export function useCreatePromptMutation(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreatePromptRequest) => apiCreatePrompt(projectId, data),
    onSuccess: (newPrompt) => {
      qc.invalidateQueries({ queryKey: promptQueryKeys.lists() })
      toast.success('Prompt Created', {
        description: `"${newPrompt.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const msg = error instanceof Error ? error.message : 'Unknown error'
      toast.error('Failed to Create Prompt', { description: msg })
    },
  })
}

export function useUpdatePromptMutation(projectId: string, promptId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: UpdatePromptRequest) =>
      apiUpdatePrompt(projectId, promptId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: promptQueryKeys.detail(projectId, promptId) })
      qc.invalidateQueries({ queryKey: promptQueryKeys.lists() })
      toast.success('Prompt Updated')
    },
    onError: (error: unknown) => {
      const msg = error instanceof Error ? error.message : 'Unknown error'
      toast.error('Failed to Update Prompt', { description: msg })
    },
  })
}

export function useDeletePromptMutation(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      promptId,
      promptName,
    }: {
      promptId: string
      promptName: string
    }) => {
      await apiDeletePrompt(projectId, promptId)
      return { promptId, promptName }
    },
    onMutate: async ({ promptId }) => {
      await qc.cancelQueries({ queryKey: promptQueryKeys.lists() })
      const previous = qc.getQueriesData<PromptListResponse>({
        queryKey: promptQueryKeys.lists(),
      })
      qc.setQueriesData<PromptListResponse>(
        { queryKey: promptQueryKeys.lists() },
        (old) => {
          if (!old) return old
          return {
            ...old,
            data: old.data.filter((p: PromptListItem) => p.id !== promptId),
            total: Math.max(0, old.total - 1),
          }
        },
      )
      return { previous }
    },
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: promptQueryKeys.lists() })
      toast.success('Prompt Deleted', {
        description: `"${variables.promptName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _v, context) => {
      if (context?.previous) {
        context.previous.forEach(([key, data]) => {
          qc.setQueryData(key, data)
        })
      }
      const msg = error instanceof Error ? error.message : 'Unknown error'
      toast.error('Failed to Delete Prompt', { description: msg })
    },
  })
}

export function useCreateVersionMutation(projectId: string, promptId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateVersionRequest) =>
      apiCreateVersion(projectId, promptId, data),
    onSuccess: (newVersion) => {
      qc.invalidateQueries({ queryKey: promptQueryKeys.versions(projectId, promptId) })
      qc.invalidateQueries({ queryKey: promptQueryKeys.detail(projectId, promptId) })
      qc.invalidateQueries({ queryKey: promptQueryKeys.lists() })
      toast.success('Version Created', {
        description: `Version ${newVersion.version} has been created.`,
      })
    },
    onError: (error: unknown) => {
      const msg = error instanceof Error ? error.message : 'Unknown error'
      toast.error('Failed to Create Version', { description: msg })
    },
  })
}

export function useSetLabelsMutation(projectId: string, promptId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      versionId,
      labels,
    }: {
      versionId: string
      labels: string[]
    }) => {
      await apiSetLabels(projectId, promptId, versionId, labels)
      return { versionId, labels }
    },
    onMutate: async ({ versionId, labels }) => {
      await qc.cancelQueries({
        queryKey: promptQueryKeys.versions(projectId, promptId),
      })
      const previous = qc.getQueryData<PromptVersion[]>(
        promptQueryKeys.versions(projectId, promptId),
      )
      qc.setQueryData<PromptVersion[]>(
        promptQueryKeys.versions(projectId, promptId),
        (old) => {
          if (!old) return old
          return old.map((v) => (v.id === versionId ? { ...v, labels } : v))
        },
      )
      return { previous }
    },
    onSuccess: (_data, { labels }) => {
      qc.invalidateQueries({ queryKey: promptQueryKeys.versions(projectId, promptId) })
      qc.invalidateQueries({ queryKey: promptQueryKeys.detail(projectId, promptId) })
      qc.invalidateQueries({ queryKey: promptQueryKeys.lists() })
      toast.success('Labels Updated', {
        description:
          labels.length > 0
            ? `Labels set: ${labels.join(', ')}`
            : 'Labels cleared.',
      })
    },
    onError: (error: unknown, _v, context) => {
      if (context?.previous) {
        qc.setQueryData(
          promptQueryKeys.versions(projectId, promptId),
          context.previous,
        )
      }
      const msg = error instanceof Error ? error.message : 'Unknown error'
      toast.error('Failed to Update Labels', { description: msg })
    },
  })
}

export function useSetProtectedLabelsMutation(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (labels: string[]) => {
      await apiSetProtectedLabels(projectId, labels)
      return labels
    },
    onSuccess: (labels) => {
      qc.invalidateQueries({ queryKey: promptQueryKeys.protectedLabels(projectId) })
      toast.success('Protected Labels Updated', {
        description:
          labels.length > 0
            ? `Protected labels: ${labels.join(', ')}`
            : 'Protected labels cleared.',
      })
    },
    onError: (error: unknown) => {
      const msg = error instanceof Error ? error.message : 'Unknown error'
      toast.error('Failed to Update Protected Labels', { description: msg })
    },
  })
}
