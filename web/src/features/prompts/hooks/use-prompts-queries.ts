'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  getPrompts,
  getPromptById,
  createPrompt,
  updatePrompt,
  deletePrompt,
  getVersions,
  createVersion,
  getVersion,
  setLabels,
  getVersionDiff,
  getProtectedLabels,
  setProtectedLabels,
} from '../api/prompts-api'
import type {
  PromptListItem,
  PromptVersion,
  CreatePromptRequest,
  UpdatePromptRequest,
  CreateVersionRequest,
  PromptType,
} from '../types'

export const promptQueryKeys = {
  all: ['prompts'] as const,

  // Lists
  lists: () => [...promptQueryKeys.all, 'list'] as const,
  list: (projectId: string, filters?: PromptFilters) =>
    [...promptQueryKeys.lists(), projectId, filters] as const,

  // Details
  details: () => [...promptQueryKeys.all, 'detail'] as const,
  detail: (projectId: string, promptId: string) =>
    [...promptQueryKeys.details(), projectId, promptId] as const,

  // Versions
  versions: (projectId: string, promptId: string) =>
    [...promptQueryKeys.detail(projectId, promptId), 'versions'] as const,
  version: (projectId: string, promptId: string, versionId: string) =>
    [...promptQueryKeys.versions(projectId, promptId), versionId] as const,

  // Diff
  diff: (projectId: string, promptId: string, from: number, to: number) =>
    [...promptQueryKeys.detail(projectId, promptId), 'diff', from, to] as const,

  // Protected Labels
  protectedLabels: (projectId: string) =>
    [...promptQueryKeys.all, 'protected-labels', projectId] as const,
}

export interface PromptFilters {
  type?: PromptType
  tags?: string[]
  search?: string
  page?: number
  limit?: number
}

/**
 * Query hook to list prompts for a project with filtering and pagination
 */
export function usePromptsQuery(
  projectId: string | undefined,
  filters?: PromptFilters,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: promptQueryKeys.list(projectId || '', filters),
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return getPrompts({
        projectId,
        ...filters,
      })
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30_000, // 30 seconds
    gcTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Query hook to get a single prompt by ID
 */
export function usePromptQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: promptQueryKeys.detail(projectId || '', promptId || ''),
    queryFn: async () => {
      if (!projectId || !promptId) {
        throw new Error('Project ID and Prompt ID are required')
      }
      return getPromptById(projectId, promptId)
    },
    enabled: !!projectId && !!promptId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Query hook to get all versions of a prompt
 */
export function useVersionsQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: promptQueryKeys.versions(projectId || '', promptId || ''),
    queryFn: async () => {
      if (!projectId || !promptId) {
        throw new Error('Project ID and Prompt ID are required')
      }
      return getVersions(projectId, promptId)
    },
    enabled: !!projectId && !!promptId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Query hook to get a single version
 */
export function useVersionQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  versionId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: promptQueryKeys.version(projectId || '', promptId || '', versionId || ''),
    queryFn: async () => {
      if (!projectId || !promptId || !versionId) {
        throw new Error('Project ID, Prompt ID, and Version ID are required')
      }
      return getVersion(projectId, promptId, versionId)
    },
    enabled: !!projectId && !!promptId && !!versionId && (options.enabled ?? true),
    staleTime: 60_000, // 1 minute - versions are immutable
    gcTime: 10 * 60 * 1000,
  })
}

/**
 * Query hook to get version diff
 */
export function useVersionDiffQuery(
  projectId: string | undefined,
  promptId: string | undefined,
  fromVersion: number | undefined,
  toVersion: number | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: promptQueryKeys.diff(
      projectId || '',
      promptId || '',
      fromVersion ?? 0,
      toVersion ?? 0
    ),
    queryFn: async () => {
      if (!projectId || !promptId || fromVersion === undefined || toVersion === undefined) {
        throw new Error('All parameters are required for diff')
      }
      return getVersionDiff(projectId, promptId, fromVersion, toVersion)
    },
    enabled:
      !!projectId &&
      !!promptId &&
      fromVersion !== undefined &&
      toVersion !== undefined &&
      (options.enabled ?? true),
    staleTime: 5 * 60 * 1000, // 5 minutes - diffs are computed from immutable data
    gcTime: 15 * 60 * 1000,
  })
}

/**
 * Query hook to get protected labels for a project
 */
export function useProtectedLabelsQuery(
  projectId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: promptQueryKeys.protectedLabels(projectId || ''),
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return getProtectedLabels(projectId)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 60_000, // 1 minute
    gcTime: 10 * 60 * 1000,
  })
}

/**
 * Mutation hook to create a new prompt
 */
export function useCreatePromptMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreatePromptRequest) => {
      return createPrompt(projectId, data)
    },
    onSuccess: (newPrompt) => {
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.lists(),
      })
      toast.success('Prompt Created', {
        description: `"${newPrompt.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Prompt', {
        description: apiError?.message || 'Could not create prompt. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to update a prompt's metadata
 */
export function useUpdatePromptMutation(projectId: string, promptId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: UpdatePromptRequest) => {
      return updatePrompt(projectId, promptId, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.detail(projectId, promptId),
      })
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.lists(),
      })
      toast.success('Prompt Updated', {
        description: 'Prompt metadata has been updated.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Prompt', {
        description: apiError?.message || 'Could not update prompt. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to delete a prompt
 */
export function useDeletePromptMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ promptId, promptName }: { promptId: string; promptName: string }) => {
      await deletePrompt(projectId, promptId)
      return { promptId, promptName }
    },
    onMutate: async ({ promptId }) => {
      await queryClient.cancelQueries({
        queryKey: promptQueryKeys.lists(),
      })

      const previousPrompts = queryClient.getQueriesData({
        queryKey: promptQueryKeys.lists(),
      })

      // Optimistic update
      queryClient.setQueriesData<{ prompts: PromptListItem[] }>(
        { queryKey: promptQueryKeys.lists() },
        (old) => {
          if (!old) return old
          return {
            ...old,
            prompts: old.prompts.filter((p) => p.id !== promptId),
          }
        }
      )

      return { previousPrompts }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.lists(),
      })
      toast.success('Prompt Deleted', {
        description: `"${variables.promptName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      if (context?.previousPrompts) {
        context.previousPrompts.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Prompt', {
        description: apiError?.message || 'Could not delete prompt. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to create a new version
 */
export function useCreateVersionMutation(projectId: string, promptId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateVersionRequest) => {
      return createVersion(projectId, promptId, data)
    },
    onSuccess: (newVersion) => {
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.versions(projectId, promptId),
      })
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.detail(projectId, promptId),
      })
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.lists(),
      })
      toast.success('Version Created', {
        description: `Version ${newVersion.version} has been created.`,
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

/**
 * Mutation hook to set labels on a version
 */
export function useSetLabelsMutation(projectId: string, promptId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      versionId,
      labels,
    }: {
      versionId: string
      labels: string[]
    }) => {
      await setLabels(projectId, promptId, versionId, labels)
      return { versionId, labels }
    },
    onMutate: async ({ versionId, labels }) => {
      await queryClient.cancelQueries({
        queryKey: promptQueryKeys.versions(projectId, promptId),
      })

      const previousVersions = queryClient.getQueryData<PromptVersion[]>(
        promptQueryKeys.versions(projectId, promptId)
      )

      queryClient.setQueryData<PromptVersion[]>(
        promptQueryKeys.versions(projectId, promptId),
        (old) => {
          if (!old) return old
          return old.map((v) =>
            v.id === versionId ? { ...v, labels } : v
          )
        }
      )

      return { previousVersions }
    },
    onSuccess: (_data, { labels }) => {
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.versions(projectId, promptId),
      })
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.detail(projectId, promptId),
      })
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.lists(),
      })
      toast.success('Labels Updated', {
        description: labels.length > 0
          ? `Labels set: ${labels.join(', ')}`
          : 'Labels cleared.',
      })
    },
    onError: (error: unknown, _variables, context) => {
      if (context?.previousVersions) {
        queryClient.setQueryData(
          promptQueryKeys.versions(projectId, promptId),
          context.previousVersions
        )
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Update Labels', {
        description: apiError?.message || 'Could not update labels. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to set protected labels for a project
 */
export function useSetProtectedLabelsMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (labels: string[]) => {
      await setProtectedLabels(projectId, labels)
      return labels
    },
    onSuccess: (labels) => {
      queryClient.invalidateQueries({
        queryKey: promptQueryKeys.protectedLabels(projectId),
      })
      toast.success('Protected Labels Updated', {
        description: labels.length > 0
          ? `Protected labels: ${labels.join(', ')}`
          : 'Protected labels cleared.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Protected Labels', {
        description: apiError?.message || 'Could not update protected labels. Please try again.',
      })
    },
  })
}

