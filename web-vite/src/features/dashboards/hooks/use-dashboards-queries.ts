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

export const dashboardQueryKeys = {
  all: ['dashboards'] as const,
  lists: () => [...dashboardQueryKeys.all, 'list'] as const,
  list: (projectId: string, filter?: DashboardFilter) =>
    [...dashboardQueryKeys.lists(), projectId, filter] as const,
  details: () => [...dashboardQueryKeys.all, 'detail'] as const,
  detail: (projectId: string, dashboardId: string) =>
    [...dashboardQueryKeys.details(), projectId, dashboardId] as const,
}

export function useDashboardsQuery(
  projectId: string | undefined,
  filter?: DashboardFilter,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: dashboardQueryKeys.list(projectId ?? '', filter),
    queryFn: async () => {
      if (!projectId) throw new Error('Project ID is required')
      return getDashboards(projectId, filter)
    },
    enabled: !!projectId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60_000,
  })
}

export function useDashboardQuery(
  projectId: string | undefined,
  dashboardId: string | undefined,
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: dashboardQueryKeys.detail(projectId ?? '', dashboardId ?? ''),
    queryFn: async () => {
      if (!projectId || !dashboardId) {
        throw new Error('Project ID and Dashboard ID are required')
      }
      return getDashboardById(projectId, dashboardId)
    },
    enabled: !!projectId && !!dashboardId && (options.enabled ?? true),
    staleTime: 30_000,
    gcTime: 5 * 60_000,
  })
}

function messageOf(error: unknown): string | undefined {
  if (error && typeof error === 'object' && 'message' in error) {
    const m = (error as { message?: unknown }).message
    if (typeof m === 'string') return m
  }
  return undefined
}

export function useCreateDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateDashboardRequest) =>
      createDashboard(projectId, data),
    onSuccess: (newDashboard) => {
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
      toast.success('Dashboard Created', {
        description: `"${newDashboard.name}" has been created successfully.`,
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Create Dashboard', {
        description:
          messageOf(error) ?? 'Could not create dashboard. Please try again.',
      })
    },
  })
}

export function useUpdateDashboardMutation(
  projectId: string,
  dashboardId: string,
) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: UpdateDashboardRequest) =>
      updateDashboard(projectId, dashboardId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
      toast.success('Dashboard Updated', {
        description: 'Dashboard has been updated successfully.',
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Update Dashboard', {
        description:
          messageOf(error) ?? 'Could not update dashboard. Please try again.',
      })
    },
  })
}

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
      await queryClient.cancelQueries({ queryKey: dashboardQueryKeys.lists() })
      const previousDashboards = queryClient.getQueriesData({
        queryKey: dashboardQueryKeys.lists(),
      })
      queryClient.setQueriesData<{ dashboards: Dashboard[] }>(
        { queryKey: dashboardQueryKeys.lists() },
        (old) => {
          if (!old) return old
          return {
            ...old,
            dashboards: old.dashboards.filter((d) => d.id !== dashboardId),
          }
        },
      )
      return { previousDashboards }
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
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
      toast.error('Failed to Delete Dashboard', {
        description:
          messageOf(error) ?? 'Could not delete dashboard. Please try again.',
      })
    },
  })
}

export function useLockDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (dashboardId: string) => lockDashboard(projectId, dashboardId),
    onMutate: async (dashboardId) => {
      await queryClient.cancelQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })
      const previousDashboard = queryClient.getQueryData<Dashboard>(
        dashboardQueryKeys.detail(projectId, dashboardId),
      )
      if (previousDashboard) {
        queryClient.setQueryData<Dashboard>(
          dashboardQueryKeys.detail(projectId, dashboardId),
          { ...previousDashboard, is_locked: true },
        )
      }
      return { previousDashboard }
    },
    onSuccess: (updatedDashboard) => {
      queryClient.setQueryData(
        dashboardQueryKeys.detail(projectId, updatedDashboard.id),
        updatedDashboard,
      )
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
      toast.success('Dashboard Locked', {
        description: 'Dashboard is now protected from modifications.',
      })
    },
    onError: (error: unknown, dashboardId, context) => {
      if (context?.previousDashboard) {
        queryClient.setQueryData(
          dashboardQueryKeys.detail(projectId, dashboardId),
          context.previousDashboard,
        )
      }
      toast.error('Failed to Lock Dashboard', {
        description:
          messageOf(error) ?? 'Could not lock dashboard. Please try again.',
      })
    },
  })
}

export function useUnlockDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (dashboardId: string) =>
      unlockDashboard(projectId, dashboardId),
    onMutate: async (dashboardId) => {
      await queryClient.cancelQueries({
        queryKey: dashboardQueryKeys.detail(projectId, dashboardId),
      })
      const previousDashboard = queryClient.getQueryData<Dashboard>(
        dashboardQueryKeys.detail(projectId, dashboardId),
      )
      if (previousDashboard) {
        queryClient.setQueryData<Dashboard>(
          dashboardQueryKeys.detail(projectId, dashboardId),
          { ...previousDashboard, is_locked: false },
        )
      }
      return { previousDashboard }
    },
    onSuccess: (updatedDashboard) => {
      queryClient.setQueryData(
        dashboardQueryKeys.detail(projectId, updatedDashboard.id),
        updatedDashboard,
      )
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
      toast.success('Dashboard Unlocked', {
        description: 'Dashboard can now be modified.',
      })
    },
    onError: (error: unknown, dashboardId, context) => {
      if (context?.previousDashboard) {
        queryClient.setQueryData(
          dashboardQueryKeys.detail(projectId, dashboardId),
          context.previousDashboard,
        )
      }
      toast.error('Failed to Unlock Dashboard', {
        description:
          messageOf(error) ?? 'Could not unlock dashboard. Please try again.',
      })
    },
  })
}

export function useImportDashboardMutation(projectId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: DashboardImportRequest) =>
      importDashboard(projectId, data),
    onSuccess: (newDashboard) => {
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.lists() })
      toast.success('Dashboard Imported', {
        description: `"${newDashboard.name}" has been imported successfully.`,
      })
    },
    onError: (error: unknown) => {
      toast.error('Failed to Import Dashboard', {
        description:
          messageOf(error) ?? 'Could not import dashboard. Please try again.',
      })
    },
  })
}
