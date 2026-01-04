'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  getDashboards,
  getDashboardById,
  createDashboard,
  updateDashboard,
  deleteDashboard,
  lockDashboard,
  unlockDashboard,
  importDashboard,
} from '../api/dashboards-api'
import type {
  Dashboard,
  DashboardFilter,
  CreateDashboardRequest,
  UpdateDashboardRequest,
  DashboardImportRequest,
} from '../types'

/**
 * Query keys for dashboard queries
 */
export const dashboardQueryKeys = {
  all: ['dashboards'] as const,

  // Lists
  lists: () => [...dashboardQueryKeys.all, 'list'] as const,
  list: (projectId: string, filter?: DashboardFilter) =>
    [...dashboardQueryKeys.lists(), projectId, filter] as const,

  // Details
  details: () => [...dashboardQueryKeys.all, 'detail'] as const,
  detail: (projectId: string, dashboardId: string) =>
    [...dashboardQueryKeys.details(), projectId, dashboardId] as const,
}

/**
 * Query hook to list dashboards for a project with filtering
 */
export function useDashboardsQuery(
  projectId: string | undefined,
  filter?: DashboardFilter,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: dashboardQueryKeys.list(projectId || '', filter),
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required')
      }
      return getDashboards(projectId, filter)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30_000, // 30 seconds
    gcTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Query hook to get a single dashboard by ID
 */
export function useDashboardQuery(
  projectId: string | undefined,
  dashboardId: string | undefined,
  options: { enabled?: boolean } = {}
) {
  return useQuery({
    queryKey: dashboardQueryKeys.detail(projectId || '', dashboardId || ''),
    queryFn: async () => {
      if (!projectId || !dashboardId) {
        throw new Error('Project ID and Dashboard ID are required')
      }
      return getDashboardById(projectId, dashboardId)
    },
    enabled: !!projectId && !!dashboardId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60 * 1000,
  })
}

/**
 * Mutation hook to create a new dashboard
 */
export function useCreateDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: CreateDashboardRequest) => {
      return createDashboard(projectId, data)
    },
    onSuccess: (newDashboard) => {
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Created', {
        description: `"${newDashboard.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Create Dashboard', {
        description:
          apiError?.message || 'Could not create dashboard. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to update a dashboard
 */
export function useUpdateDashboardMutation(
  projectId: string,
  dashboardId: string
) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: UpdateDashboardRequest) => {
      return updateDashboard(projectId, dashboardId, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Updated', {
        description: 'Dashboard has been updated successfully.',
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Update Dashboard', {
        description:
          apiError?.message || 'Could not update dashboard. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to delete a dashboard
 */
export function useDeleteDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      dashboardId,
      dashboardName,
    }: {
      dashboardId: string
      dashboardName: string
    }) => {
      await deleteDashboard(projectId, dashboardId)
      return { dashboardId, dashboardName }
    },
    onMutate: async ({ dashboardId }) => {
      await queryClient.cancelQueries({
        queryKey: dashboardQueryKeys.lists(),
      })

      const previousDashboards = queryClient.getQueriesData({
        queryKey: dashboardQueryKeys.lists(),
      })

      // Optimistic update
      queryClient.setQueriesData<{ dashboards: Dashboard[] }>(
        { queryKey: dashboardQueryKeys.lists() },
        (old) => {
          if (!old) return old
          return {
            ...old,
            dashboards: old.dashboards.filter((d) => d.id !== dashboardId),
          }
        }
      )

      return { previousDashboards }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Deleted', {
        description: `"${variables.dashboardName}" has been deleted.`,
      })
    },
    onError: (error: unknown, _variables, context) => {
      if (context?.previousDashboards) {
        context.previousDashboards.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Delete Dashboard', {
        description:
          apiError?.message || 'Could not delete dashboard. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to lock a dashboard
 */
export function useLockDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (dashboardId: string) => {
      return lockDashboard(projectId, dashboardId)
    },
    onMutate: async (dashboardId) => {
      // Optimistically update the dashboard
      await queryClient.cancelQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })

      const previousDashboard = queryClient.getQueryData<Dashboard>(
        dashboardQueryKeys.detail(projectId, dashboardId)
      )

      if (previousDashboard) {
        queryClient.setQueryData<Dashboard>(
          dashboardQueryKeys.detail(projectId, dashboardId),
          { ...previousDashboard, is_locked: true }
        )
      }

      return { previousDashboard }
    },
    onSuccess: (updatedDashboard) => {
      queryClient.setQueryData(
        dashboardQueryKeys.detail(projectId, updatedDashboard.id),
        updatedDashboard
      )
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Locked', {
        description: 'Dashboard is now protected from modifications.',
      })
    },
    onError: (error: unknown, dashboardId, context) => {
      if (context?.previousDashboard) {
        queryClient.setQueryData(
          dashboardQueryKeys.detail(projectId, dashboardId),
          context.previousDashboard
        )
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Lock Dashboard', {
        description:
          apiError?.message || 'Could not lock dashboard. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to unlock a dashboard
 */
export function useUnlockDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (dashboardId: string) => {
      return unlockDashboard(projectId, dashboardId)
    },
    onMutate: async (dashboardId) => {
      // Optimistically update the dashboard
      await queryClient.cancelQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })

      const previousDashboard = queryClient.getQueryData<Dashboard>(
        dashboardQueryKeys.detail(projectId, dashboardId)
      )

      if (previousDashboard) {
        queryClient.setQueryData<Dashboard>(
          dashboardQueryKeys.detail(projectId, dashboardId),
          { ...previousDashboard, is_locked: false }
        )
      }

      return { previousDashboard }
    },
    onSuccess: (updatedDashboard) => {
      queryClient.setQueryData(
        dashboardQueryKeys.detail(projectId, updatedDashboard.id),
        updatedDashboard
      )
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Unlocked', {
        description: 'Dashboard can now be modified.',
      })
    },
    onError: (error: unknown, dashboardId, context) => {
      if (context?.previousDashboard) {
        queryClient.setQueryData(
          dashboardQueryKeys.detail(projectId, dashboardId),
          context.previousDashboard
        )
      }
      const apiError = error as { message?: string }
      toast.error('Failed to Unlock Dashboard', {
        description:
          apiError?.message || 'Could not unlock dashboard. Please try again.',
      })
    },
  })
}

/**
 * Mutation hook to import a dashboard
 */
export function useImportDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (data: DashboardImportRequest) => {
      return importDashboard(projectId, data)
    },
    onSuccess: (newDashboard) => {
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.lists(),
      })
      toast.success('Dashboard Imported', {
        description: `"${newDashboard.name}" has been imported successfully.`,
      })
    },
    onError: (error: unknown) => {
      const apiError = error as { message?: string }
      toast.error('Failed to Import Dashboard', {
        description:
          apiError?.message || 'Could not import dashboard. Please try again.',
      })
    },
  })
}
