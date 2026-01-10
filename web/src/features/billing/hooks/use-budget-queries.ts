'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listBudgets,
  getBudget,
  createBudget,
  updateBudget,
  deleteBudget,
  getAlerts,
  acknowledgeAlert,
} from '../api/budget-api'
import type { CreateBudgetRequest, UpdateBudgetRequest } from '../types'

// Query keys for cache management
export const budgetQueryKeys = {
  all: ['budgets'] as const,
  lists: () => [...budgetQueryKeys.all, 'list'] as const,
  list: (orgId: string) => [...budgetQueryKeys.lists(), orgId] as const,
  details: () => [...budgetQueryKeys.all, 'detail'] as const,
  detail: (orgId: string, budgetId: string) =>
    [...budgetQueryKeys.details(), orgId, budgetId] as const,
  alerts: (orgId: string) => [...budgetQueryKeys.all, 'alerts', orgId] as const,
}

export function useBudgetsQuery(organizationId: string | undefined) {
  return useQuery({
    queryKey: budgetQueryKeys.list(organizationId ?? ''),
    queryFn: () => listBudgets(organizationId!),
    enabled: !!organizationId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useBudgetQuery(
  organizationId: string | undefined,
  budgetId: string | undefined
) {
  return useQuery({
    queryKey: budgetQueryKeys.detail(organizationId ?? '', budgetId ?? ''),
    queryFn: () => getBudget(organizationId!, budgetId!),
    enabled: !!organizationId && !!budgetId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useCreateBudgetMutation(organizationId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBudgetRequest) => createBudget(organizationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: budgetQueryKeys.list(organizationId),
      })
    },
  })
}

export function useUpdateBudgetMutation(organizationId: string, budgetId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateBudgetRequest) =>
      updateBudget(organizationId, budgetId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: budgetQueryKeys.list(organizationId),
      })
      queryClient.invalidateQueries({
        queryKey: budgetQueryKeys.detail(organizationId, budgetId),
      })
    },
  })
}

export function useDeleteBudgetMutation(organizationId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (budgetId: string) => deleteBudget(organizationId, budgetId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: budgetQueryKeys.list(organizationId),
      })
    },
  })
}

export function useAlertsQuery(organizationId: string | undefined, limit?: number) {
  return useQuery({
    queryKey: budgetQueryKeys.alerts(organizationId ?? ''),
    queryFn: () => getAlerts(organizationId!, limit),
    enabled: !!organizationId,
    staleTime: 60 * 1000, // 1 minute - alerts should refresh more frequently
    refetchInterval: 60 * 1000,
  })
}

export function useAcknowledgeAlertMutation(organizationId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (alertId: string) => acknowledgeAlert(organizationId, alertId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: budgetQueryKeys.alerts(organizationId),
      })
    },
  })
}
